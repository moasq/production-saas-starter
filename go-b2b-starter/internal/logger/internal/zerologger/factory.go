package zerolog

import (
	"github.com/moasq/go-b2b-starter/internal/logger/domain"
)

func NewLogger(opts *domain.Options) domain.Logger {
	return newZerologLogger(opts)
}
