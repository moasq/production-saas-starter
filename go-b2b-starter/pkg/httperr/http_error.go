package httperr

type HTTPError struct {
	StatusCode int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func (e HTTPError) Error() string {
	return e.Message
}

func NewHTTPError(statusCode int, code, message string) HTTPError {
	return HTTPError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
	}
}
