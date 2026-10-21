package utils

import (
	"regexp"
	"strings"
)

// Validator handles validation logic
type Validator struct{}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateEmail validates an email address
func (v *Validator) ValidateEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, err := regexp.MatchString(pattern, email)
	return err == nil && matched
}

// ValidateUsername validates a username
func (v *Validator) ValidateUsername(username string) bool {
	if len(username) < 3 || len(username) > 20 {
		return false
	}
	pattern := `^[a-zA-Z0-9_-]+$`
	matched, err := regexp.MatchString(pattern, username)
	return err == nil && matched
}

// ValidatePassword validates a password
func (v *Validator) ValidatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	return hasUpper && hasLower && hasDigit
}

// ValidateMangaID validates a manga ID
func (v *Validator) ValidateMangaID(id string) bool {
	return len(strings.TrimSpace(id)) > 0
}

// ValidateChapterNumber validates a chapter number
func (v *Validator) ValidateChapterNumber(chapter, maxChapters int) bool {
	return chapter > 0 && chapter <= maxChapters
}

// Package-level validation functions that return errors (for convenience)

// ValidateEmail validates an email address and returns an error if invalid
func ValidateEmail(email string) error {
	v := NewValidator()
	if !v.ValidateEmail(email) {
		return ErrInvalidEmail
	}
	return nil
}

// ValidateUsername validates a username and returns an error if invalid
func ValidateUsername(username string) error {
	v := NewValidator()
	if !v.ValidateUsername(username) {
		return ErrInvalidUsername
	}
	return nil
}

// ValidatePassword validates a password and returns an error if invalid
func ValidatePassword(password string) error {
	v := NewValidator()
	if !v.ValidatePassword(password) {
		return ErrInvalidPassword
	}
	return nil
}

var (
	ErrInvalidEmail    = &validationError{message: "invalid email format"}
	ErrInvalidUsername = &validationError{message: "username must be 3-20 characters, alphanumeric with - or _"}
	ErrInvalidPassword = &validationError{message: "password must be at least 8 characters with uppercase, lowercase, and digits"}
)

type validationError struct {
	message string
}

func (e *validationError) Error() string {
	return e.message
}
