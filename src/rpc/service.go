package rpc

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	sequencerBindings "github.com/shutter-network/gnosh-contracts/gnoshcontracts/sequencer"
	shopContractBindings "github.com/shutter-network/shop-contracts/bindings"
)

type Processor struct {
	URL                      string
	RPCUrl                   string
	SigningKey               *ecdsa.PrivateKey
	SigningAddress           *common.Address
	KeyperSetChangeLookAhead int
	Client                   *ethclient.Client
	KeyBroadcastContract     *shopContractBindings.KeyBroadcastContract
	SequencerContract        *sequencerBindings.Sequencer
	KeyperSetManagerContract *shopContractBindings.KeyperSetManager
}

type RPCService interface {
	Name() string
	InjectProcessor(Processor)
}
