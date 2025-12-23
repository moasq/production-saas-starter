# Adding a New Module

This guide shows how to create a new feature module following Clean Architecture and the idiomatic Go project layout.

## Modules vs Platform Decision

Before creating a new module, determine whether it should be a **feature module** or a **platform component**.

### Create a Module (`internal/modules/`) when:
✅ Implementing a business domain feature
✅ Has domain entities with business rules
✅ Exposes API endpoints
✅ Contains use cases and workflows
✅ **Examples**: billing, documents, organizations, invoices, products

### Create a Platform Component (`internal/platform/`) when:
✅ Infrastructure or cross-cutting concern
✅ Used by multiple modules
✅ No business logic (pure infrastructure)
✅ Provides technical capability
✅ **Examples**: logger, eventbus, redis, http server

### Decision Tree

```
Is it a business domain feature?
├─ Yes → Create in internal/modules/{name}/
└─ No → Is it used by multiple modules?
    ├─ Yes → Create in internal/platform/{name}/
    └─ No → Should it be part of an existing module?
```

## Module Location

All feature modules live in `internal/modules/` which enforces Go's import boundary:

```
internal/
├── modules/               # Feature modules (business domains)
│   ├── auth/             # Authentication & RBAC
│   ├── billing/          # Subscription & billing
│   ├── organizations/    # Multi-tenant organizations
│   ├── documents/        # Document management
│   ├── cognitive/        # AI/RAG features
│   ├── files/            # File storage
│   ├── paywall/          # Subscription middleware
│   └── products/         # ← Your new module here
│
├── platform/             # Cross-cutting infrastructure
│   ├── server/           # HTTP server
│   ├── eventbus/         # Event pub/sub
│   ├── logger/           # Structured logging
│   ├── redis/            # Redis client
│   └── ...
│
├── db/                   # Database layer
└── bootstrap/            # App initialization
```

## Module Structure

Each module follows **Clean Architecture** with these layers:

```
internal/modules/products/
├── cmd/                      # Module initialization (DI wiring)
│   └── init.go
│
├── app/                      # Application Layer (Use Cases)
│   └── services/
│       └── product_service.go
│
├── domain/                   # Domain Layer (Core Business Logic)
│   ├── entity.go             # Data structures
│   └── repository.go         # Interface definitions
│
├── infra/                    # Infrastructure Layer (External)
│   └── repositories/
│       └── product_repository.go
│
├── handler.go                # HTTP handlers (Delivery Layer)
├── routes.go                 # Route registration
└── module.go                 # Dependency injection setup
```

## Step-by-Step Guide

### 1. Define the Entity (`domain/entity.go`)

Start with your core business objects:

```go
package domain

import "time"

type Product struct {
    ID             int32     `json:"id"`
    Name           string    `json:"name"`
    Description    string    `json:"description"`
    Price          float64   `json:"price"`
    OrganizationID int32     `json:"organization_id"`
    CreatedAt      time.Time `json:"created_at"`
    UpdatedAt      time.Time `json:"updated_at"`
}

// Validate validates the product data
func (p *Product) Validate() error {
    if p.Name == "" {
        return ErrInvalidProductName
    }
    if p.Price < 0 {
        return ErrInvalidPrice
    }
    return nil
}
```

### 2. Define the Repository Interface (`domain/repository.go`)

Define what operations your module needs:

```go
package domain

import "context"

type ProductRepository interface {
    Create(ctx context.Context, p *Product) (*Product, error)
    GetByID(ctx context.Context, orgID, id int32) (*Product, error)
    ListByOrganization(ctx context.Context, orgID int32, limit, offset int32) ([]*Product, error)
    Update(ctx context.Context, p *Product) error
    Delete(ctx context.Context, orgID, id int32) error
}
```

