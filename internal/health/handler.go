package health

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/middleware/header"
	"gorm.io/gorm"
)

type Handler struct {
	db     *gorm.DB
	logger *slog.Logger
}

func NewHandler(db *gorm.DB, logger *slog.Logger) *Handler {
	return &Handler{
		db:     db,
		logger: logger,
	}
}

func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok"}`))
}

func (h *Handler) Readyz(w http.ResponseWriter, r *http.Request) {
	rID := header.GetRequestIDFromContext(r.Context())

	w.Header().Set("Content-Type", "application/json")

	sqlDB, err := h.db.DB()
	if err != nil {
		h.logger.Error(
			"Readyz::\n",
			slog.String("X-Request-ID", rID),
			slog.String("error", err.Error()),
		)

		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "not_ready",
			"error":  "failed to initilize db",
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		h.logger.Error(
			"Readyz::\n",
			slog.String("X-Request-ID", rID),
			slog.String("error", err.Error()),
		)

		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "not_ready",
			"error":  "failed to connect db",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ready"}`))
}
