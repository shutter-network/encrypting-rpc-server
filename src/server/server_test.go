package server_test

import (
	"context"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/rs/zerolog/log"
	rpc "github.com/shutter-network/encrypting-rpc-server/rpc"
	server "github.com/shutter-network/encrypting-rpc-server/server"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/url"
	medleyService "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

func setupGanacheServer() *os.Process {
	ctx := context.Background()
	ganachePath, err := exec.Command("which", "ganache").Output()
	if err != nil {
		log.Fatal().Err(err).Msg("can not get ganache path")
	}
	args := []string{"-d", "-b", "3000"}
	procAttr := new(os.ProcAttr)
	procAttr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}
	ganachePathAsString := strings.TrimRight(string(ganachePath), "\n")
	proc, err := os.StartProcess(ganachePathAsString, args, procAttr)
	if err != nil {
		log.Fatal().Err(err).Msg("can not start ganache")
	}
	for {
		client, _ := ethclient.Dial("http://localhost:8545")
		_, err := client.ChainID(ctx)
		if err != nil {
			continue
		} else {
			break
		}
	}
	return proc
}

func setupProcessor() rpc.Processor {
	privKey, err := crypto.HexToECDSA("b0057716d5917badaf911b193b12b910811c1497b5bada8d7711f758981c3773")
	if err != nil {
		log.Fatal().Err(err).Msg("can not create ecdsa privkey")
	}
	address := common.HexToAddress("0x1dF62f291b2E969fB0849d99D9Ce41e2F137006e")
	rpcUrl := "http://localhost:8545"
	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		log.Fatal().Err(err).Msg("can not connect to rpcUrl")
	}
	processor := rpc.Processor{
		URL:                      ":8546",
		RPCUrl:                   rpcUrl,
		SigningKey:               privKey,
		SigningAddress:           &address,
		KeyperSetChangeLookAhead: 50,
		Client:                   client,
		KeyBroadcastContract:     nil,
		SequencerContract:        nil,
		KeyperSetManagerContract: nil,
	}
	return processor
}

func setupServer() *os.Process {

	ctx := context.Background()
	proc := setupGanacheServer()
	processor := setupProcessor()
	backendUrl := &url.URL{}
	err := backendUrl.UnmarshalText([]byte(processor.RPCUrl))
	if err != nil {
		log.Fatal().Err(err).Msg("can not parse rpcUrl")
	}
	config := server.Config{
		BackendURL:        backendUrl,
		HTTPListenAddress: processor.URL,
	}
	service := server.NewRPCService(processor, &config)
	go medleyService.Run(ctx, service)
	return proc
}

func captureLog(f func()) []byte {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	log.Info().Interface("test", r).Interface("test2", w).Msg("TEST")
	os.Stdout = w
	f()
	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = oldStdout
	return out
}

func TestServer(t *testing.T) {
	ctx := context.Background()
	proc := setupServer()
	client, err := ethclient.Dial("http://localhost:8546")
	if err != nil {
		log.Fatal().Err(err).Msg("can not connect to server")
	}
	output := captureLog(func() {
		client.ChainID(ctx)
	})
	log.Info().Str("out", string(output)).Msg("test")
	proc.Kill()

}
