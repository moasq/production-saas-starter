# Backend Architecture

The backend is a **Modular Monolith** using **idiomatic Go project layout** with Clear separation between business features and infrastructure.

## High-Level Structure

```
go-b2b-starter/
├── cmd/                  # Application entry points
│   └── api/
│       └── main.go       # Main entry point
│
├── internal/             # Private application code (import boundary)
│   ├── modules/          # Feature modules (business domains)
│   ├── platform/         # Cross-cutting infrastructure
│   ├── db/               # Database layer (SQLC, DI registration)
│   ├── bootstrap/        # Application initialization
│   └── api/              # API route registration
│
├── pkg/                  # Public reusable packages
│   ├── httperr/          # HTTP error types
│   ├── pagination/       # Pagination helpers
│   └── response/         # API response utilities
│
└── go.mod                # Single consolidated module
```

## Architectural Layers

### Feature Modules (`internal/modules/`)

Business domain modules following **Clean Architecture**. Each module represents a distinct business capability:

```
internal/modules/
├── auth/             # Authentication & RBAC
├── billing/          # Polar.sh subscriptions & quota management
├── organizations/    # Multi-tenant organization management
├── documents/        # PDF document processing
├── cognitive/        # RAG (Retrieval-Augmented Generation) & embeddings
├── files/            # File storage (R2 + metadata)
└── paywall/          # Subscription middleware
```

**Characteristics of Feature Modules:**
- ✅ Business domain logic
- ✅ Domain entities with business rules
- ✅ API endpoints
- ✅ Use cases and workflows
- ✅ Follow Clean Architecture layers (domain → app → infra)

### Platform Services (`internal/platform/`)

Cross-cutting infrastructure components used by multiple modules:

```
internal/platform/
├── server/           # HTTP server & middleware
├── eventbus/         # Event pub/sub system
├── logger/           # Structured logging
├── redis/            # Redis cache client
├── stytch/           # Stytch auth provider client
├── polar/            # Polar.sh billing provider client
├── llm/              # LLM integration (OpenAI)
└── ocr/              # OCR service (Mistral)
```

**Characteristics of Platform Components:**
- ✅ Infrastructure concerns
- ✅ Used by multiple modules
- ✅ No business logic
- ✅ Provide technical capabilities

### Database Layer (`internal/db/`)

Centralized database layer using SQLC:

```
internal/db/
├── postgres/
│   ├── sqlc/
│   │   ├── migrations/    # SQL migration files
│   │   ├── query/         # SQL queries with SQLC annotations
│   │   └── gen/          # Generated Go code
│   └── postgres.go       # DB connection and pooling
├── inject.go             # DI registration for all repositories
└── core/
    └── errors.go         # Database error types
```

**Key Responsibilities:**
- SQLC code generation
- Repository DI registration
- Database connection management
- Transaction support

## Module Structure (Clean Architecture)

Each feature module in `internal/modules/` follows **Clean Architecture**:

```mermaid
graph TD
    A[Handler] --> B[Service]
    B --> C[Repository Interface]
    C --> D[Repository Implementation]
    D --> E[SQLC Store]
    E --> F[PostgreSQL]

    style A fill:#e1f5fe
    style B fill:#fff9c4
    style C fill:#f3e5f5
    style D fill:#e8f5e9
    style E fill:#fce4ec
    style F fill:#f5f5f5
```

### Layer Details

```
internal/modules/billing/
├── cmd/                  # Module initialization (DI wiring)
│   └── init.go
│
├── app/                  # Application Layer (Use Cases)
│   └── services/
│       └── billing_service.go
│
├── domain/               # Domain Layer (Core Business Logic)
│   ├── entity.go         # Data structures
│   ├── repository.go     # Interface definitions
│   ├── errors.go         # Domain errors
│   └── events/           # Domain events
│
├── infra/                # Infrastructure Layer (External)
│   ├── repositories/     # Repository implementations
│   │   └── subscription_repository.go
│   └── polar/           # Polar.sh adapter
│       └── polar_adapter.go
│
├── handler.go            # HTTP handlers (Delivery Layer)
├── routes.go             # Route registration
└── module.go             # Dependency injection setup
```

## Key Principles

### 1. `internal/` Boundary

Code in `internal/` cannot be imported by external packages. This enforces encapsulation and prevents unintended dependencies.

### 2. Dependency Rule (Clean Architecture)

