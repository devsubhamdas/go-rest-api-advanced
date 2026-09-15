package response

import "net/http"

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

var StatusByCode = map[Code]int{
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
