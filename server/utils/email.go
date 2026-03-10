package utils

import (
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"

	"github.com/google/uuid"
	"github.com/pranaovs/qashare/config"
)

// sanitizeHeader removes CR and LF characters to prevent email header injection.
func sanitizeHeader(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	return s
}

// sanitizeEmailAddress normalizes and sanitizes an email address to prevent injection.
// It trims whitespace, strips CR/LF characters, and re-validates the address.
func sanitizeEmailAddress(email string) (string, error) {
	// First, remove any CR/LF characters that could be used for header injection.
	email = sanitizeHeader(email)

	// Reuse existing email validation/normalization logic.
	safeEmail, err := ValidateEmail(email)
	if err != nil {
		return "", err
	}

	return safeEmail, nil
}

// ErrEmailSendFailed indicates that the verification email could not be sent
var ErrEmailSendFailed = &UtilsError{
	Code:    "EMAIL_SEND_FAILED",
	Message: "failed to send verification email",
}

// SendVerificationEmail sends a link-based verification email to the given address.
func SendVerificationEmail(emailConfig config.EmailConfig, apiConfig config.APIConfig, to string, token uuid.UUID) error {
	// Sanitize and validate the recipient email to prevent header injection.
	safeTo, err := sanitizeEmailAddress(to)
	if err != nil {
		return ErrEmailSendFailed.WithError(err)
	}

	subject := "Qashare - Verify your email address"

	link := fmt.Sprintf("%s%s/v1/auth/verify?token=%s", apiConfig.PublicURL, apiConfig.BasePath, token.String())

	body := fmt.Sprintf(
		"<html><body>"+
			"<p>Welcome to Qashare!</p>"+
			"<p>Please verify your email address by clicking the link below:</p>"+
			"<p><a href=\"%s\">Verify Email</a></p>"+
			"<p>If you did not create an account, you can ignore this email.</p>"+
			"<p>This link expires in %s.</p>"+
			"</body></html>",
		link, emailConfig.Expiry,
	)

	msg := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=\"UTF-8\"\r\n"+
			"\r\n"+
			"%s",
		sanitizeHeader(emailConfig.From.String()), safeTo, subject, body,
	)

	auth := smtp.PlainAuth("", emailConfig.Username, emailConfig.Password, emailConfig.Host)

	err = smtp.SendMail(emailConfig.Host+":"+fmt.Sprint(emailConfig.Port), auth, emailConfig.From.Address, []string{safeTo}, []byte(msg))
	if err != nil {
		slog.Error("Failed to send verification email", "to", safeTo, "error", err)
		return ErrEmailSendFailed.WithError(err)
	}

	return nil
}