```mermaid
graph LR
    A[Domain] --> B[Application]
    B --> C[Infrastructure]
    C --> D[Delivery/Handler]

    style A fill:#c8e6c9
    style B fill:#fff9c4
    style C fill:#ffccbc
    style D fill:#b3e5fc
```

**Direction of dependencies:**
- Domain → Nothing (pure business logic)
- Application → Domain (uses domain interfaces)
- Infrastructure → Domain (implements domain interfaces)
- Handlers → Application (calls services)

**Key Point**: Inner layers never depend on outer layers. Infrastructure implements interfaces defined in domain.

### 3. Feature-Based Organization

Modules are organized by **business feature** (billing, auth, organizations), not by technical layer (controllers, services, models).

This promotes:
- High cohesion within features
- Low coupling between features
- Easy to understand and navigate
- Clear ownership and boundaries

### 4. Single `go.mod`

One module for the entire project eliminates workspace complexity and simplifies dependency management.

## Request Flow

```mermaid
sequenceDiagram
    participant Client
    participant Middleware
    participant Handler
    participant Service
    participant Repository
    participant Database

    Client->>Middleware: HTTP Request
    Middleware->>Middleware: Auth & Validation
    Middleware->>Handler: Authenticated Request
    Handler->>Handler: Parse & Validate
    Handler->>Service: Call Use Case
    Service->>Service: Business Logic
    Service->>Repository: Get/Save Data (Interface)
    Repository->>Database: SQL Query (SQLC)
    Database-->>Repository: Result
    Repository-->>Service: Domain Entity
    Service-->>Handler: DTO Response
    Handler-->>Client: JSON Response
```

### Flow Breakdown

1. **Client Request**: HTTP request arrives at server
2. **Middleware**: Auth, logging, rate limiting, CORS
3. **Handler**: Parse request, extract context (org ID, user ID)
4. **Service**: Execute business logic, orchestrate operations
5. **Repository**: Data access using domain interfaces
6. **Database**: SQLC-generated type-safe queries

## Initialization Flow

```mermaid
graph TD
    A[cmd/api/main.go] --> B[bootstrap.Execute]
    B --> C[InitMods]
    C --> D[Infrastructure Layer]
    C --> E[Platform Services]
    C --> F[Feature Modules]
    C --> G[API Routes]

    D --> D1[Logger]
    D --> D2[Server]
    D --> D3[Database]

    E --> E1[Redis]
    E --> E2[EventBus]
    E --> E3[Stytch]
    E --> E4[Polar]

    F --> F1[Auth Module]
    F --> F2[Billing Module]
    F --> F3[Organizations Module]
    F --> F4[Documents Module]

    G --> G1[Register Routes]
    G --> G2[Setup Middleware]

    style A fill:#e3f2fd
    style D fill:#fff9c4
    style E fill:#f3e5f5
    style F fill:#e8f5e9
    style G fill:#fce4ec
```

### Initialization Order (Critical)

```
cmd/api/main.go
    └── bootstrap.Execute()
        ├── 1. Infrastructure (no dependencies)
        │   ├── logger.Inject()
        │   ├── server.Inject()
        │   └── db.Inject()              # Registers all domain repositories
        │
        ├── 2. Platform Services
        │   ├── redis.Inject()
        │   ├── llm.Inject()
        │   ├── ocr.Inject()
        │   ├── polar.Inject()
        │   └── eventbus.Inject()
        │
        ├── 3. Feature Modules (order matters!)
        │   ├── files.SetupDependencies()
        │   ├── auth.SetupDependencies()
        │   ├── organizations.RegisterDependencies()
        │   ├── billing.Configure()
        │   ├── cognitive.RegisterDependencies()
        │   └── documents.RegisterDependencies()
        │
        ├── 4. Event Subscriptions
        │   └── cognitive.SetupEventSubscriptions()
        │
        └── 5. HTTP Server Setup
            ├── server.SetupMiddleware()
            └── api.RegisterRoutes()
```

**Why Order Matters:**
- `db.Inject()` must run early (registers all repositories)
- `auth` must be before modules that need auth middleware
- `files` must be before `documents` (documents depend on files)
- Event subscriptions must be after all modules are loaded

## Dependency Injection

The project uses **uber-go/dig** for dependency injection:

