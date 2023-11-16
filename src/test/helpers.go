package test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/integralist/go-findroot/find"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"gotest.tools/assert"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
	"github.com/shutter-network/encrypting-rpc-server/contracts"
	"github.com/shutter-network/encrypting-rpc-server/rpc"
	"github.com/shutter-network/encrypting-rpc-server/server"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/url"
	medleyService "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	medleyKeygen "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testkeygen"
	"github.com/shutter-network/shutter/shlib/shcrypto"
)

func init() {
	TestKeygen = medleyKeygen.NewTestKeyGenerator(&testing.T{}, 3, 2, true)
	TestEonKey = TestKeygen.EonPublicKey([]byte("test"))
	stat, err := find.Repo()
	if err != nil {
		log.Fatal().Err(err).Msg("can't find repository root path")
	}
	RootDir = stat.Path
}

var (
	RootDir         string
	TestKeygen      *medleyKeygen.TestKeyGenerator
	TestEonKey      *shcrypto.EonPublicKey
	DeployerKey     = "a26ebb1df46424945009db72c7a7ba034027450784b93f34000169b35fd3adaa"
	DeployerAddress = "0xA868bC7c1AF08B8831795FAC946025557369F69C"
	BackendURL      = "http://localhost:8545"
	ServerURL       = "http://localhost:8546"
	TxPrivKey       = "bbfbee4961061d506ffbb11dfea64eba16355cbf1d9c29613126ba7fec0aed5d"
	TxFromAddress   = "0x66aB6D9362d4F35596279692F0251Db635165871"
	TxToAddress     = "0x33A4622B82D4c04a53e170c638B944ce27cffce3"
)

type DeployTxData struct {
	Type       string   `json:"type"`
	From       string   `json:"from"`
	Gas        string   `json:"gas"`
	Value      string   `json:"value"`
	Data       string   `json:"data"`
	Nonce      string   `json:"nonce"`
	AccessList []string `json:"accessList"`
}

type DeployTx struct {
	Hash                string       `json:"hash"`
	TransactionType     string       `json:"transactionType"`
	ContractName        string       `json:"contractName"`
	ContractAddress     string       `json:"contractAddress"`
	Function            string       `json:"function"`
	Arguments           []string     `json:"arguments"`
	Transaction         DeployTxData `json:"transaction"`
	AdditionalContracts []string     `json:"additionalContracts"`
	IsFixedGasLimit     bool         `json:"isFixedGasLimit"`
}

type ReceiptLog struct {
	Address          string   `json:"address"`
	Topics           []string `json:"topics"`
	Data             string   `json:"data"`
	BlockHash        string   `json:"blockHash"`
	BlockNumber      string   `json:"blockNumber"`
	TransactionHash  string   `json:"transactionHash"`
	TransactionIndex string   `json:"transactionIndex"`
	LogIndex         string   `json:"logIndex"`
	Removed          bool     `json:"removed"`
}

type DeployReceipt struct {
	TransactionHash   string       `json:"transactionHash"`
	TransactionIndex  string       `json:"transactionIndex"`
	BlockHash         string       `json:"blockHash"`
	From              string       `json:"from"`
	To                *string      `json:"to"`
	CumulativeGasUsed string       `json:"cumulativeGasUsed"`
	GasUsed           string       `json:"gasUsed"`
	ContractAddress   string       `json:"contractAddress"`
	Logs              []ReceiptLog `json:"logs"`
	Status            string       `json:"status"`
	LogsBloom         string       `json:"logsBloom"`
	Type              string       `json:"type"`
	EffectiveGasPrice string       `json:"effectiveGasPrice"`
}

type DeployData struct {
	Transactions []DeployTx        `json:"transactions"`
	Receipts     []DeployReceipt   `json:"receipts"`
	Libraries    []string          `json:"libraries"`
	Pending      []string          `json:"pending"`
	Returns      map[string]string `json:"returns"`
	Timestamp    int               `json:"timestamp"`
	Chain        int               `json:"chain"`
	Multi        bool              `json:"multi"`
	Commit       string            `json:"commit"`
}

