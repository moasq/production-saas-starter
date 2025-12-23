// middleware/ip_protection.go

package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type IPProtection struct {
	whitelist      map[string]bool
	blacklist      map[string]struct{}
	failedAttempts map[string]*FailedAttempt
	mu             sync.RWMutex
	cleanupTicker  *time.Ticker
	done           chan struct{}
}

type FailedAttempt struct {
	count     int
	firstFail time.Time
	lastSeen  time.Time
}

func NewIPProtection() *IPProtection {
	ip := &IPProtection{
		whitelist:      make(map[string]bool),
		blacklist:      make(map[string]struct{}),
		failedAttempts: make(map[string]*FailedAttempt),
		cleanupTicker:  time.NewTicker(5 * time.Minute),
		done:           make(chan struct{}),
	}

	// Start cleanup goroutine
	go ip.periodicCleanup()

	return ip
}

// Stop should be called when the server is shutting down
func (ip *IPProtection) Stop() {
	ip.cleanupTicker.Stop()
	close(ip.done)
}

// periodicCleanup removes old entries from maps to prevent memory leaks
func (ip *IPProtection) periodicCleanup() {
	for {
		select {
		case <-ip.cleanupTicker.C:
			ip.cleanupFailedAttempts()
		case <-ip.done:
			return
		}
	}
}

// cleanupFailedAttempts removes old entries from the failedAttempts map
func (ip *IPProtection) cleanupFailedAttempts() {
	cutoff := time.Now().Add(-15 * time.Minute)

	ip.mu.Lock()
	defer ip.mu.Unlock()

	for clientIP, attempt := range ip.failedAttempts {
		// Remove entries that haven't been seen in the last 15 minutes
		if attempt.lastSeen.Before(cutoff) {
			delete(ip.failedAttempts, clientIP)
		}
	}
}

func (ip *IPProtection) Protect() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// Check if IP is whitelisted
		if ip.isWhitelisted(clientIP) {
			c.Next()
			return
		}

		// Check if IP is blacklisted
		if ip.isBlacklisted(clientIP) {
			c.AbortWithStatusJSON(403, gin.H{"error": "Access denied"})
			return
		}

		// Check for suspicious activity
		if ip.isSuspicious(clientIP) {
			ip.addToBlacklist(clientIP)
			c.AbortWithStatusJSON(403, gin.H{"error": "Suspicious activity detected"})
			return
		}

		c.Next()
	}
}

func (ip *IPProtection) isWhitelisted(clientIP string) bool {
	ip.mu.RLock()
	defer ip.mu.RUnlock()
	return ip.whitelist[clientIP]
}

func (ip *IPProtection) isBlacklisted(clientIP string) bool {
	ip.mu.RLock()
	defer ip.mu.RUnlock()
	_, exists := ip.blacklist[clientIP]
	return exists
}

func (ip *IPProtection) isSuspicious(clientIP string) bool {
	ip.mu.Lock()
	defer ip.mu.Unlock()

	attempt, exists := ip.failedAttempts[clientIP]
	if !exists {
		return false
	}

	// Check if more than 10 failed attempts in 5 minutes
	if attempt.count > 10 && time.Since(attempt.firstFail) < 5*time.Minute {
		return true
	}

	return false
}

func (ip *IPProtection) RecordFailedAttempt(clientIP string) {
	ip.mu.Lock()
	defer ip.mu.Unlock()

	now := time.Now()
	attempt, exists := ip.failedAttempts[clientIP]
	if !exists {
		ip.failedAttempts[clientIP] = &FailedAttempt{
			count:     1,
			firstFail: now,
			lastSeen:  now,
		}
		return
	}

	// Reset if last attempt was more than 5 minutes ago
	if time.Since(attempt.firstFail) > 5*time.Minute {
		attempt.count = 1
		attempt.firstFail = now
	} else {
		attempt.count++
	}
	attempt.lastSeen = now
}

func (ip *IPProtection) addToBlacklist(clientIP string) {
	ip.mu.Lock()
	defer ip.mu.Unlock()
	ip.blacklist[clientIP] = struct{}{}
}

// AddToWhitelist adds an IP to the whitelist
func (ip *IPProtection) AddToWhitelist(clientIP string) {
	ip.mu.Lock()
	defer ip.mu.Unlock()
	ip.whitelist[clientIP] = true
}

// RemoveFromBlacklist removes an IP from the blacklist
func (ip *IPProtection) RemoveFromBlacklist(clientIP string) {
	ip.mu.Lock()
	defer ip.mu.Unlock()
	delete(ip.blacklist, clientIP)
}
