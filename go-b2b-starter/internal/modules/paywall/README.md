# Paywall Middleware Package

Provider-agnostic access gating middleware for B2B SaaS applications. This package provides the **"Payment Bouncer"** - checking if an organization has an active subscription before allowing access to protected features.

## Key Concept: Separation of Concerns

This package ONLY handles **access gating** (the "can they use this feature?" question). It does NOT manage subscriptions directly.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          PAYWALL (This Package)                             │
│                                                                             │
│  "Can this organization access premium features right now?"                 │
│                                                                             │
│  - Reads subscription status from LOCAL DATABASE                            │
│  - Makes NO external API calls (fast, reliable)                             │
│  - Returns 402 if payment required                                          │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                          BILLING MODULE (app/billing)                       │
│                                                                             │
│  "Manage subscription lifecycle via webhooks"                               │
│                                                                             │
│  - Processes Polar.sh webhooks (subscription created, updated, canceled)    │
│  - Updates LOCAL DATABASE with subscription state                           │
│  - Tracks quotas and usage                                                  │
│  - No direct API calls during request handling                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Architecture: Event-Driven Integration

The paywall middleware and billing module are decoupled via **event-driven architecture**:

```
┌─────────────┐   webhook   ┌─────────────────┐   writes   ┌─────────────┐
│  Polar.sh   │ ─────────►  │ Billing Module  │ ─────────► │  Local DB   │
└─────────────┘             └─────────────────┘            └─────────────┘
                                                                  │
                                                           reads  │
                                                                  ▼
┌─────────────┐             ┌─────────────────┐            ┌─────────────┐
│   Request   │ ─────────►  │    Paywall      │ ◄───────── │  Local DB   │
└─────────────┘             └─────────────────┘            └─────────────┘
                                    │
                                    ▼
                            Pass (200) or Block (402)
```

**Why Event-Driven?**
- No external API calls during request handling (fast responses)
- Billing provider outage doesn't block your users
- Clean separation between access control and subscription management
- Easy to swap billing providers without touching access logic

## Request Flow

```
HTTP Request
     │
     ▼
┌────────────────┐
│ Auth Middleware│ ── Verify JWT, extract Identity
└────────────────┘
     │
     ▼
┌────────────────┐
│ Org Middleware │ ── Resolve OrganizationID from Identity
└────────────────┘
     │
     ▼
┌────────────────┐
│Paywall Middleware│ ── Check subscription status (LOCAL DB)
└────────────────┘
     │
     ├── Active? ──► Handler ──► 200 OK
     │
     └── Inactive? ──► 402 Payment Required
```

## Usage

### 1. Setup (Already configured in init_mods.go)

```go
import (
    "github.com/moasq/go-b2b-starter/pkg/paywall"
)

// Setup middleware (after billing module is initialized)
if err := paywall.SetupMiddleware(container); err != nil {
    panic(err)
}

// Register named middlewares for route configuration
if err := paywall.RegisterNamedMiddlewares(container); err != nil {
    panic(err)
}
```

### 2. Protecting Routes

**Using Named Middleware (Recommended):**

```go
// In routes.go
func (r *Routes) Routes(router *gin.RouterGroup, resolver serverDomain.MiddlewareResolver) {
    // Premium features - require active subscription
    premiumGroup := router.Group("/premium")
    premiumGroup.Use(
        resolver.Get("auth"),           // Verify JWT
        resolver.Get("org_context"),    // Resolve org/account IDs
        resolver.Get("paywall"),        // Block if no active subscription
    )
    {
        premiumGroup.POST("/ai/generate", r.handler.Generate)
        premiumGroup.POST("/reports/export", r.handler.Export)
    }

    // Basic features - auth only, no paywall
    basicGroup := router.Group("/basic")
    basicGroup.Use(
        resolver.Get("auth"),
        resolver.Get("org_context"),
        // No paywall - allow access to fix billing issues
    )
    {
        basicGroup.GET("/billing/status", r.handler.GetBillingStatus)
        basicGroup.POST("/billing/portal", r.handler.CreatePortalSession)
    }
}
```

