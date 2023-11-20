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
	cd ../gnosh-contracts && forge build --extra-output-files abi --force

abigen:
	abigen --abi ../gnosh-contracts/out/KeyBroadcastContract.sol/KeyBroadcastContract.abi.json --pkg contracts --out contracts/KeyBroadcastContract.go --type KeyBroadcastContract && \
	abigen --abi ../gnosh-contracts/out/Sequencer.sol/Sequencer.abi.json --pkg contracts --out contracts/SequencerContract.go --type SequencerContract && \
	abigen --abi ../gnosh-contracts/out/KeyperSetManager.sol/KeyperSetManager.abi.json --pkg contracts --out contracts/KeyperSetManagerContract.go --type KeyperSetManagerContract

deploy:
	cd ../gnosh-contracts && BROADCASTER=0xA868bC7c1AF08B8831795FAC946025557369F69C PRIVATE_KEY=0xa26ebb1df46424945009db72c7a7ba034027450784b93f34000169b35fd3adaa forge script script/deploy.s.sol --rpc-url http://localhost:8545 --broadcast -vvvv

clean:
	rm -f ${EXECUTABLE}

lint-changes:
	base=`git merge-base HEAD origin/main`; \
	golangci-lint run --new-from-rev $${base}

.PHONY: build clean compile-contracts abigen deploy lint-changes