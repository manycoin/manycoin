# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: gokc android ios gokc-cross swarm evm all test clean
.PHONY: gokc-linux gokc-linux-386 gokc-linux-amd64 gokc-linux-mips64 gokc-linux-mips64le
.PHONY: gokc-linux-arm gokc-linux-arm-5 gokc-linux-arm-6 gokc-linux-arm-7 gokc-linux-arm64
.PHONY: gokc-darwin gokc-darwin-386 gokc-darwin-amd64
.PHONY: gokc-windows gokc-windows-386 gokc-windows-amd64

GOBIN = $(shell pwd)/build/bin
GO ?= latest

gokc:
	build/env.sh go run build/ci.go install ./cmd/gokc
	@echo "Done building."
	@echo "Run \"$(GOBIN)/gokc\" to launch gokc."

swarm:
	build/env.sh go run build/ci.go install ./cmd/swarm
	@echo "Done building."
	@echo "Run \"$(GOBIN)/swarm\" to launch swarm."

all:
	build/env.sh go run build/ci.go install

android:
	build/env.sh go run build/ci.go aar --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/gokc.aar\" to use the library."

ios:
	build/env.sh go run build/ci.go xcode --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/Gokc.framework\" to use the library."

test: all
	build/env.sh go run build/ci.go test

clean:
	rm -fr build/_workspace/pkg/ $(GOBIN)/*

# The devtools target installs tools required for 'go generate'.
# You need to put $GOBIN (or $GOPATH/bin) in your PATH to use 'go generate'.

devtools:
	env GOBIN= go get -u golang.org/x/tools/cmd/stringer
	env GOBIN= go get -u github.com/kevinburke/go-bindata/go-bindata
	env GOBIN= go get -u github.com/fjl/gencodec
	env GOBIN= go get -u github.com/golang/protobuf/protoc-gen-go
	env GOBIN= go install ./cmd/abigen
	@type "npm" 2> /dev/null || echo 'Please install node.js and npm'
	@type "solc" 2> /dev/null || echo 'Please install solc'
	@type "protoc" 2> /dev/null || echo 'Please install protoc'

# Cross Compilation Targets (xgo)

gokc-cross: gokc-linux gokc-darwin gokc-windows gokc-android gokc-ios
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/gokc-*

gokc-linux: gokc-linux-386 gokc-linux-amd64 gokc-linux-arm gokc-linux-mips64 gokc-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/gokc-linux-*

gokc-linux-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/386 -v ./cmd/gokc
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/gokc-linux-* | grep 386

gokc-linux-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/amd64 -v ./cmd/gokc
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gokc-linux-* | grep amd64

gokc-linux-arm: gokc-linux-arm-5 gokc-linux-arm-6 gokc-linux-arm-7 gokc-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/gokc-linux-* | grep arm

gokc-linux-arm-5:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-5 -v ./cmd/gokc
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/gokc-linux-* | grep arm-5

gokc-linux-arm-6:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-6 -v ./cmd/gokc
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/gokc-linux-* | grep arm-6

gokc-linux-arm-7:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-7 -v ./cmd/gokc
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/gokc-linux-* | grep arm-7

gokc-linux-arm64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm64 -v ./cmd/gokc
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/gokc-linux-* | grep arm64

gokc-linux-mips:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips --ldflags '-extldflags "-static"' -v ./cmd/gokc
	@echo "Linux MIPS cross compilation done:"
	@ls -ld $(GOBIN)/gokc-linux-* | grep mips

gokc-linux-mipsle:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mipsle --ldflags '-extldflags "-static"' -v ./cmd/gokc
	@echo "Linux MIPSle cross compilation done:"
	@ls -ld $(GOBIN)/gokc-linux-* | grep mipsle

gokc-linux-mips64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64 --ldflags '-extldflags "-static"' -v ./cmd/gokc
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/gokc-linux-* | grep mips64

gokc-linux-mips64le:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64le --ldflags '-extldflags "-static"' -v ./cmd/gokc
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/gokc-linux-* | grep mips64le

gokc-darwin: gokc-darwin-386 gokc-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/gokc-darwin-*

gokc-darwin-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/386 -v ./cmd/gokc
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/gokc-darwin-* | grep 386

gokc-darwin-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./cmd/gokc
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gokc-darwin-* | grep amd64

gokc-windows: gokc-windows-386 gokc-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/gokc-windows-*

gokc-windows-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/386 -v ./cmd/gokc
	@echo "Windows 386 cross compilation done:"
	@ls -ld $(GOBIN)/gokc-windows-* | grep 386

gokc-windows-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/amd64 -v ./cmd/gokc
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gokc-windows-* | grep amd64
