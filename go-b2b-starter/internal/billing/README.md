# Billing Module

Hybrid subscription lifecycle management for B2B SaaS applications. This module combines **event-driven webhooks** with **active verification** for maximum reliability.

## Key Principle: Hybrid Synchronization Strategy

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                     HYBRID BILLING SYNC (This Module)                       │
│                                                                             │
│  "Primary: Webhooks + Fallback: Active Verification + Self-Healing"        │
│                                                                             │
│  1. VERIFICATION ON REDIRECT (Initial Payment):                             │
│     User pays → Frontend calls /verify-payment → Instant access            │
│                                                                             │
│  2. WEBHOOKS (Renewals):                                                    │
│     Polar.sh sends webhook → Billing module processes → DB updated         │
│                                                                             │
│  3. LAZY GUARDING (Missed Webhooks):                                        │
│     DB says expired → Middleware checks Polar API → Self-healing           │
│                                                                             │
│  Result: Fast, reliable, self-healing subscription management               │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Architecture

```
                                EXTERNAL
┌────────────────────────────────────────────────────────────────────────────┐
│                              Polar.sh                                      │
│                                                                            │
│   - Handles checkout, payment processing, subscription management          │
│   - Sends webhooks on state changes                                        │
└────────────────────────────────────────────────────────────────────────────┘
                    │
                    │ Webhooks (subscription.created, subscription.updated, etc.)
                    ▼
┌────────────────────────────────────────────────────────────────────────────┐
│                         BILLING MODULE                                     │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                            │
│  ┌──────────────────┐     ┌──────────────────┐     ┌──────────────────┐   │
│  │  Webhook Handler │ ──► │  BillingService  │ ──► │   Repository     │   │
│  │   (API Layer)    │     │  (Domain Logic)  │     │  (Infra Layer)   │   │
│  └──────────────────┘     └──────────────────┘     └──────────────────┘   │
│                                    │                        │              │
│                                    ▼                        ▼              │
│                           ┌──────────────────┐     ┌──────────────────┐   │
│                           │  Quota Tracking  │     │    Local DB      │   │
│                           └──────────────────┘     └──────────────────┘   │
│                                                                            │
└────────────────────────────────────────────────────────────────────────────┘
                    │
                    │ Reads subscription status
                    ▼
┌────────────────────────────────────────────────────────────────────────────┐
│                      PAYWALL MIDDLEWARE (pkg/paywall)                      │
│                                                                            │
│   - Reads from local DB (fast, no external calls)                          │
│   - Blocks requests if subscription inactive (402)                         │
│   - Provider-agnostic access gating                                        │
└────────────────────────────────────────────────────────────────────────────┘
```

## Synchronization Mechanisms

### 1. Verification on Redirect (Initial Payment)

**Use Case:** User completes payment and returns to app

**Problem:** Webhooks may not arrive immediately (delays, failures)

**Solution:** Frontend triggers backend verification

```
User Pays → Polar Redirects → Frontend → POST /verify-payment → Backend → Polar API → DB Updated → Instant Access
```

**Benefits:**
- ✅ Instant access (5 seconds vs. minutes)
- ✅ No webhook dependency
- ✅ User sees immediate result

**Implementation:**
- Endpoint: `POST /api/subscriptions/verify-payment`
- Service: `src/app/billing/app/services/verify_payment_service.go`
- Adapter: `src/app/billing/infra/polar/polar_adapter.go` → `GetCheckoutSession()`

### 2. Webhooks (Renewals & Updates)

**Use Case:** Monthly renewals, cancellations, plan changes

**Standard Flow:** Polar sends webhook → Backend processes → DB updated

**Supported Events:**
- `subscription.created`, `subscription.updated`, `subscription.canceled`
- `customer.updated`, `meter.grant.updated`

### 3. Lazy Guarding (Self-Healing)

**Use Case:** Webhook failed or delayed for renewal

**How It Works:**
1. User makes request
2. Middleware checks DB → Status: "expired"
3. Middleware calls Polar API to verify
4. If Polar says "active" → Grant access + Update DB
5. If Polar says "inactive" → Block access (truly expired)

**Code (Automatic in Middleware):**
```go
if !status.IsActive && status.Status != StatusNone {
    freshStatus, err := provider.RefreshSubscriptionStatus(ctx, orgID)
    if err == nil && freshStatus.IsActive {
        status = freshStatus  // Self-healed!
    }
}
```

**Benefits:**
- ✅ Self-healing: No manual intervention
- ✅ Fast: Only calls API in edge cases (<1% of requests)
- ✅ Reliable: Paying users never locked out

## Why Hybrid Approach?

| Scenario | Mechanism | Benefit |
|----------|-----------|---------|
| Initial Payment | Verification on Redirect | Instant access |
| Monthly Renewal | Webhooks | No user action needed |
| Missed Webhook | Lazy Guarding | Self-healing |
| Normal Requests | Database Read | Fast (no API calls) |

## Module Structure

