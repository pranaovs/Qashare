package db

import (
	"context"
	"math"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Settlement represents a simplified debt between two users
type Settlement struct {
	FromUserID string  `json:"from_user_id"`
	ToUserID   string  `json:"to_user_id"`
	Amount     float64 `json:"amount"`
}

// CalculateSettlements calculates the simplified settlements for a group
// It uses a greedy algorithm to minimize the number of transactions
func CalculateSettlements(ctx context.Context, pool *pgxpool.Pool, groupID string) ([]Settlement, error) {
	// Get all splits for expenses in this group
	rows, err := pool.Query(ctx, `
		SELECT s.user_id, s.amount, s.is_paid
		FROM expense_splits s
		JOIN expenses e ON s.expense_id = e.expense_id
		WHERE e.group_id = $1
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Calculate net balance for each user
	balances := make(map[string]float64)

	for rows.Next() {
		var userID string
		var amount float64
		var isPaid bool

		if err := rows.Scan(&userID, &amount, &isPaid); err != nil {
			return nil, err
		}

		if isPaid {
			// User paid, they are owed this amount
			balances[userID] += amount
		} else {
			// User owes this amount
			balances[userID] -= amount
		}
	}

	// Round balances to 2 decimal places to avoid floating point issues
	for userID := range balances {
		balances[userID] = math.Round(balances[userID]*100) / 100
	}

	// Separate creditors (positive balance) and debtors (negative balance)
	var creditors []struct {
		userID string
		amount float64
	}
	var debtors []struct {
		userID string
		amount float64
	}

	for userID, balance := range balances {
		if balance > 0.01 { // Ignore very small amounts
			creditors = append(creditors, struct {
				userID string
				amount float64
			}{userID, balance})
		} else if balance < -0.01 {
			debtors = append(debtors, struct {
				userID string
				amount float64
			}{userID, -balance}) // Store as positive amount
		}
	}

	// Calculate settlements using greedy algorithm
	settlements := make([]Settlement, 0)
	i, j := 0, 0

	for i < len(debtors) && j < len(creditors) {
		debtor := &debtors[i]
		creditor := &creditors[j]

		// Settle the minimum of what debtor owes and what creditor is owed
		settleAmount := math.Min(debtor.amount, creditor.amount)
		settleAmount = math.Round(settleAmount*100) / 100

		if settleAmount > 0.01 { // Only add if amount is significant
			settlements = append(settlements, Settlement{
				FromUserID: debtor.userID,
				ToUserID:   creditor.userID,
				Amount:     settleAmount,
			})
		}

		// Update balances
		debtor.amount -= settleAmount
		creditor.amount -= settleAmount

		// Move to next debtor or creditor if settled
		if debtor.amount < 0.01 {
			i++
		}
		if creditor.amount < 0.01 {
			j++
		}
	}

	return settlements, nil
}
