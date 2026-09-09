package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Subham-Das-98/go-rest-api-advanced/internal/platform/config"
	"github.com/Subham-Das-98/go-rest-api-advanced/internal/platform/middleware/header"
	"github.com/Subham-Das-98/go-rest-api-advanced/internal/platform/storage"
	"gorm.io/gorm"
)

type Application struct {
	cfg    *config.Config
	server *http.Server
	db     *gorm.DB
	Logger *slog.Logger
}

func New(cfg *config.Config) (*Application, error) {
	db, err := storage.NewPostgres(cfg)
	if err != nil {
		return nil, err
	}

	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	})

	logger := slog.New(logHandler)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		// rID := header.GetRequestIDFromContext(r.Context())
		// fmt.Printf("X-Request-ID: %s\n", rID)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	handler := header.SetRequestID(mux)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Application{
		cfg:    cfg,
		server: server,
		db:     db,
		Logger: logger,
	}, nil
}

func (a *Application) Run() error {
	errCh := make(chan error, 1)
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info(
			"server started listening",
			slog.String("port", a.cfg.Port),
			slog.String("address", a.server.Addr),
		)
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case sig := <-quit:
		slog.Info("shutting down", slog.String("signal", sig.String()))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: \n%w", err)
	}

	slog.Info("server shutdown completed!!")
	return nil
}

func (a *Application) Close() {
	if err := storage.CloseConnection(a.db); err != nil {
		a.Logger.Error("application.Close::", slog.String("error", err.Error()))
	}

	slog.Info("application closed!!")
}
