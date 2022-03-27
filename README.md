# Oracle EVM

[Avalanche](https://docs.avax.network/learn/platform-overview) is a network composed of multiple blockchains.
Each blockchain is an instance of a Virtual Machine (VM), much like an object in an object-oriented language is an instance of a class.
That is, the VM defines the behavior of the blockchain.

Oracle EVM is a showcase for a Virtual Machine (VM) where price data is written into the stateDb without contract calls.

Instead price data is an extra field in the block header and is propogated across the network

This chain implements the Ethereum Virtual Machine and supports Solidity smart contracts as well as most other Ethereum client functionality.

The Subnet EVM runs in a separate process from the main AvalancheGo process and communicates with it over a local gRPC connection.

## Running test network

You can intitialize a local test network using run.sh script in `scripts/run.sh`

