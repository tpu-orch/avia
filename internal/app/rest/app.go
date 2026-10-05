package restapp

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tpu-orch/avia/internal/delivery/rest"
)

type App struct {
	server   *http.Server
	listener net.Listener
	engine   *gin.Engine
}

func New(address string, handler *rest.Handler) (*App, error) {
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
	}, nil
}

func (a *App) Run() error {
	log.Printf("starting HTTP server on %s", a.listener.Addr().String())
	return a.server.Serve(a.listener)
}

func (a *App) Address() string {
	return a.listener.Addr().String()
}

func (a *App) Stop(ctx context.Context) error {
	log.Println("stopping HTTP server")
	return a.server.Shutdown(ctx)
}
