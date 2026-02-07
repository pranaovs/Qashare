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
//   - If field is an embedded struct, recursively merge it
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

	// Create merged struct starting with original
	mergedVal := reflect.New(origVal.Type()).Elem()
	mergedVal.Set(origVal)

	// Merge fields from patch
	if err := mergeStructFields(mergedVal, patchVal); err != nil {
		return nil, err
	}

	// Return pointer to merged struct
	result := mergedVal.Addr().Interface().(*T)
	return result, nil
}

// mergeStructFields recursively merges fields from patchVal into mergedVal.
// It handles embedded structs by recursively merging their fields.
func mergeStructFields(mergedVal, patchVal reflect.Value) error {
	patchType := patchVal.Type()

	for i := 0; i < patchVal.NumField(); i++ {
		patchField := patchVal.Field(i)
		mergedField := mergedVal.Field(i)
		fieldInfo := patchType.Field(i)

		// Skip if field is not settable
		if !mergedField.CanSet() {
			continue
		}

		// Skip if field is immutable (keep original value)
		if tag := fieldInfo.Tag.Get("immutable"); tag == "true" {
			continue
		}

		// Handle embedded structs recursively
		if fieldInfo.Anonymous && patchField.Kind() == reflect.Struct {
			if err := mergeStructFields(mergedField, patchField); err != nil {
				return err
			}
			continue
		}

		// For non-embedded fields: if patch field is non-zero, use it
		if !patchField.IsZero() {
			mergedField.Set(patchField)
		}
	}

	return nil
}
