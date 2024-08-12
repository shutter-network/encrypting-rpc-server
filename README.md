# Encrypting RPC Server

This server is using Gnosis Chain Configuration and Shutter Network Mempool encryption to protect your transactions against MEV.

## Requirements

To start and build the server within a docker container:

1. Clone the repository

2. Change to the docker directory

3. Copy the environment variables from `template.env` and adjust them as needed.
   Use the values below to run the server against the Chiado testnet with the deployed contracts as of Aug 1.

    ```env
    SERVICE_DOMAIN_NAME=localhost
    # an url for a gnosis chain RPC service (which we will proxy with encryption)
    UPSTREAM_RPC=https://rpc.chiadochain.net
    # the private key for Sequencer submissions as non-0x-prefixed hex
    SIGNING_KEY=YOUR_PRIVATE_KEY
    # the ethereum address of the key broadcast contract
    KEY_BROADCAST_CONTRACT_ADDRESS=0xDd9Ea21f682a6484ac40D36c97Fa056Fbce9004f
    # the ethereum address of the sequencer contract
    SEQUENCER_CONTRACT_ADDRESS=0xAC3209DCBced710Dc2612bD714b9EC947a6d1e8f
    # the ethereum address of the keyper set manager contract
    KEYPER_SET_MANAGER_CONTRACT_ADDRESS=0x6759Ab83de6f7d5bc4cf02d41BbB3Bd1500712E1
    # the delay in seconds
    DELAY_IN_SECONDS=100
    # the db url
    DB_URL="postgres://postgres:postgres@postgres:5432/postgres?sslmode=disable"
    WAIT_MINED_INTERVAL=10
    ```

4. Build and run the application:

    ```sh
    docker-compose up --build -d
    ```

5. Alternatively, you can run the server locally following the below steps:

    ```sh
    cd src
    go build
    ```

   And then:

    ```sh
    ./encrypting-rpc-server \
    --signing-key YOUR_PRIVATE_KEY \
    --key-broadcast-contract-address 0xDd9Ea21f682a6484ac40D36c97Fa056Fbce9004f \
    --sequencer-address 0xAC3209DCBced710Dc2612bD714b9EC947a6d1e8f \
    --keyper-set-manager-address 0x6759Ab83de6f7d5bc4cf02d41BbB3Bd1500712E1 \
    --rpc-url https://rpc.chiadochain.net \
    --dbUrl "host=localhost port=5432 user=postgres dbname=postgres password=postgres sslmode=disable"
    --wait-mined-interval 10
    ```
