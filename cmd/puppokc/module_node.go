// Copyright 2017 The go-okcoin Authors
// This file is part of go-okcoin.
//
// go-okcoin is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-okcoin is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-okcoin. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/okcoin/go-okcoin/common"
	"github.com/okcoin/go-okcoin/log"
)

// nodeDockerfile is the Dockerfile required to run an Okcoin node.
var nodeDockerfile = `
FROM okcoin/client-go:latest

ADD genesis.json /genesis.json
{{if .Unlock}}
	ADD signer.json /signer.json
	ADD signer.pass /signer.pass
{{end}}
RUN \
  echo 'gokc --cache 512 init /genesis.json' > gokc.sh && \{{if .Unlock}}
	echo 'mkdir -p /root/.okcoin/keystore/ && cp /signer.json /root/.okcoin/keystore/' >> gokc.sh && \{{end}}
	echo $'gokc --networkid {{.NetworkID}} --cache 512 --port {{.Port}} --maxpeers {{.Peers}} {{.LightFlag}} --okcstats \'{{.Okcstats}}\' {{if .Bootnodes}}--bootnodes {{.Bootnodes}}{{end}} {{if .Okcerbase}}--okcerbase {{.Okcerbase}} --mine --minerthreads 1{{end}} {{if .Unlock}}--unlock 0 --password /signer.pass --mine{{end}} --targetgaslimit {{.GasTarget}} --gasprice {{.GasPrice}}' >> gokc.sh

ENTRYPOINT ["/bin/sh", "gokc.sh"]
`

// nodeComposefile is the docker-compose.yml file required to deploy and maintain
// an Okcoin node (bootnode or miner for now).
var nodeComposefile = `
version: '2'
services:
  {{.Type}}:
    build: .
    image: {{.Network}}/{{.Type}}
    ports:
      - "{{.Port}}:{{.Port}}"
      - "{{.Port}}:{{.Port}}/udp"
    volumes:
      - {{.Datadir}}:/root/.okcoin{{if .Okcashdir}}
      - {{.Okcashdir}}:/root/.okcash{{end}}
    environment:
      - PORT={{.Port}}/tcp
      - TOTAL_PEERS={{.TotalPeers}}
      - LIGHT_PEERS={{.LightPeers}}
      - STATS_NAME={{.Okcstats}}
      - MINER_NAME={{.Okcerbase}}
      - GAS_TARGET={{.GasTarget}}
      - GAS_PRICE={{.GasPrice}}
    logging:
      driver: "json-file"
      options:
        max-size: "1m"
        max-file: "10"
    restart: always
`

// deployNode deploys a new Okcoin node container to a remote machine via SSH,
// docker and docker-compose. If an instance with the specified network name
// already exists there, it will be overwritten!
func deployNode(client *sshClient, network string, bootnodes []string, config *nodeInfos, nocache bool) ([]byte, error) {
	kind := "sealnode"
	if config.keyJSON == "" && config.okcerbase == "" {
		kind = "bootnode"
		bootnodes = make([]string, 0)
	}
	// Generate the content to upload to the server
	workdir := fmt.Sprintf("%d", rand.Int63())
	files := make(map[string][]byte)

	lightFlag := ""
	if config.peersLight > 0 {
		lightFlag = fmt.Sprintf("--lightpeers=%d --lightserv=50", config.peersLight)
	}
	dockerfile := new(bytes.Buffer)
	template.Must(template.New("").Parse(nodeDockerfile)).Execute(dockerfile, map[string]interface{}{
		"NetworkID": config.network,
		"Port":      config.port,
		"Peers":     config.peersTotal,
		"LightFlag": lightFlag,
		"Bootnodes": strings.Join(bootnodes, ","),
		"Okcstats":  config.okcstats,
		"Okcerbase": config.okcerbase,
		"GasTarget": uint64(1000000 * config.gasTarget),
		"GasPrice":  uint64(1000000000 * config.gasPrice),
		"Unlock":    config.keyJSON != "",
	})
	files[filepath.Join(workdir, "Dockerfile")] = dockerfile.Bytes()

	composefile := new(bytes.Buffer)
	template.Must(template.New("").Parse(nodeComposefile)).Execute(composefile, map[string]interface{}{
		"Type":       kind,
		"Datadir":    config.datadir,
		"Okcashdir":  config.okcashdir,
		"Network":    network,
		"Port":       config.port,
		"TotalPeers": config.peersTotal,
		"Light":      config.peersLight > 0,
		"LightPeers": config.peersLight,
		"Okcstats":   config.okcstats[:strings.Index(config.okcstats, ":")],
		"Okcerbase":  config.okcerbase,
		"GasTarget":  config.gasTarget,
		"GasPrice":   config.gasPrice,
	})
	files[filepath.Join(workdir, "docker-compose.yaml")] = composefile.Bytes()

	files[filepath.Join(workdir, "genesis.json")] = config.genesis
	if config.keyJSON != "" {
		files[filepath.Join(workdir, "signer.json")] = []byte(config.keyJSON)
		files[filepath.Join(workdir, "signer.pass")] = []byte(config.keyPass)
	}
	// Upload the deployment files to the remote server (and clean up afterwards)
	if out, err := client.Upload(files); err != nil {
		return out, err
	}
	defer client.Run("rm -rf " + workdir)

	// Build and deploy the boot or seal node service
	if nocache {
		return nil, client.Stream(fmt.Sprintf("cd %s && docker-compose -p %s build --pull --no-cache && docker-compose -p %s up -d --force-recreate", workdir, network, network))
	}
	return nil, client.Stream(fmt.Sprintf("cd %s && docker-compose -p %s up -d --build --force-recreate", workdir, network))
}

