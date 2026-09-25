package domain

import "time"

// BillingStatus is a fresh provider snapshot, not a locally cached entitlement.
type BillingStatus struct {
	BillingEnabled        bool
	OrganizationID        int32
	ExternalID            string
	HasActiveSubscription bool
	SubscriptionID        string
	SubscriptionStatus    string
	ProductID             string
	CurrentPeriodEnd      *time.Time
	CancelAtPeriodEnd     bool
	Reason                string
	CheckedAt             time.Time
}

type Subscription struct {
	ID                string     `json:"id"`
	Status            string     `json:"status"`
	ProductID         string     `json:"product_id"`
	CurrentPeriodEnd  *time.Time `json:"current_period_end"`
	EndsAt            *time.Time `json:"ends_at"`
	CancelAtPeriodEnd bool       `json:"cancel_at_period_end"`
}

type CustomerState struct {
	ExternalID          string         `json:"external_id"`
	ActiveSubscriptions []Subscription `json:"active_subscriptions"`
}

type CheckoutSession struct {
	ProductID          string `json:"product_id"`
	Status             string `json:"status"`
	CustomerExternalID string `json:"external_customer_id"`
}
