package health

import "net/http"

func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("GET /api/healthz", h.Healthz)
	mux.HandleFunc("GET /api/readyz", h.Readyz)
}
