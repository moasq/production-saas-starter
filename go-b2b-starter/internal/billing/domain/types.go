package domain

import "time"

// Subscription represents a billing subscription from Polar
type Subscription struct {
	ID                 int32
	OrganizationID     int32
	ExternalCustomerID string
	SubscriptionID     string
	SubscriptionStatus string
	ProductID          string
	ProductName        string
	PlanName           string
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
	CancelAtPeriodEnd  bool
	CanceledAt         *time.Time
	Metadata           map[string]any
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// QuotaTracking represents usage quota tracking for an organization
type QuotaTracking struct {
	ID             int32
	OrganizationID int32
	InvoiceCount   int32 // Remaining invoices (decremented on use)
	MaxSeats       int32
	PeriodStart    time.Time
	PeriodEnd      time.Time
	LastSyncedAt   *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// QuotaStatus represents the combined subscription and quota status
// This is returned from the GetQuotaStatus database query
type QuotaStatus struct {
	SubscriptionStatus string
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
	CancelAtPeriodEnd  bool
	InvoiceCount       int32 // Remaining invoices
	MaxSeats           int32
	CanProcessInvoice  bool
}

// BillingStatus represents the overall billing status for quota verification
type BillingStatus struct {
	OrganizationID        int32
	ExternalID            string
	HasActiveSubscription bool
	CanProcessInvoices    bool
	InvoiceCount          int32 // Remaining invoices
	Reason                string
	CheckedAt             time.Time
}

// WebhookEvent represents a Polar webhook event
type WebhookEvent struct {
	EventType string
	Payload   map[string]any
}

// SubscriptionEventData represents parsed subscription data from webhook
type SubscriptionEventData struct {
	SubscriptionID     string
	ExternalCustomerID string
	ProductID          string
	ProductName        string
	Status             string
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
	CancelAtPeriodEnd  bool
	CanceledAt         *time.Time
	ProductMetadata    map[string]string
	CustomerMetadata   map[string]string
}

// MeterGrantEventData represents meter grant payload details from Polar webhooks
type MeterGrantEventData struct {
	MeterSlug          string
	ExternalCustomerID string
	AvailableCredits   int32
}

// CheckoutSessionResponse represents a Polar checkout session
type CheckoutSessionResponse struct {
	ID             string
	Status         string // "succeeded", "pending", "expired", "failed"
	CustomerID     string
	SubscriptionID string
	ProductID      string
	Amount         int64
	CreatedAt      time.Time
}
