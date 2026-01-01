# CLAUDE.md

AI instructions for working with this Go B2B SaaS Starter Kit codebase.

## Project Overview

Go B2B SaaS Starter Kit - Invoice-to-pay lifecycle automation with AI-powered data extraction, duplicate detection, and payment optimization via Polar.sh integration.

**Architecture**: Modular monolith following Clean Architecture with clear module boundaries.

**Layers**:
- `internal/modules/*/` - Feature modules (domain/app/infra layers per module)
- `internal/platform/` - Cross-cutting concerns (logger, server, etc.)
- `internal/db/` - Database layer (SQLC adapters, DI registration)
- `internal/bootstrap/` - Application initialization
- `pkg/` - Shared utility packages (httperr, response)
- `cmd/api/` - Application entry point

**Core Patterns**:
- Clean Architecture (domain → app → infra)
- Dependency Injection (uber-go/dig)
- Repository Pattern (domain interfaces → store adapters)
- Adapter Pattern (limit SQLC exposure)

## Commands

```bash
# Development
make server          # Run dev server
make build          # Build binary
make deps           # Update Go dependencies

# Database
make run-deps       # Start PostgreSQL in Docker
make migrateup      # Apply migrations
make migratedown    # Rollback migrations
make sqlc           # Generate Go code from SQL

# Testing
make test           # Run tests with coverage

# Docker
make run-stack      # Start full stack
make restart-app    # Restart app container
```

## Database Layer (internal/db/)

**Structure**:
```
internal/db/
├── adapters/           # Legacy adapter interfaces (being phased out)
├── postgres/
│   ├── sqlc/
│   │   ├── migrations/  # SQL migration files
│   │   ├── query/       # SQL queries with SQLC annotations
│   │   └── gen/        # Generated code (DO NOT EDIT)
│   ├── adapter_impl/   # Legacy adapter implementations
│   └── postgres.go     # DB connection and pooling
└── inject.go           # DI registration for all domain repositories
```

**Key Principle**: Domain interfaces (defined in `internal/modules/*/domain/`) are implemented by repositories in `internal/modules/*/infra/repositories/`. The `internal/db/inject.go` registers these implementations in the DI container.

**SQLC Workflow**:
```
internal/db/postgres/sqlc/
├── migrations/     # SQL migration files
├── query/          # SQL queries with SQLC annotations
└── gen/           # Generated code (DO NOT EDIT)
```

### Adding Database Operations

**1. Define domain interface** in your module (`internal/modules/{module}/domain/repository.go`):
```go
package domain

type UserRepository interface {
    GetByID(ctx context.Context, orgID, userID int32) (*User, error)
    Create(ctx context.Context, user *User) (*User, error)
    Update(ctx context.Context, user *User) (*User, error)
    Delete(ctx context.Context, orgID, userID int32) error
}
```

**2. Write SQL query** (`internal/db/postgres/sqlc/query/{domain}.sql`):
```sql
-- name: GetUserByID :one
SELECT * FROM users WHERE organization_id = $1 AND id = $2;

-- name: CreateUser :one
INSERT INTO users (organization_id, email, full_name)
VALUES ($1, $2, $3)
RETURNING *;
```

**3. Generate SQLC code**:
```bash
make sqlc
```

**4. Implement repository** (`internal/modules/{module}/infra/repositories/{domain}_repository.go`):
```go
package repositories

import (
    "github.com/moasq/go-b2b-starter/internal/modules/{module}/domain"
    sqlc "github.com/moasq/go-b2b-starter/internal/db/postgres/sqlc/gen"
)

type userRepository struct {
    store sqlc.Store
}

func NewUserRepository(store sqlc.Store) domain.UserRepository {
    return &userRepository{store: store}
}

func (r *userRepository) GetByID(ctx context.Context, orgID, userID int32) (*domain.User, error) {
    dbUser, err := r.store.GetUserByID(ctx, sqlc.GetUserByIDParams{
        OrganizationID: orgID,
        ID: userID,
    })
    if err != nil {
        return nil, err
    }

    // Map SQLC type to domain type
    return &domain.User{
        ID:             dbUser.ID,
        OrganizationID: dbUser.OrganizationID,
        Email:          dbUser.Email,
        FullName:       dbUser.FullName,
    }, nil
}
```

**5. Register in DI** (`internal/db/inject.go`):
```go
import (
    userDomain "github.com/moasq/go-b2b-starter/internal/modules/users/domain"
    userRepos "github.com/moasq/go-b2b-starter/internal/modules/users/infra/repositories"
)

// In registerDomainStores function:
if err := container.Provide(func(sqlcStore sqlc.Store) userDomain.UserRepository {
    return userRepos.NewUserRepository(sqlcStore)
}); err != nil {
    return fmt.Errorf("failed to provide user repository: %w", err)
}
```

