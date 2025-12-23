# Database Guide

The database layer uses PostgreSQL with SQLC for type-safe SQL operations. Domain modules define repository interfaces in their `domain/` layer, which are implemented by repositories in the `infra/` layer using SQLC.

## Architecture

The database layer follows the **Repository Pattern**:

**1. Domain Interfaces** (`internal/modules/{module}/domain/repository.go`) - Repository contracts defined by the domain
**2. Repository Implementations** (`internal/modules/{module}/infra/repositories/`) - Implement interfaces using SQLC
**3. SQLC Generated Code** (`internal/db/postgres/sqlc/gen/`) - Auto-generated type-safe queries
**4. DI Registration** (`internal/db/inject.go`) - Wire repositories to domain interfaces

### Why This Pattern?

- **Dependency Inversion** - Domain defines what it needs, infrastructure provides it
- **SQLC Isolation** - SQLC types never leak out of the `infra/` layer
- **Easy Testing** - Mock domain interfaces, not SQLC
- **Clean Boundaries** - Domain stays pure, no database knowledge
- **Type Safety** - SQLC generates type-safe Go code from SQL

### Legacy Note

> **Note**: Earlier versions used an `adapters/` and `adapter_impl/` pattern. This has been phased out in favor of the simpler repository pattern where domain interfaces are implemented directly by repository classes.

## SQLC Workflow

### 1. Write SQL Query

Create queries in `internal/db/postgres/sqlc/query/{domain}.sql`:

```sql
-- name: GetDocumentByID :one
SELECT * FROM documents
WHERE organization_id = $1 AND id = $2;

-- name: CreateDocument :one
INSERT INTO documents (organization_id, title, file_path, status)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListDocuments :many
SELECT * FROM documents
WHERE organization_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
```

**SQLC Annotations:**
- `:one` - Returns single row
- `:many` - Returns slice of rows
- `:exec` - Returns error only (no data)

### 2. Generate Code

```bash
make sqlc
```

Generates Go code in `internal/db/postgres/sqlc/gen/`.

**Never edit generated files** - they are regenerated on every run.

### 3. Define Repository Interface in Domain

Define interface in `internal/modules/documents/domain/repository.go`:

```go
package domain

import "context"

type DocumentRepository interface {
    GetByID(ctx context.Context, orgID, docID int32) (*Document, error)
    Create(ctx context.Context, doc *Document) (*Document, error)
    List(ctx context.Context, orgID int32, limit, offset int32) ([]*Document, error)
    Update(ctx context.Context, doc *Document) error
    Delete(ctx context.Context, orgID, docID int32) error
}
```

**Key Points:**
- Interface uses **domain types** (`*Document`), not SQLC types
- Defined where it's used (in the domain layer)
- Independent of implementation details

### 4. Implement Repository

Create repository in `internal/modules/documents/infra/repositories/document_repository.go`:

```go
package repositories

import (
    "context"
    "fmt"

    "github.com/moasq/go-b2b-starter/internal/modules/documents/domain"
    sqlc "github.com/moasq/go-b2b-starter/internal/db/postgres/sqlc/gen"
)

type documentRepository struct {
    store sqlc.Store
}

func NewDocumentRepository(store sqlc.Store) domain.DocumentRepository {
    return &documentRepository{store: store}
}

func (r *documentRepository) GetByID(ctx context.Context, orgID, docID int32) (*domain.Document, error) {
    // Call SQLC-generated method
    dbDoc, err := r.store.GetDocumentByID(ctx, sqlc.GetDocumentByIDParams{
        OrganizationID: orgID,
        ID:             docID,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to get document: %w", err)
    }

    // Map SQLC type to domain type
    return &domain.Document{
        ID:             dbDoc.ID,
        OrganizationID: dbDoc.OrganizationID,
        Title:          dbDoc.Title,
        FilePath:       dbDoc.FilePath,
        Status:         dbDoc.Status,
        CreatedAt:      dbDoc.CreatedAt,
        UpdatedAt:      dbDoc.UpdatedAt,
    }, nil
}

func (r *documentRepository) Create(ctx context.Context, doc *domain.Document) (*domain.Document, error) {
    dbDoc, err := r.store.CreateDocument(ctx, sqlc.CreateDocumentParams{
        OrganizationID: doc.OrganizationID,
        Title:          doc.Title,
        FilePath:       doc.FilePath,
        Status:         doc.Status,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create document: %w", err)
    }

    // Map back to domain
    return &domain.Document{
        ID:             dbDoc.ID,
        OrganizationID: dbDoc.OrganizationID,
        Title:          dbDoc.Title,
        FilePath:       dbDoc.FilePath,
        Status:         dbDoc.Status,
        CreatedAt:      dbDoc.CreatedAt,
        UpdatedAt:      dbDoc.UpdatedAt,
    }, nil
}
```

