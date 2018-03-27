// Copyright 2015 The go-okcoin Authors
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

package okc

import (
	"context"
	"math/big"

	"github.com/okcoin/go-okcoin/accounts"
	"github.com/okcoin/go-okcoin/common"
	"github.com/okcoin/go-okcoin/common/math"
	"github.com/okcoin/go-okcoin/core"
	"github.com/okcoin/go-okcoin/core/bloombits"
	"github.com/okcoin/go-okcoin/core/state"
	"github.com/okcoin/go-okcoin/core/types"
	"github.com/okcoin/go-okcoin/core/vm"
	"github.com/okcoin/go-okcoin/okc/downloader"
	"github.com/okcoin/go-okcoin/okc/gasprice"
	"github.com/okcoin/go-okcoin/okcdb"
	"github.com/okcoin/go-okcoin/event"
	"github.com/okcoin/go-okcoin/params"
	"github.com/okcoin/go-okcoin/rpc"
)

// OkcApiBackend implements okcapi.Backend for full nodes
type OkcApiBackend struct {
	okc *Okcoin
	gpo *gasprice.Oracle
}

func (b *OkcApiBackend) ChainConfig() *params.ChainConfig {
	return b.okc.chainConfig
}

func (b *OkcApiBackend) CurrentBlock() *types.Block {
	return b.okc.blockchain.CurrentBlock()
}

func (b *OkcApiBackend) SetHead(number uint64) {
	b.okc.protocolManager.downloader.Cancel()
	b.okc.blockchain.SetHead(number)
}

func (b *OkcApiBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.okc.miner.PendingBlock()
		return block.Header(), nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.okc.blockchain.CurrentBlock().Header(), nil
	}
	return b.okc.blockchain.GetHeaderByNumber(uint64(blockNr)), nil
}

func (b *OkcApiBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.okc.miner.PendingBlock()
		return block, nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.okc.blockchain.CurrentBlock(), nil
	}
	return b.okc.blockchain.GetBlockByNumber(uint64(blockNr)), nil
}

func (b *OkcApiBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// Pending state is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block, state := b.okc.miner.Pending()
		return state, block.Header(), nil
	}
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	stateDb, err := b.okc.BlockChain().StateAt(header.Root)
	return stateDb, header, err
}

func (b *OkcApiBackend) GetBlock(ctx context.Context, blockHash common.Hash) (*types.Block, error) {
	return b.okc.blockchain.GetBlockByHash(blockHash), nil
}

func (b *OkcApiBackend) GetReceipts(ctx context.Context, blockHash common.Hash) (types.Receipts, error) {
	return core.GetBlockReceipts(b.okc.chainDb, blockHash, core.GetBlockNumber(b.okc.chainDb, blockHash)), nil
}

func (b *OkcApiBackend) GetLogs(ctx context.Context, blockHash common.Hash) ([][]*types.Log, error) {
	receipts := core.GetBlockReceipts(b.okc.chainDb, blockHash, core.GetBlockNumber(b.okc.chainDb, blockHash))
	if receipts == nil {
		return nil, nil
	}
	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

func (b *OkcApiBackend) GetTd(blockHash common.Hash) *big.Int {
	return b.okc.blockchain.GetTdByHash(blockHash)
}

func (b *OkcApiBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	state.SetBalance(msg.From(), math.MaxBig256)
	vmError := func() error { return nil }

	context := core.NewEVMContext(msg, header, b.okc.BlockChain(), nil)
	return vm.NewEVM(context, state, b.okc.chainConfig, vmCfg), vmError, nil
}

func (b *OkcApiBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return b.okc.BlockChain().SubscribeRemovedLogsEvent(ch)
}

func (b *OkcApiBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return b.okc.BlockChain().SubscribeChainEvent(ch)
}

func (b *OkcApiBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return b.okc.BlockChain().SubscribeChainHeadEvent(ch)
}

func (b *OkcApiBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	return b.okc.BlockChain().SubscribeChainSideEvent(ch)
}

func (b *OkcApiBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.okc.BlockChain().SubscribeLogsEvent(ch)
}

func (b *OkcApiBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.okc.txPool.AddLocal(signedTx)
}

func (b *OkcApiBackend) GetPoolTransactions() (types.Transactions, error) {
	pending, err := b.okc.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs types.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *OkcApiBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	return b.okc.txPool.Get(hash)
}

func (b *OkcApiBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.okc.txPool.State().GetNonce(addr), nil
}

func (b *OkcApiBackend) Stats() (pending int, queued int) {
	return b.okc.txPool.Stats()
}

func (b *OkcApiBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.okc.TxPool().Content()
}

func (b *OkcApiBackend) SubscribeTxPreEvent(ch chan<- core.TxPreEvent) event.Subscription {
	return b.okc.TxPool().SubscribeTxPreEvent(ch)
}

func (b *OkcApiBackend) Downloader() *downloader.Downloader {
	return b.okc.Downloader()
}

func (b *OkcApiBackend) ProtocolVersion() int {
	return b.okc.OkcVersion()
}

func (b *OkcApiBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *OkcApiBackend) ChainDb() okcdb.Database {
	return b.okc.ChainDb()
}

func (b *OkcApiBackend) EventMux() *event.TypeMux {
	return b.okc.EventMux()
}

func (b *OkcApiBackend) AccountManager() *accounts.Manager {
	return b.okc.AccountManager()
}

func (b *OkcApiBackend) BloomStatus() (uint64, uint64) {
	sections, _, _ := b.okc.bloomIndexer.Sections()
	return params.BloomBitsBlocks, sections
}

func (b *OkcApiBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.okc.bloomRequests)
	}
}