**Why This Architecture**:
- Domain defines interfaces (Dependency Inversion Principle)
- Infra implements these interfaces using SQLC
- SQLC types never leak out of the infra layer
- Easy to mock for testing (depend on interface, not implementation)
- Clear separation of concerns

**Error Handling** (`core/errors.go`):
- `ErrNoRows`, `ErrTxClosed`, `ErrPoolClosed`, `ErrInvalidConnection`, `ErrTimeout`
- Helpers: `IsNoRowsError()`, `IsConstraintError()`, `IsTimeoutError()`

## Authentication (internal/modules/auth/)

**Provider-agnostic auth with Stytch integration**. Type-safe middleware for JWT verification, RBAC, and multi-tenant org context.

**Core Types**:
- `Identity` - User info from auth provider (email, roles, permissions)
- `RequestContext` - Resolved database IDs (OrganizationID, AccountID)
- `Permission` - Format `"resource:action"` (e.g., `"invoice:create"`)

**Middleware Setup**:
```go
authMiddleware := auth.NewMiddleware(authProvider, orgResolver, accResolver, nil)

// Apply middleware
router.Use(authMiddleware.RequireAuth())          // Verify JWT
router.Use(authMiddleware.RequireOrganization())  // Resolve org/account IDs
```

**Route Protection**:
```go
// Permission-based
router.POST("/invoices",
    auth.RequirePermissionFunc("invoice", "create"),
    handler.CreateInvoice)

// Role-based
router.DELETE("/orgs/:id",
    authMiddleware.RequireRole(auth.RoleAdmin),
    handler.DeleteOrg)
```

**Handler Context Access**:
```go
func (h *Handler) MyHandler(c *gin.Context) {
    reqCtx := auth.GetRequestContext(c)
    orgID := reqCtx.OrganizationID      // int32
    accountID := reqCtx.AccountID       // int32
    email := reqCtx.Identity.Email      // string

    // Or use convenience functions
    orgID := auth.GetOrganizationID(c)  // Safe: returns 0 if not set
    accountID := auth.GetAccountID(c)   // Safe: returns 0 if not set
}
```

**Common Permissions** (`permissions.go`):
```go
auth.PermInvoiceCreate    // "invoice:create"
auth.PermInvoiceView      // "invoice:view"
auth.PermInvoiceDelete    // "invoice:delete"
auth.PermOrgView          // "org:view"
auth.PermOrgManage        // "org:manage"
```

**Configuration** (environment variables):
```env
STYTCH_PROJECT_ID=project-test-xxx-xxx    # Required
STYTCH_SECRET=secret-test-xxx             # Required
STYTCH_ENV=test                           # Optional: "test" or "live"
```

**See**: `internal/modules/auth/README.md` for detailed usage patterns and examples.

## File Manager (internal/modules/files/)

**Dual architecture**: Cloudflare R2 (object storage) + PostgreSQL (searchable metadata).

**Components**:
- `FileRepository` - Combined operations (upload, download, delete, search)
- `R2Repository` - R2 object storage operations
- `FileMetadataRepository` - Database metadata operations
- `FileService` - Business logic with validation

**Upload with Entity Linking**:
```go
req := &domain.FileUploadRequest{
    Filename:    "invoice_001.pdf",
    ContentType: "application/pdf",
    Context:     file_manager.ContextInvoice,
}

file := &domain.FileAsset{
    EntityType: "invoice",
    EntityID:   invoiceID,
}

uploadedFile, err := fileService.UploadFile(ctx, req, fileReader)
```

**Search Operations**:
```go
files, err := fileRepo.GetByEntity(ctx, "invoice", invoiceID)
documents, err := fileRepo.GetByCategory(ctx, file_manager.CategoryDocument, 10, 0)
receipts, err := fileRepo.GetByContext(ctx, file_manager.ContextReceipt, 20, 0)
```

**Atomic Transactions**:
1. Save metadata to DB (get ID)
2. Upload file to R2 (using DB ID in key)
3. Update metadata with storage path
4. Automatic rollback on failure

## Go Coding Standards

### Core Rules

**Use `any` instead of `interface{}`** (Go 1.18+):
```go
// ✅ Good
func ProcessData(data any) error
type Request struct { Metadata map[string]any }

// ❌ Bad
func ProcessData(data interface{}) error
```

**Error Wrapping**:
```go
// ✅ Good
if err := repo.Create(ctx, invoice); err != nil {
    return fmt.Errorf("failed to create invoice %d: %w", invoice.ID, err)
}
```