**Why Map Types?**
- SQLC types (`sqlc.Document`) are generated and may change
- Domain types (`domain.Document`) are stable and business-focused
- Mapping keeps SQLC isolated in the infrastructure layer

### 5. Register in DI

Add to `internal/db/inject.go`:

```go
import (
    documentDomain "github.com/moasq/go-b2b-starter/internal/modules/documents/domain"
    documentRepos "github.com/moasq/go-b2b-starter/internal/modules/documents/infra/repositories"
)

// In registerDomainStores function:
if err := container.Provide(func(sqlcStore sqlc.Store) documentDomain.DocumentRepository {
    return documentRepos.NewDocumentRepository(sqlcStore)
}); err != nil {
    return fmt.Errorf("failed to provide document repository: %w", err)
}
```

**Why Centralize DI?**
- All repository registrations in one place
- Easy to see all database dependencies
- Consistent pattern across modules

## Database Migrations

### File Structure

Migrations live in `internal/db/postgres/sqlc/migrations/`:

```
000001_create_schema.up.sql
000001_create_schema.down.sql
000002_add_indexes.up.sql
000002_add_indexes.down.sql
```

### Naming Convention

Format: `{6-digit-number}_{description}.{up|down}.sql`

- `.up.sql` - Apply the migration
- `.down.sql` - Rollback the migration

### Example Migration

**Up migration** (`000005_create_documents.up.sql`):

```sql
CREATE SCHEMA IF NOT EXISTS app;

CREATE TABLE app.documents (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    file_path VARCHAR(512) NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_organization
        FOREIGN KEY (organization_id)
        REFERENCES app.organizations(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_documents_org_id ON app.documents(organization_id);
CREATE INDEX idx_documents_status ON app.documents(status);
```

**Down migration** (`000005_create_documents.down.sql`):

```sql
DROP TABLE IF EXISTS app.documents;
```

### Running Migrations

```bash
make migrateup      # Apply all pending migrations
make migratedown    # Rollback last migration
```

## Type Conversions

PostgreSQL types need conversion to Go types.

### Nullable Fields

SQLC uses `pgtype` for nullable fields:

```go
// Convert pgtype.Text to string
str := postgres.StringFromPgText(dbRecord.NullableField)

// Convert string to pgtype.Text
pgText := postgres.ToPgText(str)

// Convert pgtype.Int4 to int32
num := postgres.Int32FromPgInt4(dbRecord.NullableInt)
```

Helper functions in `internal/db/postgres/types_transform.go`.

### JSONB Fields

```go
// Convert map to JSONB
jsonbData := postgres.ToJSONB(map[string]any{"key": "value"})

// Convert JSONB to map
data := postgres.JSONBToMap(dbRecord.Metadata)
```

## Error Handling

The database layer provides specific error types in `internal/db/core/errors.go`:

**Common Errors:**
- `ErrNoRows` - Query returned no results
- `ErrTxClosed` - Transaction already committed/rolled back
- `ErrTimeout` - Operation exceeded timeout
- `ErrPoolClosed` - Connection pool is closed

**Helper Functions:**

```go
if core.IsNoRowsError(err) {
    return domain.ErrDocumentNotFound
}

if core.IsConstraintError(err, "unique_title") {
    return domain.ErrDocumentAlreadyExists
}

if core.IsTimeoutError(err) {
    return domain.ErrDatabaseTimeout
}
```

## Transactions

Use transactions for multi-step operations that must be atomic.

### Basic Transaction

```go
func (r *repository) CreateWithRelation(ctx context.Context, doc *domain.Document) error {
    return r.db.WithTx(ctx, func(tx core.Transaction) error {
        // Step 1: Create document
        created, err := tx.CreateDocument(ctx, params)
        if err != nil {
            return err
        }

        // Step 2: Create embeddings
        _, err = tx.CreateEmbedding(ctx, embeddingParams)
        if err != nil {
            return err // Transaction auto-rolls back on error
        }

        return nil // Transaction commits on success
    })
}
```

### Transaction Options

```go
// Read-only transaction
err := r.db.WithTxOptions(ctx, &sql.TxOptions{ReadOnly: true}, func(tx core.Transaction) error {
    // Read operations only
})

// Custom isolation level
err := r.db.WithTxOptions(ctx, &sql.TxOptions{
    Isolation: sql.LevelSerializable,
}, func(tx core.Transaction) error {
    // Operations
})
```

## Best Practices

### Always Use Context

```go
// ✅ Good
func (r *repository) GetDocument(ctx context.Context, id int32) (*Document, error)

// ❌ Bad
func (r *repository) GetDocument(id int32) (*Document, error)
```

### Handle Errors Appropriately

```go
// ✅ Convert database errors to domain errors
doc, err := r.store.GetDocumentByID(ctx, params)
if err != nil {
    if core.IsNoRowsError(err) {
        return nil, domain.ErrDocumentNotFound
    }
    return nil, fmt.Errorf("failed to get document: %w", err)
}
```

