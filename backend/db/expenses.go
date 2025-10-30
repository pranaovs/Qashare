package db

import (
	"context"
	"errors"
	"shared-expenses-app/models"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateExpense(
	ctx context.Context,
	pool *pgxpool.Pool,
	expense models.Expense,
) (string, error) {
	if expense.Title == "" {
		return "", errors.New("title required")
	}
	if expense.Amount <= 0 {
		return "", errors.New("invalid amount")
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	// Insert expense details
	var expenseID string
	err = tx.QueryRow(
		ctx,
		`INSERT INTO expenses (
			group_id, added_by, title, description, amount,
			is_incomplete_amount, is_incomplete_split, latitude, longitude, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING expense_id`,
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
		return "", err
	}

	// Batch insert splits for better performance
	if len(expense.Splits) > 0 {
		batch := &pgx.Batch{}
		for _, split := range expense.Splits {
			batch.Queue(`
				INSERT INTO expense_splits (expense_id, user_id, amount, is_paid)
				VALUES ($1, $2, $3, $4)
			`, expenseID, split.UserID, split.Amount, split.IsPaid)
		}
		br := tx.SendBatch(ctx, batch)
		defer br.Close()
		
		// Execute all batched queries and check for errors
		for i := 0; i < len(expense.Splits); i++ {
			_, err = br.Exec()
			if err != nil {
				return "", err
			}
		}
	}
	err = tx.Commit(ctx)
	if err != nil {
		return "", err
	}

	return expenseID, nil
}

func UpdateExpense(ctx context.Context, pool *pgxpool.Pool, expense models.Expense) error {
	if expense.ExpenseID == "" {
		return errors.New("expense_id required")
	}
	if expense.Title == "" {
		return errors.New("title required")
	}
	if expense.Amount <= 0 {
		return errors.New("invalid amount")
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Update main expense fields
	_, err = tx.Exec(
		ctx,
		`UPDATE expenses
			SET title = $2,
				description = $3,
				amount = $4,
				added_by = $5,
				is_incomplete_amount = $6,
				is_incomplete_split = $7,
				latitude = $8,
				longitude = $9
			WHERE expense_id = $1`,
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
		return err
	}

	// Remove old splits first
	_, err = tx.Exec(ctx, `DELETE FROM expense_splits WHERE expense_id = $1`, expense.ExpenseID)
	if err != nil {
		return err
	}

	// Batch insert updated splits for better performance
	if len(expense.Splits) > 0 {
		batch := &pgx.Batch{}
		for _, split := range expense.Splits {
			batch.Queue(`
				INSERT INTO expense_splits (expense_id, user_id, amount, is_paid)
				VALUES ($1, $2, $3, $4)
			`, expense.ExpenseID, split.UserID, split.Amount, split.IsPaid)
		}
		br := tx.SendBatch(ctx, batch)
		defer br.Close()
		
		// Execute all batched queries and check for errors
		for i := 0; i < len(expense.Splits); i++ {
			_, err = br.Exec()
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}

func GetExpense(ctx context.Context, pool *pgxpool.Pool, expenseID string) (models.Expense, error) {
	var expense models.Expense

	err := pool.QueryRow(
		ctx,
		`SELECT expense_id,
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
		 WHERE expense_id = $1`,
		expenseID,
	).Scan(
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
		return models.Expense{}, errors.New("expense not found")
	}
	if err != nil {
		return models.Expense{}, err
	}

	// Fetch splits
	rows, err := pool.Query(ctx, `SELECT user_id, amount, is_paid FROM expense_splits WHERE expense_id = $1`, expenseID)
	if err != nil {
		return models.Expense{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var split models.ExpenseSplit
		err = rows.Scan(&split.UserID, &split.Amount, &split.IsPaid)
		if err != nil {
			return models.Expense{}, err
		}
		expense.Splits = append(expense.Splits, split)
	}

	return expense, nil
}

func DeleteExpense(ctx context.Context, pool *pgxpool.Pool, expenseID string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Delete Expense
	cmd, err := tx.Exec(ctx, `DELETE FROM expenses WHERE expense_id = $1`, expenseID)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return errors.New("expense not found")
	}

	return tx.Commit(ctx)
}