**Context First**:
```go
// ✅ Good
func (s *service) ProcessInvoice(ctx context.Context, invoiceID int32) error
```

**Naming**:
- Packages: lowercase, single word (`invoice`, not `invoice_mgmt`)
- Interfaces: noun/adjective + "er" (`Repository`, `Handler`)
- Structs: PascalCase (`InvoiceService`, `PaymentRequest`)
- Methods: PascalCase verbs (`CreateInvoice`, `ValidateData`)

### Struct Organization
```go
type Invoice struct {
    // Identifiers first
    ID            int32  `json:"id" db:"id"`
    InvoiceNumber string `json:"invoice_number"`

    // Core business data
    Amount        decimal.Decimal `json:"amount"`
    DueDate       time.Time       `json:"due_date"`

    // References
    VendorID      int32 `json:"vendor_id"`

    // Timestamps last
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}
```

### Dependency Injection
```go
// ✅ Constructor returns interface
func NewInvoiceService(
    repo domain.InvoiceRepository,
    logger logger.Logger,
) domain.InvoiceService {
    return &invoiceService{repo: repo, logger: logger}
}

// ✅ Register in DI
container.Provide(NewInvoiceService)
```

### Event-Driven Patterns
```go
// ✅ Events are past tense, include all data
type InvoiceCreatedEvent struct {
    BaseEvent
    InvoiceID int32           `json:"invoice_id"`
    Amount    decimal.Decimal `json:"amount"`
    CreatedAt time.Time       `json:"created_at"`
}
```

### Testing
```go
// ✅ Table-driven tests
func TestInvoiceValidation(t *testing.T) {
    tests := []struct {
        name    string
        invoice *Invoice
        wantErr bool
    }{
        {"valid", &Invoice{Amount: decimal.NewFromInt(100)}, false},
        {"negative", &Invoice{Amount: decimal.NewFromInt(-100)}, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.invoice.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("got error = %v, want %v", err, tt.wantErr)
            }
        })
    }
}
```

## API Development Pattern

**ALWAYS follow this pattern** for consistency and Clean Architecture compliance.

### Implementation Steps

**1. Database Layer** (if new data needed):
- Add SQLC queries: `internal/db/postgres/sqlc/query/{domain}.sql`
- Run `make sqlc`
- Define repository interface in `internal/modules/{module}/domain/repository.go`
- Implement repository in `internal/modules/{module}/infra/repositories/{domain}_repository.go`
- Register in DI: `internal/db/inject.go`

**2. Domain Layer** (`internal/modules/{module}/domain/`):
- Create entities with business types
- Define repository interfaces
- Add validation methods

**3. Infrastructure** (`internal/modules/{module}/infra/repositories/`):
- Implement repository interfaces using SQLC
- Map SQLC types ↔ domain types (never expose SQLC outside infra)
- Handle transactions

**4. Application** (`internal/modules/{module}/app/services/`):
- Define request/response DTOs
- Add service interface
- Implement business logic

**5. API Layer** (`internal/modules/{module}/`):
- Add handler with validation (`handler.go`)
- Add Swagger annotations (see Swagger Best Practices below)
- Register routes (`routes.go`)
- Wire dependencies in module initialization

### Required Handler Pattern
```go
func (h *Handler) OperationName(c *gin.Context) {
    // 1. Extract path params
    var entityID int32
    if _, err := fmt.Sscanf(c.Param("id"), "%d", &entityID); err != nil {
        c.JSON(400, httperr.NewHTTPError(400, "invalid_id", "Invalid ID"))
        return
    }

    // 2. Get auth context
    reqCtx := auth.GetRequestContext(c)
    if reqCtx == nil {
        c.JSON(401, httperr.NewHTTPError(401, "unauthorized", "Auth required"))
        return
    }

    // 3. Bind request (if needed)
    var req models.RequestDto
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, httperr.NewHTTPError(400, "invalid_request", err.Error()))
        return
    }

    // 4. Call service
    response, err := h.service.Operation(c.Request.Context(), reqCtx.OrganizationID, req)
    if err != nil {
        c.JSON(500, httperr.NewHTTPError(500, "operation_failed", err.Error()))
        return
    }

    // 5. Return response
    c.JSON(200, response)
}
```

### Swagger Best Practices

**CRITICAL**: Always use local type references in swagger annotations. Never use full package paths.

**✅ Correct**:
```go
// @Success 200 {object} domain.User "User details"
// @Success 201 {object} services.CreateUserResponse "Created user"
// @Failure 400 {object} httperr.HTTPError "Bad request"
// @Param request body services.CreateUserRequest true "User data"
```

