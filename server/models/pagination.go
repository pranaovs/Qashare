package models

import "github.com/google/uuid"

// PaginatedResponse wraps any list response with pagination metadata.
type PaginatedResponse[T any] struct {
	Data       []T            `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

// PaginationMeta contains cursor-based pagination metadata.
type PaginationMeta struct {
	NextCursor *uuid.UUID `json:"next_cursor"`
	HasNext    bool       `json:"has_next"`
}
