package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// AppError represents application error
type AppError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	StatusCode int                    `json:"-"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Cause      error                  `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// APIResponse represents standard API response format
type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

// Meta represents pagination metadata
type Meta struct {
	Page      int `json:"page"`
	Limit     int `json:"limit"`
	Total     int `json:"total"`
	TotalPage int `json:"total_page"`
}

// ValidationError represents field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// ValidationErrors collection
type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return "validation failed"
	}
	return fmt.Sprintf("validation failed: %s", v[0].Message)
}

// Error codes constants
const (
	// Client errors (4xx)
	CodeBadRequest          = "BAD_REQUEST"
	CodeUnauthorized        = "UNAUTHORIZED"
	CodeForbidden           = "FORBIDDEN"
	CodeNotFound            = "NOT_FOUND"
	CodeMethodNotAllowed    = "METHOD_NOT_ALLOWED"
	CodeConflict            = "CONFLICT"
	CodeUnprocessableEntity = "UNPROCESSABLE_ENTITY"
	CodeTooManyRequests     = "TOO_MANY_REQUESTS"
	CodeRequestTimeout      = "REQUEST_TIMEOUT"

	// Server errors (5xx)
	CodeInternalServer     = "INTERNAL_SERVER_ERROR"
	CodeNotImplemented     = "NOT_IMPLEMENTED"
	CodeBadGateway         = "BAD_GATEWAY"
	CodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	CodeGatewayTimeout     = "GATEWAY_TIMEOUT"

	// Business logic errors
	CodeValidationFailed   = "VALIDATION_FAILED"
	CodeDuplicateEntry     = "DUPLICATE_ENTRY"
	CodeInsufficientFunds  = "INSUFFICIENT_FUNDS"
	CodeExpiredToken       = "EXPIRED_TOKEN"
	CodeInvalidCredentials = "INVALID_CREDENTIALS"

	// Database errors
	CodeDatabaseConnection = "DATABASE_CONNECTION_ERROR"
	CodeDatabaseQuery      = "DATABASE_QUERY_ERROR"
	CodeDatabaseConstraint = "DATABASE_CONSTRAINT_ERROR"

	// External service errors
	CodeExternalService = "EXTERNAL_SERVICE_ERROR"
	CodePaymentFailed   = "PAYMENT_FAILED"
	CodeEmailFailed     = "EMAIL_FAILED"
)

// Status constants
const (
	StatusSuccess = "success"
	StatusError   = "error"
	StatusFail    = "fail"
)

// Constructor functions for common errors

// 4xx Client Errors
func NewBadRequestError(message string, cause error) *AppError {
	return &AppError{
		Code:       CodeBadRequest,
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Cause:      cause,
	}
}

func NewUnauthorizedError(message string, cause error) *AppError {
	return &AppError{
		Code:       CodeUnauthorized,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
		Cause:      cause,
	}
}

func NewForbiddenError(message string, cause error) *AppError {
	return &AppError{
		Code:       CodeForbidden,
		Message:    message,
		StatusCode: http.StatusForbidden,
		Cause:      cause,
	}
}

func NewNotFoundError(message string, cause error) *AppError {
	return &AppError{
		Code:       CodeNotFound,
		Message:    message,
		StatusCode: http.StatusNotFound,
		Cause:      cause,
	}
}

func NewMethodNotAllowedError(message string, cause error) *AppError {
	return &AppError{
		Code:       CodeMethodNotAllowed,
		Message:    message,
		StatusCode: http.StatusMethodNotAllowed,
		Cause:      cause,
	}
}

func NewConflictError(message string, cause error) *AppError {
	return &AppError{
		Code:       CodeConflict,
		Message:    message,
		StatusCode: http.StatusConflict,
		Cause:      cause,
	}
}

func NewUnprocessableEntityError(message string, cause error) *AppError {
	return &AppError{
		Code:       CodeUnprocessableEntity,
		Message:    message,
		StatusCode: http.StatusUnprocessableEntity,
		Cause:      cause,
	}
}

func NewTooManyRequestsError(message string, cause error) *AppError {
	return &AppError{
		Code:       CodeTooManyRequests,
		Message:    message,
		StatusCode: http.StatusTooManyRequests,
		Cause:      cause,
	}
}

func NewRequestTimeoutError(message string, cause error) *AppError {
	return &AppError{
		Code:       CodeRequestTimeout,
		Message:    message,
		StatusCode: http.StatusRequestTimeout,
		Cause:      cause,
	}
}

// 5xx Server Errors
func NewInternalServerError(message string, cause error) *AppError {
	return &AppError{
		Code:       CodeInternalServer,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Cause:      cause,
	}
}

func NewNotImplementedError(message string, cause error) *AppError {
	return &AppError{
		Code:       CodeNotImplemented,
		Message:    message,
		StatusCode: http.StatusNotImplemented,
		Cause:      cause,
	}
}

func NewBadGatewayError(message string, cause error) *AppError {
	return &AppError{
		Code:       CodeBadGateway,
		Message:    message,
		StatusCode: http.StatusBadGateway,
		Cause:      cause,
	}
}

