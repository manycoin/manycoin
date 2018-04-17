// Copyright 2014 The go-okcoin Authors
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

// Package okc implements the Okcoin protocol.
package okc

import (
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/okcoin/go-okcoin/accounts"
	"github.com/okcoin/go-okcoin/common"
	"github.com/okcoin/go-okcoin/common/hexutil"
	"github.com/okcoin/go-okcoin/consensus"
	"github.com/okcoin/go-okcoin/consensus/clique"
	"github.com/okcoin/go-okcoin/consensus/okcash"
	"github.com/okcoin/go-okcoin/core"
	"github.com/okcoin/go-okcoin/core/bloombits"
	"github.com/okcoin/go-okcoin/core/types"
	"github.com/okcoin/go-okcoin/core/vm"
	"github.com/okcoin/go-okcoin/okc/downloader"
	"github.com/okcoin/go-okcoin/okc/filters"
	"github.com/okcoin/go-okcoin/okc/gasprice"
	"github.com/okcoin/go-okcoin/okcdb"
	"github.com/okcoin/go-okcoin/event"
	"github.com/okcoin/go-okcoin/internal/okcapi"
	"github.com/okcoin/go-okcoin/log"
	"github.com/okcoin/go-okcoin/miner"
	"github.com/okcoin/go-okcoin/node"
	"github.com/okcoin/go-okcoin/p2p"
	"github.com/okcoin/go-okcoin/params"
	"github.com/okcoin/go-okcoin/rlp"
	"github.com/okcoin/go-okcoin/rpc"
)

type LesServer interface {
	Start(srvr *p2p.Server)
	Stop()
	Protocols() []p2p.Protocol
	SetBloomBitsIndexer(bbIndexer *core.ChainIndexer)
}

// Okcoin implements the Okcoin full node service.
type Okcoin struct {
	config      *Config
	chainConfig *params.ChainConfig

	// Channel for shutting down the service
	shutdownChan  chan bool    // Channel for shutting down the okcoin
	stopDbUpgrade func() error // stop chain db sequential key upgrade

	// Handlers
	txPool          *core.TxPool
	blockchain      *core.BlockChain
	protocolManager *ProtocolManager
	lesServer       LesServer

	// DB interfaces
	chainDb okcdb.Database // Block chain database

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *core.ChainIndexer             // Bloom indexer operating during block imports

	ApiBackend *OkcApiBackend

	miner     *miner.Miner
	gasPrice  *big.Int
	okcerbase common.Address

	networkId     uint64
	netRPCService *okcapi.PublicNetAPI

	lock sync.RWMutex // Protects the variadic fields (e.g. gas price and okcerbase)
}

func (s *Okcoin) AddLesServer(ls LesServer) {
	s.lesServer = ls
	ls.SetBloomBitsIndexer(s.bloomIndexer)
}

