package test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
	sequencerBindings "github.com/shutter-network/gnosh-contracts/gnoshcontracts/sequencer"
	shopContractBindings "github.com/shutter-network/shop-contracts/bindings"
	"github.com/shutter-network/encrypting-rpc-server/rpc"
	"github.com/shutter-network/encrypting-rpc-server/server"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/url"
	medleyService "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	medleyKeygen "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testkeygen"
	"github.com/shutter-network/shutter/shlib/shcrypto"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func init() {
	TestKeygen = medleyKeygen.NewTestKeyGenerator(&testing.T{}, 3, 2, true)

	TestEonKey = TestKeygen.EonPublicKey([]byte("test"))
}

var (
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
	defer func() {
		server.Logger = oldLogger
	}()
	server.Logger = newLogger
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

	processor := setupProcessor(ctx)
	backendUrl := &url.URL{}
	err = backendUrl.UnmarshalText([]byte(processor.RPCUrl))
	if err != nil {
		log.Info().Msg("can not unmarshal rpcurl")
		return err
	}
	config := server.Config{
		BackendURL:        backendUrl,
		HTTPListenAddress: processor.URL,
	}
	service := server.NewRPCService(processor, &config)
	go func() {
		err := medleyService.Run(ctx, service)
		if err != nil {
			log.Info().Err(err).Msg("server can not run")
		}
	}()
	return nil
}
