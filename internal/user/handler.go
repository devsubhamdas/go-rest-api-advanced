package user

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/errorsx"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/response"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/user/dto"
)

type Handler struct {
	svc *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{
		svc: s,
	}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateUserInput

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		var typeErr *json.UnmarshalTypeError
		if errors.Is(err, typeErr) && errors.As(err, &typeErr) {
			_ = response.WriteErrorFromCodeWithDetails(
				w,
				response.CodeUnprocessableEntity,
				"failed to decode json payload",
				errorsx.ValidationErrorDetails{
					Field:   typeErr.Field,
					Message: "invalid type of value",
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
		var vErr *errorsx.ValidationError
		if errors.Is(err, vErr) && errors.As(err, &vErr) {
			_ = response.WriteErrorFromCodeWithDetails(
				w,
				response.CodeUnprocessableEntity,
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
	id := r.PathValue("id")
	if id == "" {
		_ = response.WriteErrorFromCode(w, response.CodeBadRequest, "id path value is missing")
	}

	var req dto.UpdateUserInput

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		var typeErr *json.UnmarshalTypeError
		if errors.Is(err, typeErr) && errors.As(err, &typeErr) {
			response.WriteErrorFromCodeWithDetails(
				w,
				response.CodeUnprocessableEntity,
				"failed to decode json",
				errorsx.ValidationErrorDetails{
					Field:   typeErr.Field,
					Message: "invalid type of value",
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

	user, err := h.svc.Update(r.Context(), id, &req)

	if err != nil {
		var vErr *errorsx.ValidationError
		if errors.Is(err, vErr) && errors.As(err, &vErr) {
			_ = response.WriteErrorFromCodeWithDetails(
				w,
				response.CodeUnprocessableEntity,
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
	id := r.PathValue("id")
	if id == "" {
		_ = response.WriteErrorFromCode(w, response.CodeBadRequest, "id path value is missing")
	}

	err := h.svc.Delete(r.Context(), id)
	if err != nil {
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
	id := r.PathValue("id")
	if id == "" {
		_ = response.WriteErrorFromCode(w, response.CodeBadRequest, "id path value is missing")
	}

	user, err := h.svc.GetByID(r.Context(), id)

	if err != nil {
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

	_ = response.WriteJSONWithData(w, http.StatusOK, "", user)
}

func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.svc.GetMany(r.Context())

	if err != nil {
		_ = response.WriteErrorFromCode(
			w,
			response.CodeInternalServerError,
			"something went wrong",
		)
		return
	}

	_ = response.WriteJSONWithData(w, http.StatusOK, "", users)
}
