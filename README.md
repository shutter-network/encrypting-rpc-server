# Encrypting RPC Server

This server is using Gnosis Chain Configuration and Shutter Network Mempool encryption to protect your transactions against MEV.

## Requirements

To deploy and use for requests other than eth_sendRawTransaction
* Ganache >= 7.0.0
* Foundry >= 0.2.0

To start and build the server:
* Docker

## How to run example

1. First you need to use below command to get submodules:
`git submodule update --init`
2. Start a ganache server by:
`ganache -b 5 -t 2021-12-08T20:55:40 --wallet.mnemonic brownie`
3. Deploy contracts in /src folder by using:
`make compile-contracts && make deploy`
4. Build the container image:
`docker build . -t encrypting-rpc-server`
5. Run the server:
`docker run -p 8546:8546 encrypting-rpc-server --signing-key bbfbee4961061d506ffbb11dfea64eba16355cbf1d9c29613126ba7fec0aed5d --key-broadcast-contract-address {KEY_BROADCAST_CONTRACT_ADDRESS} --sequencer-address {SEQUENCER_CONTRACT_ADDRESS} --keyper-set-manager-address {KEYPER_SET_MANAGER_CONTRACT_ADDRESS} --rpc-url http://host.docker.internal:8545`

Contract  addresses can be obtained from `gnosh-contracts/broadcast/deploy.s.sol/1337/run-latest.json`. There are `CREATE` `transactionType` and select `contractName` as what you need then  you have the `contractAddress`.  

There are other options you can use like:

* `rpc-url`: RPC URL from alchemy/infura or other providers. Default: http://localhost:8545
* `http-listen-address`: Which address this server runs. Default: :8546
* `keyper-set-change-look-ahead`: How much ahead your transactions should be revealed.
* `dbUrl`: configuration for postgres db, eg: --dbUrl "host=localhost port=5432 user=postgres dbname=postgres password=postgres sslmode=disable"

