package utils

import (
	"encoding/json"
	"net/http"

	"github.com/dhekaag/golang-microservices/shared/pkg/errors"
)

// SendSuccess sends a success response
func SendSuccess(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	errors.WriteSuccessResponse(w, statusCode, message, data)
}

// SendError sends an error response
func SendError(w http.ResponseWriter, statusCode int, message string) {
	var appErr *errors.AppError

	switch statusCode {
	case http.StatusBadRequest:
		appErr = errors.NewBadRequestError(message, nil)
	case http.StatusUnauthorized:
		appErr = errors.NewUnauthorizedError(message, nil)
	case http.StatusForbidden:
		appErr = errors.NewForbiddenError(message, nil)
	case http.StatusNotFound:
		appErr = errors.NewNotFoundError(message, nil)
	case http.StatusMethodNotAllowed:
		appErr = errors.NewMethodNotAllowedError(message, nil)
	case http.StatusConflict:
		appErr = errors.NewConflictError(message, nil)
	case http.StatusUnprocessableEntity:
		appErr = errors.NewUnprocessableEntityError(message, nil)
	case http.StatusTooManyRequests:
		appErr = errors.NewTooManyRequestsError(message, nil)
	case http.StatusRequestTimeout:
		appErr = errors.NewRequestTimeoutError(message, nil)
	case http.StatusInternalServerError:
		appErr = errors.NewInternalServerError(message, nil)
	case http.StatusNotImplemented:
		appErr = errors.NewNotImplementedError(message, nil)
	case http.StatusBadGateway:
		appErr = errors.NewBadGatewayError(message, nil)
	case http.StatusServiceUnavailable:
		appErr = errors.NewServiceUnavailableError(message, nil)
	case http.StatusGatewayTimeout:
		appErr = errors.NewGatewayTimeoutError(message, nil)
	default:
		appErr = errors.NewInternalServerError(message, nil)
	}

	errors.WriteErrorResponse(w, appErr)
}

// SendPaginated sends a paginated response
func SendPaginated(w http.ResponseWriter, message string, data interface{}, page, limit, total int) {
	errors.WritePaginatedResponse(w, message, data, page, limit, total)
}

// SendValidationError sends validation error response
func SendValidationError(w http.ResponseWriter, validationErrors []errors.ValidationError) {
	errors.WriteValidationErrorResponse(w, validationErrors)
}

// SendJSON sends a generic JSON response
func SendJSON(w http.ResponseWriter, statusCode int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
