package test

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
	"github.com/shutter-network/encrypting-rpc-server/db"
	"github.com/shutter-network/encrypting-rpc-server/rpc"
	"github.com/shutter-network/encrypting-rpc-server/server"
	"github.com/shutter-network/encrypting-rpc-server/utils"
	sequencerBindings "github.com/shutter-network/gnosh-contracts/gnoshcontracts/sequencer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/url"
	medleyService "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	medleyKeygen "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testkeygen"
	shopContractBindings "github.com/shutter-network/shop-contracts/bindings"
	"github.com/shutter-network/shutter/shlib/shcrypto"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gorm_logger "gorm.io/gorm/logger"
)

func init() {
	TestKeygen, _ = medleyKeygen.NewEonKeys(rand.Reader, 3, 2)
	TestEonKey = TestKeygen.EonPublicKey()
}

var (
	TestKeygen      *medleyKeygen.EonKeys
	TestEonKey      *shcrypto.EonPublicKey
	DeployerKey     = "a26ebb1df46424945009db72c7a7ba034027450784b93f34000169b35fd3adaa"
	DeployerAddress = "0xA868bC7c1AF08B8831795FAC946025557369F69C"
	BackendURL      = "http://localhost:8545"
	ServerURL       = "http://localhost:8546"
	TxPrivKey       = "bbfbee4961061d506ffbb11dfea64eba16355cbf1d9c29613126ba7fec0aed5d"
	TxFromAddress   = "0x66aB6D9362d4F35596279692F0251Db635165871"
	TxToAddress     = "0x33A4622B82D4c04a53e170c638B944ce27cffce3"
)

type DeployTx struct {
	ContractName    string `json:"contractName"`
	ContractAddress string `json:"contractAddress"`
}

type DeployData struct {
	Transactions []DeployTx `json:"transactions"`
}

func GetContractData() (map[string]common.Address, error) {
	contractInfo := make(map[string]common.Address)
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	deployDataPath := wd + "/../../gnosh-contracts/broadcast/deploy.s.sol/1337/run-latest.json"
	jsonFile, err := os.Open(deployDataPath)
	if err != nil {
		return nil, err
	}

	byteData, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var data DeployData

	err = json.Unmarshal(byteData, &data)
	if err != nil {
		return nil, err
	}

	for _, transaction := range data.Transactions {
		contractInfo[transaction.ContractName] = common.HexToAddress(transaction.ContractAddress)
	}

	defer jsonFile.Close()
	return contractInfo, nil
}

func setupGanacheServer(ctx context.Context, t *testing.T) error {
	ganachePath, err := exec.Command("which", "ganache").Output()
	if err != nil {
		log.Info().Msg("can not get ganache path")
		return err
	}
	args := []string{"--miner.blockTime=5", "--chain.time=2021-12-08T20:55:40", "--logging.verbose", "--wallet.mnemonic=brownie"}
	procAttr := new(os.ProcAttr)
	procAttr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}
	ganachePathAsString := strings.TrimRight(string(ganachePath), "\n")
	proc, err := os.StartProcess(ganachePathAsString, args, procAttr)
	if err != nil {
		log.Info().Msg("can not start ganache process")
		return err
	}
	t.Cleanup(func() {
		log.Info().Msg("kill ganache")
		if proc != nil {
			err := proc.Kill()
			if err != nil {
				log.Fatal().Err(err).Msg("can not kill ganache")
			}
		}
	})
	for {
		client, _ := ethclient.Dial("http://localhost:8545")
		_, err := client.ChainID(ctx)
		if err != nil {
			log.Info().Msg("can not dial")
			continue
		} else {
			break
		}
	}
	return nil
}