func NewServiceUnavailableError(message string, cause error) *AppError {
	return &AppError{
		Code:       CodeServiceUnavailable,
		Message:    message,
		StatusCode: http.StatusServiceUnavailable,
		Cause:      cause,
	}
}

func NewGatewayTimeoutError(message string, cause error) *AppError {
	return &AppError{
		Code:       CodeGatewayTimeout,
		Message:    message,
		StatusCode: http.StatusGatewayTimeout,
		Cause:      cause,
	}
}

// Business Logic Errors
func NewValidationError(message string, validationErrors ValidationErrors) *AppError {
	data := make(map[string]interface{})
	if len(validationErrors) > 0 {
		data["validation_errors"] = validationErrors
	}

	return &AppError{
		Code:       CodeValidationFailed,
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Data:       data,
	}
}

func NewDuplicateEntryError(message string, field string, value string) *AppError {
	return &AppError{
		Code:       CodeDuplicateEntry,
		Message:    message,
		StatusCode: http.StatusConflict,
		Data: map[string]interface{}{
			"field": field,
			"value": value,
		},
	}
}

func NewInsufficientFundsError(message string, required, available float64) *AppError {
	return &AppError{
		Code:       CodeInsufficientFunds,
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Data: map[string]interface{}{
			"required":  required,
			"available": available,
		},
	}
}

func NewExpiredTokenError(message string, cause error) *AppError {
	return &AppError{
		Code:       CodeExpiredToken,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
		Cause:      cause,
	}
}

func NewInvalidCredentialsError(message string, cause error) *AppError {
	return &AppError{
		Code:       CodeInvalidCredentials,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
		Cause:      cause,
	}
}

// Database Errors
func NewDatabaseConnectionError(message string, cause error) *AppError {
	return &AppError{
		Code:       CodeDatabaseConnection,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Cause:      cause,
	}
}

func NewDatabaseQueryError(message string, cause error) *AppError {
	return &AppError{
		Code:       CodeDatabaseQuery,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Cause:      cause,
	}
}

func NewDatabaseConstraintError(message string, constraint string, cause error) *AppError {
	return &AppError{
		Code:       CodeDatabaseConstraint,
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Data: map[string]interface{}{
			"constraint": constraint,
		},
		Cause: cause,
	}
}

// External Service Errors
func NewExternalServiceError(service, message string, cause error) *AppError {
	return &AppError{
		Code:       CodeExternalService,
		Message:    message,
		StatusCode: http.StatusBadGateway,
		Data: map[string]interface{}{
			"service": service,
		},
		Cause: cause,
	}
}

func NewPaymentFailedError(message string, cause error) *AppError {
	return &AppError{
		Code:       CodePaymentFailed,
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Cause:      cause,
	}
}

func NewEmailFailedError(message string, cause error) *AppError {
	return &AppError{
		Code:       CodeEmailFailed,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Cause:      cause,
	}
}

// Response helper functions
func WriteErrorResponse(w http.ResponseWriter, err *AppError) {
	response := APIResponse{
		Status:  StatusError,
		Message: err.Message,
		Error:   err.Code,
	}

	// Add data if exists
	if err.Data != nil && len(err.Data) > 0 {
		response.Data = err.Data
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.StatusCode)
	json.NewEncoder(w).Encode(response)
}

func WriteValidationErrorResponse(w http.ResponseWriter, validationErrors ValidationErrors) {
	err := NewValidationError("Validation failed", validationErrors)
	WriteErrorResponse(w, err)
}

func WriteSuccessResponse(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	response := APIResponse{
		Status:  StatusSuccess,
		Message: message,
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func WriteSuccessResponseWithMeta(w http.ResponseWriter, statusCode int, message string, data interface{}, meta *Meta) {
	response := APIResponse{
		Status:  StatusSuccess,
		Message: message,
		Data:    data,
		Meta:    meta,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func WritePaginatedResponse(w http.ResponseWriter, message string, data interface{}, page, limit, total int) {
	totalPage := (total + limit - 1) / limit // Calculate total pages

	meta := &Meta{
		Page:      page,
		Limit:     limit,
		Total:     total,
		TotalPage: totalPage,
	}

	WriteSuccessResponseWithMeta(w, http.StatusOK, message, data, meta)
}

// Error checking helpers
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

func GetAppError(err error) (*AppError, bool) {
	appErr, ok := err.(*AppError)
	return appErr, ok
}

func IsClientError(err error) bool {
	if appErr, ok := GetAppError(err); ok {
		return appErr.StatusCode >= 400 && appErr.StatusCode < 500
	}
	return false
}

func IsServerError(err error) bool {
	if appErr, ok := GetAppError(err); ok {
		return appErr.StatusCode >= 500
	}
	return false
}

// Wrap standard errors
func FromError(err error) *AppError {
	if appErr, ok := GetAppError(err); ok {
		return appErr
	}
	return NewInternalServerError("Internal server error", err)
}

func WrapError(err error, code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Cause:      err,
	}
}
