package domain

import "errors"

var (
	ErrInvalidInput    = errors.New("invalid OCR input")
	ErrQuotaExceeded   = errors.New("OCR quota exceeded") 
	ErrUnsupportedFile = errors.New("unsupported file type")
	ErrAsyncJobFailed  = errors.New("async OCR job failed")
	ErrJobNotFound     = errors.New("OCR job not found")
	ErrAuthFailed      = errors.New("OCR authentication failed")
	ErrTransientError  = errors.New("OCR transient error")
	ErrNotFound        = errors.New("OCR resource not found")
)