package bootstrap

import (
	"go.uber.org/dig"

	"github.com/moasq/go-b2b-starter/internal/api"
	"github.com/moasq/go-b2b-starter/internal/modules/auth"
	authCmd "github.com/moasq/go-b2b-starter/internal/modules/auth/cmd"
	cognitive "github.com/moasq/go-b2b-starter/internal/modules/cognitive/cmd"
	db "github.com/moasq/go-b2b-starter/internal/db/cmd"
	docs "github.com/moasq/go-b2b-starter/internal/docs/cmd"
	documents "github.com/moasq/go-b2b-starter/internal/modules/documents/cmd"
	eventbus "github.com/moasq/go-b2b-starter/internal/platform/eventbus/cmd"
	files "github.com/moasq/go-b2b-starter/internal/modules/files/cmd"
	llm "github.com/moasq/go-b2b-starter/internal/platform/llm/cmd"
	logger "github.com/moasq/go-b2b-starter/internal/platform/logger/cmd"
	ocr "github.com/moasq/go-b2b-starter/internal/platform/ocr/cmd"
	redisCmd "github.com/moasq/go-b2b-starter/internal/platform/redis/cmd"
	server "github.com/moasq/go-b2b-starter/internal/platform/server/cmd"
	users "github.com/moasq/go-b2b-starter/internal/modules/users/cmd"
)

func InitMods(container *dig.Container) {

	// Platform services
	server.Init(container)
	logger.Init(container)
	db.Init(container)
	files.Init(container)
	if err := eventbus.Init(container); err != nil {
		panic(err)
	}
	if err := llm.Init(container); err != nil {
		panic(err)
	}

	// Redis for caching
	if err := redisCmd.Init(container); err != nil {
		panic(err)
	}

	// Users module (provides UserRepository and RefreshTokenRepository)
	if err := users.Init(container); err != nil {
		panic(err)
	}

	// Auth module with self-hosted JWT authentication
	if err := authCmd.Init(container); err != nil {
		panic(err)
	}

	// Initialize auth middleware
	if err := authCmd.InitMiddleware(container); err != nil {
		panic(err)
	}

	// Register auth middleware as named middlewares for use in routes
	if err := auth.RegisterNamedMiddlewares(container); err != nil {
		panic(err)
	}

	// docs
	docs.Init(container)

	// NOTE: Billing and Paywall modules are B2B-specific and depend on organizations.
	// For B2C, implement user-based billing if needed.

	// OCR service (Mistral API for document text extraction)
	// Must be initialized before documents module (documents depends on OCR)
	if err := ocr.Init(container); err != nil {
		panic(err)
	}

	// Documents module (PDF upload and text extraction)
	if err := documents.Init(container); err != nil {
		panic(err)
	}

	// Cognitive module (AI/RAG with embeddings and vector search)
	// Note: This also wires the event listener for DocumentUploaded events
	if err := cognitive.Init(container); err != nil {
		panic(err)
	}

	// api
	api.Init(container)
}
