package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"time"

	ethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pkg/errors"

	"github.com/shutter-network/encrypting-rpc-server/rpc"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/url"
	medleyService "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

type JSONRPCProxy struct {
	backend   http.Handler
	processor http.Handler
}

func (p *JSONRPCProxy) SelectHandler(method string) http.Handler {
	// route the eth_namespace to the l2-backend
	switch method {
	case "eth_sendTransaction":
		return p.processor
	case "eth_sendRawTransaction":
		return p.processor
	default:
		return p.backend
	}
}

func (p *JSONRPCProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	rpcreq := medley.RPCRequest{}
	err = json.Unmarshal(body, &rpcreq)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	selectedHandler := p.SelectHandler(rpcreq.Method)
	if selectedHandler == p.processor {
		Logger.Info().Str("method", rpcreq.Method).Msg("dispatching to processor")
	} else {
		Logger.Info().Str("method", rpcreq.Method).Msg("dispatching to backend")
	}

	// make the body available again before letting reverse proxy handle the rest
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	selectedHandler.ServeHTTP(w, r)
}

type Config struct {
	BackendURL        *url.URL
	HTTPListenAddress string
}

type server struct {
	processor rpc.Processor
	config    *Config
}

func NewRPCService(processor rpc.Processor, config *Config) medleyService.Service {
	return &server{
		processor: processor,
		config:    config,
	}
}

func (srv *server) rpcHandler() (http.Handler, error) {
	rpcServices := []rpc.RPCService{
		&rpc.EthService{},
	}

	rpcServer := ethrpc.NewServer()
	for _, service := range rpcServices {
		service.InjectProcessor(srv.processor)
		err := rpcServer.RegisterName(service.Name(), service)
		if err != nil {
			return nil, errors.Wrap(err, "error while trying to register RPCService")
		}
	}

	p := &JSONRPCProxy{
		backend:   httputil.NewSingleHostReverseProxy(srv.config.BackendURL.URL),
		processor: rpcServer,
	}
	// handler := injectHTTPLogger(p)
	return p, nil
}

func (srv *server) setupRouter() (*chi.Mux, error) {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	handler, err := srv.rpcHandler()
	if err != nil {
		return nil, err
	}
	router.Mount("/", handler)
	return router, nil
}

func (srv *server) Start(ctx context.Context, runner medleyService.Runner) error {
	handler, err := srv.setupRouter()
	if err != nil {
		return err
	}
	httpServer := &http.Server{
		Addr:              srv.config.HTTPListenAddress,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	runner.Go(httpServer.ListenAndServe)
	runner.Go(func() error {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	})
	return nil
}
