package user

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/errorsx"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/middleware/header"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/response"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/user/dto"
)

type Handler struct {
	svc    *Service
	logger *slog.Logger
}

func NewHandler(s *Service, logger *slog.Logger) *Handler {
	return &Handler{
		svc:    s,
		logger: logger,
	}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	rID := header.GetRequestIDFromContext(r.Context())

	var req dto.CreateUserInput

	defer r.Body.Close()

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logger.Error(
			"CreateUser::\n",
			slog.String("X-Request-ID", rID),
			slog.String("error", err.Error()),
		)

		if typeErr, ok := errors.AsType[*json.UnmarshalTypeError](err); ok {
			_ = response.WriteErrorFromCodeWithDetails(
				w,
				response.CodeUnprocessableEntity,
				"failed to decode json payload",
				[]errorsx.ValidationErrorDetails{
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
			response.CodeUnprocessableEntity,
			"failed to decode json payload",
		)
		return
	}

	user, err := h.svc.Create(r.Context(), &req)

	if err != nil {
		h.logger.Error(
			"CreateUser::\n",
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

	_ = response.WriteJSONWithData(w, http.StatusCreated, "user created successfully", user)

}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	rID := header.GetRequestIDFromContext(r.Context())

	id := r.PathValue("id")
	if id == "" {
		_ = response.WriteErrorFromCode(w, response.CodeBadRequest, "id path value is missing")
		return
	}

	var req dto.UpdateUserInput

	defer r.Body.Close()

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logger.Error(
			"UpdateUser::\n",
			slog.String("X-Request-ID", rID),
			slog.String("error", err.Error()),
		)

		if typeErr, ok := errors.AsType[*json.UnmarshalTypeError](err); ok {
			_ = response.WriteErrorFromCodeWithDetails(
				w,
				response.CodeUnprocessableEntity,
				"failed to decode json",
				[]errorsx.ValidationErrorDetails{
					{
						Field:   typeErr.Field,
						Message: "invalid type of value",
					},
				},
			)
			return
		}

		if errors.Is(err, errorsx.ErrNotFound) {
			_ = response.WriteErrorFromCode(
				w,
				response.CodeNotFoundError,
				errorsx.ErrNotFound.Error(),
			)
		}

		_ = response.WriteErrorFromCode(
			w,
			response.CodeUnprocessableEntity,
			"failed to decode json payload",
		)
		return
	}

	user, err := h.svc.Update(r.Context(), id, &req)

	if err != nil {
		h.logger.Error(
			"UpdateUser::\n",
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

		if errors.Is(err, errorsx.ErrInvalidID) {
			_ = response.WriteErrorFromCode(
				w,
				response.CodeBadRequest,
				errorsx.ErrInvalidID.Error(),
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

	_ = response.WriteJSONWithData(w, http.StatusOK, "user updated successfully", user)

}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	rID := header.GetRequestIDFromContext(r.Context())

	id := r.PathValue("id")
	if id == "" {
		_ = response.WriteErrorFromCode(w, response.CodeBadRequest, "id path value is missing")
		return
	}

	err := h.svc.Delete(r.Context(), id)
	if err != nil {
		h.logger.Error(
			"DeleteUser::\n",
			slog.String("X-Request-ID", rID),
			slog.String("error", err.Error()),
		)

		if errors.Is(err, errorsx.ErrInvalidID) {
			_ = response.WriteErrorFromCode(
				w,
				response.CodeBadRequest,
				errorsx.ErrInvalidID.Error(),
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

	_ = response.WriteJSON(w, http.StatusOK, "user deleted successfully")
}

func (h *Handler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	rID := header.GetRequestIDFromContext(r.Context())

	id := r.PathValue("id")
	if id == "" {
		_ = response.WriteErrorFromCode(w, response.CodeBadRequest, "id path value is missing")
		return
	}

	user, err := h.svc.GetByID(r.Context(), id)

	if err != nil {
		h.logger.Error(
			"GetUserByID::\n",
			slog.String("X-Request-ID", rID),
			slog.String("error", err.Error()),
		)

		if errors.Is(err, errorsx.ErrInvalidID) {
			_ = response.WriteErrorFromCode(
				w,
				response.CodeBadRequest,
				errorsx.ErrInvalidID.Error(),
			)
			return
		}

		if errors.Is(err, errorsx.ErrNotFound) {
			_ = response.WriteErrorFromCode(
				w,
				response.CodeNotFoundError,
				errorsx.ErrNotFound.Error(),
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

	_ = response.WriteJSONWithData(w, http.StatusOK, "", user)
}

func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	rID := header.GetRequestIDFromContext(r.Context())

	users, err := h.svc.GetMany(r.Context())

	if err != nil {
		h.logger.Error(
			"GetUsers::\n",
			slog.String("X-Request-ID", rID),
			slog.String("error", err.Error()),
		)

		_ = response.WriteErrorFromCode(
			w,
			response.CodeInternalServerError,
			"something went wrong",
		)
		return
	}

	_ = response.WriteJSONWithData(w, http.StatusOK, "", users)
}