```
src/app/billing/
├── domain/
│   ├── subscription.go      # Subscription entity
│   ├── quota.go             # Quota tracking entity
│   ├── billing_status.go    # Combined status for API responses
│   ├── repository.go        # Repository interfaces
│   ├── service.go           # Service interface
│   └── errors.go            # Domain errors
│
├── app/services/
│   ├── subscription_service_dec.go  # BillingService interface
│   ├── sync_service.go              # Sync subscription from Polar
│   ├── webhook_service.go           # Process webhook events
│   └── quota_service.go             # Quota management
│
├── infra/
│   ├── adapters/
│   │   └── status_provider.go       # Bridge to paywall middleware
│   ├── repositories/
│   │   ├── subscription_repository.go   # Subscription DB operations
│   │   └── organization_adapter.go      # Org ID lookups
│   └── polar/
│       └── polar_adapter.go         # Polar API client (webhook only)
│
└── cmd/
    └── init.go              # DI initialization
```

## Data Flow

### 1. Subscription Created (Webhook)

```
Polar.sh                    Billing Module                Local DB
    │                            │                            │
    │  subscription.created      │                            │
    │ ────────────────────────►  │                            │
    │                            │  UpsertSubscription()      │
    │                            │ ────────────────────────►  │
    │                            │                            │
    │                            │  UpsertQuota()             │
    │                            │ ────────────────────────►  │
    │                            │                            │
    │         200 OK             │                            │
    │ ◄────────────────────────  │                            │
```

### 2. User Accesses Premium Feature

```
User                        Paywall                    Local DB
  │                            │                          │
  │  GET /ai/generate          │                          │
  │ ────────────────────────►  │                          │
  │                            │  GetSubscriptionStatus() │
  │                            │ ──────────────────────►  │
  │                            │                          │
  │                            │  {status: "active"}      │
  │                            │ ◄──────────────────────  │
  │                            │                          │
  │         Pass through       │                          │
  │ ◄────────────────────────  │                          │
```

### 3. Quota Consumption (Invoice Processing)

```
User                      BillingService                Local DB
  │                            │                           │
  │  POST /invoices/process    │                           │
  │ ────────────────────────►  │                           │
  │                            │  DecrementInvoiceCount()  │
  │                            │ ──────────────────────►   │
  │                            │                           │
  │                            │  {remaining: 42}          │
  │                            │ ◄──────────────────────   │
  │                            │                           │
  │         200 OK             │                           │
  │ ◄────────────────────────  │                           │
```

## Key Components

### BillingService

```go
// BillingService handles subscription management and quota verification.
//
// This service manages the billing lifecycle with Polar.sh via event-driven webhooks.
// It does NOT expose direct API calls to Polar during request handling:
//
//  1. WEBHOOK PROCESSING (async, event-driven):
//     - subscription.created, subscription.updated, subscription.canceled
//     - Updates local database with subscription state
//
//  2. LOCAL DB QUERIES (sync, during requests):
//     - GetBillingStatus: Check subscription status from local DB
//     - GetQuotaStatus: Check quota limits from local DB
//
//  3. QUOTA CONSUMPTION (sync, during requests):
//     - ConsumeInvoiceQuota: Decrement invoice count in local DB
type BillingService interface {
    // Webhook processing (called by webhook handler)
    ProcessWebhookEvent(ctx context.Context, eventType string, payload map[string]any) error

    // Status queries (from local DB only)
    GetBillingStatus(ctx context.Context, organizationID int32) (*BillingStatus, error)
    CheckQuotaAvailability(ctx context.Context, organizationID int32) (*BillingStatus, error)

    // Quota consumption (local DB update)
    ConsumeInvoiceQuota(ctx context.Context, organizationID int32) (*BillingStatus, error)

    // NEW: Verification on Redirect (makes Polar API call)
    VerifyPaymentFromCheckout(ctx context.Context, sessionID string) (*BillingStatus, error)

    // NEW: Lazy Guarding (makes Polar API call when DB says expired)
    RefreshSubscriptionStatus(ctx context.Context, organizationID int32) (*BillingStatus, error)

    // Manual sync (for admin/debug - makes Polar API call)
    SyncSubscriptionFromPolar(ctx context.Context, organizationID int32) error
}
```

### StatusProviderAdapter

Bridges the billing module to the paywall middleware:

```go
// In app/billing/infra/adapters/status_provider.go
type StatusProviderAdapter struct {
    service services.BillingService
}

// Implements paywall.SubscriptionStatusProvider
func (a *StatusProviderAdapter) GetSubscriptionStatus(ctx context.Context, orgID int32) (*paywall.SubscriptionStatus, error) {
    billingStatus, err := a.service.GetBillingStatus(ctx, orgID)
    if err != nil {
        return nil, err
    }

    return &paywall.SubscriptionStatus{
        OrganizationID: orgID,
        Status:         billingStatus.SubscriptionStatus,
        IsActive:       billingStatus.HasActiveSubscription,
        // Maps billing status to access status
    }, nil
}

// NEW: Implements lazy guarding - refreshes from Polar API when DB says expired
func (a *StatusProviderAdapter) RefreshSubscriptionStatus(ctx context.Context, orgID int32) (*paywall.SubscriptionStatus, error) {
    billingStatus, err := a.service.RefreshSubscriptionStatus(ctx, orgID)
    if err != nil {
        return nil, err
    }

    return &paywall.SubscriptionStatus{
        OrganizationID: orgID,
        Status:         billingStatus.SubscriptionStatus,
        IsActive:       billingStatus.HasActiveSubscription,
    }, nil
}
```