**❌ Wrong**:
```go
// @Success 200 {object} github_com_moasq_go-b2b-starter_internal_modules_users_domain.User
// @Failure 400 {object} errors.HTTPError  // Wrong package name
```

**Common Patterns**:
```go
// Handler in internal/modules/users/handler.go
import (
    "github.com/moasq/go-b2b-starter/internal/modules/users/domain"
    "github.com/moasq/go-b2b-starter/internal/modules/users/app/services"
    "github.com/moasq/go-b2b-starter/pkg/httperr"
)

// @Summary Create user
// @Description Creates a new user in the organization
// @Tags users
// @Accept json
// @Produce json
// @Param request body services.CreateUserRequest true "User data"
// @Success 201 {object} domain.User "Created user"
// @Failure 400 {object} httperr.HTTPError "Invalid request"
// @Failure 500 {object} httperr.HTTPError "Internal error"
// @Router /api/users [post]
func (h *Handler) CreateUser(c *gin.Context) {
    // Implementation
}
```

**Docker Compose Best Practices**:

When using docker-compose for CLI tools, use `${PWD}` for volume mounts:
```yaml
cli:
  volumes:
    - ${PWD}:/workspace  # ✅ Correct - uses current directory
    # - ../:/workspace   # ❌ Wrong - relative paths don't work consistently
  working_dir: /workspace
```

### Required Service Pattern
```go
func (s *service) Operation(ctx context.Context, orgID int32, req *Request) (*Response, error) {
    // 1. Validate
    if err := req.Validate(); err != nil {
        return nil, err
    }

    // 2. Execute business logic
    result, err := s.repo.Operation(ctx, orgID, req)
    if err != nil {
        return nil, fmt.Errorf("operation failed: %w", err)
    }

    // 3. Return
    return result, nil
}
```

### Required Middleware Pattern
```go
// In routes.go
apiGroup := router.Group("/{domain}")
apiGroup.Use(
    authMiddleware.RequireAuth(),
    authMiddleware.RequireOrganization(),
)
{
    apiGroup.POST("/path",
        auth.RequirePermissionFunc("{resource}", "{action}"),
        h.HandlerMethod)
}
```

### Mandatory Requirements
- **Context**: First parameter in all I/O operations
- **Error Wrapping**: `fmt.Errorf("context: %w", err)`
- **Validation**: At both entity and service levels
- **Logging**: Structured logging for operations
- **Middleware**: Auth, org context, permissions
- **Swagger**: Complete API documentation
- **Transactions**: Atomic operations where needed

### Testing Requirements
- Unit tests with mocked dependencies
- Integration tests with database
- Validation tests for all error scenarios
- Permission tests for access control

## Dependency Management

**Rule**: Use interface abstractions when dependencies aren't ready.

```go
// ✅ Good - Depend on interface
type OCRService interface {
    ExtractData(ctx context.Context, fileID int32) (map[string]any, error)
}

// ✅ Good - Event-driven integration
func NewInvoiceService(eventBus eventbus.EventBus) InvoiceService {
    // Publish events; other modules subscribe when ready
}

// ❌ Bad - Don't inject concrete types that don't exist
func NewInvoiceService(ocrService *OCRServiceImpl) InvoiceService {
    // Fails if OCRServiceImpl doesn't exist
}
```

## Project Structure

```
go-b2b-starter/
├── cmd/api/                    # Application entry point
├── internal/
│   ├── bootstrap/             # App initialization and wiring
│   ├── db/                    # Database layer (SQLC, DI registration)
│   │   ├── postgres/sqlc/     # SQLC queries and generated code
│   │   ├── adapters/          # Legacy adapters (being phased out)
│   │   └── inject.go          # Repository DI registration
│   ├── modules/               # Feature modules
│   │   ├── {module}/
│   │   │   ├── domain/        # Entities, interfaces, validation
│   │   │   ├── app/services/  # Business logic (use cases)
│   │   │   ├── infra/         # Repository implementations
│   │   │   ├── handler.go     # HTTP handlers
│   │   │   ├── routes.go      # Route definitions
│   │   │   └── module.go      # Module DI setup
│   │   ├── auth/              # Authentication & RBAC
│   │   ├── billing/           # Polar.sh subscriptions
│   │   ├── organizations/     # Multi-tenant org management
│   │   ├── documents/         # PDF document management
│   │   ├── cognitive/         # RAG and embeddings
│   │   ├── files/             # File storage (R2 + metadata)
│   │   └── paywall/           # Subscription access gating
│   └── platform/              # Cross-cutting concerns
│       ├── logger/            # Structured logging
│       ├── server/            # HTTP server
│       ├── eventbus/          # Event pub-sub
│       ├── llm/               # LLM client
│       ├── ocr/               # OCR service
│       ├── redis/             # Redis client
│       └── stytch/            # Stytch B2B client
└── pkg/                       # Public shared utilities
    ├── httperr/               # HTTP error responses
    ├── pagination/            # Pagination helpers
    ├── response/              # Standard API responses
    └── slugify/               # Slug generation utilities
```

