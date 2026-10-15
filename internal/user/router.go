package user

import "net/http"

func RegisterRoutes(mux *http.ServeMux, h *Handler, requireAuth func(http.Handler) http.Handler) {
	mux.HandleFunc("POST /api/users", h.CreateUser)
	mux.HandleFunc("PATCH /api/users/{id}", h.UpdateUser)
	mux.HandleFunc("DELETE /api/users/{id}", h.DeleteUser)
	mux.HandleFunc("GET /api/users/{id}", h.GetUserByID)
	mux.HandleFunc("GET /api/users", h.GetUsers)
	mux.Handle("GET /api/users/profile", requireAuth(http.HandlerFunc(h.GetProfile)))
}
