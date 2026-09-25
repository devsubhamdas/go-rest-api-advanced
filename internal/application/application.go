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

	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/config"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/logger"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/middleware"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/middleware/header"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/storage"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/user"
	"gorm.io/gorm"
)

type Application struct {
	cfg             *config.Config
	server          *http.Server
	db              *gorm.DB
	Logger          *slog.Logger
	stopRateLimiter context.CancelFunc
}

func New(cfg *config.Config) (*Application, error) {
	db, err := storage.NewPostgres(cfg)
	if err != nil {
		return nil, err
	}

	// app logger setup
	logger := logger.NewJSONLogger(cfg)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		// rID := header.GetRequestIDFromContext(r.Context())
		// fmt.Printf("X-Request-ID: %s\n", rID)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	// User module
	userRepo := user.NewRepository(db)
	userSvc := user.NewService(userRepo)
	userHandler := user.NewHandler(userSvc, logger)
	user.RegisterRoutes(mux, userHandler)

	// cors config
	corsCfg := middleware.DefaultCORSConfig(cfg.AllowedOrigins...)

	// rate limiter setup with canel/stop
	ctx, cancel := context.WithCancel(context.Background())
	rateLimiter := middleware.NewRateLimiter(ctx, 5, 10)

	// Middleware setup - order matters: outer most runs first on the way in,
	// last on the way out
	handler := http.Handler(mux)

	chain := []func(http.Handler) http.Handler{
		middleware.SecurityHeaders,
		rateLimiter.Limit,
		middleware.CORS(corsCfg),
		header.SetRequestID,
		middleware.Logging,
		middleware.Recover,
	}

	for _, mw := range chain {
		handler = mw(handler)
	}

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Application{
		cfg:             cfg,
		server:          server,
		db:              db,
		Logger:          logger,
		stopRateLimiter: cancel,
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
	// stop rate limiter routinge
	a.stopRateLimiter()

	// close db connection
	if err := storage.CloseConnection(a.db); err != nil {
		a.Logger.Error("application.Close::", slog.String("error", err.Error()))
	}

	slog.Info("application closed!!")
}
