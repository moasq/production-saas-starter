package bootstrap

import (
	"context"

	"go.uber.org/dig"

	"github.com/moasq/go-b2b-starter/internal/api"
	"github.com/moasq/go-b2b-starter/internal/modules/auth"
	authCmd "github.com/moasq/go-b2b-starter/internal/modules/auth/cmd"
	billing "github.com/moasq/go-b2b-starter/internal/modules/billing/cmd"
	cognitive "github.com/moasq/go-b2b-starter/internal/modules/cognitive/cmd"
	db "github.com/moasq/go-b2b-starter/internal/db/cmd"
	docs "github.com/moasq/go-b2b-starter/internal/docs/cmd"
	documents "github.com/moasq/go-b2b-starter/internal/modules/documents/cmd"
	eventbus "github.com/moasq/go-b2b-starter/internal/platform/eventbus/cmd"
	files "github.com/moasq/go-b2b-starter/internal/modules/files/cmd"
	llm "github.com/moasq/go-b2b-starter/internal/platform/llm/cmd"
	logger "github.com/moasq/go-b2b-starter/internal/platform/logger/cmd"
	ocr "github.com/moasq/go-b2b-starter/internal/platform/ocr/cmd"
	orgDomain "github.com/moasq/go-b2b-starter/internal/modules/organizations/domain"
	organizations "github.com/moasq/go-b2b-starter/internal/modules/organizations/cmd"
	paywall "github.com/moasq/go-b2b-starter/internal/modules/paywall/cmd"
	polar "github.com/moasq/go-b2b-starter/internal/platform/polar/cmd"
	redisCmd "github.com/moasq/go-b2b-starter/internal/platform/redis/cmd"
	server "github.com/moasq/go-b2b-starter/internal/platform/server/cmd"
	stytchCmd "github.com/moasq/go-b2b-starter/internal/platform/stytch/cmd"
)

// orgLookupAdapter adapts orgDomain.OrganizationRepository to auth.OrganizationLookup
type orgLookupAdapter struct {
	repo orgDomain.OrganizationRepository
}

func (a *orgLookupAdapter) GetByStytchID(ctx context.Context, stytchOrgID string) (auth.OrganizationEntity, error) {
	return a.repo.GetByStytchID(ctx, stytchOrgID)
}

// accLookupAdapter adapts orgDomain.AccountRepository to auth.AccountLookup
type accLookupAdapter struct {
	repo orgDomain.AccountRepository
}

func (a *accLookupAdapter) GetByEmail(ctx context.Context, orgID int32, email string) (auth.AccountEntity, error) {
	return a.repo.GetByEmail(ctx, orgID, email)
}

func InitMods(container *dig.Container) {

	// pkg
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

	// Polar package must be initialized before payment module (payment depends on Polar client)
	if err := polar.Init(container); err != nil {
		panic(err)
	}

	// Redis must be initialized before auth (Stytch repositories rely on Redis-backed clients upstream)
	if err := redisCmd.Init(container); err != nil {
		panic(err)
	}

	// Stytch client package must be initialized before app/auth (for organization/member management)
	// This provides: stytch.Config, stytch.Client, stytch.RBACPolicyService
	if err := stytchCmd.ProvideStytchDependencies(container); err != nil {
		panic(err)
	}

	// Auth package (pkg/auth) must be initialized before app/auth
	// This provides: auth.AuthProvider (authentication/authorization)
	if err := authCmd.Init(container); err != nil {
		panic(err)
	}

	// docs
	docs.Init(container)

	// app
	if err := organizations.Init(container); err != nil {
		panic(err)
	}

	// Register auth resolvers (bridges organizations domain to auth package)
	if err := auth.ProvideResolvers(container,
		func(repo orgDomain.OrganizationRepository) auth.OrganizationResolver {
			return auth.NewOrganizationResolver(&orgLookupAdapter{repo: repo})
		},
		func(repo orgDomain.AccountRepository) auth.AccountResolver {
			return auth.NewAccountResolver(&accLookupAdapter{repo: repo})
		},
	); err != nil {
		panic(err)
	}

	// Initialize auth middleware (requires resolvers to be registered)
	if err := authCmd.InitMiddleware(container); err != nil {
		panic(err)
	}

	// Register auth middleware as named middlewares for use in routes
	if err := auth.RegisterNamedMiddlewares(container); err != nil {
		panic(err)
	}

	// Billing module (subscription lifecycle, quotas, webhooks)
	if err := billing.Init(container); err != nil {
		panic(err)
	}

	// Paywall middleware (access gating based on subscription status)
	if err := paywall.SetupMiddleware(container); err != nil {
		panic(err)
	}
	if err := paywall.RegisterNamedMiddlewares(container); err != nil {
		panic(err)
	}

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
