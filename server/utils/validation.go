package utils

import (
	"net/mail"
	"regexp"
	"strings"

	"github.com/pranaovs/qashare/apierrors"
)

var nameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z .'\-]{1,62}[a-zA-Z]$`)

// ValidateName validates a user's name.
func ValidateName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", apierrors.ErrInvalidName.Msg("name cannot be empty")
	}
	if !nameRegex.MatchString(name) {
		return "", apierrors.ErrInvalidName.Msg("name must be 3-64 characters, start and end with a letter, and contain only letters, spaces, periods, apostrophes, and hyphens")
	}
	return name, nil
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail validates and normalizes an email. Returns a cleaned, lowercase email string or an error.
func ValidateEmail(email string) (string, error) {
	email = strings.TrimSpace(email)
	email = strings.ToLower(email)

	if email == "" {
		return "", apierrors.ErrInvalidEmail.Msg("email cannot be empty")
	}

	if !emailRegex.MatchString(email) {
		return "", apierrors.ErrInvalidEmail.Msg("email does not match required format")
	}

	addr, err := mail.ParseAddress(email)
	if err != nil {
		return "", apierrors.ErrInvalidEmail.Msg("invalid email syntax").WithInternal(err) // Send error with internal details
	}

	return addr.Address, nil
}
