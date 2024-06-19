package rpc

import (
	"context"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	sequencerBindings "github.com/shutter-network/gnosh-contracts/gnoshcontracts/sequencer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/url"
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

type Config struct {
	BackendURL        *url.URL
	WebsocketURL      *url.URL
	HTTPListenAddress string
	DelayFactor       int
}

type RPCService interface {
	Name() string
	InjectProcessor(Processor)
	NewBlock(ctx context.Context, blockNumber uint64)
	SendRawTransaction(ctx context.Context, s string) (*common.Hash, error)
	AddConfig(Config)
}
