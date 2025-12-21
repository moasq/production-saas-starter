package domain

import "errors"

var (
	// ErrSubscriptionNotFound is returned when a subscription cannot be found
	ErrSubscriptionNotFound = errors.New("subscription not found")

	// ErrSubscriptionNotActive is returned when a subscription exists but is not active
	ErrSubscriptionNotActive = errors.New("subscription is not active")

	// ErrQuotaNotFound is returned when quota tracking record cannot be found
	ErrQuotaNotFound = errors.New("quota not found")

	// ErrQuotaExceeded is returned when invoice quota has been exceeded
	ErrQuotaExceeded = errors.New("invoice quota exceeded")

	// ErrInvalidWebhookPayload is returned when webhook payload cannot be parsed
	ErrInvalidWebhookPayload = errors.New("invalid webhook payload")

	// ErrWebhookSignatureInvalid is returned when webhook signature verification fails
	ErrWebhookSignatureInvalid = errors.New("webhook signature invalid")

	// ErrQuotaDataStale is returned when quota data hasn't been synced recently
	ErrQuotaDataStale = errors.New("quota data is stale")
)
