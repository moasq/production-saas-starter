package domain

import "errors"

var (
	ErrInvalidPrompt    = errors.New("prompt cannot be empty")
	ErrProviderNotFound = errors.New("LLM provider not found")
	ErrAPIError         = errors.New("LLM API error")
	ErrTimeout          = errors.New("LLM request timeout")
)