func setupProcessor(ctx context.Context, t *testing.T) (rpc.Processor, sqlmock.Sqlmock) {

	contractInfo, err := GetContractData()
	if err != nil {
		log.Fatal().Err(err).Msg("can not get contract info")
	}
	privKey, err := crypto.HexToECDSA(DeployerKey)
	if err != nil {
		log.Fatal().Err(err).Msg("can not create ecdsa privkey")
	}
	address := common.HexToAddress(DeployerAddress)
	client, err := ethclient.Dial(BackendURL)
	if err != nil {
		log.Fatal().Err(err).Msg("can not connect to rpcUrl")
	}

	keyperSetManagerContract, err := shopContractBindings.NewKeyperSetManager(contractInfo["KeyperSetManager"], client)
	if err != nil {
		log.Fatal().Err(err).Msg("can not get KeyperSetManager")
	}

	broadcastContract, err := shopContractBindings.NewKeyBroadcastContract(contractInfo["KeyBroadcastContract"], client)
	if err != nil {
		log.Fatal().Err(err).Msg("can not get KeyBrodcastContract")
	}

	chainId, err := client.ChainID(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("can not get chain id")
	}

	b, _ := TestEonKey.GobEncode()
	txOps, err := bind.NewKeyedTransactorWithChainID(privKey, chainId)
	if err != nil {
		log.Fatal().Err(err).Msg("can not get txOps")
	}
	tx, err := broadcastContract.BroadcastEonKey(txOps, uint64(0), b)
	if err != nil {
		log.Fatal().Err(err).Msg("can not set eon key")
	}
	_, err = bind.WaitMined(ctx, client, tx)
	if err != nil {
		log.Fatal().Err(err).Msg("can not mine broadcasteonkey")
	}

	sequencerContract, err := sequencerBindings.NewSequencer(contractInfo["Sequencer"], client)
	if err != nil {
		log.Fatal().Err(err).Msg("can not get Sequencer")
	}

	mock, db := NewPostgresTestDB(t)

	processor := rpc.Processor{
		URL:                      ":8546",
		RPCUrl:                   BackendURL,
		SigningKey:               privKey,
		SigningAddress:           &address,
		KeyperSetChangeLookAhead: 0,
		Client:                   client,
		KeyBroadcastContract:     broadcastContract,
		SequencerContract:        sequencerContract,
		KeyperSetManagerContract: keyperSetManagerContract,
		Db:                       db,
	}
	return processor, mock
}

func CaptureOutput(f func() error) (error, string) {
	var buf bytes.Buffer
	oldLogger := utils.Logger
	newLogger := utils.Logger.Output(&buf)
	defer func() {
		utils.Logger = oldLogger
	}()
	utils.Logger = newLogger
	err := f()
	return err, buf.String()
}

func SetupServer(ctx context.Context, t *testing.T) error {
	err := setupGanacheServer(ctx, t)
	if err != nil {
		log.Info().Msg("ganache server didnt run")
		return err
	}
	cmd := exec.Command("make", "deploy")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "ETHERSCAN_API_KEY=''")
	wd, err := os.Getwd()
	if err != nil {
		log.Info().Msg("can not get wd")
		return err
	}
	cmd.Dir = wd + "/../"
	err = cmd.Run()
	if err != nil {
		log.Info().Msg("make deploy failed")
		return err
	}

	processor, _ := setupProcessor(ctx, t)
	backendUrl := &url.URL{}
	err = backendUrl.UnmarshalText([]byte(processor.RPCUrl))
	if err != nil {
		log.Info().Msg("can not unmarshal rpcurl")
		return err
	}
	config := rpc.Config{
		BackendURL:        backendUrl,
		HTTPListenAddress: processor.URL,
	}
	db, _ := db.InitialMigration("")
	service := server.NewRPCService(processor, config, db)
	go func() {
		err := medleyService.Run(ctx, service)
		if err != nil {
			log.Info().Err(err).Msg("server can not run")
		}
	}()
	return nil
}

// NewTestDB creates a new in-memory SQLite database instance for testing purposes.
func NewPostgresTestDB(t *testing.T) (sqlmock.Sqlmock, *db.PostgresDb) {

	var (
		mockDb *sql.DB
		mock   sqlmock.Sqlmock
		err    error
	)

	// Create a new database connection
	mockDb, mock, err = sqlmock.New()
	require.NoError(t, err)

	gormConfig := &gorm.Config{Logger: gorm_logger.Default.LogMode(gorm_logger.Silent)}

	dialector := postgres.New(postgres.Config{
		Conn:       mockDb,
		DriverName: "postgres",
	})
	testDb, err := gorm.Open(dialector, gormConfig)
	require.NoError(t, err)

	inclusionCh := make(chan db.TransactionDetails, 10)
	addTxCh := make(chan db.TransactionDetails, 10)

	return mock, &db.PostgresDb{DB: testDb, InclusionCh: inclusionCh, AddTxCh: addTxCh}
}
