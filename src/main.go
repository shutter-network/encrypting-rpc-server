package main

import (
	"context"
	"math/big"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/shutter-network/encrypting-rpc-server/db"
	"github.com/shutter-network/encrypting-rpc-server/metrics"
	"github.com/shutter-network/encrypting-rpc-server/utils"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/url"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/metricsserver"
	metrics_server "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/metricsserver"
	medleyService "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"

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
	DelayInSeconds              int    `mapstructure:"delay-in-seconds"`
	EncryptedGasLimit           uint64 `mapstructure:"encrypted-gas-limit"`
	DbUrl                       string `mapstructure:"dburl"`
	WaitMinedInterval           int    `mapstructure:"wait-mined-interval"`
	MetricsConfig               metrics_server.MetricsConfig
	FetchBalanceDelay           int    `mapstructure:"fetch-balance-delay"`
	GasPriceMultiplier          int    `mapstructure:"fetch-balance-delay"`
	EffectivePriorityFee        uint64 `mapstructure:"effective-priority-fee"`
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

	cmd.PersistentFlags().IntVarP(
		&Config.DelayInSeconds,
		"delay-in-seconds",
		"",
		10,
		"server cache delay in seconds",
	)

	cmd.PersistentFlags().IntVarP(
		&Config.WaitMinedInterval,
		"wait-mined-interval",
		"",
		10,
		"server wait time for checking tx inclusion in seconds",
	)

	cmd.PersistentFlags().Uint64VarP(
		&Config.EncryptedGasLimit,
		"encrypted-gas-limit",
		"",
		1000000,
		"encrypted gas limit",
	)

	cmd.PersistentFlags().StringVarP(
		&Config.DbUrl,
		"dbUrl",
		"",
		"",
		"database url to connect to postgres database",
	)

	cmd.PersistentFlags().BoolVarP(
		&Config.MetricsConfig.Enabled,
		"metrics-enabled",
		"",
		false,
		"to enable promnetheus metrics",
	)

	cmd.PersistentFlags().StringVarP(
		&Config.MetricsConfig.Host,
		"metrics-host",
		"",
		"localhost",
		"metrics host",
	)

	cmd.PersistentFlags().Uint16VarP(
		&Config.MetricsConfig.Port,
		"metrics-port",
		"",
		9090,
		"metrics port",
	)

	cmd.PersistentFlags().IntVarP(
		&Config.FetchBalanceDelay,
		"fetch-balance-delay",
		"",
		10,
		"delay after which balance of signing address is re recorded",
	)

	cmd.PersistentFlags().IntVarP(
		&Config.GasPriceMultiplier,
		"gas-price-multiplier",
		"",
		2,
		"multiple to increase suggested gas price than inherent one",
	)

	cmd.PersistentFlags().Uint64VarP(
		&Config.EffectivePriorityFee,
		"effective-priority-fee",
		"",
		1000000000,
		"effective priority fee",
	)

	return cmd
}

func Start() error {
	signingKey, err := crypto.HexToECDSA(Config.SigningKey)
	if err != nil {
		utils.Logger.Fatal().Err(err).Msg("failed to parse signing key")
	}

	if Config.KeyperSetChangeLookAhead < 1 {
		utils.Logger.Fatal().Msg("keyper set change look ahead should be positive")
	}

	utils.Logger.Info().Msgf("Starting rpc server version %s", shversion.Version())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-termChan
		utils.Logger.Info().Str("signal", sig.String()).Msg("Received signal, shutting down")
		cancel()
	}()

	publicKeyECDSA, ok := signingKey.Public().(*ecdsa.PublicKey)
	if !ok {
		utils.Logger.Fatal().Msg("can not create public key")
	}
	publicAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	client, err := ethclient.Dial(Config.RPCUrl)
	if err != nil {
		utils.Logger.Fatal().Err(err).Msg("can not connect to rpc")
	}

	broadcastContract, err := shopContractBindings.NewKeyBroadcastContract(common.HexToAddress(Config.KeyBroadcastContractAddress), client)
	if err != nil {
		utils.Logger.Fatal().Err(err).Msg("can not use Keybroadcast contract")
	}
	sequencerContract, err := sequencerBindings.NewSequencer(common.HexToAddress(Config.SequencerAddress), client)
	if err != nil {
		utils.Logger.Fatal().Err(err).Msg("can not use Sequencer contract")
	}
	keyperSetManagerContract, err := shopContractBindings.NewKeyperSetManager(common.HexToAddress(Config.KeyperSetManagerAddress), client)
	if err != nil {
		utils.Logger.Fatal().Err(err).Msg("can not use Sequencer contract")
	}

	dbInst, err := db.InitialMigration(Config.DbUrl)
	if err != nil {
		utils.Logger.Fatal().Err(err).Msg("can not instantiate postgres")
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
		Db:                       dbInst,
		MetricsConfig:            &Config.MetricsConfig,
	}

	backendURL := &url.URL{}
	err = backendURL.UnmarshalText([]byte(Config.RPCUrl))
	if err != nil {
		utils.Logger.Fatal().Err(err).Msg("failed to parse RPCUrl")
	}

	if Config.MetricsConfig.Enabled {
		metrics.InitMetrics()
		processor.MetricsServer = metricsserver.New(&Config.MetricsConfig)
	}

	config := rpc.Config{
		BackendURL:           backendURL,
		HTTPListenAddress:    Config.HTTPListenAddress,
		DelayInSeconds:       Config.DelayInSeconds,
		EncryptedGasLimit:    Config.EncryptedGasLimit,
		WaitMinedInterval:    Config.WaitMinedInterval,
		FetchBalanceDelay:    Config.FetchBalanceDelay,
		GasMultiplier:        big.NewInt(int64(Config.GasPriceMultiplier)),
		EffectivePriorityFee: Config.EffectivePriorityFee,
	}

	service := server.NewRPCService(processor, config, dbInst)
	utils.Logger.Info().Str("listen-on", Config.HTTPListenAddress).Msg("Serving JSON-RPC")

	func() {
		err = medleyService.Run(ctx, service)
		if err != nil {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	return err
}

func main() {
	status := 0

	if err := Cmd().Execute(); err != nil {
		utils.Logger.Info().Err(err).Msg("failed running server")
		status = 1
	}

	os.Exit(status)
}
