package api

import (
	cognitiveAPI "github.com/moasq/go-b2b-starter/api/example_cognitive"
	documentsAPI "github.com/moasq/go-b2b-starter/api/example_documents"
	organizations "github.com/moasq/go-b2b-starter/api/organizations"
	rbacAPI "github.com/moasq/go-b2b-starter/api/rbac"
	subscriptionsAPI "github.com/moasq/go-b2b-starter/api/subscriptions"
	server "github.com/moasq/go-b2b-starter/server/domain"
	"go.uber.org/dig"
)

// moduleRoutes holds handlers for all API modules
// 1. OrganizationRoutes - Handles organization, account, and member management routes (includes /auth routes)
// 2. RbacRoutes - Handles RBAC role and permission routes
// 3. BillingHandler - Handles billing status and subscription routes (uses app/billing module)
// 4. DocumentsRoutes - Handles PDF document upload and management routes
// 5. CognitiveRoutes - Handles AI/RAG chat and document search routes
type moduleRoutes struct {
	OrganizationRoutes  *organizations.Routes
	RbacRoutes          *rbacAPI.Routes
	SubscriptionHandler *subscriptionsAPI.Handler
	DocumentsRoutes     *documentsAPI.Routes
	CognitiveRoutes     *cognitiveAPI.Routes
}

// 1. Sets up all module dependencies
// 2. Registers API routes and handlers
// 3. Registers tools routes and handlers
// 4. Registers exchange rates routes and handlers
// 5. Registers visa routes and handlers

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
// 1. Provides moduleRoutes struct with all handlers
// 2. Invokes route registration for each module
func registerAPI(container *dig.Container) error {
	if err := container.Provide(func(
		organizationRoutes *organizations.Routes,
		rbacRoutes *rbacAPI.Routes,
		subscriptionHandler *subscriptionsAPI.Handler,
		documentsRoutes *documentsAPI.Routes,
		cognitiveRoutes *cognitiveAPI.Routes,
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
		// Note: OrganizationRoutes now includes /auth routes for member management
		srv.RegisterRoutes(modules.OrganizationRoutes.Routes, server.ApiPrefix)
		srv.RegisterRoutes(modules.RbacRoutes.Routes, server.ApiPrefix)
		srv.RegisterRoutes(modules.SubscriptionHandler.Routes, server.ApiPrefix)
		srv.RegisterRoutes(modules.DocumentsRoutes.Routes, server.ApiPrefix)
		srv.RegisterRoutes(modules.CognitiveRoutes.Routes, server.ApiPrefix)
	})
}

// setupDependencies initializes all module dependencies
// 1. Organizations API - multi-tenant organization management (includes auth/member routes)
// 2. RBAC API - role-based access control
// 3. Billing API - subscription and billing management (handler uses app/billing module)
// 4. Documents API - PDF document upload and management
// 5. Cognitive API - AI/RAG chat and document search
func setupDependencies(container *dig.Container) error {
	if err := organizations.NewProvider(container).RegisterDependencies(); err != nil {
		return err
	}

	// Initialize RBAC API (role and permission discovery)
	if err := rbacAPI.NewProvider(container).RegisterDependencies(); err != nil {
		return err
	}

	// Initialize billing API (subscription and billing status)
	if err := subscriptionsAPI.RegisterHandlers(container); err != nil {
		return err
	}

	// Initialize documents API (PDF upload and management)
	if err := documentsAPI.NewProvider(container).RegisterDependencies(); err != nil {
		return err
	}

	// Initialize cognitive API (AI/RAG chat and document search)
	if err := cognitiveAPI.NewProvider(container).RegisterDependencies(); err != nil {
		return err
	}

	return nil
}
