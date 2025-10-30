package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// UserExpenseBreakdown represents a user's net spending on a single expense
type UserExpenseBreakdown struct {
	ExpenseID   string  `json:"expense_id"`
	Title       string  `json:"title"`
	Description string  `json:"description,omitempty"`
	CreatedAt   int64   `json:"created_at"`
	TotalAmount float64 `json:"total_amount"`
	AmountPaid  float64 `json:"amount_paid"`
	AmountOwed  float64 `json:"amount_owed"`
	NetSpending float64 `json:"net_spending"` // AmountPaid - AmountOwed (positive means user spent for themselves)
}

// GetUserExpensesInGroup retrieves all expenses in a group showing the user's personal spending
// This shows how much the user spent for themselves (amount they owe)
// Only returns expenses where the user has a non-zero amount owed (spent for themselves)
func GetUserExpensesInGroup(ctx context.Context, pool *pgxpool.Pool, groupID, userID string) ([]UserExpenseBreakdown, error) {
	rows, err := pool.Query(ctx, `
		WITH user_splits AS (
			SELECT 
				e.expense_id,
				e.title,
				e.description,
				e.amount as total_amount,
				extract(epoch from e.created_at)::bigint as created_at,
				COALESCE(SUM(CASE WHEN s.is_paid = true THEN s.amount ELSE 0 END), 0) as amount_paid,
				COALESCE(SUM(CASE WHEN s.is_paid = false THEN s.amount ELSE 0 END), 0) as amount_owed
			FROM expenses e
			LEFT JOIN expense_splits s ON e.expense_id = s.expense_id AND s.user_id = $2
			WHERE e.group_id = $1
			GROUP BY e.expense_id, e.title, e.description, e.amount, e.created_at
		)
		SELECT 
			expense_id,
			title,
			description,
			created_at,
			total_amount,
			amount_paid,
			amount_owed,
			(amount_paid - amount_owed) as net_spending
		FROM user_splits
		WHERE amount_owed > 0 AND title != 'Settlement'
		ORDER BY created_at DESC
	`, groupID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	breakdowns := make([]UserExpenseBreakdown, 0)
	for rows.Next() {
		var breakdown UserExpenseBreakdown
		var description *string

		err := rows.Scan(
			&breakdown.ExpenseID,
			&breakdown.Title,
			&description,
			&breakdown.CreatedAt,
			&breakdown.TotalAmount,
			&breakdown.AmountPaid,
			&breakdown.AmountOwed,
			&breakdown.NetSpending,
		)
		if err != nil {
			return nil, err
		}

		if description != nil {
			breakdown.Description = *description
		}

		breakdowns = append(breakdowns, breakdown)
	}

	return breakdowns, nil
}
