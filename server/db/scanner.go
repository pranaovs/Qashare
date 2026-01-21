package db

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5"
)

// ScanStruct scans a database row into a struct using the 'db' tags for field mapping.
// This provides automatic mapping of database columns to struct fields.
//
// Example usage:
//
//	var user models.User
//	err := ScanStruct(rows, &user)
//
// The struct fields should have 'db' tags matching the column names:
//
//	type User struct {
//	    UserID string `db:"user_id"`
//	    Name   string `db:"user_name"`
//	}
func ScanStruct(row pgx.Row, dest interface{}) error {
	val := reflect.ValueOf(dest)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer to a struct")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("dest must be a pointer to a struct")
	}

	// Get field pointers based on db tags
	fieldPointers, err := getFieldPointers(val)
	if err != nil {
		return err
	}

	// Scan the row into the field pointers
	return row.Scan(fieldPointers...)
}

// ScanStructs scans multiple database rows into a slice of structs using 'db' tags.
// This is useful for queries that return multiple rows.
//
// Example usage:
//
//	var users []models.User
//	err := ScanStructs(rows, &users)
func ScanStructs(rows pgx.Rows, dest interface{}) error {
	defer rows.Close()

	// Get the slice value
	sliceVal := reflect.ValueOf(dest)
	if sliceVal.Kind() != reflect.Ptr || sliceVal.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("dest must be a pointer to a slice")
	}

	sliceVal = sliceVal.Elem()
	elemType := sliceVal.Type().Elem()

	// Scan each row
	for rows.Next() {
		// Create a new instance of the element type
		elemPtr := reflect.New(elemType)
		elem := elemPtr.Elem()

		// Get field pointers for this element
		fieldPointers, err := getFieldPointers(elem)
		if err != nil {
			return err
		}

		// Scan the row into the field pointers
		if err := rows.Scan(fieldPointers...); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		// Append the element to the slice
		sliceVal.Set(reflect.Append(sliceVal, elem))
	}

	return rows.Err()
}

// getFieldPointers extracts pointers to struct fields based on their 'db' tags
// Returns a slice of pointers in the order of the struct fields with db tags
func getFieldPointers(val reflect.Value) ([]interface{}, error) {
	typ := val.Type()
	var fieldPointers []interface{}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Get the db tag
		dbTag := fieldType.Tag.Get("db")

		// Skip fields without db tags or with '-'
		if dbTag == "" || dbTag == "-" {
			continue
		}

		// Check if field can be addressed (exported)
		if !field.CanAddr() {
			return nil, fmt.Errorf("field %s cannot be addressed", fieldType.Name)
		}

		fieldPointers = append(fieldPointers, field.Addr().Interface())
	}

	return fieldPointers, nil
}

// GetDBColumns returns the list of database column names from struct 'db' tags
// This is useful for building SELECT queries dynamically
//
// Example:
//
//	columns := GetDBColumns(models.User{})
//	// Returns: []string{"user_id", "user_name", "email", ...}
func GetDBColumns(model interface{}) []string {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	typ := val.Type()
	var columns []string

	for i := 0; i < typ.NumField(); i++ {
		fieldType := typ.Field(i)
		dbTag := fieldType.Tag.Get("db")

		// Skip fields without db tags or with '-'
		if dbTag == "" || dbTag == "-" {
			continue
		}

		columns = append(columns, dbTag)
	}

	return columns
}

// BuildSelectQuery builds a SELECT query string from struct db tags
// This provides a convenient way to generate queries that match struct definitions
//
// Example:
//
//	query := BuildSelectQuery("users", models.User{}, "WHERE user_id = $1")
//	// Returns: "SELECT user_id, user_name, email, ... FROM users WHERE user_id = $1"
func BuildSelectQuery(tableName string, model interface{}, whereClause string) string {
	columns := GetDBColumns(model)
	if len(columns) == 0 {
		return ""
	}

	query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(columns, ", "), tableName)
	if whereClause != "" {
		query += " " + whereClause
	}

	return query
}

// GetDBColumnMap creates a map of struct field names to db column names
// This is useful for custom field mapping logic
func GetDBColumnMap(model interface{}) map[string]string {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	typ := val.Type()
	columnMap := make(map[string]string)

	for i := 0; i < typ.NumField(); i++ {
		fieldType := typ.Field(i)
		dbTag := fieldType.Tag.Get("db")

		// Skip fields without db tags or with '-'
		if dbTag == "" || dbTag == "-" {
			continue
		}

		columnMap[fieldType.Name] = dbTag
	}

	return columnMap
}
