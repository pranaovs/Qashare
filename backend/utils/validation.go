package utils

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
)

var nameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z .'\-]{0,63}[a-zA-Z]$`)

// ValidateName validates a user's name.
func ValidateName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("name is empty")
	}
	if !nameRegex.MatchString(name) {
		return "", errors.New("invalid name")
	}
	return name, nil
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail validates and normalizes an email.
// Returns a cleaned, lowercase email string or an error.
func ValidateEmail(email string) (string, error) {
	// Email validation regex
	email = strings.TrimSpace(email)
	email = strings.ToLower(email)

	if email == "" {
		return "", errors.New("email is empty")
	}

	if !emailRegex.MatchString(email) {
		return "", errors.New("invalid email format")
	}

	addr, err := mail.ParseAddress(email)
	if err != nil {
		return "", errors.New("invalid email syntax")
	}

	return addr.Address, nil
}
