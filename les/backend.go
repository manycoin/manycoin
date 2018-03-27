// Copyright 2016 The go-okcoin Authors
// This file is part of the go-okcoin library.
//
// The go-okcoin library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-okcoin library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-okcoin library. If not, see <http://www.gnu.org/licenses/>.

// Package les implements the Light Okcoin Subprotocol.
package les

import (
	"fmt"
	"sync"
	"time"

	"github.com/okcoin/go-okcoin/accounts"
	"github.com/okcoin/go-okcoin/common"
	"github.com/okcoin/go-okcoin/common/hexutil"
	"github.com/okcoin/go-okcoin/consensus"
	"github.com/okcoin/go-okcoin/core"
	"github.com/okcoin/go-okcoin/core/bloombits"
	"github.com/okcoin/go-okcoin/core/types"
	"github.com/okcoin/go-okcoin/okc"
	"github.com/okcoin/go-okcoin/okc/downloader"
	"github.com/okcoin/go-okcoin/okc/filters"
	"github.com/okcoin/go-okcoin/okc/gasprice"
	"github.com/okcoin/go-okcoin/okcdb"
	"github.com/okcoin/go-okcoin/event"
	"github.com/okcoin/go-okcoin/internal/okcapi"
	"github.com/okcoin/go-okcoin/light"
	"github.com/okcoin/go-okcoin/log"
	"github.com/okcoin/go-okcoin/node"
	"github.com/okcoin/go-okcoin/p2p"
	"github.com/okcoin/go-okcoin/p2p/discv5"
	"github.com/okcoin/go-okcoin/params"
	rpc "github.com/okcoin/go-okcoin/rpc"
)

type LightOkcoin struct {
	config *okc.Config

	odr         *LesOdr
	relay       *LesTxRelay
	chainConfig *params.ChainConfig
	// Channel for shutting down the service
	shutdownChan chan bool
	// Handlers
	peers           *peerSet
	txPool          *light.TxPool
	blockchain      *light.LightChain
	protocolManager *ProtocolManager
	serverPool      *serverPool
	reqDist         *requestDistributor
	retriever       *retrieveManager
	// DB interfaces
	chainDb okcdb.Database // Block chain database

	bloomRequests                              chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer, chtIndexer, bloomTrieIndexer *core.ChainIndexer

	ApiBackend *LesApiBackend

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	networkId     uint64
	netRPCService *okcapi.PublicNetAPI

	wg sync.WaitGroup
}

