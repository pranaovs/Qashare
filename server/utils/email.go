package utils

import (
	"fmt"
	"log/slog"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

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

var emailCfg config.EmailConfig
var apiCfg config.APIConfig

// InitEmail initializes the email package with the given configuration.
// Must be called before any email sending functions.
func InitEmail(emailConfig config.EmailConfig, apiConfig config.APIConfig) {
	emailCfg = emailConfig
	apiCfg = apiConfig
}

// ErrEmailSendFailed indicates that the verification email could not be sent
var ErrEmailSendFailed = &UtilsError{
	Code:    "EMAIL_SEND_FAILED",
	Message: "failed to send verification email",
}

// SendVerificationEmail sends a link-based verification email to the given address.
func SendVerificationEmail(to string, token uuid.UUID, expiry time.Duration) error {
	// Sanitize and validate the recipient email to prevent header injection.
	safeTo, err := sanitizeEmailAddress(to)
	if err != nil {
		return ErrEmailSendFailed.WithError(err)
	}

	subject := "Qashare - Verify your email address"

	link := fmt.Sprintf("%s%s/v1/auth/verify?token=%s", apiCfg.PublicURL, apiCfg.BasePath, token.String())

	body := fmt.Sprintf(
		"<html><body>"+
			"<p>Welcome to Qashare!</p>"+
			"<p>Please verify your email address by clicking the link below:</p>"+
			"<p><a href=\"%s\">Verify Email</a></p>"+
			"<p>If you did not create an account, you can ignore this email.</p>"+
			"<p>This link expires in %s.</p>"+
			"</body></html>",
		link, expiry,
	)

	msg := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=\"UTF-8\"\r\n"+
			"\r\n"+
			"%s",
		sanitizeHeader(emailCfg.From.String()), safeTo, subject, body,
	)

	auth := smtp.PlainAuth("", emailCfg.Username, emailCfg.Password, emailCfg.Host)

	err = smtp.SendMail(emailCfg.Host+":"+fmt.Sprint(emailCfg.Port), auth, emailCfg.From.Address, []string{safeTo}, []byte(msg))
	if err != nil {
		slog.Error("Failed to send verification email", "to", safeTo, "error", err)
		return ErrEmailSendFailed.WithError(err)
	}

	return nil
}

// SendGuestsInvitationEmail sends an invitation email to the given email id
func SendGuestsInvitationEmail(to string, from mail.Address) error {
	// Sanitize and validate the recipient email to prevent header injection.
	safeTo, err := sanitizeEmailAddress(to)
	if err != nil {
		return ErrEmailSendFailed.WithError(err)
	}

	subject := "Qashare - Invitation to join an expense group"

	link := apiCfg.PublicURL

	body := fmt.Sprintf(
		"<html><body>"+
			"<p>%s (%s) has invited you to join a shared expense group on Qashare</p>"+
			"<p>Click to join</p>"+
			"<p><a href=\"%s\">Join Now</a></p>"+
			"</body></html>",
		from.Name, from.Address, link,
	)

	msg := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=\"UTF-8\"\r\n"+
			"\r\n"+
			"%s",
		sanitizeHeader(emailCfg.From.String()), safeTo, subject, body,
	)

	auth := smtp.PlainAuth("", emailCfg.Username, emailCfg.Password, emailCfg.Host)

	err = smtp.SendMail(emailCfg.Host+":"+fmt.Sprint(emailCfg.Port), auth, emailCfg.From.Address, []string{safeTo}, []byte(msg))
	if err != nil {
		slog.Error("Failed to send invitation email", "to", safeTo, "error", err)
		return ErrEmailSendFailed.WithError(err)
	}

	return nil
}
