package rpc

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shutter-network/encrypting-rpc-server/contracts"
)

type Processor struct {
	URL                      string
	RPCUrl                   string
	SigningKey               *ecdsa.PrivateKey
	SigningAddress           *common.Address
	KeyperSetChangeLookAhead int
	Client                   *ethclient.Client
	KeyBroadcastContract     *contracts.KeyBroadcastContract
	SequencerContract        *contracts.SequencerContract
	KeyperSetManagerContract *contracts.KeyperSetManagerContract
}

type RPCService interface {
	Name() string
	InjectProcessor(Processor)
}