// nodeInfos is returned from a boot or seal node status check to allow reporting
// various configuration parameters.
type nodeInfos struct {
	genesis    []byte
	network    int64
	datadir    string
	okcashdir  string
	okcstats   string
	port       int
	enode      string
	peersTotal int
	peersLight int
	okcerbase  string
	keyJSON    string
	keyPass    string
	gasTarget  float64
	gasPrice   float64
}

// Report converts the typed struct into a plain string->string map, containing
// most - but not all - fields for reporting to the user.
func (info *nodeInfos) Report() map[string]string {
	report := map[string]string{
		"Data directory":           info.datadir,
		"Listener port":            strconv.Itoa(info.port),
		"Peer count (all total)":   strconv.Itoa(info.peersTotal),
		"Peer count (light nodes)": strconv.Itoa(info.peersLight),
		"Okcstats username":        info.okcstats,
	}
	if info.gasTarget > 0 {
		// Miner or signer node
		report["Gas limit (baseline target)"] = fmt.Sprintf("%0.3f MGas", info.gasTarget)
		report["Gas price (minimum accepted)"] = fmt.Sprintf("%0.3f GWei", info.gasPrice)

		if info.okcerbase != "" {
			// Okcash proof-of-work miner
			report["Okcash directory"] = info.okcashdir
			report["Miner account"] = info.okcerbase
		}
		if info.keyJSON != "" {
			// Clique proof-of-authority signer
			var key struct {
				Address string `json:"address"`
			}
			if err := json.Unmarshal([]byte(info.keyJSON), &key); err == nil {
				report["Signer account"] = common.HexToAddress(key.Address).Hex()
			} else {
				log.Error("Failed to retrieve signer address", "err", err)
			}
		}
	}
	return report
}

// checkNode does a health-check against an boot or seal node server to verify
// whokcer it's running, and if yes, whokcer it's responsive.
func checkNode(client *sshClient, network string, boot bool) (*nodeInfos, error) {
	kind := "bootnode"
	if !boot {
		kind = "sealnode"
	}
	// Inspect a possible bootnode container on the host
	infos, err := inspectContainer(client, fmt.Sprintf("%s_%s_1", network, kind))
	if err != nil {
		return nil, err
	}
	if !infos.running {
		return nil, ErrServiceOffline
	}
	// Resolve a few types from the environmental variables
	totalPeers, _ := strconv.Atoi(infos.envvars["TOTAL_PEERS"])
	lightPeers, _ := strconv.Atoi(infos.envvars["LIGHT_PEERS"])
	gasTarget, _ := strconv.ParseFloat(infos.envvars["GAS_TARGET"], 64)
	gasPrice, _ := strconv.ParseFloat(infos.envvars["GAS_PRICE"], 64)

	// Container available, retrieve its node ID and its genesis json
	var out []byte
	if out, err = client.Run(fmt.Sprintf("docker exec %s_%s_1 gokc --exec admin.nodeInfo.id attach", network, kind)); err != nil {
		return nil, ErrServiceUnreachable
	}
	id := bytes.Trim(bytes.TrimSpace(out), "\"")

	if out, err = client.Run(fmt.Sprintf("docker exec %s_%s_1 cat /genesis.json", network, kind)); err != nil {
		return nil, ErrServiceUnreachable
	}
	genesis := bytes.TrimSpace(out)

	keyJSON, keyPass := "", ""
	if out, err = client.Run(fmt.Sprintf("docker exec %s_%s_1 cat /signer.json", network, kind)); err == nil {
		keyJSON = string(bytes.TrimSpace(out))
	}
	if out, err = client.Run(fmt.Sprintf("docker exec %s_%s_1 cat /signer.pass", network, kind)); err == nil {
		keyPass = string(bytes.TrimSpace(out))
	}
	// Run a sanity check to see if the devp2p is reachable
	port := infos.portmap[infos.envvars["PORT"]]
	if err = checkPort(client.server, port); err != nil {
		log.Warn(fmt.Sprintf("%s devp2p port seems unreachable", strings.Title(kind)), "server", client.server, "port", port, "err", err)
	}
	// Assemble and return the useful infos
	stats := &nodeInfos{
		genesis:    genesis,
		datadir:    infos.volumes["/root/.okcoin"],
		okcashdir:  infos.volumes["/root/.okcash"],
		port:       port,
		peersTotal: totalPeers,
		peersLight: lightPeers,
		okcstats:   infos.envvars["STATS_NAME"],
		okcerbase:  infos.envvars["MINER_NAME"],
		keyJSON:    keyJSON,
		keyPass:    keyPass,
		gasTarget:  gasTarget,
		gasPrice:   gasPrice,
	}
	stats.enode = fmt.Sprintf("enode://%s@%s:%d", id, client.address, stats.port)

	return stats, nil
}
