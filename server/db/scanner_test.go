package db

import (
	"reflect"
	"testing"
)

// TestBuildUpdateClauses tests the BuildUpdateClauses function
func TestBuildUpdateClauses(t *testing.T) {
	// Test with UserUpdate struct
	name := "John Doe"
	email := "john@example.com"
	guestTrue := true

	update := UserUpdate{
		UserID:       "550e8400-e29b-41d4-a716-446655440000",
		Name:         &name,
		Email:        &email,
		Guest:        &guestTrue,
		PasswordHash: nil, // Not set
	}

	clauses, args := BuildUpdateClauses(update, []string{"UserID"})

	// Should have 3 clauses (name, email, guest - not password or UserID)
	if len(clauses) != 3 {
		t.Errorf("Expected 3 clauses, got %d: %v", len(clauses), clauses)
	}

	// Should have 3 args
	if len(args) != 3 {
		t.Errorf("Expected 3 args, got %d: %v", len(args), args)
	}

	// Check that clauses use db tag names
	expectedClauses := []string{"user_name = $1", "email = $2", "is_guest = $3"}
	if !reflect.DeepEqual(clauses, expectedClauses) {
		t.Errorf("Expected clauses %v, got %v", expectedClauses, clauses)
	}

	// Check args
	expectedArgs := []interface{}{"John Doe", "john@example.com", true}
	if !reflect.DeepEqual(args, expectedArgs) {
		t.Errorf("Expected args %v, got %v", expectedArgs, args)
	}
}

// TestBuildUpdateClausesOnlyGuest tests updating only the guest field
func TestBuildUpdateClausesOnlyGuest(t *testing.T) {
	guestFalse := false
	update := UserUpdate{
		UserID: "550e8400-e29b-41d4-a716-446655440000",
		Guest:  &guestFalse, // Only this field is set
	}

	clauses, args := BuildUpdateClauses(update, []string{"UserID"})

	// Should have 1 clause
	if len(clauses) != 1 {
		t.Errorf("Expected 1 clause, got %d: %v", len(clauses), clauses)
	}

	// Should be able to set guest to false
	if clauses[0] != "is_guest = $1" {
		t.Errorf("Expected 'is_guest = $1', got %s", clauses[0])
	}

	if args[0] != false {
		t.Errorf("Expected false, got %v", args[0])
	}
}

// TestBuildUpdateClausesNoFields tests with no fields set
func TestBuildUpdateClausesNoFields(t *testing.T) {
	update := UserUpdate{
		UserID: "550e8400-e29b-41d4-a716-446655440000",
		// No fields set
	}

	clauses, args := BuildUpdateClauses(update, []string{"UserID"})

	// Should have 0 clauses
	if len(clauses) != 0 {
		t.Errorf("Expected 0 clauses, got %d: %v", len(clauses), clauses)
	}

	// Should have 0 args
	if len(args) != 0 {
		t.Errorf("Expected 0 args, got %d: %v", len(args), args)
	}
}
