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

// Contains the metrics collected by the downloader.

package downloader

import (
	"github.com/okcoin/go-okcoin/metrics"
)

var (
	headerInMeter      = metrics.NewRegisteredMeter("okc/downloader/headers/in", nil)
	headerReqTimer     = metrics.NewRegisteredTimer("okc/downloader/headers/req", nil)
	headerDropMeter    = metrics.NewRegisteredMeter("okc/downloader/headers/drop", nil)
	headerTimeoutMeter = metrics.NewRegisteredMeter("okc/downloader/headers/timeout", nil)

	bodyInMeter      = metrics.NewRegisteredMeter("okc/downloader/bodies/in", nil)
	bodyReqTimer     = metrics.NewRegisteredTimer("okc/downloader/bodies/req", nil)
	bodyDropMeter    = metrics.NewRegisteredMeter("okc/downloader/bodies/drop", nil)
	bodyTimeoutMeter = metrics.NewRegisteredMeter("okc/downloader/bodies/timeout", nil)

	receiptInMeter      = metrics.NewRegisteredMeter("okc/downloader/receipts/in", nil)
	receiptReqTimer     = metrics.NewRegisteredTimer("okc/downloader/receipts/req", nil)
	receiptDropMeter    = metrics.NewRegisteredMeter("okc/downloader/receipts/drop", nil)
	receiptTimeoutMeter = metrics.NewRegisteredMeter("okc/downloader/receipts/timeout", nil)

	stateInMeter   = metrics.NewRegisteredMeter("okc/downloader/states/in", nil)
	stateDropMeter = metrics.NewRegisteredMeter("okc/downloader/states/drop", nil)
)
