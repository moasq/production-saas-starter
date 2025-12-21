package paywall

import (
	"context"

	"github.com/gin-gonic/gin"
)

// Context keys for storing subscription data.
// Using unexported type to prevent collisions with other packages.
type contextKey string

const (
	// subscriptionStatusKey is the context key for storing the SubscriptionStatus.
	subscriptionStatusKey contextKey = "subscription_status"
)

// SetSubscriptionStatus stores the SubscriptionStatus in the Gin context.
//
// This is called by the RequireActiveSubscription middleware after checking
// subscription status. Application code should not call this directly.
func SetSubscriptionStatus(c *gin.Context, status *SubscriptionStatus) {
	c.Set(string(subscriptionStatusKey), status)
}

// GetSubscriptionStatus retrieves the SubscriptionStatus from the Gin context.
//
// Returns nil if no subscription status is set (middleware not applied).
// Use MustGetSubscriptionStatus if you expect subscription middleware to have run.
//
// Example:
//
//	status := subscription.GetSubscriptionStatus(c)
//	if status == nil || !status.IsActive {
//	    // Handle inactive subscription
//	}
func GetSubscriptionStatus(c *gin.Context) *SubscriptionStatus {
	if val, exists := c.Get(string(subscriptionStatusKey)); exists {
		if status, ok := val.(*SubscriptionStatus); ok {
			return status
		}
	}
	return nil
}

// MustGetSubscriptionStatus retrieves the SubscriptionStatus from the Gin context.
//
// Panics if no subscription status is set. Only use this after subscription middleware.
// For handlers where subscription status is optional, use GetSubscriptionStatus instead.
func MustGetSubscriptionStatus(c *gin.Context) *SubscriptionStatus {
	status := GetSubscriptionStatus(c)
	if status == nil {
		panic("subscription: MustGetSubscriptionStatus called without SubscriptionStatus in context - ensure RequireActiveSubscription middleware is applied")
	}
	return status
}

// IsSubscriptionActive is a convenience function to check if the subscription is active.
//
// Returns false if no subscription status is set or if the subscription is inactive.
// Use this for quick checks in handlers.
//
// Example:
//
//	if !subscription.IsSubscriptionActive(c) {
//	    // Handle inactive subscription
//	}
func IsSubscriptionActive(c *gin.Context) bool {
	if status := GetSubscriptionStatus(c); status != nil {
		return status.IsActive
	}
	return false
}

// WithSubscriptionStatus adds the SubscriptionStatus to a context.Context.
//
// This is useful for passing subscription context through service layers
// that don't use Gin context directly.
func WithSubscriptionStatus(ctx context.Context, status *SubscriptionStatus) context.Context {
	return context.WithValue(ctx, subscriptionStatusKey, status)
}

// SubscriptionStatusFromContext retrieves the SubscriptionStatus from a context.Context.
//
// Returns nil if no subscription status is set.
func SubscriptionStatusFromContext(ctx context.Context) *SubscriptionStatus {
	if val := ctx.Value(subscriptionStatusKey); val != nil {
		if status, ok := val.(*SubscriptionStatus); ok {
			return status
		}
	}
	return nil
}