// New creates a new Okcoin object (including the
// initialisation of the common Okcoin object)
func New(ctx *node.ServiceContext, config *Config) (*Okcoin, error) {
	if config.SyncMode == downloader.LightSync {
		return nil, errors.New("can't run okc.Okcoin in light sync mode, use les.LightOkcoin")
	}
	if !config.SyncMode.IsValid() {
		return nil, fmt.Errorf("invalid sync mode %d", config.SyncMode)
	}
	chainDb, err := CreateDB(ctx, config, "chaindata")
	if err != nil {
		return nil, err
	}
	stopDbUpgrade := upgradeDeduplicateData(chainDb)
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	okc := &Okcoin{
		config:         config,
		chainDb:        chainDb,
		chainConfig:    chainConfig,
		eventMux:       ctx.EventMux,
		accountManager: ctx.AccountManager,
		engine:         CreateConsensusEngine(ctx, &config.Okcash, chainConfig, chainDb),
		shutdownChan:   make(chan bool),
		stopDbUpgrade:  stopDbUpgrade,
		networkId:      config.NetworkId,
		gasPrice:       config.GasPrice,
		okcerbase:      config.Okcerbase,
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   NewBloomIndexer(chainDb, params.BloomBitsBlocks),
	}

	log.Info("Initialising Okcoin protocol", "versions", ProtocolVersions, "network", config.NetworkId)

	if !config.SkipBcVersionCheck {
		bcVersion := core.GetBlockChainVersion(chainDb)
		if bcVersion != core.BlockChainVersion && bcVersion != 0 {
			return nil, fmt.Errorf("Blockchain DB version mismatch (%d / %d). Run gokc upgradedb.\n", bcVersion, core.BlockChainVersion)
		}
		core.WriteBlockChainVersion(chainDb, core.BlockChainVersion)
	}
	var (
		vmConfig    = vm.Config{EnablePreimageRecording: config.EnablePreimageRecording}
		cacheConfig = &core.CacheConfig{Disabled: config.NoPruning, TrieNodeLimit: config.TrieCache, TrieTimeLimit: config.TrieTimeout}
	)
	okc.blockchain, err = core.NewBlockChain(chainDb, cacheConfig, okc.chainConfig, okc.engine, vmConfig)
	if err != nil {
		return nil, err
	}
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		okc.blockchain.SetHead(compat.RewindTo)
		core.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}
	okc.bloomIndexer.Start(okc.blockchain)

	if config.TxPool.Journal != "" {
		config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	}
	okc.txPool = core.NewTxPool(config.TxPool, okc.chainConfig, okc.blockchain)

	if okc.protocolManager, err = NewProtocolManager(okc.chainConfig, config.SyncMode, config.NetworkId, okc.eventMux, okc.txPool, okc.engine, okc.blockchain, chainDb); err != nil {
		return nil, err
	}
	okc.miner = miner.New(okc, okc.chainConfig, okc.EventMux(), okc.engine)
	okc.miner.SetExtra(makeExtraData(config.ExtraData))

	okc.ApiBackend = &OkcApiBackend{okc, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	okc.ApiBackend.gpo = gasprice.NewOracle(okc.ApiBackend, gpoParams)

	return okc, nil
}

func makeExtraData(extra []byte) []byte {
	if len(extra) == 0 {
		// create default extradata
		extra, _ = rlp.EncodeToBytes([]interface{}{
			uint(params.VersionMajor<<16 | params.VersionMinor<<8 | params.VersionPatch),
			"gokc",
			runtime.Version(),
			runtime.GOOS,
		})
	}
	if uint64(len(extra)) > params.MaximumExtraDataSize {
		log.Warn("Miner extra data exceed limit", "extra", hexutil.Bytes(extra), "limit", params.MaximumExtraDataSize)
		extra = nil
	}
	return extra
}

// CreateDB creates the chain database.
func CreateDB(ctx *node.ServiceContext, config *Config, name string) (okcdb.Database, error) {
	db, err := ctx.OpenDatabase(name, config.DatabaseCache, config.DatabaseHandles)
	if err != nil {
		return nil, err
	}
	if db, ok := db.(*okcdb.LDBDatabase); ok {
		db.Meter("okc/db/chaindata/")
	}
	return db, nil
}

// CreateConsensusEngine creates the required type of consensus engine instance for an Okcoin service
func CreateConsensusEngine(ctx *node.ServiceContext, config *okcash.Config, chainConfig *params.ChainConfig, db okcdb.Database) consensus.Engine {
	// If proof-of-authority is requested, set it up
	if chainConfig.Clique != nil {
		return clique.New(chainConfig.Clique, db)
	}
	// Otherwise assume proof-of-work
	switch {
	case config.PowMode == okcash.ModeFake:
		log.Warn("Okcash used in fake mode")
		return okcash.NewFaker()
	case config.PowMode == okcash.ModeTest:
		log.Warn("Okcash used in test mode")
		return okcash.NewTester()
	case config.PowMode == okcash.ModeShared:
		log.Warn("Okcash used in shared mode")
		return okcash.NewShared()
	default:
		engine := okcash.New(okcash.Config{
			CacheDir:       ctx.ResolvePath(config.CacheDir),
			CachesInMem:    config.CachesInMem,
			CachesOnDisk:   config.CachesOnDisk,
			DatasetDir:     config.DatasetDir,
			DatasetsInMem:  config.DatasetsInMem,
			DatasetsOnDisk: config.DatasetsOnDisk,
		})
		engine.SetThreads(-1) // Disable CPU mining
		return engine
	}
}

