package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

var ErrImmutableFieldSet = &UtilsError{Code: "IMMUTABLE_FIELD_SET", Message: "cannot set immutable field"}

// StripImmutableFields sets any field tagged with immutable:"true" to its zero value.
// This prevents clients from tampering with immutable fields in PATCH requests.
// It recursively handles anonymous embedded structs.
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

// MergeWithJSON merges a JSON patch into an original struct with proper PATCH semantics.
// Unlike MergeStructs, this function correctly handles zero values by tracking which
// fields were actually present in the JSON payload.
//
// This solves the "zero value problem" where clients cannot:
//   - Set booleans to false
//   - Set numbers to 0
//   - Set strings to ""
//   - Clear pointers to nil
//
// The function recursively handles:
//   - Anonymous embedded structs (JSON fields are flattened)
//   - Named nested structs (JSON fields are nested objects)
//
// Usage:
//
//	current := &models.Expense{Title: "Old", Amount: 100, IsIncomplete: true}
//	jsonBytes := []byte(`{"amount": 0, "is_incomplete": false}`)
//	merged, err := MergeWithJSON(current, jsonBytes)
//	// merged.Title = "Old" (not in JSON, kept from original)
//	// merged.Amount = 0 (explicitly set in JSON)
//	// merged.IsIncomplete = false (explicitly set in JSON)
func MergeWithJSON[T any](original *T, jsonData []byte) (*T, error) {
	if original == nil {
		return nil, fmt.Errorf("original must be non-nil")
	}

	// Parse JSON to see which fields are present (keeps raw values for nested objects)
	var rawFields map[string]json.RawMessage
	if err := json.Unmarshal(jsonData, &rawFields); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Unmarshal JSON into patch struct
	var patch T
	if err := json.Unmarshal(jsonData, &patch); err != nil {
		return nil, fmt.Errorf("failed to unmarshal patch: %w", err)
	}

	origVal := reflect.ValueOf(original).Elem()
	patchVal := reflect.ValueOf(&patch).Elem()

	// Create merged struct starting with original
	mergedVal := reflect.New(origVal.Type()).Elem()
	mergedVal.Set(origVal)

	// Merge fields from patch, only if present in JSON
	if err := mergeStructFieldsWithPresence(mergedVal, patchVal, rawFields); err != nil {
		return nil, err
	}

	result := mergedVal.Addr().Interface().(*T)
	return result, nil
}

// mergeStructFieldsWithPresence recursively merges fields that were present in the JSON.
// For anonymous embedded structs, fields are flattened in JSON so we pass the same rawFields.
// For named nested structs, we parse the nested JSON object for recursive merging.
func mergeStructFieldsWithPresence(mergedVal, patchVal reflect.Value, rawFields map[string]json.RawMessage) error {
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

		// Handle anonymous embedded structs recursively (JSON flattens these)
		if fieldInfo.Anonymous && patchField.Kind() == reflect.Struct {
			if err := mergeStructFieldsWithPresence(mergedField, patchField, rawFields); err != nil {
				return err
			}
			continue
		}

		// Get JSON field name
		jsonName := getJSONFieldName(fieldInfo)
		jsonNameLower := strings.ToLower(jsonName)

		// Find the raw JSON value (case-insensitive)
		var rawValue json.RawMessage
		var found bool
		for key, val := range rawFields {
			if strings.ToLower(key) == jsonNameLower {
				rawValue = val
				found = true
				break
			}
		}

		// Skip if field was not present in JSON
		if !found {
			continue
		}

		// Handle named nested structs recursively
		if patchField.Kind() == reflect.Struct {
			// Parse nested JSON object to get its fields
			var nestedRawFields map[string]json.RawMessage
			if err := json.Unmarshal(rawValue, &nestedRawFields); err == nil {
				// Recursively merge the nested struct
				if err := mergeStructFieldsWithPresence(mergedField, patchField, nestedRawFields); err != nil {
					return err
				}
				continue
			}
			// If parsing as object fails, fall through to direct assignment
		}

		// Handle pointer to struct recursively
		if patchField.Kind() == reflect.Ptr && !patchField.IsNil() && patchField.Elem().Kind() == reflect.Struct {
			// Check if JSON value is null
			if string(rawValue) == "null" {
				mergedField.Set(reflect.Zero(mergedField.Type()))
				continue
			}
			// Parse nested JSON object
			var nestedRawFields map[string]json.RawMessage
			if err := json.Unmarshal(rawValue, &nestedRawFields); err == nil {
				// Ensure merged field has a non-nil pointer
				if mergedField.IsNil() {
					mergedField.Set(reflect.New(mergedField.Type().Elem()))
				}
				if err := mergeStructFieldsWithPresence(mergedField.Elem(), patchField.Elem(), nestedRawFields); err != nil {
					return err
				}
				continue
			}
		}

		// For all other types (primitives, slices, etc.), apply the patch value directly
		mergedField.Set(patchField)
	}

	return nil
}

// getJSONFieldName returns the JSON field name for a struct field.
func getJSONFieldName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag == "" || jsonTag == "-" {
		return field.Name
	}
	// Handle "name,omitempty" format
	name, _, _ := strings.Cut(jsonTag, ",")
	if name == "" {
		return field.Name
	}
	return name
}
