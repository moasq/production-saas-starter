# Auth Package

Provider-agnostic authentication and authorization with type-safe middleware. Supports JWT verification, RBAC permissions, and multi-tenant organization context.

## Quick Start

### Setup Middleware

```go
// In server initialization
authMiddleware := auth.NewMiddleware(authProvider, orgResolver, accResolver, nil)

// Apply to routes
router.Use(authMiddleware.RequireAuth())          // Verify JWT token
router.Use(authMiddleware.RequireOrganization())  // Resolve org/account IDs
```

### Protect Routes

```go
router.GET("/invoices",
    auth.RequirePermissionFunc("invoice", "view"),
    handler.ListInvoices)

router.POST("/invoices",
    auth.RequirePermissionFunc("invoice", "create"),
    handler.CreateInvoice)
```

### Get Context in Handlers

```go
func (h *Handler) MyHandler(c *gin.Context) {
    reqCtx := auth.GetRequestContext(c)

    orgID := reqCtx.OrganizationID      // int32 database ID
    accountID := reqCtx.AccountID       // int32 database ID
    email := reqCtx.Identity.Email      // User's email
}
```

## Core Concepts

- **Identity**: User info from auth provider (email, roles, permissions)
- **RequestContext**: Resolved database IDs (OrganizationID, AccountID)
- **Permissions**: Format `"resource:action"` (e.g., `"invoice:create"`, `"org:manage"`)

## Common Patterns

### Pattern 1: Public Route

No authentication required:

```go
router.POST("/auth/signup", handler.Signup)
router.GET("/auth/check-email", handler.CheckEmail)
```

### Pattern 2: Authenticated Route

Any logged-in user can access:

```go
router.GET("/profile/me",
    authMiddleware.RequireAuth(),
    handler.GetProfile)
```

### Pattern 3: Organization Route

Requires organization context (most common pattern):

```go
orgGroup := router.Group("/organizations")
orgGroup.Use(
    authMiddleware.RequireAuth(),
    authMiddleware.RequireOrganization(),
)
{
    orgGroup.GET("", handler.GetOrganization)
    orgGroup.PUT("", handler.UpdateOrganization)
    orgGroup.GET("/stats", handler.GetOrganizationStats)
}
```

### Pattern 4: Permission-Protected Route

Requires specific permission:

```go
router.POST("/invoices",
    authMiddleware.RequireAuth(),
    authMiddleware.RequireOrganization(),
    auth.RequirePermissionFunc("invoice", "create"),
    handler.CreateInvoice)

router.DELETE("/invoices/:id",
    authMiddleware.RequireAuth(),
    authMiddleware.RequireOrganization(),
    auth.RequirePermissionFunc("invoice", "delete"),
    handler.DeleteInvoice)
```

### Pattern 5: Role-Protected Route

Requires specific role:

```go
router.DELETE("/organizations/:id",
    authMiddleware.RequireAuth(),
    authMiddleware.RequireRole(auth.RoleAdmin),
    handler.DeleteOrganization)
```

## Stytch Project Setup

### Create Stytch Account & Project

