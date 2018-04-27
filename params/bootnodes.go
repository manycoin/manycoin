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

package params

// MainnetBootnodes are the enode URLs of the P2P bootstrap nodes running on
// the main Okcoin network.
var MainnetBootnodes = []string{
	// Okcoin Foundation Go Bootnodes
	"enode://31c56abd7a8f84f3931565a586018f8f244d85c44ce54ec5a413139f5646246b2bfc420ed7ef984b57a2bc14deb0c309ca620c3b9317a33938cbb4882fe66b32@[101.132.164.100]:5000",
	"enode://16d202b7f5d230d4729067d65a706e576a774a1f5ea4931b9a1db8a2308822663d2d1aaf3bcea34530b84a1eb2b3c63acbc7d1b42142a58ab8e62e028ac5cc33@[47.100.197.171]:5000",
	"enode://a49f2816b1900b371b61f46c10ca79795d2a69401f41bc91a5b7c5310006f8a7653d8966026967d7e45f0b09e0743a77c06b552cf1911c4aaacbbbf51b5d5d36@[47.100.196.186]:5000",
	"enode://21e5f65ad6f85ae83407fd417a5731221197f5169c9abdd5018b982ebb216771ded0f2fe7e2af9b4052b42f0052ab0c6504814a61ba99a351267625918206448@[13.230.43.125]:5003",
}

// TestnetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Ropsten test network.
var TestnetBootnodes = []string{
	"enode://30b7ab30a01c124a6cceca36863ece12c4f5fa68e3ba9b0b51407ccc002eeed3b3102d20a88f1c1d3c3154e2449317b8ef95090e77b312d5cc39354f86d5d606@52.176.7.10:30303", // US-Azure gokc
}

// RinkebyBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Rinkeby test network.
var RinkebyBootnodes = []string{
	"enode://a24ac7c5484ef4ed0c5eb2d36620ba4e4aa13b8c84684e1b4aab0cebea2ae45cb4d375b77eab56516d34bfbd3c1a833fc51296ff084b770b94fb9028c4d25ccf@52.169.42.101:30303", // IE
}

// DiscoveryV5Bootnodes are the enode URLs of the P2P bootstrap nodes for the
// experimental RLPx v5 topic-discovery network.
var DiscoveryV5Bootnodes = []string{
	"enode://06051a5573c81934c9554ef2898eb13b33a34b94cf36b202b69fde139ca17a85051979867720d4bdae4323d4943ddf9aeeb6643633aa656e0be843659795007a@35.177.226.168:30303",
	"enode://0cc5f5ffb5d9098c8b8c62325f3797f56509bff942704687b6530992ac706e2cb946b90a34f1f19548cd3c7baccbcaea354531e5983c7d1bc0dee16ce4b6440b@40.118.3.223:30304",
	"enode://1c7a64d76c0334b0418c004af2f67c50e36a3be60b5e4790bdac0439d21603469a85fad36f2473c9a80eb043ae60936df905fa28f1ff614c3e5dc34f15dcd2dc@40.118.3.223:30306",
	"enode://85c85d7143ae8bb96924f2b54f1b3e70d8c4d367af305325d30a61385a432f247d2c75c45c6b4a60335060d072d7f5b35dd1d4c45f76941f62a4f83b6e75daaf@40.118.3.223:30307",
}
