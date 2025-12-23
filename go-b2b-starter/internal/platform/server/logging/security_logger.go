// logging/security_logger.go

package logging

import (
	"time"

	"go.uber.org/zap"
)

type SecurityLogger struct {
	logger *zap.SugaredLogger
}

type SecurityEvent struct {
	EventType   string
	IP          string
	UserID      string
	Description string
	Severity    string
	Timestamp   time.Time
	RequestPath string
	RequestID   string
}

func NewSecurityLogger(baseLogger *zap.SugaredLogger) *SecurityLogger {
	return &SecurityLogger{
		logger: baseLogger,
	}
}

func (sl *SecurityLogger) LogSecurityEvent(event SecurityEvent) {
	sl.logger.Warnw("Security Event",
		"event_type", event.EventType,
		"ip", event.IP,
		"user_id", event.UserID,
		"description", event.Description,
		"severity", event.Severity,
		"timestamp", event.Timestamp,
		"request_path", event.RequestPath,
		"request_id", event.RequestID,
	)
}

func (sl *SecurityLogger) LogFailedAuth(ip string, userID string, reason string) {
	sl.LogSecurityEvent(SecurityEvent{
		EventType:   "AUTH_FAILED",
		IP:          ip,
		UserID:      userID,
		Description: reason,
		Severity:    "WARNING",
		Timestamp:   time.Now(),
	})
}

func (sl *SecurityLogger) LogSuspiciousActivity(ip string, description string) {
	sl.LogSecurityEvent(SecurityEvent{
		EventType:   "SUSPICIOUS_ACTIVITY",
		IP:          ip,
		Description: description,
		Severity:    "WARNING",
		Timestamp:   time.Now(),
	})
}

func (sl *SecurityLogger) LogBlacklisted(ip string) {
	sl.LogSecurityEvent(SecurityEvent{
		EventType:   "IP_BLACKLISTED",
		IP:          ip,
		Description: "IP address has been blacklisted",
		Severity:    "HIGH",
		Timestamp:   time.Now(),
	})
}