**Direct Middleware Usage:**

```go
// Get middleware from DI container
var paywallMiddleware *paywall.Middleware
container.Invoke(func(m *paywall.Middleware) {
    paywallMiddleware = m
})

// Apply to routes
router.Use(paywallMiddleware.RequireActiveSubscription())
```

### 3. Accessing Subscription Status in Handlers

```go
func (h *Handler) MyHandler(c *gin.Context) {
    // Safe get - returns nil if not set
    status := paywall.GetSubscriptionStatus(c)
    if status != nil {
        log.Printf("Org %d status: %s", status.OrganizationID, status.Status)
    }

    // Quick boolean check
    if paywall.IsSubscriptionActive(c) {
        // Show premium features
    }

    // Must get - panics if not set (use after RequireActiveSubscription)
    status := paywall.MustGetSubscriptionStatus(c)
}
```

### 4. Optional Status Check (No Blocking)

When you want to know status without blocking access:

```go
// Show "upgrade" prompts to free users
dashboardGroup.Use(
    resolver.Get("auth"),
    resolver.Get("org_context"),
    resolver.Get("paywall_optional"),  // Sets status, doesn't block
)
```

## Configuration

```go
config := &paywall.MiddlewareConfig{
    // URL included in 402 responses for upgrading
    UpgradeURL: "/billing",

    // Allow trialing subscriptions (default: true)
    AllowTrialing: true,

    // Custom error handler (optional)
    ErrorHandler: func(c *gin.Context, statusCode int, response *paywall.ErrorResponse) {
        c.JSON(statusCode, response)
    },
}

middleware := paywall.NewMiddleware(provider, config)
```

## Subscription Status Mapping

| DB Status     | IsActive | HTTP Response        |
|---------------|----------|----------------------|
| `active`      | true     | Pass through         |
| `trialing`    | true     | Pass through         |
| `past_due`    | false    | 402 Payment Required |
| `canceled`    | false    | 402 Payment Required |
| `unpaid`      | false    | 402 Payment Required |
| No subscription | false  | 402 Payment Required |

## Error Response Format

```json
{
    "error": "subscription_required",
    "message": "An active subscription is required to access this feature",
    "upgrade_url": "/billing",
    "status": "past_due"
}
```

## The "Swiss Cheese" Strategy

Not all routes should require a subscription. Allow users to fix billing issues:

**Protected (require paywall):**
- AI/ML features
- OCR processing
- Report generation
- Premium API endpoints
- Advanced analytics

**Unprotected (auth only):**
- Billing status and portal
- Account settings
- Profile management
- Subscription upgrade flow
- Webhooks
- Public pages

## Payment Verification Flow ("Verification on Redirect")

When users complete payment on Polar, they're redirected back to your app with a checkout session ID. To provide instant access (instead of waiting for webhooks), use the **verification endpoint**.

### Frontend Integration

**1. Configure Polar Success URL:**

Set your Polar checkout success URL to:
```
https://yoursaas.com/payment/success?session_id={CHECKOUT_SESSION_ID}
```

Polar will replace `{CHECKOUT_SESSION_ID}` with the actual session ID.

**2. Handle Redirect (Frontend):**

```javascript
// On /payment/success page
const params = new URLSearchParams(window.location.search);
const sessionId = params.get('session_id');

if (sessionId) {
    // Show loading spinner
    const response = await fetch('/api/subscriptions/verify-payment', {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ session_id: sessionId })
    });

    if (response.ok) {
        // Payment verified! Redirect to dashboard
        window.location.href = '/dashboard';
    } else {
        // Show error message
        const error = await response.json();
        showError(error.message);
    }
}
```

**3. Expected Response:**

```json
{
    "organization_id": 123,
    "has_active_subscription": true,
    "can_process_invoices": true,
    "invoice_count": 100,
    "reason": "Payment verified successfully",
    "checked_at": "2025-12-12T10:30:00Z"
}
```

