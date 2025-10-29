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

	var expenseID string
	err := pool.QueryRow(
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

	return expenseID, nil
}

func UpdateExpense(ctx context.Context, pool *pgxpool.Pool, expense models.Expense) error {
	if expense.Title == "" {
		return errors.New("title required")
	}
	if expense.Amount <= 0 {
		return errors.New("invalid amount")
	}
	_, err := pool.Exec(
		ctx,
		`UPDATE expenses
			SET title = $2,
			description = $3,
			amount = $4,
			added_by = $5,
			is_incomplete_amount = $6,
			is_incomplete_split = $7
			latitude = $8
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
	if err == pgx.ErrNoRows {
		return errors.New("expense not found")
	}
	if err != nil {
		return err
	}

	return nil
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

	return expense, nil
}

func DeleteExpense(ctx context.Context, pool *pgxpool.Pool, expenseID string) error {
	cmd, err := pool.Exec(ctx, `DELETE FROM expenses WHERE expense_id = $1`, expenseID)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return errors.New("no rows deleted")
	}

	return nil
}
