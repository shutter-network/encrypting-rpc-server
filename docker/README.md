# Run encrypted-rpc for shutterized gnosis chain as a hosted service

## Installation

Clone this repository and navigate to the `docker` subfolder:

```sh
git clone https://github.com/shutter-network/encrypting-rpc-server
git submodule update --init
cd docker
```

You need to have `docker` (version `2.9.0` or greater) installed.

## Configuration

You need to copy `template.env` to `.env` and fill in the blanks:

```sh
cp template.env .env
```

The following values need to be filled in:

**Note: the snippet in this README uses dummy values. Make sure to use the correct values for
your server and the actual deployment!**

```
# a DNS entry for this server. Users will use `https://${SERVICE_DOMAIN_NAME}` (with the value given) 
# for their RPC.
SERVICE_DOMAIN_NAME=shutterized.rpc.myserver.com

# an url for a gnosis chain RPC service (which we will proxy with encryption)
UPSTREAM_RPC=rpc.myserver.com

# the private key for Sequencer submissions as non-0x-prefixed hex
SIGNING_KEY=bbfbee4961061d506ffbb11dfea64eba16355cbf1d9c29613126ba7fec0aed5d

# the ethereum address of the key broadcast contract
KEY_BROADCAST_CONTRACT_ADDRESS=0x4200000000000000000000000000000000000068

# the ethereum address of the sequencer contract
SEQUENCER_CONTRACT_ADDRESS=0x4200000000000000000000000000000000000069

# the ethereum address of the keyper set manager contract
KEYPER_SET_MANAGER_CONTRACT_ADDRESS=0x4200000000000000000000000000000000000070
```

## Running
Whenever there were updates, you should update the repository.

```sh
git pull
```

To run the service, you can call

```sh
docker compose build && docker compose up -d
```

This will build the encrypting rpc server from source and start the service in detached
mode. To inspect what is going on, run

```sh
docker compose ps
# or
docker compose logs
```