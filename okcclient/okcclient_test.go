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

package okcclient

import "github.com/okcoin/go-okcoin"

// Verify that Client implements the okcoin interfaces.
var (
	_ = okcoin.ChainReader(&Client{})
	_ = okcoin.TransactionReader(&Client{})
	_ = okcoin.ChainStateReader(&Client{})
	_ = okcoin.ChainSyncReader(&Client{})
	_ = okcoin.ContractCaller(&Client{})
	_ = okcoin.GasEstimator(&Client{})
	_ = okcoin.GasPricer(&Client{})
	_ = okcoin.LogFilterer(&Client{})
	_ = okcoin.PendingStateReader(&Client{})
	// _ = okcoin.PendingStateEventer(&Client{})
	_ = okcoin.PendingContractCaller(&Client{})
)
