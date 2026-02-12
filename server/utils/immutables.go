package utils

import (
	"fmt"
	"reflect"
)

var ErrImmutableFieldSet = &UtilsError{Code: "IMMUTABLE_FIELD_SET", Message: "cannot set immutable field"}

// StripImmutableFields sets any field tagged with immutable:"true" to its zero value.
// This helps prevent clients from tampering with immutable fields in request payloads
// (e.g., PATCH and PUT). It recursively handles anonymous embedded structs.
//
// Usage:
//
//	patch := &models.ExpenseDetails{ExpenseID: "exp123", Title: "New"}
//	StripImmutableFields(patch)  // ExpenseID becomes ""
//	// Now safe to use patch/update payload
func StripImmutableFields(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("expected non-nil pointer, got %T", v)
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %s", rv.Kind())
	}

	stripImmutableFieldsRecursive(rv)
	return nil
}

// stripImmutableFieldsRecursive recursively strips immutable fields from a struct value.
func stripImmutableFieldsRecursive(rv reflect.Value) {
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		fieldVal := rv.Field(i)

		// Skip if field is not settable (unexported)
		if !fieldVal.CanSet() {
			continue
		}

		// Recurse into anonymous embedded structs
		if field.Anonymous && fieldVal.Kind() == reflect.Struct {
			stripImmutableFieldsRecursive(fieldVal)
			continue
		}

		// Strip immutable fields
		if tag := field.Tag.Get("immutable"); tag == "true" {
			fieldVal.Set(reflect.Zero(fieldVal.Type()))
		}
	}
}

// RestoreImmutableFields copies all immutable:"true" fields from original into target.
// Both must be pointers to the same struct type. Recursively handles anonymous embedded structs.
func RestoreImmutableFields[T any](target, original *T) {
	// Guard against nil pointers to avoid reflect panics.
	if target == nil || original == nil {
		return
	}

	targetVal := reflect.ValueOf(target).Elem()
	originalVal := reflect.ValueOf(original).Elem()

	// Only operate on struct types; no-op for anything else to avoid panics.
	if targetVal.Kind() != reflect.Struct || originalVal.Kind() != reflect.Struct {
		return
	}
	restoreImmutableFieldsRecursive(targetVal, originalVal)
}

func restoreImmutableFieldsRecursive(target, original reflect.Value) {
	rt := target.Type()
	for i := 0; i < target.NumField(); i++ {
		field := rt.Field(i)
		targetField := target.Field(i)
		originalField := original.Field(i)

		if !targetField.CanSet() {
			continue
		}

		if field.Anonymous && targetField.Kind() == reflect.Struct {
			restoreImmutableFieldsRecursive(targetField, originalField)
			continue
		}

		if field.Tag.Get("immutable") == "true" {
			targetField.Set(originalField)
		}
	}
}

// ValidateNoImmutableFields returns error if any immutable field is non-zero.
// Used in PUT handler to ensure client didn't attempt to change immutable fields.
// It recursively checks anonymous embedded structs.
//
// Usage:
//
//	payload := &models.ExpenseDetails{ExpenseID: "exp123", ...}
//	if err := ValidateNoImmutableFields(payload); err != nil {
//	    return err  // Client tried to set immutable field
//	}
func ValidateNoImmutableFields(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("expected non-nil pointer, got %T", v)
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %s", rv.Kind())
	}

	return validateNoImmutableFieldsRecursive(rv)
}

// validateNoImmutableFieldsRecursive recursively validates that no immutable fields are set.
func validateNoImmutableFieldsRecursive(rv reflect.Value) error {
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		fieldVal := rv.Field(i)

		// Recurse into anonymous embedded structs
		if field.Anonymous && fieldVal.Kind() == reflect.Struct {
			if err := validateNoImmutableFieldsRecursive(fieldVal); err != nil {
				return err
			}
			continue
		}

		// Check immutable fields
		if tag := field.Tag.Get("immutable"); tag == "true" {
			if !fieldVal.IsZero() {
				return ErrImmutableFieldSet.Msgf("field %s is immutable and cannot be modified", field.Name)
			}
		}
	}

	return nil
}
