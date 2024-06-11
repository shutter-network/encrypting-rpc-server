package main

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/shutter-network/encrypting-rpc-server/config"
	"github.com/shutter-network/encrypting-rpc-server/requests"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/url"
	medleyService "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"os"
	"os/signal"
	"syscall"

	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shutter-network/encrypting-rpc-server/rpc"
	"github.com/shutter-network/encrypting-rpc-server/server"
	sequencerBindings "github.com/shutter-network/gnosh-contracts/gnoshcontracts/sequencer"
	shopContractBindings "github.com/shutter-network/shop-contracts/bindings"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/shversion"

	"github.com/spf13/cobra"
)

var Config struct {
	SigningKey                  string `mapstructure:"signing-key"`
	KeyperSetChangeLookAhead    int    `mapstructure:"keyper-set-change-look-ahead"`
	RPCUrl                      string `mapstructure:"rpc-url"`
	HTTPListenAddress           string `mapstructure:"http-listen-address"`
	KeyBroadcastContractAddress string `mapstructure:"key-broadcast-contract-address"`
	SequencerAddress            string `mapstructure:"sequencer-address"`
	KeyperSetManagerAddress     string `mapstructure:"keyperset-manager-address"`
}

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start encrypting rpc server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return Start()
		},
	}

	cmd.PersistentFlags().IntVarP(
		&Config.KeyperSetChangeLookAhead,
		"keyper-set-change-look-ahead",
		"i",
		1,
		"keyper set change look ahead",
	)

	cmd.PersistentFlags().StringVarP(
		&Config.SigningKey,
		"signing-key",
		"",
		"",
		"private key to sign and submit transactions with",
	)

	cmd.PersistentFlags().StringVarP(
		&Config.HTTPListenAddress,
		"http-listen-address",
		"",
		":8546",
		"server listening address",
	)

	cmd.PersistentFlags().StringVarP(
		&Config.RPCUrl,
		"rpc-url",
		"",
		"http://localhost:8545",
		"address to forward requests to",
	)

	cmd.PersistentFlags().StringVarP(
		&Config.KeyBroadcastContractAddress,
		"key-broadcast-contract-address",
		"",
		"",
		"key broadcast contract address",
	)

	cmd.PersistentFlags().StringVarP(
		&Config.SequencerAddress,
		"sequencer-address",
		"",
		"",
		"sequencer contract address",
	)

	cmd.PersistentFlags().StringVarP(
		&Config.KeyperSetManagerAddress,
		"keyper-set-manager-address",
		"",
		"",
		"keyper set manager contract address",
	)

	return cmd
}

func Start() error {
	signingKey, err := crypto.HexToECDSA(Config.SigningKey)
	if err != nil {
		server.Logger.Fatal().Err(err).Msg("failed to parse signing key")
	}

	if Config.KeyperSetChangeLookAhead < 1 {
		server.Logger.Fatal().Msg("keyper set change look ahead should be positive")
	}

	server.Logger.Info().Msgf("Starting rpc server version %s", shversion.Version())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-termChan
		server.Logger.Info().Str("signal", sig.String()).Msg("Received signal, shutting down")
		cancel()
	}()

	publicKeyECDSA, ok := signingKey.Public().(*ecdsa.PublicKey)
	if !ok {
		server.Logger.Fatal().Msg("can not create public key")
	}
	publicAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	client, err := ethclient.Dial(Config.RPCUrl)
	if err != nil {
		server.Logger.Fatal().Err(err).Msg("can not connect to rpc")
	}

	broadcastContract, err := shopContractBindings.NewKeyBroadcastContract(common.HexToAddress(Config.KeyBroadcastContractAddress), client)
	if err != nil {
		server.Logger.Fatal().Err(err).Msg("can not use Keybroadcast contract")
	}
	sequencerContract, err := sequencerBindings.NewSequencer(common.HexToAddress(Config.SequencerAddress), client)
	if err != nil {
		server.Logger.Fatal().Err(err).Msg("can not use Sequencer contract")
	}
	keyperSetManagerContract, err := shopContractBindings.NewKeyperSetManager(common.HexToAddress(Config.KeyperSetManagerAddress), client)
	if err != nil {
		server.Logger.Fatal().Err(err).Msg("can not use Sequencer contract")
	}

	processor := rpc.Processor{
		URL:                      Config.HTTPListenAddress,
		RPCUrl:                   Config.RPCUrl,
		SigningKey:               signingKey,
		SigningAddress:           &publicAddress,
		KeyperSetChangeLookAhead: Config.KeyperSetChangeLookAhead,
		Client:                   client,
		KeyBroadcastContract:     broadcastContract,
		SequencerContract:        sequencerContract,
		KeyperSetManagerContract: keyperSetManagerContract,
	}

	backendURL := &url.URL{}
	err = backendURL.UnmarshalText([]byte(Config.RPCUrl))
	if err != nil {
		server.Logger.Fatal().Err(err).Msg("failed to parse RPCUrl")
	}

	config := server.Config{
		BackendURL:        backendURL,
		HTTPListenAddress: Config.HTTPListenAddress,
	}

	service := server.NewRPCService(processor, &config)
	server.Logger.Info().Str("listen-on", Config.HTTPListenAddress).Msg("Serving JSON-RPC")

	func() {
		err = medleyService.Run(ctx, service)
		if err != nil {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	return err
}

func main() {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}
	go requests.FetchNewBlocks(cfg.WebSocketURL)

	status := 0
	if err := Cmd().Execute(); err != nil {
		server.Logger.Info().Err(err).Msg("failed running server")
		status = 1
	}
	os.Exit(status)
}