```go
// 1. Define interface in domain
package domain
type ProductRepository interface {
    Create(ctx context.Context, p *Product) (*Product, error)
}

// 2. Implement in infrastructure
package repositories
func NewProductRepository(store sqlc.Store) domain.ProductRepository {
    return &productRepository{store: store}
}

// 3. Register in DI (internal/db/inject.go)
container.Provide(func(sqlcStore sqlc.Store) productDomain.ProductRepository {
    return productRepos.NewProductRepository(sqlcStore)
})

// 4. Inject into service
package services
func NewProductService(repo domain.ProductRepository) ProductService {
    return &productService{repo: repo}
}
```

**Benefits:**
- Automatic dependency resolution
- Easy testing (inject mocks)
- Clear dependency graph
- Compile-time safety

## Modules vs Platform Decision

```mermaid
graph TD
    A{Is it a business domain feature?}
    A -->|Yes| B[Create in internal/modules/]
    A -->|No| C{Used by multiple modules?}
    C -->|Yes| D[Create in internal/platform/]
    C -->|No| E{Part of existing module?}
    E -->|Yes| F[Add to that module]
    E -->|No| D

    style B fill:#c8e6c9
    style D fill:#bbdefb
    style F fill:#fff9c4
```

### Examples

| Component | Location | Reason |
|-----------|----------|--------|
| Product Catalog | `modules/products/` | Business domain feature |
| Invoice Processing | `modules/invoices/` | Business domain feature |
| Subscription Management | `modules/billing/` | Business domain feature |
| Event Bus | `platform/eventbus/` | Used by all modules |
| Logging | `platform/logger/` | Cross-cutting infrastructure |
| Stytch Client | `platform/stytch/` | Auth provider client |
| Auth Module | `modules/auth/` | Business auth logic using Stytch |

## Communication Between Modules

Modules communicate through:

### 1. Event Bus (Loosely Coupled)

```go
// Module A publishes event
eventBus.Publish(ctx, events.NewDocumentUploadedEvent(docID, orgID))

// Module B subscribes
eventBus.Subscribe("document.uploaded", func(ctx context.Context, event Event) error {
    docEvent := event.(*DocumentUploadedEvent)
    return embeddingService.GenerateForDocument(ctx, docEvent.DocumentID)
})
```

**Use When:**
- Asynchronous operations
- One-to-many communication
- Loose coupling desired

### 2. Direct Service Injection (Tightly Coupled)

```go
// Service A uses Service B
func NewInvoiceService(
    repo domain.InvoiceRepository,
    billingService billing.BillingService,  // Direct dependency
) InvoiceService {
    return &invoiceService{
        repo: repo,
        billing: billingService,
    }
}
```

**Use When:**
- Synchronous operations
- One-to-one communication
- Strong dependency relationship

### 3. Shared Platform Components

```go
// Multiple modules use platform logger
func NewDocumentService(
    repo domain.DocumentRepository,
    logger logger.Logger,  // Platform component
) DocumentService {
    return &documentService{repo: repo, logger: logger}
}
```

## Multi-Tenancy

Every module supports multi-tenancy through **Organization ID**:

```go
// Organization context in middleware
type RequestContext struct {
    OrganizationID int32  // From JWT token
    AccountID      int32  // User account ID
    Identity       *Identity
}

// Used in handlers
reqCtx := auth.GetRequestContext(c)
products, err := service.ListByOrganization(ctx, reqCtx.OrganizationID)

// Enforced in repository
SELECT * FROM products WHERE organization_id = $1 AND id = $2
```

**Benefits:**
- Data isolation per organization
- Single database for all tenants
- Efficient resource usage
- Simplified deployment

## File Structure Summary

| Layer | Path | Purpose |
|-------|------|---------|
| Entry point | `cmd/api/main.go` | Application startup |
| Feature modules | `internal/modules/*/` | Business domains |
| Platform services | `internal/platform/*/` | Infrastructure |
| Database layer | `internal/db/` | SQLC, migrations, DI |
| Bootstrap | `internal/bootstrap/` | Initialization |
| API routes | `internal/api/` | Route registration |
| Shared utilities | `pkg/*/` | Public packages |

## Next Steps

- **Adding a Module**: See [Adding a Module Guide](./02-adding-a-module.md)
- **Database Operations**: See [Database Guide](./database.md)
- **API Development**: See [API Development Guide](./api-development.md)
- **Authentication**: See [Authentication Guide](./authentication.md)
- **Event Bus**: See [Event Bus Guide](./event-bus.md)
