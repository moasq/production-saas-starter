package repositories

import (
	"fmt"
	"regexp"
	"strings"
)

// generateSlug creates a URL-safe slug from an organization name
// following Stytch requirements: 2-128 chars, alphanumeric + -._~
func generateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove characters not allowed by Stytch (keep alphanumeric and -._~)
	reg := regexp.MustCompile(`[^a-z0-9\-._~]+`)
	slug = reg.ReplaceAllString(slug, "")

	// Remove leading and trailing hyphens, dots, underscores, tildes
	slug = strings.Trim(slug, "-._~")

	// Ensure minimum length of 2 characters
	if len(slug) < 2 {
		slug = "org-" + slug
	}

	// Ensure maximum length of 128 characters
	if len(slug) > 128 {
		slug = slug[:128]
		// Re-trim in case we cut off in middle of separator
		slug = strings.TrimRight(slug, "-._~")
	}

	return slug
}

// generateSlugWithSuffix generates a slug with numeric suffix for retry attempts
// attempt 1 returns base slug, attempt 2+ adds suffix
func generateSlugWithSuffix(baseSlug string, attempt int) string {
	if attempt <= 1 {
		return baseSlug
	}

	suffix := fmt.Sprintf("-%d", attempt)
	slug := baseSlug + suffix

	// Ensure we don't exceed 128 character limit
	if len(slug) > 128 {
		// Truncate base slug to make room for suffix
		maxBaseLength := 128 - len(suffix)
		slug = baseSlug[:maxBaseLength] + suffix
		// Re-trim in case we cut off in middle of separator
		slug = strings.TrimRight(slug[:len(slug)-len(suffix)], "-._~") + suffix
	}

	return slug
}