**Key Points:**
- Interface uses **domain types**, not database types
- Defined in the domain layer (where it's used)
- Independent of implementation details

### 3. Implement the Repository (`infra/repositories/product_repository.go`)

Implement the interface using SQLC:

```go
package repositories

import (
    "context"
    "fmt"

    "github.com/moasq/go-b2b-starter/internal/modules/products/domain"
    sqlc "github.com/moasq/go-b2b-starter/internal/db/postgres/sqlc/gen"
)

type productRepository struct {
    store sqlc.Store
}

func NewProductRepository(store sqlc.Store) domain.ProductRepository {
    return &productRepository{store: store}
}

func (r *productRepository) Create(ctx context.Context, p *domain.Product) (*domain.Product, error) {
    dbProduct, err := r.store.CreateProduct(ctx, sqlc.CreateProductParams{
        Name:           p.Name,
        Description:    p.Description,
        Price:          p.Price,
        OrganizationID: p.OrganizationID,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create product: %w", err)
    }

    // Map SQLC type to domain type
    return &domain.Product{
        ID:             dbProduct.ID,
        Name:           dbProduct.Name,
        Description:    dbProduct.Description,
        Price:          dbProduct.Price,
        OrganizationID: dbProduct.OrganizationID,
        CreatedAt:      dbProduct.CreatedAt,
        UpdatedAt:      dbProduct.UpdatedAt,
    }, nil
}

func (r *productRepository) GetByID(ctx context.Context, orgID, id int32) (*domain.Product, error) {
    dbProduct, err := r.store.GetProductByID(ctx, sqlc.GetProductByIDParams{
        OrganizationID: orgID,
        ID:             id,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to get product: %w", err)
    }

    return &domain.Product{
        ID:             dbProduct.ID,
        Name:           dbProduct.Name,
        Description:    dbProduct.Description,
        Price:          dbProduct.Price,
        OrganizationID: dbProduct.OrganizationID,
        CreatedAt:      dbProduct.CreatedAt,
        UpdatedAt:      dbProduct.UpdatedAt,
    }, nil
}

// ... implement other methods
```

### 4. Create the Service (`app/services/product_service.go`)

Business logic lives here:

```go
package services

import (
    "context"
    "fmt"

    "github.com/moasq/go-b2b-starter/internal/modules/products/domain"
)

type ProductService interface {
    Create(ctx context.Context, orgID int32, req *CreateProductRequest) (*domain.Product, error)
    GetByID(ctx context.Context, orgID, id int32) (*domain.Product, error)
    ListByOrganization(ctx context.Context, orgID int32, limit, offset int32) ([]*domain.Product, error)
}

type productService struct {
    repo domain.ProductRepository
}

func NewProductService(repo domain.ProductRepository) ProductService {
    return &productService{repo: repo}
}

type CreateProductRequest struct {
    Name        string  `json:"name" binding:"required"`
    Description string  `json:"description"`
    Price       float64 `json:"price" binding:"required,min=0"`
}

func (s *productService) Create(ctx context.Context, orgID int32, req *CreateProductRequest) (*domain.Product, error) {
    product := &domain.Product{
        Name:           req.Name,
        Description:    req.Description,
        Price:          req.Price,
        OrganizationID: orgID,
    }

    // Validate
    if err := product.Validate(); err != nil {
        return nil, err
    }

    // Create
    created, err := s.repo.Create(ctx, product)
    if err != nil {
        return nil, fmt.Errorf("failed to create product: %w", err)
    }

    return created, nil
}
```

### 5. Create the Handler (`handler.go`)

HTTP request handling:

```go
package products

import (
    "fmt"
    "net/http"

    "github.com/gin-gonic/gin"

    "github.com/moasq/go-b2b-starter/internal/modules/auth"
    "github.com/moasq/go-b2b-starter/internal/modules/products/app/services"
    "github.com/moasq/go-b2b-starter/pkg/httperr"
)

type Handler struct {
    service services.ProductService
}

func NewHandler(service services.ProductService) *Handler {
    return &Handler{service: service}
}

// @Summary Create product
// @Description Creates a new product in the organization
// @Tags products
// @Accept json
// @Produce json
// @Param request body services.CreateProductRequest true "Product data"
// @Success 201 {object} domain.Product "Created product"
// @Failure 400 {object} httperr.HTTPError "Invalid request"
// @Failure 500 {object} httperr.HTTPError "Internal error"
// @Router /api/products [post]
func (h *Handler) Create(c *gin.Context) {
    reqCtx := auth.GetRequestContext(c)
    if reqCtx == nil {
        c.JSON(http.StatusUnauthorized, httperr.NewHTTPError(
            http.StatusUnauthorized,
            "unauthorized",
            "Authentication required",
        ))
        return
    }

    var req services.CreateProductRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, httperr.NewHTTPError(
            http.StatusBadRequest,
            "invalid_request",
            err.Error(),
        ))
        return
    }

    product, err := h.service.Create(c.Request.Context(), reqCtx.OrganizationID, &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, httperr.NewHTTPError(
            http.StatusInternalServerError,
            "creation_failed",
            err.Error(),
        ))
        return
    }

    c.JSON(http.StatusCreated, product)
}

// @Summary Get product
// @Description Gets a product by ID
// @Tags products
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} domain.Product "Product details"
// @Failure 400 {object} httperr.HTTPError "Invalid ID"
// @Failure 404 {object} httperr.HTTPError "Product not found"
// @Router /api/products/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
    reqCtx := auth.GetRequestContext(c)
    if reqCtx == nil {
        c.JSON(http.StatusUnauthorized, httperr.NewHTTPError(
            http.StatusUnauthorized,
            "unauthorized",
            "Authentication required",
        ))
        return
    }

    var productID int32
    if _, err := fmt.Sscanf(c.Param("id"), "%d", &productID); err != nil {
        c.JSON(http.StatusBadRequest, httperr.NewHTTPError(
            http.StatusBadRequest,
            "invalid_id",
            "Product ID must be a number",
        ))
        return
    }

    product, err := h.service.GetByID(c.Request.Context(), reqCtx.OrganizationID, productID)
    if err != nil {
        c.JSON(http.StatusNotFound, httperr.NewHTTPError(
            http.StatusNotFound,
            "not_found",
            err.Error(),
        ))
        return
    }

    c.JSON(http.StatusOK, product)
}
```

### 6. Define Routes (`routes.go`)

Register your endpoints:

```go
package products

import (
    "github.com/gin-gonic/gin"

    "github.com/moasq/go-b2b-starter/internal/modules/auth"
    "github.com/moasq/go-b2b-starter/internal/platform/server/domain"
)

type Routes struct {
    handler *Handler
}

func NewRoutes(handler *Handler) *Routes {
    return &Routes{handler: handler}
}

func (r *Routes) Routes(router *gin.RouterGroup, resolver domain.MiddlewareResolver) {
    products := router.Group("/products")
    products.Use(
        resolver.Get("auth"),         // RequireAuth
        resolver.Get("org_context"),  // RequireOrganization
    )
    {
        products.POST("", r.handler.Create)
        products.GET("/:id", r.handler.GetByID)
        products.GET("", r.handler.List)
        products.PUT("/:id", r.handler.Update)
        products.DELETE("/:id", r.handler.Delete)
    }
}
```

### 7. Wire Dependencies (`cmd/init.go`)

Set up dependency injection:

```go
package cmd

import (
    "fmt"

    "go.uber.org/dig"

    "github.com/moasq/go-b2b-starter/internal/modules/products"
    "github.com/moasq/go-b2b-starter/internal/modules/products/app/services"
    "github.com/moasq/go-b2b-starter/internal/modules/products/domain"
    "github.com/moasq/go-b2b-starter/internal/modules/products/infra/repositories"
    sqlc "github.com/moasq/go-b2b-starter/internal/db/postgres/sqlc/gen"
)

func RegisterDependencies(container *dig.Container) error {
    // Repository - registered in internal/db/inject.go
    // (See step 8 below)

    // Service
    if err := container.Provide(services.NewProductService); err != nil {
        return fmt.Errorf("failed to provide product service: %w", err)
    }

    // Handler
    if err := container.Provide(products.NewHandler); err != nil {
        return fmt.Errorf("failed to provide product handler: %w", err)
    }

    // Routes
    if err := container.Provide(products.NewRoutes); err != nil {
        return fmt.Errorf("failed to provide product routes: %w", err)
    }

    return nil
}
```

### 8. Register Repository in Database Layer

**IMPORTANT**: Repositories are registered in `internal/db/inject.go`, not in the module's `cmd/init.go`.

Add to `internal/db/inject.go`:

```go
import (
    productDomain "github.com/moasq/go-b2b-starter/internal/modules/products/domain"
    productRepos "github.com/moasq/go-b2b-starter/internal/modules/products/infra/repositories"
)

// In registerDomainStores function:
if err := container.Provide(func(sqlcStore sqlc.Store) productDomain.ProductRepository {
    return productRepos.NewProductRepository(sqlcStore)
}); err != nil {
    return fmt.Errorf("failed to provide product repository: %w", err)
}
```

### 9. Register in Bootstrap

Add your module to `internal/bootstrap/init_mods.go`:

```go
import productsCmd "github.com/moasq/go-b2b-starter/internal/modules/products/cmd"

func InitMods(container *dig.Container) error {
    // ... existing modules ...

    // Products module
    if err := productsCmd.RegisterDependencies(container); err != nil {
        return fmt.Errorf("failed to register products dependencies: %w", err)
    }

    return nil
}
```

### 10. Register Routes in API

Routes are auto-registered via DI. Ensure your module's Routes struct is provided in step 7.

The `internal/api/provider.go` will automatically discover and register all route groups.

## Database Setup

### 1. Create Migration

Create a migration for your new table:

```bash
cd internal/db/postgres/sqlc/migrations
# Create files manually with next sequence number
```

**Up migration** (`000015_create_products.up.sql`):

```sql
CREATE TABLE app.products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    organization_id INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_organization
        FOREIGN KEY (organization_id)
        REFERENCES app.organizations(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_products_organization_id ON app.products(organization_id);
CREATE INDEX idx_products_name ON app.products(name);
```

**Down migration** (`000015_create_products.down.sql`):

```sql
DROP TABLE IF EXISTS app.products;
```

### 2. Create SQLC Queries

Create `internal/db/postgres/sqlc/query/products.sql`:

```sql
-- name: CreateProduct :one
INSERT INTO products (name, description, price, organization_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetProductByID :one
SELECT * FROM products
WHERE organization_id = $1 AND id = $2;

-- name: ListProductsByOrganization :many
SELECT * FROM products
WHERE organization_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateProduct :one
UPDATE products
SET name = $2, description = $3, price = $4, updated_at = NOW()
WHERE organization_id = $1 AND id = $5
RETURNING *;

-- name: DeleteProduct :exec
DELETE FROM products
WHERE organization_id = $1 AND id = $2;
```

### 3. Run Migrations and Generate Code

```bash
make migrateup  # Apply migrations
make sqlc       # Generate type-safe Go code
```

## Testing

### Unit Test Example

```go
package services_test

import (
    "context"
    "testing"

    "github.com/moasq/go-b2b-starter/internal/modules/products/domain"
    "github.com/moasq/go-b2b-starter/internal/modules/products/app/services"
)

type mockProductRepository struct {
    createFunc func(ctx context.Context, p *domain.Product) (*domain.Product, error)
}

func (m *mockProductRepository) Create(ctx context.Context, p *domain.Product) (*domain.Product, error) {
    return m.createFunc(ctx, p)
}

func TestProductService_Create(t *testing.T) {
    mockRepo := &mockProductRepository{
        createFunc: func(ctx context.Context, p *domain.Product) (*domain.Product, error) {
            p.ID = 1
            return p, nil
        },
    }

    service := services.NewProductService(mockRepo)

    req := &services.CreateProductRequest{
        Name:  "Test Product",
        Price: 99.99,
    }

    product, err := service.Create(context.Background(), 1, req)
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }

    if product.ID != 1 {
        t.Errorf("expected product ID 1, got %d", product.ID)
    }
}
```

## Best Practices

1. **Domain Layer is Pure**: No external dependencies in `domain/`
2. **Interfaces in Domain**: Repository interfaces defined where they're used
3. **Services Return Domain Types**: Not database types
4. **Handlers Are Thin**: Validation, auth, delegate to service
5. **Context First**: Always pass `context.Context` as first parameter
6. **Use `httperr.HTTPError`**: For consistent API error responses
7. **Register Repositories Centrally**: In `internal/db/inject.go`, not module init
8. **Map Database Types**: Convert SQLC types to domain types in repositories

## Common Pitfalls

### Import Cycles

```go
// ❌ Bad - Creates import cycle
// internal/modules/products/domain/entity.go
import "github.com/moasq/go-b2b-starter/internal/modules/products/app/services"

// ✅ Good - Domain has no dependencies
// internal/modules/products/domain/entity.go
package domain
```

### Wrong Repository Registration

```go
// ❌ Bad - Registering repository in module init
// internal/modules/products/cmd/init.go
container.Provide(repositories.NewProductRepository)

// ✅ Good - Register in database layer
// internal/db/inject.go
container.Provide(func(sqlcStore sqlc.Store) productDomain.ProductRepository {
    return productRepos.NewProductRepository(sqlcStore)
})
```

### Exposing SQLC Types

```go
// ❌ Bad - Service returns SQLC types
func (s *service) GetProduct(ctx context.Context, id int32) (*sqlc.Product, error)

// ✅ Good - Service returns domain types
func (s *service) GetProduct(ctx context.Context, id int32) (*domain.Product, error)
```

## File Checklist

After creating a new module, you should have:

- [ ] `domain/entity.go` - Domain entities
- [ ] `domain/repository.go` - Repository interfaces
- [ ] `infra/repositories/{entity}_repository.go` - Repository implementation
- [ ] `app/services/{entity}_service.go` - Service interface and implementation
- [ ] `handler.go` - HTTP handlers with Swagger docs
- [ ] `routes.go` - Route registration
- [ ] `cmd/init.go` - DI setup for service, handler, routes
- [ ] SQL queries in `internal/db/postgres/sqlc/query/{entity}.sql`
- [ ] Repository registration in `internal/db/inject.go`
- [ ] Module registration in `internal/bootstrap/init_mods.go`
- [ ] Database migrations in `internal/db/postgres/sqlc/migrations/`

## Next Steps

- **Architecture Details**: See [Architecture Guide](./01-architecture.md)
- **Database Operations**: See [Database Guide](./database.md)
- **API Development**: See [API Development Guide](./api-development.md)
- **Testing Guide**: See [Testing Guide](./testing.md)
