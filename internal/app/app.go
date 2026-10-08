package app

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	restapp "github.com/tpu-orch/avia/internal/app/rest"
	"github.com/tpu-orch/avia/internal/delivery/rest"
	"github.com/tpu-orch/avia/internal/domain/services/avia"
	"github.com/tpu-orch/avia/internal/infra/config"
	"github.com/tpu-orch/avia/internal/infra/db"
	"github.com/tpu-orch/avia/internal/infra/observability"
	"github.com/tpu-orch/avia/internal/infra/repos"
)

type App struct {
	httpApp *restapp.App
	dbPool  *pgxpool.Pool
	logger  observability.Logger
}

func New(configPath string) (*App, error) {
	ctx := context.Background()

	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	logger := observability.NewLogger("avia-service")

	// Run migrations
	if err := db.RunMigrations(cfg.Postgres.DSN, "migrations"); err != nil {
		logger.Error("run migrations warning", err)
	}

	dbPool, err := db.NewPostgresPool(ctx, cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("init postgres pool: %w", err)
	}

	ticketRepo := repos.NewPostgresTicketRepo(dbPool)
	infoService := service.NewInfoService(ticketRepo)
	_ = service.NewAviaService(ticketRepo)

	handler := rest.NewHandler(infoService)

	httpApp, err := restapp.New(cfg.HTTP.Address, handler, logger)
	if err != nil {
		dbPool.Close()
		return nil, fmt.Errorf("init http app: %w", err)
	}

	return &App{
		httpApp: httpApp,
		dbPool:  dbPool,
		logger:  logger,
	}, nil
}

func (a *App) Run() error {
	return a.httpApp.Run()
}

func (a *App) Stop(ctx context.Context) {
	if a.httpApp != nil {
		if err := a.httpApp.Stop(ctx); err != nil {
			a.logger.Error("error stopping http app", err)
		}
	}
	if a.dbPool != nil {
		a.dbPool.Close()
	}
}
