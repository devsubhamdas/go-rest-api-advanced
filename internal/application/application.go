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

	"github.com/devsubhamdas/go-rest-api-advanced/internal/auth"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/health"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/config"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/cookie"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/logger"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/middleware"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/middleware/header"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/storage"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/token"
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

	// App logger setup
	logger := logger.NewJSONLogger(cfg)

	// Mux setup
	mux := http.NewServeMux()

	// Health check module
	healthHandler := health.NewHandler(db, logger)
	health.RegisterRoutes(mux, healthHandler)

	// User module
	userRepo := user.NewRepository(db)
	userSvc := user.NewService(userRepo)
	userHandler := user.NewHandler(userSvc, logger)
	user.RegisterRoutes(mux, userHandler)

	// Auth module & Token Manager config
	tm, err := token.NewManager(&token.Config{
		Issuer:        cfg.JWTIssuer,
		AccessSecret:  []byte(cfg.JWTAccessSecret),
		RefreshSecret: []byte(cfg.JWTRefreshSecret),
		AccessTTL:     cfg.JWTAccessTTL,
		RefreshTTL:    cfg.JWTRefreshTTL,
	})
	if err != nil {
		logger.Error("TokenManager::\n", slog.String("error", err.Error()))
	}

	authSvc := auth.NewService(userRepo, tm)
	authHandler := auth.NewHandler(
		authSvc,
		&cookie.Config{
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteDefaultMode,
		},
		logger,
	)
	auth.RegisterRoutes(mux, authHandler)

	// CORS config
	corsCfg := middleware.DefaultCORSConfig(cfg.AllowedOrigins...)

	// Rate Limiter setup with canel/stop
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

	slog.Info("http server shutdown completed!!")
	return nil
}

func (a *Application) Close() {
	// stop rate limiter routine & clean memory
	a.stopRateLimiter()
	slog.Info("rate limiter memory cleaned...")

	// close db connection
	if err := storage.CloseConnection(a.db); err != nil {
		a.Logger.Error("application.Close::", slog.String("error", err.Error()))
	}

	slog.Info("server application closed!!")
}