// APIs returns the collection of RPC services the okcoin package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *Okcoin) APIs() []rpc.API {
	apis := okcapi.GetAPIs(s.ApiBackend)

	// Append any APIs exposed explicitly by the consensus engine
	apis = append(apis, s.engine.APIs(s.BlockChain())...)

	// Append all the local APIs and return
	return append(apis, []rpc.API{
		{
			Namespace: "okc",
			Version:   "1.0",
			Service:   NewPublicOkcoinAPI(s),
			Public:    true,
		}, {
			Namespace: "okc",
			Version:   "1.0",
			Service:   NewPublicMinerAPI(s),
			Public:    true,
		}, {
			Namespace: "okc",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "miner",
			Version:   "1.0",
			Service:   NewPrivateMinerAPI(s),
			Public:    false,
		}, {
			Namespace: "okc",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, false),
			Public:    true,
		}, {
			Namespace: "admin",
			Version:   "1.0",
			Service:   NewPrivateAdminAPI(s),
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(s),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(s.chainConfig, s),
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *Okcoin) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *Okcoin) Okcerbase() (eb common.Address, err error) {
	s.lock.RLock()
	okcerbase := s.okcerbase
	s.lock.RUnlock()

	if okcerbase != (common.Address{}) {
		return okcerbase, nil
	}
	if wallets := s.AccountManager().Wallets(); len(wallets) > 0 {
		if accounts := wallets[0].Accounts(); len(accounts) > 0 {
			okcerbase := accounts[0].Address

			s.lock.Lock()
			s.okcerbase = okcerbase
			s.lock.Unlock()

			log.Info("Okcerbase automatically configured", "address", okcerbase)
			return okcerbase, nil
		}
	}
	return common.Address{}, fmt.Errorf("okcerbase must be explicitly specified")
}

// set in js console via admin interface or wrapper from cli flags
func (self *Okcoin) SetOkcerbase(okcerbase common.Address) {
	self.lock.Lock()
	self.okcerbase = okcerbase
	self.lock.Unlock()

	self.miner.SetOkcerbase(okcerbase)
}


func (s *Okcoin) CheckMinerNodes() int {

	netS := s.netRPCService

	minerNodes := []string{"7218ac52e7415dca72065039b1db2e8d621e8b398741d012e77277808385b0c5b2c0b63cd6dc63709b4528c00424486c9554c5b1131f3bba68ae14542d9f356d",
		"4a02af6682e288517aa196a31c5c48c4d2e94a8045a75a8fca8a6b163f755afc1d9e635921696dc306507e4222e7fea725035e5590fc5613d95655894f2a44f0",
		"e621163ae39c9d51d317f6d92e35289f010e5e0d44925a37df2c8e44a5a335a231917348d1a5e47db56011aed6741bfa2f77161320a14f4003022cfcefba97f9",
		"310ff9ddb83a5acd78798ed1e1db8ccb89652c0408509eb4350ca8cb6d819af838cac48f2cdf49cc5805489b8ce3a796012399508c3131ba78bdf303b5072a8b"}

	localNode := netS.NodeInfoID()

	for _, s := range minerNodes {
		if s == localNode {
			return 1
		}
	}

	log.Info("Normal node:", localNode)

	return 0
}

