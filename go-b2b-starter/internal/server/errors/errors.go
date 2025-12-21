package errors

type APIError struct {
	Code    int
	Message string
}

func NewAPIError(code int, message string) *APIError {
	return &APIError{Code: code, Message: message}
}
