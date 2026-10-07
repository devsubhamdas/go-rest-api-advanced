package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/devsubhamdas/go-rest-api-advanced/internal/auth/dto"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/cookie"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/errorsx"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/middleware/header"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/response"
)

type Service interface {
	Login(ctx context.Context, input *dto.LoginUserInput) (*dto.LoginResponseData, error)
	Refresh(ctx context.Context, rt string) (*dto.RefreshResponseData, error)
	Singup(ctx context.Context, input *dto.SignupUserInput) (*dto.SignupResponseData, error)
}

type Handler struct {
	svc       Service
	cookieCfg *cookie.Config
	logger    *slog.Logger
}

func NewHandler(s Service, cfg *cookie.Config, logger *slog.Logger) *Handler {
	return &Handler{
		svc:       s,
		cookieCfg: cfg,
		logger:    logger,
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	rID := header.GetRequestIDFromContext(r.Context())

	var req dto.LoginUserInput

	err := json.NewDecoder(r.Body).Decode(&req)
	defer r.Body.Close()
	if err != nil {
		h.logger.Error(
			"LoginUser::\n",
			slog.String("X-Request-ID", rID),
			slog.String("error", err.Error()),
		)

		if typeErr, ok := errors.AsType[*json.UnmarshalTypeError](err); ok {
			_ = response.WriteErrorFromCodeWithDetails(
				w,
				response.CodeUnprocessableEntity,
				"failed to decode json",
				[]errorsx.FieldErrorDetails{
					{
						Field:   typeErr.Field,
						Message: "invalid type of value",
					},
				},
			)
			return
		}

		_ = response.WriteErrorFromCode(
			w,
			response.CodeBadRequest,
			"failed to decode json payload",
		)
		return
	}

	data, err := h.svc.Login(r.Context(), &req)
	if err != nil {
		h.logger.Error(
			"LoginUser::\n",
			slog.String("X-Request-ID", rID),
			slog.String("error", err.Error()),
		)

		if vErr, ok := errors.AsType[*errorsx.ValidationError](err); ok {
			_ = response.WriteErrorFromCodeWithDetails(
				w,
				vErr.Code,
				vErr.Error(),
				vErr.Details,
			)
			return
		}

		if errors.Is(err, errorsx.ErrInvalidEmail) {
			_ = response.WriteErrorFromCodeWithDetails(
				w,
				response.CodeInvalidCredentials,
				errorsx.ErrInvalidCredentials.Error(),
				errorsx.FieldErrorDetails{
					Field:   "email",
					Message: "worng email",
				},
			)
			return
		}

		if errors.Is(err, errorsx.ErrInvalidPassword) {
			_ = response.WriteErrorFromCodeWithDetails(
				w,
				response.CodeInvalidCredentials,
				errorsx.ErrInvalidCredentials.Error(),
				errorsx.FieldErrorDetails{
					Field:   "password",
					Message: "worng password",
				},
			)
			return
		}

		_ = response.WriteErrorFromCode(
			w,
			response.CodeInternalServerError,
			"something went wrong",
		)
		return
	}

	// set session token cookie
	cookie.SetSessionToken(
		h.cookieCfg,
		w,
		data.RefreshToken,
		data.RefreshTokenExp,
	)

	_ = response.WriteJSONWithData(w, http.StatusOK, "user login successful", data)

}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	rID := header.GetRequestIDFromContext(r.Context())

	rt, err := cookie.ExtractSessionToken(r)
	if err != nil {
		h.logger.Error(
			"RefreshUser::\n",
			slog.String("X-Request-ID", rID),
			slog.String("error", err.Error()),
		)

		if errors.Is(err, http.ErrNoCookie) {
			_ = response.WriteError(
				w,
				http.StatusBadRequest,
				response.CodeBadRequest,
				"session token is missing",
			)
			return
		}
		_ = response.WriteErrorFromCode(
			w,
			response.CodeInternalServerError,
			"something went wrong",
		)
		return
	}

	data, err := h.svc.Refresh(r.Context(), rt)
	if err != nil {

		if errors.Is(err, errorsx.ErrInvalidToken) {
			_ = response.WriteError(
				w,
				http.StatusBadRequest,
				response.CodeBadRequest,
				errorsx.ErrInvalidToken.Error(),
			)
			return
		}

		if errors.Is(err, errorsx.ErrInvalidCredentials) {
			_ = response.WriteError(
				w,
				http.StatusBadRequest,
				response.CodeBadRequest,
				errorsx.ErrInvalidCredentials.Error(),
			)
			return
		}

		_ = response.WriteErrorFromCode(
			w,
			response.CodeInternalServerError,
			"something went wrong",
		)
		return
	}

	// set session token cookie
	cookie.SetSessionToken(
		h.cookieCfg,
		w,
		data.RefreshToken,
		data.RefreshTokenExp,
	)

	_ = response.WriteJSONWithData(w, http.StatusOK, "refresh login successful", data)

}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	rID := header.GetRequestIDFromContext(r.Context())

	var req dto.SignupUserInput

	err := json.NewDecoder(r.Body).Decode(&req)
	defer r.Body.Close()
	if err != nil {
		h.logger.Error(
			"SignupUser::\n",
			slog.String("X-Request-ID", rID),
			slog.String("error", err.Error()),
		)

		if typeErr, ok := errors.AsType[*json.UnmarshalTypeError](err); ok {
			_ = response.WriteErrorFromCodeWithDetails(
				w,
				response.CodeUnprocessableEntity,
				"failed to decode json payload",
				[]errorsx.FieldErrorDetails{
					{
						Field:   typeErr.Field,
						Message: "invalid type of value",
					},
				},
			)
			return
		}

		_ = response.WriteErrorFromCode(
			w,
			response.CodeBadRequest,
			"failed to decode json payload",
		)
		return
	}

	user, err := h.svc.Singup(r.Context(), &req)
	if err != nil {
		h.logger.Error(
			"SignupUser::\n",
			slog.String("X-Request-ID", rID),
			slog.String("error", err.Error()),
		)

		if vErr, ok := errors.AsType[*errorsx.ValidationError](err); ok {
			_ = response.WriteErrorFromCodeWithDetails(
				w,
				vErr.Code,
				vErr.Error(),
				vErr.Details,
			)
			return
		}

		if errors.Is(err, errorsx.ErrEmailAlreadyExists) {
			_ = response.WriteErrorFromCode(
				w,
				response.CodeConflict,
				errorsx.ErrEmailAlreadyExists.Error(),
			)
			return
		}

		if errors.Is(err, errorsx.ErrDuplicateKey) {
			_ = response.WriteErrorFromCode(
				w,
				response.CodeDuplicateKeyError,
				errorsx.ErrDuplicateKey.Error(),
			)
			return
		}

		_ = response.WriteErrorFromCode(
			w,
			response.CodeInternalServerError,
			"something went wrong",
		)
		return
	}

	_ = response.WriteJSONWithData(w, http.StatusCreated, "user singup successful", user)

}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie.ClearSessionToken(h.cookieCfg, w)
	w.WriteHeader(http.StatusNoContent)
}
