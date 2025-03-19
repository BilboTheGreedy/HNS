package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidateHostname validates a generated hostname
func ValidateHostname(hostname string, maxLength int) error {
	if hostname == "" {
		return nil
	}

	// Check length
	if maxLength > 0 && len(hostname) > maxLength {
		return fmt.Errorf("hostname exceeds maximum length of %d characters", maxLength)
	}

	// Validate format (alphanumeric and hyphens only)
	validHostname := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
	if !validHostname.MatchString(hostname) {
		return fmt.Errorf("hostname contains invalid characters (only alphanumeric and hyphens allowed)")
	}

	// Hostname cannot start or end with hyphen
	if strings.HasPrefix(hostname, "-") || strings.HasSuffix(hostname, "-") {
		return fmt.Errorf("hostname cannot start or end with a hyphen")
	}

	// Validate no consecutive hyphens
	if strings.Contains(hostname, "--") {
		return fmt.Errorf("hostname cannot contain consecutive hyphens")
	}

	return nil
}

// ValidateEmail validates an email address
func ValidateEmail(email string) error {
	if email == "" {
		return nil
	}

	// Simplified email validation
	validEmail := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !validEmail.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// ValidateUsername validates a username
func ValidateUsername(username string) error {
	if username == "" {
		return nil
	}

	// Username must be alphanumeric with underscores and hyphens
	validUsername := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validUsername.MatchString(username) {
		return fmt.Errorf("username can only contain letters, numbers, underscores and hyphens")
	}

	// Username must be at least 3 characters
	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters")
	}

	// Username must be at most 50 characters
	if len(username) > 50 {
		return fmt.Errorf("username cannot exceed 50 characters")
	}

	return nil
}

// ValidatePassword validates a password
func ValidatePassword(password string) error {
	if password == "" {
		return nil
	}

	// Password must be at least 8 characters
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	// Password should contain at least one uppercase letter
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	// Password should contain at least one lowercase letter
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	// Password should contain at least one digit
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	// Password should contain at least one special character
	hasSpecial := regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password)

	// At least 3 of the 4 requirements must be met
	requirementsMet := 0
	if hasUpper {
		requirementsMet++
	}
	if hasLower {
		requirementsMet++
	}
	if hasDigit {
		requirementsMet++
	}
	if hasSpecial {
		requirementsMet++
	}

	if requirementsMet < 3 {
		return fmt.Errorf("password must meet at least 3 of the following: uppercase letter, lowercase letter, digit, special character")
	}

	return nil
}
