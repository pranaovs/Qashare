package utils

import (
	"fmt"
	"reflect"
)

var ErrImmutableFieldSet = &UtilsError{Code: "IMMUTABLE_FIELD_SET", Message: "cannot set immutable field"}

// StripImmutableFields sets any field tagged with immutable:"true" to its zero value.
// This prevents clients from tampering with immutable fields in PATCH requests.
//
// Usage:
//
//	patch := &models.ExpenseDetails{ExpenseID: "exp123", Title: "New"}
//	StripImmutableFields(patch)  // ExpenseID becomes ""
//	// Now safe to use patch
func StripImmutableFields(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("expected non-nil pointer, got %T", v)
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %s", rv.Kind())
	}

	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		if tag := field.Tag.Get("immutable"); tag == "true" {
			rv.Field(i).Set(reflect.Zero(rv.Field(i).Type()))
		}
	}

	return nil
}

// ValidateNoImmutableFields returns error if any immutable field is non-zero.
// Used in PUT handler to ensure client didn't attempt to change immutable fields.
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

	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		if tag := field.Tag.Get("immutable"); tag == "true" {
			if !rv.Field(i).IsZero() {
				return ErrImmutableFieldSet.Msgf("field %s is immutable and cannot be modified", field.Name)
			}
		}
	}

	return nil
}

// MergeStructs merges patch into original, returning a new merged struct.
// This function automatically handles immutable fields:
// 1. Strips immutable fields from patch (sets them to zero)
// 2. For each remaining field in patch:
//   - If field is zero value, use original value
//   - If field is non-zero, use patch value
//
// 3. Ensures immutable fields in result always come from original
//
// Uses generics for type safety - returns same type as inputs.
//
// Usage:
//
//	current := &models.Expense{ExpenseID: "exp1", Title: "Old", Amount: 100}
//	patch := &models.Expense{ExpenseID: "hacker", Title: "New", Amount: 0}
//	merged, err := MergeStructs(current, patch)
//	// merged.ExpenseID = "exp1" (immutable, from original)
//	// merged.Title = "New" (from patch)
//	// merged.Amount = 100 (zero in patch, from original)
func MergeStructs[T any](original, patch *T) (*T, error) {
	if original == nil {
		return nil, fmt.Errorf("original must be non-nil")
	}
	if patch == nil {
		return nil, fmt.Errorf("patch must be non-nil")
	}

	origVal := reflect.ValueOf(original).Elem()
	patchVal := reflect.ValueOf(patch).Elem()

	if origVal.Kind() != reflect.Struct {
		return nil, fmt.Errorf("original must be struct, got %s", origVal.Kind())
	}
	if patchVal.Kind() != reflect.Struct {
		return nil, fmt.Errorf("patch must be struct, got %s", patchVal.Kind())
	}

	// Step 1: Strip immutable fields from patch
	patchType := patchVal.Type()
	for i := 0; i < patchVal.NumField(); i++ {
		field := patchType.Field(i)
		if tag := field.Tag.Get("immutable"); tag == "true" {
			patchField := patchVal.Field(i)
			if patchField.CanSet() {
				patchField.Set(reflect.Zero(patchField.Type()))
			}
		}
	}

	// Step 2: Create merged struct starting with original
	mergedVal := reflect.New(origVal.Type()).Elem()
	mergedVal.Set(origVal)

	// Step 3: Merge non-zero fields from patch
	for i := 0; i < patchVal.NumField(); i++ {
		patchField := patchVal.Field(i)
		mergedField := mergedVal.Field(i)

		// Skip if field is not settable
		if !mergedField.CanSet() {
			continue
		}

		// Skip if field is immutable (already has original value)
		field := patchType.Field(i)
		if tag := field.Tag.Get("immutable"); tag == "true" {
			continue
		}

		// If patch field is non-zero, use it; otherwise keep original
		if !patchField.IsZero() {
			mergedField.Set(patchField)
		}
	}

	// Return pointer to merged struct
	result := mergedVal.Addr().Interface().(*T)
	return result, nil
}
