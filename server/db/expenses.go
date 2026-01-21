// Package db provides database operations for expense management.
// This file contains all expense-related database operations including creating,
// updating, retrieving, and deleting expenses with their associated splits.
package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pranaovs/qashare/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateExpense creates a new expense with associated splits in the database.
// This operation is atomic - either both the expense and all splits are created,
// or neither is (using a transaction).
//
// The expense parameter should contain:
//   - GroupID: The group this expense belongs to
//   - AddedBy: The user who added the expense
//   - Title: The expense title (required)
//   - Amount: The total amount (must be > 0 unless IsIncompleteAmount is true)
//   - Splits: List of expense splits (who paid and who owes)
//
// Returns the newly created expense's ID or an error if validation fails or the operation fails.
func CreateExpense(
	ctx context.Context,
	pool *pgxpool.Pool,
	expense models.Expense,
) (string, error) {
	// Validate input
	if expense.Title == "" {
		return "", ErrTitleRequired
	}
	if !expense.IsIncompleteAmount && expense.Amount <= 0 {
		return "", ErrInvalidAmount
	}

	log.Printf("[DB] Creating expense: %s (amount: %.2f, group: %s)", 
		expense.Title, expense.Amount, expense.GroupID)

	var expenseID string

	// Use WithTransaction helper for consistent transaction management
	err := WithTransaction(ctx, pool, func(ctx context.Context, tx pgx.Tx) error {
		// Insert expense record
		insertQuery := `INSERT INTO expenses (
			group_id, added_by, title, description, amount,
			is_incomplete_amount, is_incomplete_split, latitude, longitude, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING expense_id`
		
		err := tx.QueryRow(
			ctx,
			insertQuery,
			expense.GroupID,
			expense.AddedBy,
			expense.Title,
			expense.Description,
			expense.Amount,
			expense.IsIncompleteAmount,
			expense.IsIncompleteSplit,
			expense.Latitude,
			expense.Longitude,
			time.Now(),
		).Scan(&expenseID)
		if err != nil {
			return fmt.Errorf("failed to insert expense: %w", err)
		}

		// Batch insert splits for better performance
		if len(expense.Splits) > 0 {
			batch := &pgx.Batch{}
			splitQuery := `INSERT INTO expense_splits (expense_id, user_id, amount, is_paid)
				VALUES ($1, $2, $3, $4)`
			
			for _, split := range expense.Splits {
				batch.Queue(splitQuery, expenseID, split.UserID, split.Amount, split.IsPaid)
			}
			
			br := tx.SendBatch(ctx, batch)
			defer br.Close()

			// Execute all batched queries and check for errors
			for i := 0; i < len(expense.Splits); i++ {
				_, err = br.Exec()
				if err != nil {
					return fmt.Errorf("failed to insert split %d of %d: %w", i+1, len(expense.Splits), err)
				}
			}

			log.Printf("[DB] Inserted %d splits for expense", len(expense.Splits))
		}

		return nil
	})

	if err != nil {
		return "", NewDBError("CreateExpense", err, "failed to create expense")
	}

	log.Printf("[DB] ✓ Expense created successfully with ID: %s", expenseID)
	return expenseID, nil
}

// UpdateExpense updates an existing expense and replaces all its splits.
// This operation is atomic - either both the expense and all splits are updated,
// or neither is (using a transaction).
//
// The old splits are deleted and replaced with the new splits provided.
// Returns an error if validation fails or the operation fails.
func UpdateExpense(ctx context.Context, pool *pgxpool.Pool, expense models.Expense) error {
	// Validate input
	if expense.ExpenseID == "" {
		return ErrExpenseIDRequired
	}
	if expense.Title == "" {
		return ErrTitleRequired
	}
	if !expense.IsIncompleteAmount && expense.Amount <= 0 {
		return ErrInvalidAmount
	}

	log.Printf("[DB] Updating expense: %s (ID: %s)", expense.Title, expense.ExpenseID)

	// Use WithTransaction helper for consistent transaction management
	err := WithTransaction(ctx, pool, func(ctx context.Context, tx pgx.Tx) error {
		// Update main expense fields
		updateQuery := `UPDATE expenses
			SET title = $2,
				description = $3,
				amount = $4,
				added_by = $5,
				is_incomplete_amount = $6,
				is_incomplete_split = $7,
				latitude = $8,
				longitude = $9
			WHERE expense_id = $1`
		
		result, err := tx.Exec(
			ctx,
			updateQuery,
			expense.ExpenseID,
			expense.Title,
			expense.Description,
			expense.Amount,
			expense.AddedBy,
			expense.IsIncompleteAmount,
			expense.IsIncompleteSplit,
			expense.Latitude,
			expense.Longitude,
		)
		if err != nil {
			return fmt.Errorf("failed to update expense: %w", err)
		}

		// Check if expense was found
		if result.RowsAffected() == 0 {
			return ErrExpenseNotFound
		}

		// Remove old splits
		_, err = tx.Exec(ctx, `DELETE FROM expense_splits WHERE expense_id = $1`, expense.ExpenseID)
		if err != nil {
			return fmt.Errorf("failed to delete old splits: %w", err)
		}

		// Batch insert updated splits for better performance
		if len(expense.Splits) > 0 {
			batch := &pgx.Batch{}
			splitQuery := `INSERT INTO expense_splits (expense_id, user_id, amount, is_paid)
				VALUES ($1, $2, $3, $4)`
			
			for _, split := range expense.Splits {
				batch.Queue(splitQuery, expense.ExpenseID, split.UserID, split.Amount, split.IsPaid)
			}
			
			br := tx.SendBatch(ctx, batch)
			defer br.Close()

			// Execute all batched queries and check for errors
			for i := 0; i < len(expense.Splits); i++ {
				_, err = br.Exec()
				if err != nil {
					return fmt.Errorf("failed to insert split %d of %d: %w", i+1, len(expense.Splits), err)
				}
			}

			log.Printf("[DB] Updated %d splits for expense", len(expense.Splits))
		}

		return nil
	})

	if err != nil {
		return NewDBError("UpdateExpense", err, "failed to update expense")
	}

	log.Printf("[DB] ✓ Expense updated successfully: %s", expense.ExpenseID)
	return nil
}

