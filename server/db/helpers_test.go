package db

import (
	"errors"
	"strings"
	"testing"
)

// TestValidateUserForCreate tests the ValidateUserForCreate function
func TestValidateUserForCreate(t *testing.T) {
	tests := []struct {
		name         string
		userName     string
		email        string
		passwordHash string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "valid user data",
			userName:     "John Doe",
			email:        "john@example.com",
			passwordHash: "hashed_password_123",
			wantErr:      false,
		},
		{
			name:         "missing name",
			userName:     "",
			email:        "john@example.com",
			passwordHash: "hashed_password_123",
			wantErr:      true,
			errContains:  "name is required",
		},
		{
			name:         "missing email",
			userName:     "John Doe",
			email:        "",
			passwordHash: "hashed_password_123",
			wantErr:      true,
			errContains:  "email is required",
		},
		{
			name:         "missing password",
			userName:     "John Doe",
			email:        "john@example.com",
			passwordHash: "",
			wantErr:      true,
			errContains:  "password is required",
		},
		{
			name:         "whitespace only name",
			userName:     "   ",
			email:        "john@example.com",
			passwordHash: "hashed_password_123",
			wantErr:      true,
			errContains:  "name is required",
		},
		{
			name:         "whitespace only email",
			userName:     "John Doe",
			email:        "   ",
			passwordHash: "hashed_password_123",
			wantErr:      true,
			errContains:  "email is required",
		},
		{
			name:         "whitespace only password",
			userName:     "John Doe",
			email:        "john@example.com",
			passwordHash: "   ",
			wantErr:      true,
			errContains:  "password is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUserForCreate(tt.userName, tt.email, tt.passwordHash)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateUserForCreate() expected error but got nil")
					return
				}
				if !errors.Is(err, ErrMissingRequiredField) {
					t.Errorf("ValidateUserForCreate() error should wrap ErrMissingRequiredField, got: %v", err)
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateUserForCreate() error = %v, should contain %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateUserForCreate() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestValidateUserForUpdate tests the ValidateUserForUpdate function
func TestValidateUserForUpdate(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	invalidUUID := "not-a-uuid"

	tests := []struct {
		name         string
		userID       string
		userName     string
		email        string
		passwordHash string
		wantErr      bool
		errType      error
		errContains  string
	}{
		{
			name:         "valid update with name",
			userID:       validUUID,
			userName:     "New Name",
			email:        "",
			passwordHash: "",
			wantErr:      false,
		},
		{
			name:         "valid update with email",
			userID:       validUUID,
			userName:     "",
			email:        "newemail@example.com",
			passwordHash: "",
			wantErr:      false,
		},
		{
			name:         "valid update with password",
			userID:       validUUID,
			userName:     "",
			email:        "",
			passwordHash: "new_hashed_password",
			wantErr:      false,
		},
		{
			name:         "valid update with all fields",
			userID:       validUUID,
			userName:     "New Name",
			email:        "newemail@example.com",
			passwordHash: "new_hashed_password",
			wantErr:      false,
		},
		{
			name:         "missing userID",
			userID:       "",
			userName:     "New Name",
			email:        "newemail@example.com",
			passwordHash: "",
			wantErr:      true,
			errType:      ErrMissingRequiredField,
			errContains:  "user_id is required",
		},
		{
			name:         "invalid userID format",
			userID:       invalidUUID,
			userName:     "New Name",
			email:        "",
			passwordHash: "",
			wantErr:      true,
			errType:      ErrInvalidFieldValue,
			errContains:  "invalid user_id format",
		},
		{
			name:         "no fields to update",
			userID:       validUUID,
			userName:     "",
			email:        "",
			passwordHash: "",
			wantErr:      true,
			errType:      ErrMissingRequiredField,
			errContains:  "at least one field",
		},
		{
			name:         "whitespace only fields",
			userID:       validUUID,
			userName:     "   ",
			email:        "   ",
			passwordHash: "   ",
			wantErr:      true,
			errType:      ErrMissingRequiredField,
			errContains:  "at least one field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUserForUpdate(tt.userID, tt.userName, tt.email, tt.passwordHash)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateUserForUpdate() expected error but got nil")
					return
				}
				if tt.errType != nil && !errors.Is(err, tt.errType) {
					t.Errorf("ValidateUserForUpdate() error should wrap %v, got: %v", tt.errType, err)
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateUserForUpdate() error = %v, should contain %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateUserForUpdate() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestValidateUUID tests the ValidateUUID function
func TestValidateUUID(t *testing.T) {
	tests := []struct {
		name  string
		uuid  string
		valid bool
	}{
		{
			name:  "valid UUID",
			uuid:  "550e8400-e29b-41d4-a716-446655440000",
			valid: true,
		},
		{
			name:  "valid UUID uppercase",
			uuid:  "550E8400-E29B-41D4-A716-446655440000",
			valid: true,
		},
		{
			name:  "invalid UUID - too short",
			uuid:  "550e8400-e29b-41d4-a716",
			valid: false,
		},
		{
			name:  "invalid UUID - missing dashes",
			uuid:  "550e8400e29b41d4a716446655440000",
			valid: false,
		},
		{
			name:  "invalid UUID - invalid characters",
			uuid:  "550e8400-e29b-41d4-a716-44665544000g",
			valid: false,
		},
		{
			name:  "invalid UUID - empty string",
			uuid:  "",
			valid: false,
		},
		{
			name:  "invalid UUID - wrong format",
			uuid:  "not-a-uuid",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateUUID(tt.uuid)
			if result != tt.valid {
				t.Errorf("ValidateUUID(%q) = %v, want %v", tt.uuid, result, tt.valid)
			}
		})
	}
}
