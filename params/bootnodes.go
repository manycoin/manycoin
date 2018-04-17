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
	"enode://34551bd860482cc9c012480201728b95f9c813d3c4dcebd8081f52414578d24f371f001130475bbff259969f06d0d36d3b49792eb7cfd58346e146f510a58889@[13.230.43.125]:5060", // exploror
	"enode://46c9c03d10dee434f9e02e054a5d517cbbcf4550f2f0ad169de0d6cf3e2ba808f6f162558cc66c4f12691170386413f90321797845ce95daaa6a4196c79c322d@[13.230.73.233]:5003",
	"enode://d3035ee8cadb5ab184adbea2882d85279ebc9d10e1c46a196bdbb26f957e92665db947a5b0a04f697523ee40604106f22995b2607de2cc186282ec5097930c7a@[13.230.38.146]:5003",
	"enode://42c98ea9c6da2a67d5f346e046071e9cd75138943c28f3ee516145e89f302a25618bc2580c7a0b0e2076bd2a664b408c7f40399b0d6d0455e9cf7bfd561db5b7@[13.230.43.97]:5003",

"enode://7218ac52e7415dca72065039b1db2e8d621e8b398741d012e77277808385b0c5b2c0b63cd6dc63709b4528c00424486c9554c5b1131f3bba68ae14542d9f356d@[13.230.43.125]:5060",
"enode://4a02af6682e288517aa196a31c5c48c4d2e94a8045a75a8fca8a6b163f755afc1d9e635921696dc306507e4222e7fea725035e5590fc5613d95655894f2a44f0@[13.230.73.233]:5003",
"enode://e621163ae39c9d51d317f6d92e35289f010e5e0d44925a37df2c8e44a5a335a231917348d1a5e47db56011aed6741bfa2f77161320a14f4003022cfcefba97f9@[13.230.38.146]:5003",
"enode://310ff9ddb83a5acd78798ed1e1db8ccb89652c0408509eb4350ca8cb6d819af838cac48f2cdf49cc5805489b8ce3a796012399508c3131ba78bdf303b5072a8b@[13.230.43.97]:5003",
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
