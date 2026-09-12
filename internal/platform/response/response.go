package response

import (
	"encoding/json"
	"net/http"
)

type Code string

const (
	// 400 - Client/Validation errors
	CodeBadRequest      Code = "bad_request"
	CodeValidationError Code = "validation_error"
	CodeInvalidInput    Code = "invalid_input"
	CodeMissingField    Code = "missing_field"

	// 401 - Authentication
	CodeUnauthorized       Code = "unauthorized"
	CodeInvalidToken       Code = "invalid_token"
	CodeTokenExpired       Code = "token_expired"
	CodeInvalidCredentials Code = "invalid_credentials"

	// 403 - Authorization
	CodeForbidden        Code = "forbidden"
	CodePermissionDenied Code = "permission_denied"

	// 404
	CodeNotFoundError Code = "not_found_error"

	// 408 - Request Timeout
	CodeRequestTimeoutError Code = "request_timeout_error"

	// 409 - Conflict
	CodeDuplicateKeyError Code = "duplicate_key_error"
	CodeConflict          Code = "conflict_error"
	CodeAlreadyExists     Code = "already_exists"

	// 422
	CodeUnprocessableEntity Code = "unprocessable_entity"

	// 429
	CodeTooManyRequests Code = "too_many_requests"

	// 500
	CodeInternalServerError Code = "internal_server_error"
	CodeDatabaseError       Code = "database_error"

	// 502/503/504
	CodeBadGateway         Code = "bad_gateway"
	CodeServiceUnavailable Code = "service_unavailable"
	CodeGatewayTimeout     Code = "gateway_timeout"
)

var errCodeHTTPStatusMap = map[Code]int{
	// 400 - Client/Validation errors
	CodeBadRequest:      http.StatusBadRequest,
	CodeValidationError: http.StatusBadRequest,
	CodeInvalidInput:    http.StatusBadRequest,
	CodeMissingField:    http.StatusBadRequest,

	// 401 - Authentication
	CodeUnauthorized:       http.StatusUnauthorized,
	CodeInvalidToken:       http.StatusUnauthorized,
	CodeTokenExpired:       http.StatusUnauthorized,
	CodeInvalidCredentials: http.StatusUnauthorized,

	// 403 - Authorization
	CodeForbidden:        http.StatusForbidden,
	CodePermissionDenied: http.StatusForbidden,

	// 404
	CodeNotFoundError: http.StatusNotFound,

	// 408 - Request Timeout
	CodeRequestTimeoutError: http.StatusRequestTimeout,

	// 409 - Conflict
	CodeDuplicateKeyError: http.StatusConflict,
	CodeConflict:          http.StatusConflict,
	CodeAlreadyExists:     http.StatusConflict,

	// 422
	CodeUnprocessableEntity: http.StatusUnprocessableEntity,

	// 429
	CodeTooManyRequests: http.StatusTooManyRequests,

	// 500
	CodeInternalServerError: http.StatusInternalServerError,
	CodeDatabaseError:       http.StatusInternalServerError,

	// 502/503/504
	CodeBadGateway:         http.StatusBadGateway,
	CodeServiceUnavailable: http.StatusServiceUnavailable,
	CodeGatewayTimeout:     http.StatusGatewayTimeout,
}

type JSONResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type ErrorBody struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type ErrorResponse struct {
	Success bool      `json:"success"`
	Error   ErrorBody `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, message string, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(JSONResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// WriteError writes an error response with an explicit HTTP status and code.
func WriteError(w http.ResponseWriter, status int, code Code, message string, details any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(ErrorResponse{
		Success: false,
		Error: ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// WriteErrorFromCode writes an error response, deriving the HTTP status from code's default mapping.
func WriteErrorFromCode(w http.ResponseWriter, code Code, message string, details any) error {
	status, ok := errCodeHTTPStatusMap[code]
	if !ok {
		status = http.StatusInternalServerError
	}

	return WriteError(w, status, code, message, details)
}