### Webhook Events

Supported Polar.sh webhook events:

| Event | Description | Action |
|-------|-------------|--------|
| `subscription.created` | New subscription | Create/update subscription + quota |
| `subscription.updated` | Status change | Update subscription status |
| `subscription.canceled` | Subscription canceled | Mark as canceled |
| `checkout.completed` | Checkout finished | Trigger subscription sync |

## Usage

### Webhook Handler (API Layer)

```go
// POST /api/webhooks/polar
func (h *Handler) HandlePolarWebhook(c *gin.Context) {
    var event domain.WebhookEvent
    if err := c.ShouldBindJSON(&event); err != nil {
        c.JSON(400, gin.H{"error": "invalid payload"})
        return
    }

    if err := h.billingService.ProcessSubscriptionWebhook(c.Request.Context(), &event); err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"status": "processed"})
}
```

### Getting Billing Status

```go
// In any handler that needs billing info
func (h *Handler) GetBillingStatus(c *gin.Context) {
    reqCtx := auth.GetRequestContext(c)

    status, err := h.billingService.GetBillingStatus(c.Request.Context(), reqCtx.OrganizationID)
    if err != nil {
        if err == domain.ErrSubscriptionNotFound {
            // No subscription yet - return appropriate response
            c.JSON(200, domain.BillingStatus{
                HasActiveSubscription: false,
                CanProcessInvoices:    false,
                Reason:                "No active subscription",
            })
            return
        }
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, status)
}
```

### Consuming Quota

```go
// Before processing an invoice
func (h *Handler) ProcessInvoice(c *gin.Context) {
    reqCtx := auth.GetRequestContext(c)

    // Check quota
    quotaStatus, err := h.billingService.GetQuotaStatus(c.Request.Context(), reqCtx.OrganizationID)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    if !quotaStatus.CanProcessInvoice {
        c.JSON(402, gin.H{
            "error": "quota_exceeded",
            "message": "Invoice processing quota exhausted",
            "upgrade_url": "/billing",
        })
        return
    }

    // Process the invoice...

    // Consume quota
    if err := h.billingService.ConsumeInvoiceQuota(c.Request.Context(), reqCtx.OrganizationID); err != nil {
        // Log but don't fail - invoice was processed
        log.Printf("failed to consume quota: %v", err)
    }
}
```

## Configuration

Environment variables for Polar.sh integration:

```env
POLAR_API_KEY=your_polar_api_key
POLAR_WEBHOOK_SECRET=your_webhook_secret
POLAR_ORGANIZATION_ID=your_polar_org_id
```

## Database Schema

```sql
-- Subscription tracking
CREATE TABLE subscription_billing.subscriptions (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations.organizations(id),
    external_customer_id TEXT NOT NULL,      -- Polar customer ID
    subscription_id TEXT NOT NULL,           -- Polar subscription ID
    subscription_status TEXT NOT NULL,       -- active, trialing, past_due, canceled, unpaid
    product_id TEXT NOT NULL,
    product_name TEXT,
    plan_name TEXT,
    current_period_start TIMESTAMP,
    current_period_end TIMESTAMP,
    cancel_at_period_end BOOLEAN DEFAULT FALSE,
    canceled_at TIMESTAMP,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Quota tracking
CREATE TABLE subscription_billing.quota_tracking (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations.organizations(id),
    invoice_count INTEGER DEFAULT 0,         -- Remaining invoices
    max_seats INTEGER,                       -- Seat limit
    period_start TIMESTAMP,
    period_end TIMESTAMP,
    last_synced_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

## Related Modules

- **pkg/paywall**: Access gating middleware (reads from this module's DB)
- **pkg/polar**: Polar.sh API client (used for webhook validation)
- **app/organizations**: Organization management (links subscription to org)

## Testing

```go
// Mock the billing service for unit tests
type MockBillingService struct {
    mock.Mock
}

func (m *MockBillingService) GetBillingStatus(ctx context.Context, orgID int32) (*domain.BillingStatus, error) {
    args := m.Called(ctx, orgID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*domain.BillingStatus), args.Error(1)
}

// In your test
func TestHandler_RequiresActiveSubscription(t *testing.T) {
    mockService := new(MockBillingService)
    mockService.On("GetBillingStatus", mock.Anything, int32(1)).Return(&domain.BillingStatus{
        HasActiveSubscription: false,
    }, nil)

    // Test that handler returns 402
}
```