1. Go to [https://stytch.com](https://stytch.com) and sign up
2. Create a new **B2B project**
3. Choose **"Test"** environment for development

### Get Your Credentials

From your Stytch project dashboard:

```env
STYTCH_PROJECT_ID=project-test-xxx-xxx    # Project Settings → Project ID
STYTCH_SECRET=secret-test-xxx             # API Keys → Secret
STYTCH_ENV=test                           # "test" or "live"
```

### Configure RBAC Policies

Go to **Dashboard → RBAC → Policies** and create these resources:

| Resource | Actions | Description |
|----------|---------|-------------|
| `resource` | `view`, `create`, `edit`, `delete`, `approve` | Your domain entity (rename to your business) |
| `org` | `view`, `manage` | Organization settings |

> **Tip**: Rename "resource" to your domain entity (e.g., `invoice`, `patient`, `project`).

### Set Up Roles

Go to **Dashboard → RBAC → Roles**:

- **stytch_admin** (built-in): Auto-assigned to organization creator, has all permissions
- **stytch_member** (built-in): Default role for all members
- **manager** (custom): Create for elevated access

| Role | Permissions |
|------|-------------|
| member | `resource:view`, `resource:create` |
| manager | All resource permissions + `org:view` |
| admin | All permissions |

Assign permissions to roles based on your business needs.

### Test Your Setup

```bash
# Add credentials to app.env
STYTCH_PROJECT_ID=project-test-xxx-xxx
STYTCH_SECRET=secret-test-xxx
STYTCH_ENV=test
STYTCH_SESSION_DURATION_MINUTES=1440  # Optional: 24 hours

# Run the application
make server

# First user to sign up becomes stytch_admin
# Subsequent users get stytch_member role
```

## Configuration

After completing setup above, your `app.env` should have:

```env
STYTCH_PROJECT_ID=project-test-xxx-xxx    # Required
STYTCH_SECRET=secret-test-xxx             # Required
STYTCH_ENV=test                           # Optional: "test" or "live"
STYTCH_SESSION_DURATION_MINUTES=1440      # Optional: 24 hours (default)
STYTCH_API_TIMEOUT=15s                   # Optional: 15 seconds (default)
```

## Adding New Permissions

**Step 1:** Define in `rbac.go`

```go
// In the permissions section of rbac.go
var (
    // Add your new permissions
    PermReportView   = NewPermission("report", "view")
    PermReportExport = NewPermission("report", "export")
)

// Don't forget to add to AllPermissions
var AllPermissions = []Permission{
    // ... existing permissions
    PermReportView,
    PermReportExport,
}
```

**Step 2:** Use in routes

```go
router.GET("/reports",
    auth.RequirePermissionFunc("report", "view"),
    handler.ListReports)

router.POST("/reports/export",
    auth.RequirePermissionFunc("report", "export"),
    handler.ExportReport)
```

**Step 3:** Configure in Stytch Dashboard
- Go to Stytch Dashboard → RBAC → Policies
- Add resource: `report`
- Add actions: `view`, `export`
- Assign to roles

## Common Handler Patterns

### Check Permission

```go
identity := auth.GetIdentity(c)
if identity.HasResourcePermission("invoice", "delete") {
    // Show delete button
}
```

### Check Role

```go
identity := auth.GetIdentity(c)
if identity.HasRole(auth.RoleAdmin) {
    // Show admin panel
}
```

### Get Organization ID

```go
// Safe: returns 0 if not set
orgID := auth.GetOrganizationID(c)

// Panics if not set (use only after RequireOrganization)
reqCtx := auth.MustGetRequestContext(c)
```

### Get Account ID

```go
// Safe: returns 0 if not set
accountID := auth.GetAccountID(c)
```

### Get Full Context

```go
reqCtx := auth.GetRequestContext(c)
if reqCtx == nil {
    // Handle missing context
    return
}

// Access all fields
orgID := reqCtx.OrganizationID
accountID := reqCtx.AccountID
email := reqCtx.Identity.Email
roles := reqCtx.Identity.Roles
permissions := reqCtx.Identity.Permissions
```

## Multiple Permission Checks

### Require Any Permission

At least one permission required:

```go
router.GET("/reports",
    auth.RequireAnyPermissionFunc(
        auth.PermReportView,
        auth.PermReportExport,
    ),
    handler.GetReports)
```

### Require All Permissions

All permissions required:

```go
router.POST("/admin/dangerous",
    authMiddleware.RequireAllPermissions(
        auth.NewPermission("admin", "access"),
        auth.NewPermission("admin", "write"),
    ),
    handler.DangerousOperation)
```

## Troubleshooting

**"authentication required"**
- Check `Authorization: Bearer <token>` header format
- Verify token not expired
- Check STYTCH_PROJECT_ID matches token issuer

**"organization not found"**
- Organization must exist in database before authentication
- Check provider org ID is mapped to database ID

**"insufficient permissions"**
- Verify user has required permission in Stytch RBAC dashboard
- Check permission format: `"resource:action"` (not `resource.action`)
- Check role-based permissions in `roles.go`

## Predefined Permissions

Generic permissions available as constants (rename "resource" to your domain):

```go
// Resource (rename to your domain: invoice, patient, project, etc.)
auth.PermResourceView     // "resource:view"
auth.PermResourceCreate   // "resource:create"
auth.PermResourceEdit     // "resource:edit"
auth.PermResourceDelete   // "resource:delete"
auth.PermResourceApprove  // "resource:approve"

// Organization
auth.PermOrgView          // "org:view"
auth.PermOrgManage        // "org:manage"

// See rbac.go for complete list and customization instructions
```

## Custom Middleware Example

Create custom auth middleware for specific needs:

```go
func RequireAdminOrOwner() gin.HandlerFunc {
    return func(c *gin.Context) {
        identity := auth.GetIdentity(c)
        if identity == nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
            c.Abort()
            return
        }

        if !identity.HasRole(auth.RoleAdmin) && !identity.HasRole(auth.RoleOwner) {
            c.JSON(http.StatusForbidden, gin.H{"error": "admin or owner required"})
            c.Abort()
            return
        }

        c.Next()
    }
}

// Usage
router.DELETE("/organizations/:id", RequireAdminOrOwner(), handler.Delete)
```

## Learn More

- **RBAC Definitions**: See `rbac.go` for all roles, permissions, and customization instructions
- **Stytch Setup**: See `STYTCH_SETUP.md` for dashboard configuration
- **API reference**: Run `go doc github.com/moasq/go-b2b-starter/pkg/auth`
- **Examples**: See `src/api/organizations/routes.go` for real-world usage
- **Stytch B2B RBAC**: https://stytch.com/docs/b2b/guides/rbac/overview
