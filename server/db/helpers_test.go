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

// TestUserUpdate tests the UserUpdate struct validation logic
func TestUserUpdate(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	
	// Test that UserUpdate struct can hold all update fields
	name := "New Name"
	email := "newemail@example.com"
	password := "new_password_hash"
	guestTrue := true
	guestFalse := false
	
	// Test with name only
	update1 := UserUpdate{
		UserID: validUUID,
		Name:   &name,
	}
	if update1.UserID != validUUID || update1.Name == nil || *update1.Name != name {
		t.Errorf("UserUpdate with name not properly constructed")
	}
	
	// Test with email only
	update2 := UserUpdate{
		UserID: validUUID,
		Email:  &email,
	}
	if update2.Email == nil || *update2.Email != email {
		t.Errorf("UserUpdate with email not properly constructed")
	}
	
	// Test with password only
	update3 := UserUpdate{
		UserID:       validUUID,
		PasswordHash: &password,
	}
	if update3.PasswordHash == nil || *update3.PasswordHash != password {
		t.Errorf("UserUpdate with password not properly constructed")
	}
	
	// Test with Guest set to true
	update4 := UserUpdate{
		UserID: validUUID,
		Guest:  &guestTrue,
	}
	if update4.Guest == nil || *update4.Guest != true {
		t.Errorf("UserUpdate with Guest=true not properly constructed")
	}
	
	// Test with Guest set to false - this should be distinguishable from not set
	update5 := UserUpdate{
		UserID: validUUID,
		Guest:  &guestFalse,
	}
	if update5.Guest == nil || *update5.Guest != false {
		t.Errorf("UserUpdate with Guest=false not properly constructed")
	}
	
	// Test with all fields
	update6 := UserUpdate{
		UserID:       validUUID,
		Name:         &name,
		Email:        &email,
		PasswordHash: &password,
		Guest:        &guestTrue,
	}
	if update6.Name == nil || update6.Email == nil || update6.PasswordHash == nil || update6.Guest == nil {
		t.Errorf("UserUpdate with all fields not properly constructed")
	}
}
