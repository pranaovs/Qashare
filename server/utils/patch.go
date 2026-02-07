package utils

import (
	"fmt"
	"reflect"
)

// INFO: lmao i dont know what black magic this file does. its all claude

// Patch applies a patch struct to a target struct.
// The patch struct should have pointer fields - nil means "not provided",
// non-nil means "apply this value" (even if the value is zero).
//
// Fields are matched by JSON tag name. Immutable fields (tagged with immutable:"true")
// in the target are skipped. Embedded structs in the target are handled recursively.
//
// Usage:
//
//	patch := &models.ExpensePatch{Amount: ptr(0.0), Title: ptr("New")}
//	target := &models.Expense{Amount: 100, Title: "Old"}
//	Patch(target, patch)
//	// target.Amount = 0 (applied from patch)
//	// target.Title = "New" (applied from patch)
func Patch(target, patch any) error {
	targetVal := reflect.ValueOf(target)
	patchVal := reflect.ValueOf(patch)

	if targetVal.Kind() != reflect.Pointer || targetVal.IsNil() {
		return fmt.Errorf("target must be non-nil pointer, got %T", target)
	}
	if patchVal.Kind() != reflect.Pointer || patchVal.IsNil() {
		return fmt.Errorf("patch must be non-nil pointer, got %T", patch)
	}

	targetVal = targetVal.Elem()
	patchVal = patchVal.Elem()

	if targetVal.Kind() != reflect.Struct {
		return fmt.Errorf("target must be struct, got %s", targetVal.Kind())
	}
	if patchVal.Kind() != reflect.Struct {
		return fmt.Errorf("patch must be struct, got %s", patchVal.Kind())
	}

	return applyPatchFields(targetVal, patchVal)
}

// applyPatchFields applies patch fields to target fields recursively.
func applyPatchFields(targetVal, patchVal reflect.Value) error {
	patchType := patchVal.Type()

	// Build a map of target fields by JSON name for efficient lookup
	targetFields := buildFieldMap(targetVal)

	for i := 0; i < patchVal.NumField(); i++ {
		patchFieldInfo := patchType.Field(i)
		patchField := patchVal.Field(i)

		// Handle embedded structs in patch (like ExpenseDetailsPatch embedding ExpensePatch)
		if patchFieldInfo.Anonymous {
			if patchField.Kind() == reflect.Struct {
				if err := applyPatchFields(targetVal, patchField); err != nil {
					return err
				}
			}
			continue
		}

		// Skip nil pointer fields (not provided in patch)
		if patchField.Kind() != reflect.Pointer || patchField.IsNil() {
			continue
		}

		// Get JSON field name from patch
		jsonName := getJSONName(patchFieldInfo)

		// Find corresponding target field
		targetFieldEntry, ok := targetFields[jsonName]
		if !ok {
			continue // No matching field in target
		}

		// Skip immutable fields
		if tag := targetFieldEntry.structField.Tag.Get("immutable"); tag == "true" {
			continue
		}

		targetField := targetFieldEntry.value
		if !targetField.CanSet() {
			continue
		}

		// Get the value from patch pointer
		patchValue := patchField.Elem()

		// Handle type compatibility
		if err := applyValue(targetField, patchValue); err != nil {
			return fmt.Errorf("field %s: %w", patchFieldInfo.Name, err)
		}
	}

	return nil
}

// fieldEntry holds both the reflect.Value and StructField for a target field
type fieldEntry struct {
	value       reflect.Value
	structField reflect.StructField
}

// buildFieldMap builds a map of JSON field names to field entries, handling embedded structs.
func buildFieldMap(val reflect.Value) map[string]fieldEntry {
	fields := make(map[string]fieldEntry)
	buildFieldMapRecursive(val, fields)
	return fields
}

func buildFieldMapRecursive(val reflect.Value, fields map[string]fieldEntry) {
	valType := val.Type()

	for i := 0; i < val.NumField(); i++ {
		fieldInfo := valType.Field(i)
		fieldVal := val.Field(i)

		// Recurse into anonymous embedded structs
		if fieldInfo.Anonymous && fieldVal.Kind() == reflect.Struct {
			buildFieldMapRecursive(fieldVal, fields)
			continue
		}

		jsonName := getJSONName(fieldInfo)
		fields[jsonName] = fieldEntry{
			value:       fieldVal,
			structField: fieldInfo,
		}
	}
}

// applyValue applies a patch value to a target field, handling type differences.
func applyValue(targetField, patchValue reflect.Value) error {
	targetType := targetField.Type()
	patchType := patchValue.Type()

	// Direct assignment if types match
	if patchType.AssignableTo(targetType) {
		targetField.Set(patchValue)
		return nil
	}

	// Handle pointer target with non-pointer patch (e.g., target is *string, patch is string)
	if targetType.Kind() == reflect.Pointer && patchType == targetType.Elem() {
		ptr := reflect.New(patchType)
		ptr.Elem().Set(patchValue)
		targetField.Set(ptr)
		return nil
	}

	// Handle slice assignment (e.g., patch has *[]T, target has []T)
	if targetType.Kind() == reflect.Slice && patchType.Kind() == reflect.Slice {
		if patchType.Elem().AssignableTo(targetType.Elem()) {
			targetField.Set(patchValue)
			return nil
		}
	}

	return fmt.Errorf("cannot assign %s to %s", patchType, targetType)
}

// getJSONName extracts the JSON field name from struct field tags.
func getJSONName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag == "" || jsonTag == "-" {
		return field.Name
	}
	// Handle "name,omitempty" format
	for i := 0; i < len(jsonTag); i++ {
		if jsonTag[i] == ',' {
			return jsonTag[:i]
		}
	}
	return jsonTag
}
