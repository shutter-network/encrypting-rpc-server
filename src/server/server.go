package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/shutter-network/encrypting-rpc-server/db"
	"github.com/shutter-network/encrypting-rpc-server/metrics"
	"github.com/shutter-network/encrypting-rpc-server/utils"

	ethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pkg/errors"

	"github.com/shutter-network/encrypting-rpc-server/rpc"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
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
		utils.Logger.Info().Str("method", rpcreq.Method).Msg("dispatching to processor")
	} else {
		utils.Logger.Info().Str("method", rpcreq.Method).Msg("dispatching to backend")
	}

	// make the body available again before letting reverse proxy handle the rest
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	startTime := time.Now()

	selectedHandler.ServeHTTP(w, r)

	if selectedHandler == p.backend {
		metrics.UpstreamRequestDuration.WithLabelValues(rpcreq.Method).Observe(time.Since(startTime).Seconds())
	}
}

type server struct {
	processor        rpc.Processor
	config           rpc.Config
	postgresDatabase *db.PostgresDb
}

func NewRPCService(processor rpc.Processor, config rpc.Config, pgDb *db.PostgresDb) medleyService.Service {
	return &server{
		processor:        processor,
		config:           config,
		postgresDatabase: pgDb,
	}
}

func (srv *server) rpcHandler(ctx context.Context) (http.Handler, *[]rpc.RPCService, error) {
	rpcServices := []rpc.RPCService{
		&rpc.EthService{},
	}

	rpcServer := ethrpc.NewServer()
	for _, service := range rpcServices {
		service.Init(srv.processor, srv.config)
		err := rpcServer.RegisterName(service.Name(), service)
		if err != nil {
			return nil, nil, errors.Wrap(err, "error while trying to register RPCService")
		}
	}

	p := &JSONRPCProxy{
		backend:   NewReverseProxy(srv.config.BackendURL),
		processor: rpcServer,
	}
	return p, &rpcServices, nil
}

func (srv *server) setupRouter(ctx context.Context) (*chi.Mux, *[]rpc.RPCService, error) {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	handler, services, err := srv.rpcHandler(ctx)
	if err != nil {
		return nil, nil, err
	}
	router.Mount("/", handler)
	return router, services, nil
}

func (srv *server) Start(ctx context.Context, runner medleyService.Runner) error {
	handler, services, err := srv.setupRouter(ctx)

	if err != nil {
		return err
	}
	httpServer := &http.Server{
		Addr:              srv.config.HTTPListenAddress,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	runner.Go(func() error {
		srv.postgresDatabase.Start(ctx)
		return nil
	})

	for _, service := range *services {
		if err := runner.StartService(service); err != nil {
			return err
		}
	}

	if srv.processor.MetricsConfig.Enabled {
		if err := runner.StartService(srv.processor.MetricsServer); err != nil {
			return err
		}
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