// GetExpense retrieves a complete expense record including all its splits.
// Returns ErrExpenseNotFound if no expense with the ID exists.
func GetExpense(ctx context.Context, pool *pgxpool.Pool, expenseID string) (models.Expense, error) {
	log.Printf("[DB] Fetching expense: %s", expenseID)

	var expense models.Expense

	// Fetch expense details
	expenseQuery := `SELECT expense_id,
		group_id,
		added_by,
		title,
		description,
		extract(epoch from created_at)::bigint,
		amount,
		is_incomplete_amount,
		is_incomplete_split,
		latitude,
		longitude
	FROM expenses
	WHERE expense_id = $1`
	
	err := pool.QueryRow(ctx, expenseQuery, expenseID).Scan(
		&expense.ExpenseID,
		&expense.GroupID,
		&expense.AddedBy,
		&expense.Title,
		&expense.Description,
		&expense.CreatedAt,
		&expense.Amount,
		&expense.IsIncompleteAmount,
		&expense.IsIncompleteSplit,
		&expense.Latitude,
		&expense.Longitude,
	)
	if err == pgx.ErrNoRows {
		log.Printf("[DB] Expense not found: %s", expenseID)
		return models.Expense{}, ErrExpenseNotFound
	}
	if err != nil {
		return models.Expense{}, NewDBError("GetExpense", err, "failed to query expense")
	}

	// Fetch expense splits
	splitsQuery := `SELECT user_id, amount, is_paid 
		FROM expense_splits 
		WHERE expense_id = $1
		ORDER BY is_paid DESC, user_id`
	
	rows, err := pool.Query(ctx, splitsQuery, expenseID)
	if err != nil {
		return models.Expense{}, NewDBError("GetExpense", err, "failed to query splits")
	}
	defer rows.Close()

	// Scan splits into the expense
	expense.Splits = make([]models.ExpenseSplit, 0)
	for rows.Next() {
		var split models.ExpenseSplit
		split.ExpenseID = expenseID
		err = rows.Scan(&split.UserID, &split.Amount, &split.IsPaid)
		if err != nil {
			return models.Expense{}, NewDBError("GetExpense", err, "failed to scan split row")
		}
		expense.Splits = append(expense.Splits, split)
	}

	// Check for any errors during iteration
	if err := rows.Err(); err != nil {
		return models.Expense{}, NewDBError("GetExpense", err, "error iterating split rows")
	}

	log.Printf("[DB] ✓ Expense retrieved: %s with %d splits", expense.Title, len(expense.Splits))
	return expense, nil
}

// DeleteExpense deletes an expense from the database.
// This operation is atomic and uses a transaction.
// Note: The database will handle cascading deletes for expense_splits if configured.
// Returns ErrExpenseNotFound if no expense with the ID exists.
func DeleteExpense(ctx context.Context, pool *pgxpool.Pool, expenseID string) error {
	log.Printf("[DB] Deleting expense: %s", expenseID)

	// Use WithTransaction helper for consistent transaction management
	err := WithTransaction(ctx, pool, func(ctx context.Context, tx pgx.Tx) error {
		// Delete the expense (splits will be cascade deleted)
		deleteQuery := `DELETE FROM expenses WHERE expense_id = $1`
		
		result, err := tx.Exec(ctx, deleteQuery, expenseID)
		if err != nil {
			return fmt.Errorf("failed to delete expense: %w", err)
		}

		// Check if expense was found
		if result.RowsAffected() == 0 {
			return ErrExpenseNotFound
		}

		return nil
	})

	if err != nil {
		return NewDBError("DeleteExpense", err, "failed to delete expense")
	}

	log.Printf("[DB] ✓ Expense deleted successfully: %s", expenseID)
	return nil
}
