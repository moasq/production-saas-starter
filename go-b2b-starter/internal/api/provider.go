package api

import (
	"go.uber.org/dig"

	"github.com/moasq/go-b2b-starter/internal/auth"
	"github.com/moasq/go-b2b-starter/internal/billing"
	"github.com/moasq/go-b2b-starter/internal/cognitive"
	"github.com/moasq/go-b2b-starter/internal/documents"
	"github.com/moasq/go-b2b-starter/internal/organizations"
	server "github.com/moasq/go-b2b-starter/internal/server/domain"
)

// moduleRoutes holds handlers for all API modules
// 1. OrganizationRoutes - Handles organization, account, and member management routes (includes /auth routes)
// 2. RbacRoutes - Handles RBAC role and permission routes
// 3. BillingHandler - Handles billing status and subscription routes (uses billing module)
// 4. DocumentsRoutes - Handles PDF document upload and management routes
// 5. CognitiveRoutes - Handles AI/RAG chat and document search routes
type moduleRoutes struct {
	OrganizationRoutes  *organizations.Routes
	RbacRoutes          *auth.Routes
	SubscriptionHandler *billing.Handler
	DocumentsRoutes     *documents.Routes
	CognitiveRoutes     *cognitive.Routes
}

// Init sets up all module dependencies and registers API routes
func Init(container *dig.Container) error {
	if err := setupDependencies(container); err != nil {
		return err
	}

	if err := registerAPI(container); err != nil {
		return err
	}
	return nil
}

// registerAPI registers all module handlers and routes
func registerAPI(container *dig.Container) error {
	if err := container.Provide(func(
		organizationRoutes *organizations.Routes,
		rbacRoutes *auth.Routes,
		subscriptionHandler *billing.Handler,
		documentsRoutes *documents.Routes,
		cognitiveRoutes *cognitive.Routes,
	) *moduleRoutes {
		return &moduleRoutes{
			OrganizationRoutes:  organizationRoutes,
			RbacRoutes:          rbacRoutes,
			SubscriptionHandler: subscriptionHandler,
			DocumentsRoutes:     documentsRoutes,
			CognitiveRoutes:     cognitiveRoutes,
		}
	}); err != nil {
		return err
	}

	return container.Invoke(func(
		srv server.Server,
		modules *moduleRoutes,
	) {
		// Register each module's routes
		srv.RegisterRoutes(modules.OrganizationRoutes.Routes, server.ApiPrefix)
		srv.RegisterRoutes(modules.RbacRoutes.Routes, server.ApiPrefix)
		srv.RegisterRoutes(modules.SubscriptionHandler.Routes, server.ApiPrefix)
		srv.RegisterRoutes(modules.DocumentsRoutes.Routes, server.ApiPrefix)
		srv.RegisterRoutes(modules.CognitiveRoutes.Routes, server.ApiPrefix)
	})
}

// setupDependencies initializes all module dependencies
func setupDependencies(container *dig.Container) error {
	if err := organizations.NewProvider(container).RegisterDependencies(); err != nil {
		return err
	}

	// Initialize RBAC API (role and permission discovery)
	if err := auth.NewProvider(container).RegisterDependencies(); err != nil {
		return err
	}

	// Initialize billing API (subscription and billing status)
	if err := billing.RegisterHandlers(container); err != nil {
		return err
	}

	// Initialize documents API (PDF upload and management)
	if err := documents.NewProvider(container).RegisterDependencies(); err != nil {
		return err
	}

	// Initialize cognitive API (AI/RAG chat and document search)
	if err := cognitive.NewProvider(container).RegisterDependencies(); err != nil {
		return err
	}

	return nil
}
