.PHONY: build cluster test

default: build

build:
	go build -o bin/swarm    github.com/okcoin/go-okcoin/cmd/swarm
	go build -o bin/gokc     github.com/okcoin/go-okcoin/cmd/gokc
	go build -o bin/bootnode github.com/okcoin/go-okcoin/cmd/bootnode

cluster: build
	scripts/boot-cluster.sh

test:
	go test -v github.com/okcoin/go-okcoin/swarm/...