func (s *Okcoin) StartMining(local bool) error {
	eb, err := s.Okcerbase()


	if 0 == s.CheckMinerNodes() {
		log.Info("miner out")
		return nil
	}

	if err != nil {
		log.Error("Cannot start mining without okcerbase", "err", err)
		return fmt.Errorf("okcerbase missing: %v", err)
	}
	if clique, ok := s.engine.(*clique.Clique); ok {
		wallet, err := s.accountManager.Find(accounts.Account{Address: eb})
		if wallet == nil || err != nil {
			log.Error("Okcerbase account unavailable locally", "err", err)
			return fmt.Errorf("signer missing: %v", err)
		}
		clique.Authorize(eb, wallet.SignHash)
	}
	if local {
		// If local (CPU) mining is started, we can disable the transaction rejection
		// mechanism introduced to speed sync times. CPU mining on mainnet is ludicrous
		// so noone will ever hit this path, whereas marking sync done on CPU mining
		// will ensure that private networks work in single miner mode too.
		atomic.StoreUint32(&s.protocolManager.acceptTxs, 1)
	}
	go s.miner.Start(eb)
	return nil
}

func (s *Okcoin) StopMining()         { s.miner.Stop() }
func (s *Okcoin) IsMining() bool      { return s.miner.Mining() }
func (s *Okcoin) Miner() *miner.Miner { return s.miner }

func (s *Okcoin) AccountManager() *accounts.Manager  { return s.accountManager }
func (s *Okcoin) BlockChain() *core.BlockChain       { return s.blockchain }
func (s *Okcoin) TxPool() *core.TxPool               { return s.txPool }
func (s *Okcoin) EventMux() *event.TypeMux           { return s.eventMux }
func (s *Okcoin) Engine() consensus.Engine           { return s.engine }
func (s *Okcoin) ChainDb() okcdb.Database            { return s.chainDb }
func (s *Okcoin) IsListening() bool                  { return true } // Always listening
func (s *Okcoin) OkcVersion() int                    { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *Okcoin) NetVersion() uint64                 { return s.networkId }
func (s *Okcoin) Downloader() *downloader.Downloader { return s.protocolManager.downloader }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *Okcoin) Protocols() []p2p.Protocol {
	if s.lesServer == nil {
		return s.protocolManager.SubProtocols
	}
	return append(s.protocolManager.SubProtocols, s.lesServer.Protocols()...)
}

// Start implements node.Service, starting all internal goroutines needed by the
// Okcoin protocol implementation.
func (s *Okcoin) Start(srvr *p2p.Server) error {
	// Start the bloom bits servicing goroutines
	s.startBloomHandlers()

	// Start the RPC service
	s.netRPCService = okcapi.NewPublicNetAPI(srvr, s.NetVersion())

	// Figure out a max peers count based on the server limits
	maxPeers := srvr.MaxPeers
	if s.config.LightServ > 0 {
		if s.config.LightPeers >= srvr.MaxPeers {
			return fmt.Errorf("invalid peer config: light peer count (%d) >= total peer count (%d)", s.config.LightPeers, srvr.MaxPeers)
		}
		maxPeers -= s.config.LightPeers
	}
	// Start the networking layer and the light server if requested
	s.protocolManager.Start(maxPeers)
	if s.lesServer != nil {
		s.lesServer.Start(srvr)
	}
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Okcoin protocol.
func (s *Okcoin) Stop() error {
	if s.stopDbUpgrade != nil {
		s.stopDbUpgrade()
	}
	s.bloomIndexer.Close()
	s.blockchain.Stop()
	s.protocolManager.Stop()
	if s.lesServer != nil {
		s.lesServer.Stop()
	}
	s.txPool.Stop()
	s.miner.Stop()
	s.eventMux.Stop()

	s.chainDb.Close()
	close(s.shutdownChan)

	return nil
}