### Backend Verification Endpoint

Endpoint: `POST /api/subscriptions/verify-payment`

**Request:**
```json
{
    "session_id": "cs_test_xxx"
}
```

**Flow:**
1. Fetch checkout session from Polar API
2. Verify status is "succeeded"
3. Extract customer_id and map to organization
4. Fetch full subscription details from Polar
5. Update local database (subscription + quota)
6. Return updated billing status

**Error Codes:**
- `400` - Checkout session failed or expired
- `404` - Checkout session not found
- `500` - Internal error (failed to sync)

### Lazy Guarding (Renewal Webhooks)

For monthly renewals, the middleware includes **lazy guarding** to handle missed webhooks:

**How It Works:**

1. User makes request → Middleware checks DB status
2. If DB says "expired" BUT subscription exists:
   - Middleware calls Polar API to refresh status
   - If Polar says "active" → Grant access and update DB
   - If Polar says "inactive" → Block access (truly expired)

```go
// Automatic in paywall middleware - no configuration needed
if !status.IsActive && status.Status != StatusNone {
    freshStatus, err := provider.RefreshSubscriptionStatus(ctx, orgID)
    if err == nil && freshStatus.IsActive {
        // Webhook was missed! Allow access
        status = freshStatus
    }
}
```

**Benefits:**
- Self-healing: Missed webhooks don't lock out paying users
- Fast: Only checks API when DB says inactive (edge case)
- Reliable: Users always get access if they've paid

### Architecture Overview

| Scenario | Primary Mechanism | Fallback Mechanism |
|----------|-------------------|-------------------|
| **User Just Paid** | **Verification on Redirect** (Frontend → Backend → Polar API) | Webhook (processed if arrives later) |
| **Monthly Renewal** | **Webhooks** (Standard processing) | **Lazy Guarding** (Middleware checks API if DB says expired) |

**Success Metrics:**
- Payment to access time: ~5 seconds (verification) vs. ~minutes (webhook only)
- Webhook miss recovery: Automatic via lazy guarding
- User complaints: Eliminated "I paid but still locked out" issues

## Implementing SubscriptionStatusProvider

The middleware requires a `SubscriptionStatusProvider` implementation. This is provided by the billing module:

```go
// Interface defined in this package
type SubscriptionStatusProvider interface {
    GetSubscriptionStatus(ctx context.Context, organizationID int32) (*SubscriptionStatus, error)
}

// Implemented in app/billing/infra/adapters/status_provider.go
type StatusProviderAdapter struct {
    service services.BillingService
}

func (a *StatusProviderAdapter) GetSubscriptionStatus(ctx context.Context, orgID int32) (*paywall.SubscriptionStatus, error) {
    billingStatus, err := a.service.GetBillingStatus(ctx, orgID)
    if err != nil {
        return nil, err
    }
    return &paywall.SubscriptionStatus{
        OrganizationID: orgID,
        Status:         billingStatus.SubscriptionStatus,
        IsActive:       billingStatus.HasActiveSubscription,
        // ...
    }, nil
}
```

## Named Middleware Reference

| Name                  | Function                    | Description                    |
|-----------------------|-----------------------------|--------------------------------|
| `paywall`             | RequireActiveSubscription   | Block if no active subscription|
| `paywall_optional`    | OptionalSubscriptionStatus  | Set status, don't block        |
| `subscription` (legacy)| RequireActiveSubscription  | Deprecated, use `paywall`      |

## Files

```
src/pkg/paywall/
├── subscription.go    # Core types and SubscriptionStatusProvider interface
├── middleware.go      # Gin middleware (RequireActiveSubscription)
├── context.go         # Context helpers (Get/Set SubscriptionStatus)
├── errors.go          # Error types (ErrNoSubscription, etc.)
├── provider.go        # DI registration and named middleware
└── README.md          # This file
```

## Related: Billing Module

See `app/billing/README.md` for:
- Polar.sh webhook integration
- Subscription lifecycle management
- Quota tracking
- Event-driven architecture details
