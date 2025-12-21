package events

import (
	"time"

	"github.com/moasq/go-b2b-starter/internal/organizations/domain"
)

const (
	OrganizationCreatedEventType = "organization.created"
	OrganizationUpdatedEventType = "organization.updated"
	AccountCreatedEventType      = "account.created"
	AccountUpdatedEventType      = "account.updated"
	AccountDeletedEventType      = "account.deleted"
	AccountLoginEventType        = "account.login"
)

type OrganizationCreatedEvent struct {
	EventID       string                `json:"event_id"`
	EventType     string                `json:"event_type"`
	Timestamp     time.Time             `json:"timestamp"`
	Organization  *domain.Organization  `json:"organization"`
	OwnerAccount  *domain.Account       `json:"owner_account"`
}

type OrganizationUpdatedEvent struct {
	EventID      string               `json:"event_id"`
	EventType    string               `json:"event_type"`
	Timestamp    time.Time            `json:"timestamp"`
	Organization *domain.Organization `json:"organization"`
	PreviousName string               `json:"previous_name"`
}

type AccountCreatedEvent struct {
	EventID        string              `json:"event_id"`
	EventType      string              `json:"event_type"`
	Timestamp      time.Time           `json:"timestamp"`
	Account        *domain.Account     `json:"account"`
	OrganizationID int32               `json:"organization_id"`
}

type AccountUpdatedEvent struct {
	EventID        string              `json:"event_id"`
	EventType      string              `json:"event_type"`
	Timestamp      time.Time           `json:"timestamp"`
	Account        *domain.Account     `json:"account"`
	OrganizationID int32               `json:"organization_id"`
	PreviousRole   string              `json:"previous_role"`
	PreviousStatus string              `json:"previous_status"`
}

type AccountDeletedEvent struct {
	EventID        string              `json:"event_id"`
	EventType      string              `json:"event_type"`
	Timestamp      time.Time           `json:"timestamp"`
	AccountID      int32               `json:"account_id"`
	OrganizationID int32               `json:"organization_id"`
	Email          string              `json:"email"`
}

type AccountLoginEvent struct {
	EventID        string              `json:"event_id"`
	EventType      string              `json:"event_type"`
	Timestamp      time.Time           `json:"timestamp"`
	AccountID      int32               `json:"account_id"`
	OrganizationID int32               `json:"organization_id"`
	Email          string              `json:"email"`
}