## Billing & Paywall (internal/modules/billing/)

**Polar.sh integration** with hybrid sync for subscriptions.

**Sync Strategy**:
1. **Webhooks** (primary) - Real-time updates from Polar.sh
2. **Active Verification** - Poll after checkout redirect
3. **Lazy Guarding** - Verify with API if local data suggests expired

**Paywall Middleware**:
```go
// Require active subscription
premiumGroup := router.Group("/premium")
premiumGroup.Use(
    resolver.Get("auth"),
    resolver.Get("org_context"),
    resolver.Get("paywall"),  // RequireActiveSubscription
)

// Optional subscription info
publicGroup.Use(resolver.Get("paywall_optional"))
```

**Quota Management**:
```go
// Check and consume quota
status, err := billingService.ConsumeInvoiceQuota(ctx, orgID)
if err == domain.ErrInsufficientQuota {
    // Handle quota exhausted
}
```

## Event Bus (internal/platform/eventbus/)

**In-memory event bus** for loose coupling between modules.

**Define Event**:
```go
type DocumentUploadedEvent struct {
    eventbus.BaseEvent
    DocumentID     int32 `json:"document_id"`
    OrganizationID int32 `json:"organization_id"`
}

func NewDocumentUploadedEvent(docID, orgID int32) *DocumentUploadedEvent {
    return &DocumentUploadedEvent{
        BaseEvent:      eventbus.NewBaseEvent("document.uploaded"),
        DocumentID:     docID,
        OrganizationID: orgID,
    }
}
```

**Publish**:
```go
event := NewDocumentUploadedEvent(doc.ID, orgID)
eventBus.Publish(ctx, event)  // Fire-and-forget
```

**Subscribe**:
```go
eventBus.Subscribe("document.uploaded", func(ctx context.Context, event eventbus.Event) error {
    docEvent := event.(*DocumentUploadedEvent)
    return embeddingService.GenerateForDocument(ctx, docEvent.DocumentID)
})
```

## Module Initialization Order

**File**: `internal/bootstrap/bootstrap.go`

Order matters due to dependencies:

```go
// Phase 1: Infrastructure (no dependencies)
logger.Inject(container)
server.Inject(container)
db.Inject(container)           // Registers all domain repositories

// Phase 2: Platform Services
redis.Inject(container)
llm.Inject(container)
ocr.Inject(container)
polar.Inject(container)
eventbus.Inject(container)

// Phase 3: Module Dependencies (order critical!)
files.SetupDependencies(container)     // File storage
auth.SetupDependencies(container)      // Auth, RBAC, resolvers
organizations.RegisterDependencies(container)
billing.Configure(container)
cognitive.RegisterDependencies(container)
documents.RegisterDependencies(container)

// Phase 4: Event Subscriptions
cognitive.SetupEventSubscriptions(container)

// Phase 5: HTTP Server Setup
server.SetupMiddleware(container)
```

## Named Middlewares

Access middleware by name in routes:

```go
func (r *Routes) Routes(router *gin.RouterGroup, resolver server.MiddlewareResolver) {
    group := router.Group("/api")
    group.Use(
        resolver.Get("auth"),          // RequireAuth
        resolver.Get("org_context"),   // RequireOrganization
        resolver.Get("paywall"),       // RequireActiveSubscription
        resolver.Get("paywall_optional"), // OptionalSubscriptionStatus
        resolver.Get("subscription"),  // Deprecated alias
    )
}
```

## Configuration

Environment-based configuration using `app.env` and `example.env`. Docker Compose for local dependencies.

## Documentation

Comprehensive documentation available in `docs/`:

- `docs/README.md` - Overview and quick start
- `docs/architecture.md` - Clean Architecture patterns
- `docs/database.md` - SQLC workflow and migrations
- `docs/authentication.md` - Auth, RBAC, Stytch integration
- `docs/billing.md` - Polar.sh and paywall
- `docs/file-manager.md` - R2 file storage
- `docs/event-bus.md` - Event-driven patterns
- `docs/api-development.md` - Step-by-step API guide
- `docs/modules/` - Module-specific documentation
