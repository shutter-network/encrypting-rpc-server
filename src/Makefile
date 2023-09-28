.POSIX:

GO		?= go
GOFLAGS		?=
TINYGO		?= tinygo
WASMOPT     ?= wasm-opt
BINDIR		?= ./bin
EXECUTABLE	?= ${BINDIR}/encrypting-rpc-server
GOPATH          ?= $(${GO} env GOPATH)

build:
	${GO} build ${GOFLAGS} -o ${EXECUTABLE}

compile-contracts:
	cd ../gnosh-contracts/src && forge build --extra-output-files abi --force

abigen:
	abigen --abi ../gnosh-contracts/out/KeyBroadcastContract.sol/KeyBroadcastContract.abi.json --pkg contracts --out contracts/KeyBroadcastContract.go --type KeyBroadcastContract && \
	abigen --abi ../gnosh-contracts/out/Sequencer.sol/Sequencer.abi.json --pkg contracts --out contracts/SequencerContract.go --type SequencerContract && \
	abigen --abi ../gnosh-contracts/out/KeyperSetManager.sol/KeyperSetManager.abi.json --pkg contracts --out contracts/KeyperSetManagerContract.go --type KeyperSetManagerContract


clean:
	rm -f ${EXECUTABLE}

.PHONY: build clean compile-contracts abigen