func New(ctx *node.ServiceContext, config *okc.Config) (*LightOkcoin, error) {
	chainDb, err := okc.CreateDB(ctx, config, "lightchaindata")
	if err != nil {
		return nil, err
	}
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, isCompat := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !isCompat {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	peers := newPeerSet()
	quitSync := make(chan struct{})

	lokc := &LightOkcoin{
		config:           config,
		chainConfig:      chainConfig,
		chainDb:          chainDb,
		eventMux:         ctx.EventMux,
		peers:            peers,
		reqDist:          newRequestDistributor(peers, quitSync),
		accountManager:   ctx.AccountManager,
		engine:           okc.CreateConsensusEngine(ctx, &config.Okcash, chainConfig, chainDb),
		shutdownChan:     make(chan bool),
		networkId:        config.NetworkId,
		bloomRequests:    make(chan chan *bloombits.Retrieval),
		bloomIndexer:     okc.NewBloomIndexer(chainDb, light.BloomTrieFrequency),
		chtIndexer:       light.NewChtIndexer(chainDb, true),
		bloomTrieIndexer: light.NewBloomTrieIndexer(chainDb, true),
	}

	lokc.relay = NewLesTxRelay(peers, lokc.reqDist)
	lokc.serverPool = newServerPool(chainDb, quitSync, &lokc.wg)
	lokc.retriever = newRetrieveManager(peers, lokc.reqDist, lokc.serverPool)
	lokc.odr = NewLesOdr(chainDb, lokc.chtIndexer, lokc.bloomTrieIndexer, lokc.bloomIndexer, lokc.retriever)
	if lokc.blockchain, err = light.NewLightChain(lokc.odr, lokc.chainConfig, lokc.engine); err != nil {
		return nil, err
	}
	lokc.bloomIndexer.Start(lokc.blockchain)
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		lokc.blockchain.SetHead(compat.RewindTo)
		core.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}

	lokc.txPool = light.NewTxPool(lokc.chainConfig, lokc.blockchain, lokc.relay)
	if lokc.protocolManager, err = NewProtocolManager(lokc.chainConfig, true, ClientProtocolVersions, config.NetworkId, lokc.eventMux, lokc.engine, lokc.peers, lokc.blockchain, nil, chainDb, lokc.odr, lokc.relay, quitSync, &lokc.wg); err != nil {
		return nil, err
	}
	lokc.ApiBackend = &LesApiBackend{lokc, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	lokc.ApiBackend.gpo = gasprice.NewOracle(lokc.ApiBackend, gpoParams)
	return lokc, nil
}

func lesTopic(genesisHash common.Hash, protocolVersion uint) discv5.Topic {
	var name string
	switch protocolVersion {
	case lpv1:
		name = "LES"
	case lpv2:
		name = "LES2"
	default:
		panic(nil)
	}
	return discv5.Topic(name + "@" + common.Bytes2Hex(genesisHash.Bytes()[0:8]))
}

type LightDummyAPI struct{}

// Okcerbase is the address that mining rewards will be send to
func (s *LightDummyAPI) Okcerbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Coinbase is the address that mining rewards will be send to (alias for Okcerbase)
func (s *LightDummyAPI) Coinbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Hashrate returns the POW hashrate
func (s *LightDummyAPI) Hashrate() hexutil.Uint {
	return 0
}

// Mining returns an indication if this node is currently mining.
func (s *LightDummyAPI) Mining() bool {
	return false
}

// APIs returns the collection of RPC services the okcoin package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *LightOkcoin) APIs() []rpc.API {
	return append(okcapi.GetAPIs(s.ApiBackend), []rpc.API{
		{
			Namespace: "okc",
			Version:   "1.0",
			Service:   &LightDummyAPI{},
			Public:    true,
		}, {
			Namespace: "okc",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "okc",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, true),
			Public:    true,
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *LightOkcoin) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *LightOkcoin) BlockChain() *light.LightChain      { return s.blockchain }
func (s *LightOkcoin) TxPool() *light.TxPool              { return s.txPool }
func (s *LightOkcoin) Engine() consensus.Engine           { return s.engine }
func (s *LightOkcoin) LesVersion() int                    { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *LightOkcoin) Downloader() *downloader.Downloader { return s.protocolManager.downloader }
func (s *LightOkcoin) EventMux() *event.TypeMux           { return s.eventMux }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *LightOkcoin) Protocols() []p2p.Protocol {
	return s.protocolManager.SubProtocols
}

// Start implements node.Service, starting all internal goroutines needed by the
// Okcoin protocol implementation.
func (s *LightOkcoin) Start(srvr *p2p.Server) error {
	s.startBloomHandlers()
	log.Warn("Light client mode is an experimental feature")
	s.netRPCService = okcapi.NewPublicNetAPI(srvr, s.networkId)
	// clients are searching for the first advertised protocol in the list
	protocolVersion := AdvertiseProtocolVersions[0]
	s.serverPool.start(srvr, lesTopic(s.blockchain.Genesis().Hash(), protocolVersion))
	s.protocolManager.Start(s.config.LightPeers)
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Okcoin protocol.
func (s *LightOkcoin) Stop() error {
	s.odr.Stop()
	if s.bloomIndexer != nil {
		s.bloomIndexer.Close()
	}
	if s.chtIndexer != nil {
		s.chtIndexer.Close()
	}
	if s.bloomTrieIndexer != nil {
		s.bloomTrieIndexer.Close()
	}
	s.blockchain.Stop()
	s.protocolManager.Stop()
	s.txPool.Stop()

	s.eventMux.Stop()

	time.Sleep(time.Millisecond * 200)
	s.chainDb.Close()
	close(s.shutdownChan)

	return nil
}