### Use Prepared Statements

SQLC automatically creates prepared statements. Never concatenate SQL strings.

```go
// ✅ Good (SQLC handles this)
SELECT * FROM documents WHERE title = $1

// ❌ Bad (SQL injection risk)
query := fmt.Sprintf("SELECT * FROM documents WHERE title = '%s'", title)
```

### Indexes for Performance

Add indexes for commonly queried fields:

```sql
-- Foreign keys
CREATE INDEX idx_documents_org_id ON documents(organization_id);

-- Status fields
CREATE INDEX idx_documents_status ON documents(status);

-- Timestamps for sorting
CREATE INDEX idx_documents_created_at ON documents(created_at DESC);

-- Composite indexes for multi-column queries
CREATE INDEX idx_documents_org_status ON documents(organization_id, status);
```

### Map SQLC Types to Domain Types

Always convert SQLC types to domain types in the repository layer:

```go
// ✅ Good - Repository returns domain types
func (r *repository) GetDocument(ctx context.Context, id int32) (*domain.Document, error) {
    dbDoc, err := r.store.GetDocumentByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // Map SQLC type to domain type
    return &domain.Document{
        ID:    dbDoc.ID,
        Title: dbDoc.Title,
        // ... other fields
    }, nil
}

// ❌ Bad - Service receives SQLC types
func (s *service) GetDocument(ctx context.Context, id int32) (*sqlc.Document, error)
```

## Complete Example: Adding a New Entity

Let's add a `Comment` entity to the documents module:

### 1. Write Migration

`internal/db/postgres/sqlc/migrations/000010_create_comments.up.sql`:
```sql
CREATE TABLE app.comments (
    id SERIAL PRIMARY KEY,
    document_id INTEGER NOT NULL,
    author_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_document
        FOREIGN KEY (document_id)
        REFERENCES app.documents(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_comments_document_id ON app.comments(document_id);
```

### 2. Write SQLC Queries

`internal/db/postgres/sqlc/query/comments.sql`:
```sql
-- name: CreateComment :one
INSERT INTO comments (document_id, author_id, content)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListCommentsByDocument :many
SELECT * FROM comments
WHERE document_id = $1
ORDER BY created_at ASC;
```

### 3. Generate SQLC Code

```bash
make migrateup
make sqlc
```

### 4. Define Domain Interface

`internal/modules/documents/domain/repository.go`:
```go
type CommentRepository interface {
    Create(ctx context.Context, comment *Comment) (*Comment, error)
    ListByDocument(ctx context.Context, docID int32) ([]*Comment, error)
}
```

### 5. Implement Repository

`internal/modules/documents/infra/repositories/comment_repository.go`:
```go
package repositories

import (
    "context"
    "github.com/moasq/go-b2b-starter/internal/modules/documents/domain"
    sqlc "github.com/moasq/go-b2b-starter/internal/db/postgres/sqlc/gen"
)

type commentRepository struct {
    store sqlc.Store
}

func NewCommentRepository(store sqlc.Store) domain.CommentRepository {
    return &commentRepository{store: store}
}

func (r *commentRepository) Create(ctx context.Context, comment *domain.Comment) (*domain.Comment, error) {
    dbComment, err := r.store.CreateComment(ctx, sqlc.CreateCommentParams{
        DocumentID: comment.DocumentID,
        AuthorID:   comment.AuthorID,
        Content:    comment.Content,
    })
    if err != nil {
        return nil, err
    }

    return &domain.Comment{
        ID:         dbComment.ID,
        DocumentID: dbComment.DocumentID,
        AuthorID:   dbComment.AuthorID,
        Content:    dbComment.Content,
        CreatedAt:  dbComment.CreatedAt,
    }, nil
}
```

### 6. Register in DI

`internal/db/inject.go`:
```go
if err := container.Provide(func(sqlcStore sqlc.Store) documentDomain.CommentRepository {
    return documentRepos.NewCommentRepository(sqlcStore)
}); err != nil {
    return fmt.Errorf("failed to provide comment repository: %w", err)
}
```

## File Locations

| Component | Path |
|-----------|------|
| Domain interfaces | `internal/modules/{module}/domain/repository.go` |
| Repository implementations | `internal/modules/{module}/infra/repositories/` |
| SQL queries | `internal/db/postgres/sqlc/query/` |
| Migrations | `internal/db/postgres/sqlc/migrations/` |
| Generated code | `internal/db/postgres/sqlc/gen/` |
| Type helpers | `internal/db/postgres/types_transform.go` |
| Error types | `internal/db/core/errors.go` |
| DI registration | `internal/db/inject.go` |

## Next Steps

- **Using repositories in services**: See [Architecture Guide](./architecture.md)
- **Building APIs**: See [API Development Guide](./api-development.md)
- **Adding a new module**: See [Adding a Module Guide](./02-adding-a-module.md)
- **SQLC documentation**: https://docs.sqlc.dev/