func GetContractData() (map[string]common.Address, error) {
	contractInfo := make(map[string]common.Address)
	deployDataPath := path.Join(RootDir, "gnosh-contracts/broadcast/deploy.s.sol/1337/run-latest.json")
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

func setupGanacheServer(t *testing.T, ctx context.Context) (*os.Process, error) {
	t.Helper()

	ganachePath, err := exec.Command("which", "ganache").Output()
	assert.NilError(t, err)
	args := []string{"ganache", "-b", "5", "-t", "2021-12-08T20:55:40", "-v", "--wallet.mnemonic", "brownie"}
	procAttr := new(os.ProcAttr)
	procAttr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}
	ganachePathAsString := strings.TrimRight(string(ganachePath), "\n")
	if ganachePathAsString == "" {
		return nil, errors.New("can't find ganache path")
	}
	log.Info().Str("ganache-path", ganachePathAsString).Msg("found path")
	proc, err := os.StartProcess(ganachePathAsString, args, procAttr)
	t.Cleanup(
		func() {
			_ = proc.Signal(syscall.SIGINT)
			_ = proc.Signal(syscall.SIGTERM)
			t := time.NewTimer(3 * time.Second)
			wait := make(chan error)
			go func() {
				_, err := proc.Wait()
				wait <- err
			}()
			select {
			case <-wait:
				close(wait)
				return
			case <-t.C:
				_ = proc.Kill()
				return
			}
		},
	)
	assert.NilError(t, err)

	ticker := time.NewTicker(100 * time.Millisecond)
	var client *ethclient.Client
	for {
		select {
		case <-ticker.C:
			if client == nil {
				client, err = ethclient.Dial("http://localhost:8545")
				if err != nil {
					log.Info().Msg("could not dial blockchain")
					continue
				}
			}
			_, err := client.ChainID(ctx)
			if err != nil {
				log.Info().Err(err).Msg("could not get chainid")
				continue
			}
			return proc, nil
		case <-ctx.Done():
			return proc, ctx.Err()
		}
	}
}

func setupProcessor(ctx context.Context) rpc.Processor {
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

	keyperSetManagerContract, err := contracts.NewKeyperSetManagerContract(contractInfo["KeyperSetManager"], client)
	if err != nil {
		log.Fatal().Err(err).Msg("can not get KeyperSetManager")
	}

	broadcastContract, err := contracts.NewKeyBroadcastContract(contractInfo["KeyBroadcastContract"], client)
	if err != nil {
		log.Fatal().Err(err).Msg("can not get KeyBrodcastContract")
	}

	chainId, err := client.ChainID(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("can not get chain id")
	}

	b, err := TestEonKey.GobEncode()
	if err != nil {
		log.Fatal().Err(err).Msg("Test")
	}
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
		log.Fatal().Err(err).Msg("could not mine transactions")
	}

	sequencerContract, err := contracts.NewSequencerContract(contractInfo["Sequencer"], client)
	if err != nil {
		log.Fatal().Err(err).Msg("can not get Sequencer")
	}

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
	}
	return processor
}

func CaptureOutput(f func() error) (error, string) {
	var buf bytes.Buffer
	oldLogger := server.Logger
	newLogger := server.Logger.Output(&buf)
	server.Logger = newLogger
	err := f()
	server.Logger = oldLogger
	return err, buf.String()
}

func SetupServer(t *testing.T, ctx context.Context) (*os.Process, error) {
	t.Helper()

	group, ctx := errgroup.WithContext(ctx)
	proc, err := setupGanacheServer(t, ctx)
	if err != nil {
		err = errors.Wrap(err, "failed to setup ganache blockchain")
		return proc, err
	}

	cmd := exec.CommandContext(ctx, "make", "deploy")
	cmd.Dir = path.Join(RootDir, "src/")
	err = cmd.Run()
	if err != nil {
		err = errors.Wrap(err, "failed to run deployment")
		return proc, err
	}

	processor := setupProcessor(ctx)
	backendUrl := &url.URL{}
	err = backendUrl.UnmarshalText([]byte(processor.RPCUrl))
	assert.NilError(t, err)

	config := server.Config{
		BackendURL:        backendUrl,
		HTTPListenAddress: processor.URL,
	}
	service := server.NewRPCService(processor, &config)
	group.Go(
		func() error {
			return medleyService.Run(ctx, service)
		},
	)
	return proc, nil
}
