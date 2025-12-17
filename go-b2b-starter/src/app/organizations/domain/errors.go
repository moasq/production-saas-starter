package domain

import "errors"

// Organization errors
var (
	ErrOrganizationNotFound      = errors.New("organization not found")
	ErrOrganizationNameRequired  = errors.New("organization name is required")
	ErrOrganizationSlugRequired  = errors.New("organization slug is required")
	ErrOrganizationSlugTooShort  = errors.New("organization slug must be at least 3 characters")
	ErrOrganizationSlugTaken     = errors.New("organization slug is already taken")
	ErrOrganizationInactive      = errors.New("organization is inactive")
)

// Account errors
var (
	ErrAccountNotFound             = errors.New("account not found")
	ErrAccountEmailRequired        = errors.New("account email is required")
	ErrAccountFullNameRequired     = errors.New("account full name is required")
	ErrAccountOrganizationRequired = errors.New("account organization is required")
	ErrAccountEmailTaken           = errors.New("account email is already taken")
	ErrAccountInactive             = errors.New("account is inactive")
	ErrAccountInsufficientRole     = errors.New("account does not have sufficient permissions")
)

// Permission errors
var (
	ErrPermissionDenied = errors.New("permission denied")
	ErrInvalidRole      = errors.New("invalid role")
)

// Auth provider member-related errors
var (
	ErrAuthMemberNotFound      = errors.New("auth member not found")
	ErrAuthMemberAlreadyExists = errors.New("auth member already exists")
	ErrAuthEmailRequired       = errors.New("email is required")
	ErrAuthInvalidEmail        = errors.New("invalid email format")
	ErrAuthPasswordRequired    = errors.New("password is required")
	ErrAuthNameRequired        = errors.New("name is required")
	ErrAuthMemberIDRequired    = errors.New("member ID is required")
	ErrAuthMemberIDsRequired   = errors.New("member IDs are required")
)

// Auth provider organization-related errors
var (
	ErrAuthOrganizationNotFound            = errors.New("auth organization not found")
	ErrAuthOrganizationAlreadyExists       = errors.New("auth organization already exists")
	ErrAuthOrganizationNameRequired        = errors.New("auth organization name is required")
	ErrAuthOrganizationDisplayNameRequired = errors.New("auth organization display name is required")
	ErrAuthOrganizationNameTooShort        = errors.New("auth organization name must be at least 2 characters")
	ErrAuthOrganizationIDRequired          = errors.New("auth organization ID is required")
)

// Auth provider role-related errors
var (
	ErrAuthRoleNotFound    = errors.New("auth role not found")
	ErrAuthRoleIDsRequired = errors.New("auth role IDs are required")
)

// Auth provider integration errors
var (
	ErrAuthConnection   = errors.New("failed to connect to auth provider")
	ErrAuthOperation    = errors.New("auth provider operation failed")
	ErrAuthUnauthorized = errors.New("unauthorized auth operation")
	ErrAuthRateLimit    = errors.New("auth provider rate limit exceeded")
)

// OrganizationError represents a domain-specific organization error
type OrganizationError struct {
	Type           string `json:"type"`
	Message        string `json:"message"`
	OrganizationID *int32 `json:"organization_id,omitempty"`
	Cause          error  `json:"-"`
}

func (e *OrganizationError) Error() string {
	return e.Message
}

func (e *OrganizationError) Unwrap() error {
	return e.Cause
}

func NewOrganizationError(errorType, message string, orgID *int32, cause error) *OrganizationError {
	return &OrganizationError{
		Type:           errorType,
		Message:        message,
		OrganizationID: orgID,
		Cause:          cause,
	}
}

// AccountError represents a domain-specific account error
type AccountError struct {
	Type           string `json:"type"`
	Message        string `json:"message"`
	AccountID      *int32 `json:"account_id,omitempty"`
	OrganizationID *int32 `json:"organization_id,omitempty"`
	Cause          error  `json:"-"`
}

func (e *AccountError) Error() string {
	return e.Message
}

func (e *AccountError) Unwrap() error {
	return e.Cause
}

func NewAccountError(errorType, message string, accountID, orgID *int32, cause error) *AccountError {
	return &AccountError{
		Type:           errorType,
		Message:        message,
		AccountID:      accountID,
		OrganizationID: orgID,
		Cause:          cause,
	}
}