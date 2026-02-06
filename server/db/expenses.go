// Package db provides database operations for expense management.
// This file contains all expense-related database operations including creating,
// updating, retrieving, and deleting expenses with their associated splits.
package db

import (
	"context"
	"fmt"
	"log"

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
	expense *models.ExpenseDetails,
) error {
	// Validate input
	if expense.Title == "" {
		return ErrInvalidInput.Msg("title is required")
	}
	if !expense.IsIncompleteAmount && expense.Amount <= 0 {
		return ErrInvalidInput.Msg("amount must be greater than zero")
	}

	// Use WithTransaction helper for consistent transaction management
	err := WithTransaction(ctx, pool, func(ctx context.Context, tx pgx.Tx) error {
		// Insert expense record
		insertQuery := `INSERT INTO expenses (
			group_id, added_by, title, description, amount,
			is_incomplete_amount, is_incomplete_split, latitude, longitude
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING expense_id, extract(epoch from created_at)::bigint`

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
		).Scan(&expense.ExpenseID, &expense.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to insert expense: %w", err)
		}

		// Batch insert splits for better performance
		if len(expense.Splits) > 0 {
			batch := &pgx.Batch{}
			splitQuery := `INSERT INTO expense_splits (expense_id, user_id, amount, is_paid)
				VALUES ($1, $2, $3, $4)`

			for _, split := range expense.Splits {
				batch.Queue(splitQuery, expense.ExpenseID, split.UserID, split.Amount, split.IsPaid)
			}

			br := tx.SendBatch(ctx, batch)
			defer func() {
				if err := br.Close(); err != nil {
					log.Printf("[DB] Error closing batch: %v", err)
				}
			}()
			// Execute all batched queries and check for errors
			for i := 0; i < len(expense.Splits); i++ {
				_, err = br.Exec()
				if err != nil {
					return fmt.Errorf("failed to insert split %d of %d: %w", i+1, len(expense.Splits), err)
				}
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// UpdateExpense updates an existing expense and replaces all its splits.
// This operation is atomic - either both the expense and all splits are updated,
// or neither is (using a transaction).
//
// The old splits are deleted and replaced with the new splits provided.
// Returns an error if validation fails or the operation fails.
func UpdateExpense(ctx context.Context, pool *pgxpool.Pool, expense *models.ExpenseDetails) error {
	// Validate input
	if expense.ExpenseID == "" {
		return ErrNotFound.Msg("expense not found")
	}
	if expense.Title == "" {
		return ErrInvalidInput.Msg("title is required")
	}
	if !expense.IsIncompleteAmount && expense.Amount <= 0 {
		return ErrInvalidInput.Msg("amount must be greater than zero")
	}

	// Use WithTransaction helper for consistent transaction management
	err := WithTransaction(ctx, pool, func(ctx context.Context, tx pgx.Tx) error {
		// Update main expense fields
		updateQuery := `UPDATE expenses
			SET title = $2,
				description = $3,
				amount = $4,
				is_incomplete_amount = $5,
				is_incomplete_split = $6,
				latitude = $7,
				longitude = $8
			WHERE expense_id = $1`

		result, err := tx.Exec(
			ctx,
			updateQuery,
			expense.ExpenseID,
			expense.Title,
			expense.Description,
			expense.Amount,
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
			return ErrNotFound.Msgf("expense with id %s not found", expense.ExpenseID)
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
			defer func() {
				if err := br.Close(); err != nil {
					log.Printf("[DB] Error closing batch: %v", err)
				}
			}()

			// Execute all batched queries and check for errors
			for i := 0; i < len(expense.Splits); i++ {
				_, err = br.Exec()
				if err != nil {
					return fmt.Errorf("failed to insert split %d of %d: %w", i+1, len(expense.Splits), err)
				}
			}

		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// GetExpense retrieves a complete expense record including all its splits.
// Returns ErrExpenseNotFound if no expense with the ID exists.
func GetExpense(ctx context.Context, pool *pgxpool.Pool, expenseID string) (models.ExpenseDetails, error) {
	var expense models.ExpenseDetails

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
		return models.ExpenseDetails{}, ErrNotFound.Msgf("expense with id %s not found", expenseID)
	}
	if err != nil {
		return models.ExpenseDetails{}, err
	}

	// Fetch expense splits
	splitsQuery := `SELECT user_id, amount, is_paid
		FROM expense_splits 
		WHERE expense_id = $1
		ORDER BY is_paid DESC, user_id`

	rows, err := pool.Query(ctx, splitsQuery, expenseID)
	if err != nil {
		return models.ExpenseDetails{}, err
	}
	defer rows.Close()

	// Scan splits into the expense
	expense.Splits = make([]models.ExpenseSplit, 0)
	for rows.Next() {
		var split models.ExpenseSplit
		split.ExpenseID = expenseID
		err = rows.Scan(&split.UserID, &split.Amount, &split.IsPaid)
		if err != nil {
			return models.ExpenseDetails{}, err
		}
		expense.Splits = append(expense.Splits, split)
	}

	// Check for any errors during iteration
	if err := rows.Err(); err != nil {
		return models.ExpenseDetails{}, err
	}

	return expense, nil
}

// DeleteExpense deletes an expense from the database.
// This operation is atomic and uses a transaction.
// Note: The database will handle cascading deletes for expense_splits if configured.
// Returns ErrExpenseNotFound if no expense with the ID exists.
func DeleteExpense(ctx context.Context, pool *pgxpool.Pool, expenseID string) error {
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
			return ErrNotFound.Msgf("expense with id %s not found", expenseID)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// GetExpenses retrieves all expenses for a given group, ordered by creation time descending.
// Returns an empty slice if no expenses are found.
// Returns an error if the groupID is empty or the operation fails.
func GetExpenses(ctx context.Context, pool *pgxpool.Pool, groupID string) ([]models.Expense, error) {
	// TODO: Add pagination support for large datasets

	// Validate input
	if groupID == "" {
		return nil, ErrInvalidInput.Msg("group id missing")
	}

	// Query to get all expenses for the group
	expensesQuery := `SELECT expense_id,
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
	WHERE group_id = $1
	ORDER BY created_at DESC`

	rows, err := pool.Query(ctx, expensesQuery, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []models.Expense
	for rows.Next() {
		var expense models.Expense
		err = rows.Scan(
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
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, expense)
	}

	// Check for any errors during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return expenses, nil
}

// GetUserSpending calculates a user's spending summary in a specific group.
// It returns the total amount paid by the user, total amount owed, net spending,
// and a list of all expenses where the user either paid or owes money.
//
// Returns a *models.UserSpendings or an error if validation fails or the operation fails.
func GetUserSpending(ctx context.Context, pool *pgxpool.Pool, userID, groupID string) (*models.UserSpendings, error) {
	// Validate input
	if userID == "" {
		return nil, ErrInvalidInput.Msg("user id missing")
	}
	if groupID == "" {
		return nil, ErrInvalidInput.Msg("group id missing")
	}

	var spending models.UserSpendings

	// Calculate spending metrics:
	//
	//  TotalPaid:
	//    Sum of all expense splits for the user that are marked as paid (is_paid = true)
	//
	//  NetSpending:
	//    Sum of all expense splits for the user that are marked as owed (is_paid = false)
	//
	// NOTE: This represents settlement status and assigned consumption,
	// NOT who originally paid for the expense.
	sumQuery := `
		SELECT
			COALESCE(SUM(es.amount) FILTER (WHERE es.is_paid = true), 0)  AS net_paid,
			COALESCE(SUM(es.amount) FILTER (WHERE es.is_paid = false), 0) AS net_spending
		FROM expense_splits es
		JOIN expenses e ON e.expense_id = es.expense_id
		WHERE e.group_id = $1
			AND es.user_id = $2
	`
	err := pool.QueryRow(ctx, sumQuery, groupID, userID).
		Scan(&spending.TotalPaid, &spending.NetSpending)
	if err != nil {
		return nil, err
	}

	// Calculate TotalOwed as: TotalPaid - NetSpending
	// If user paid $100 but only consumed $80, TotalOwed = +$20 (others owe user)
	// If user paid $100 but consumed $120, TotalOwed = -$20 (user owes others)
	spending.TotalOwed = spending.TotalPaid - spending.NetSpending

	expensesQuery := `
		SELECT
			e.expense_id,
			e.group_id,
			e.added_by,
			e.title,
			e.description,
			extract(epoch from e.created_at)::bigint AS created_at,
			e.amount,
			SUM(es.amount) AS user_amount,
			e.is_incomplete_amount,
			e.is_incomplete_split,
			e.latitude,
			e.longitude
		FROM expenses e
		JOIN expense_splits es
			ON e.expense_id = es.expense_id
		WHERE e.group_id = $1
			AND es.user_id = $2
			AND es.is_paid = false
		GROUP BY
			e.expense_id,
			e.group_id,
			e.added_by,
			e.title,
			e.description,
			e.created_at,
			e.amount,
			e.is_incomplete_amount,
			e.is_incomplete_split,
			e.latitude,
			e.longitude
		ORDER BY e.created_at DESC
`

	rows, err := pool.Query(ctx, expensesQuery, groupID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	spending.Expenses = []models.UserSpendingsExpense{}
	for rows.Next() {
		var expense models.UserSpendingsExpense
		err = rows.Scan(
			&expense.ExpenseID,
			&expense.GroupID,
			&expense.AddedBy,
			&expense.Title,
			&expense.Description,
			&expense.CreatedAt,
			&expense.Amount,
			&expense.UserAmount,
			&expense.IsIncompleteAmount,
			&expense.IsIncompleteSplit,
			&expense.Latitude,
			&expense.Longitude,
		)
		if err != nil {
			return nil, err
		}
		spending.Expenses = append(spending.Expenses, expense)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &spending, nil
}
