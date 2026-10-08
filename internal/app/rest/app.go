package restapp

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tpu-orch/avia/internal/delivery/rest"
	"github.com/tpu-orch/avia/internal/infra/observability"
)

type App struct {
	server   *http.Server
	listener net.Listener
	engine   *gin.Engine
	logger   observability.Logger
}

func New(address string, handler *rest.Handler, logger observability.Logger) (*App, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("listen HTTP on %s: %w", address, err)
	}

	engine := gin.Default()
	handler.RegisterRoutes(engine)

	server := &http.Server{
		Handler: engine,
	}

	return &App{
		server:   server,
		listener: listener,
		engine:   engine,
		logger:   logger,
	}, nil
}

func (a *App) Run() error {
	a.logger.Info("starting HTTP server", "address", a.listener.Addr().String())
	return a.server.Serve(a.listener)
}

func (a *App) Address() string {
	return a.listener.Addr().String()
}

func (a *App) Stop(ctx context.Context) error {
	a.logger.Info("stopping HTTP server")
	return a.server.Shutdown(ctx)
}
