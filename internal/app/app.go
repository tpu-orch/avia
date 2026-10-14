package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	kafkaapp "github.com/tpu-orch/avia/internal/app/kafka"
	restapp "github.com/tpu-orch/avia/internal/app/rest"
	"github.com/tpu-orch/avia/internal/delivery/rest"
	service "github.com/tpu-orch/avia/internal/domain/services/avia"
	"github.com/tpu-orch/avia/internal/infra/config"
	"github.com/tpu-orch/avia/internal/infra/db"
	"github.com/tpu-orch/avia/internal/infra/observability"
	"github.com/tpu-orch/avia/internal/infra/repos"
	"golang.org/x/sync/errgroup"
)

type App struct {
	httpApp  *restapp.App
	kafkaApp *kafkaapp.App
	dbPool   *pgxpool.Pool
	logger   observability.Logger
}

func New(configPath string) (*App, error) {
	ctx := context.Background()

	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	logger := observability.NewLogger("avia-service")

	// Run migrations
	if err = db.RunMigrations(cfg.Postgres.DSN, "migrations"); err != nil {
		logger.Error("run migrations warning", err)
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	dbPool, err := db.NewPostgresPool(ctx, cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("init postgres pool: %w", err)
	}

	ticketRepo := repos.NewPostgresTicketRepo(dbPool)
	ticketQueryService := service.NewTicketQueryService(ticketRepo)
	ticketCommandService := service.NewTicketCommandService(ticketRepo)

	handler := rest.NewHandler(ticketQueryService)

	httpApp, err := restapp.New(cfg.HTTP.Address, handler, logger)
	if err != nil {
		dbPool.Close()
		return nil, fmt.Errorf("init http app: %w", err)
	}

	kafkaApp := kafkaapp.New(cfg.Kafka, ticketCommandService, logger)

	return &App{
		httpApp:  httpApp,
		kafkaApp: kafkaApp,
		dbPool:   dbPool,
		logger:   logger,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		err := a.httpApp.Run()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http app: %w", err)
		}

		if ctx.Err() == nil {
			return errors.New("http server stopped unexpectedly")
		}

		return nil
	})

	g.Go(func() error {
		err := a.kafkaApp.Run(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			return fmt.Errorf("kafka app: %w", err)
		}

		if ctx.Err() == nil {
			return errors.New("kafka app stopped unexpectedly")
		}

		return nil
	})

	g.Go(func() error {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		return a.httpApp.Stop(shutdownCtx)
	})

	err := g.Wait()

	// Здесь обе основные задачи уже завершились.
	if shutdownErr := a.kafkaApp.Shutdown(); shutdownErr != nil {
		err = errors.Join(err, shutdownErr)
	}

	if a.dbPool != nil {
		a.dbPool.Close()
	}

	return err
}

func (a *App) Stop(ctx context.Context) {
	if a.kafkaApp != nil {
		if err := a.kafkaApp.Shutdown(); err != nil {
			a.logger.Error("error stopping kafka app", err)
		}
	}
	if a.httpApp != nil {
		if err := a.httpApp.Stop(ctx); err != nil {
			a.logger.Error("error stopping http app", err)
		}
	}
	if a.dbPool != nil {
		a.dbPool.Close()
	}
}
