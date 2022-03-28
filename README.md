# Oracle EVM

[Avalanche](https://docs.avax.network/learn/platform-overview) is a network composed of multiple blockchains.
Each blockchain is an instance of a Virtual Machine (VM), much like an object in an object-oriented language is an instance of a class.
That is, the VM defines the behavior of the blockchain.

The OracleEVM is a customised EVM that uses stateful pre-compiles to create gas efficient access to high fidelity financial information in every block.

How it works:
1. Validators stream deterministic financial data from the decentralised pyth network on solana
2. During block production, the financial data is included in the block header
3. Validators vote on the validity of the block (including the contained financial data)
4. Financial data is written into the state db when blocks are accepted by nodes
5. Stateful pre-compiles make data directly accessible from smart contracts

The main benefits of the OracleEVM are:
- The conservation of block space
- Very gas efficient access to financial data via pre-compiles
- Validity of financial data is enforced by vm block verification and consensus

The Subnet EVM runs in a separate process from the main AvalancheGo process and communicates with it over a local gRPC connection.

## Running test network

You can intitialize a local test network using run.sh script in `scripts/run.sh`

