# Database Layer Guide

This guide shows you how to work with the database layer using our adapter pattern. It's designed to be simple and practical.

## What's the Adapter Pattern?

We keep database code separate from business logic:
- **`adapters/`** - Interface definitions (what operations are available)
- **`postgres/adapter_impl/`** - Actual database code (how PostgreSQL does it)

Your business modules only need to know about `adapters/`. They never touch PostgreSQL directly.

## The Workflow

Here's the typical flow when you need to add something to the database:

### 1. Create a Database Migration

Use the Makefile to create migration files:

```bash
make create-migration MIGRATION_NAME=add_users_table
```

This creates two files in `postgres/sqlc/migrations/`:
- `000XXX_add_users_table.up.sql` (creates your changes)
- `000XXX_add_users_table.down.sql` (removes your changes)

Write your SQL in these files. Keep it simple.

### 2. Apply Your Migration

Run the migration to update your database:

```bash
make migrateup
```

If something goes wrong, you can rollback:

```bash
make migratedown
```

### 3. Write SQL Queries

Create a file in `postgres/sqlc/query/` with your queries. Use SQLC's comments to tell it what to generate:

```sql
-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (email, name) VALUES ($1, $2) RETURNING *;

-- name: UpdateUserBalance :exec
UPDATE users SET balance = balance + $1 WHERE id = $2;
```

The comment annotations tell SQLC what kind of method to create (`:one` returns single row, `:many` returns multiple, `:exec` runs without returning).

### 4. Generate Go Code

Let SQLC generate type-safe Go code from your queries:

```bash
make sqlc
```

This creates Go methods in `postgres/sqlc/gen/` that you can use safely without writing SQL in Go code.

### 5. Create an Adapter Interface

Define what operations your business logic needs in `adapters/user_adapter.go`:

```go
package adapters

import (
    "context"
    db "github.com/moasq/go-b2b-starter/pkg/db/postgres/sqlc/gen"
)

type UserAdapter interface {
    GetUserByID(ctx context.Context, id int32) (db.User, error)
    CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)

    // For transactions - aggregate multiple operations
    TransferMoney(ctx context.Context, fromUserID, toUserID int32, amount int32) error
}
```

### 6. Implement the Adapter

Create the actual implementation in `postgres/adapter_impl/user_adapter.go`:

```go
package adapterimpl

import (
    "context"
    "fmt"
    "github.com/moasq/go-b2b-starter/pkg/db/adapters"
    sqlc "github.com/moasq/go-b2b-starter/pkg/db/postgres/sqlc/gen"
)

type userAdapter struct {
    store sqlc.Store
}

func NewUserAdapter(store sqlc.Store) adapters.UserAdapter {
    return &userAdapter{store: store}
}

func (a *userAdapter) GetUserByID(ctx context.Context, id int32) (sqlc.User, error) {
    return a.store.GetUserByID(ctx, id)
}

func (a *userAdapter) CreateUser(ctx context.Context, arg sqlc.CreateUserParams) (sqlc.User, error) {
    return a.store.CreateUser(ctx, arg)
}

// TransferMoney - transaction example aggregating multiple queries
func (a *userAdapter) TransferMoney(ctx context.Context, fromUserID, toUserID int32, amount int32) error {
    // Start transaction using the store's ExecTx method
    return a.store.ExecTx(ctx, func(q sqlc.Querier) error {
        // All queries run within this transaction

        // 1. Deduct from sender
        err := q.UpdateUserBalance(ctx, sqlc.UpdateUserBalanceParams{
            ID:     fromUserID,
            Amount: -amount,
        })
        if err != nil {
            return fmt.Errorf("failed to deduct balance: %w", err)
        }

        // 2. Add to receiver
        err = q.UpdateUserBalance(ctx, sqlc.UpdateUserBalanceParams{
            ID:     toUserID,
            Amount: amount,
        })
        if err != nil {
            return fmt.Errorf("failed to add balance: %w", err)
        }

        // 3. Verify sender has enough balance
        sender, err := q.GetUserByID(ctx, fromUserID)
        if err != nil {
            return fmt.Errorf("failed to get sender: %w", err)
        }
        if sender.Balance < 0 {
            return fmt.Errorf("insufficient balance")
        }

        // If any query fails, transaction auto-rollbacks
        // If all succeed, transaction auto-commits
        return nil
    })
}
```

### 7. Register in Dependency Injection

Add your adapter to `inject.go` so it's available everywhere:

```go
if err := container.Provide(func(sqlcStore sqlc.Store) adapters.UserAdapter {
    return adapterImpl.NewUserAdapter(sqlcStore)
}); err != nil {
    return fmt.Errorf("failed to provide user adapter: %w", err)
}
```

### 8. Use in Your Module

Now your business modules can request the adapter through dependency injection:

```go
type UserService struct {
    userAdapter adapters.UserAdapter
}

func NewUserService(userAdapter adapters.UserAdapter) *UserService {
    return &UserService{userAdapter: userAdapter}
}

func (s *UserService) GetUser(ctx context.Context, id int32) (*User, error) {
    user, err := s.userAdapter.GetUserByID(ctx, id)
    if err != nil {
        return nil, err
    }
    // ... your business logic here
}

func (s *UserService) TransferFunds(ctx context.Context, fromID, toID int32, amount int32) error {
    // The adapter handles the transaction internally
    return s.userAdapter.TransferMoney(ctx, fromID, toID, amount)
}
```

## Working with Transactions

When you need multiple queries to succeed or fail together, add a transaction method to your adapter:

**Key Points:**
- Your module only depends on the adapter interface (never touches the database pool)
- The adapter implementation handles transaction logic using `store.ExecTx()`
- All queries inside `ExecTx()` run in a transaction
- If any query fails, everything rolls back automatically
- If all succeed, everything commits automatically

## Common Commands

```bash
make create-migration MIGRATION_NAME=your_change  # Create migration files
make migrateup                                    # Apply migrations
make migratedown                                  # Rollback last migration
make sqlc                                         # Generate Go code from SQL
make build                                        # Verify everything compiles
```

## Why This Pattern?

- **Clean separation** - modules only know about adapters, not databases
- **Easy to test** - mock the adapter interface, no database needed
- **Type safety** - SQLC catches SQL errors at compile time
- **Transaction safety** - adapters encapsulate transaction logic
- **Single dependency** - modules only inject adapters

That's it! The pattern keeps your business logic clean and focused.
