package db

import (
	"context"
	"sort"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pranaovs/qashare/models"
	"github.com/pranaovs/qashare/utils"
)

// GetSettlements calculates the net balance between the current user and all other group members.
// It analyzes all expenses in a group and determines who owes whom, then optimizes the settlements
// using a debt minimization algorithm.
//
// Returns a slice of optimized Settlement where each entry represents a single payment:
//   - UserID: Who the current user needs to interact with (pay or receive from)
//   - Amount: Transaction amount
//   - Positive: Current user receives from UserID
//   - Negative: Current user pays to UserID
//
// Uses greedy algorithm to minimize number of transactions while settling all debts.
func GetSettlements(ctx context.Context, pool *pgxpool.Pool, userID, groupID string) ([]models.Settlement, error) {
	// Validate input
	if groupID == "" {
		return nil, ErrInvalidInput.Msg("group id missing")
	}
	if userID == "" {
		return nil, ErrInvalidInput.Msg("user id missing")
	}

	// Query to get all payer-debtor pairs in the group
	query := `
	SELECT
	  es_payer.user_id as payer_id,
	  es_debtor.user_id as debtor_id,
	  es_debtor.amount
	FROM expense_splits es_payer
	JOIN expense_splits es_debtor ON es_payer.expense_id = es_debtor.expense_id
	JOIN expenses e ON e.expense_id = es_payer.expense_id
	WHERE e.group_id = $1
	  AND es_payer.is_paid = true
	  AND es_debtor.is_paid = false
	ORDER BY es_payer.user_id, es_debtor.user_id
	`

	rows, err := pool.Query(ctx, query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Collect all payer-debtor relationships
	balances := make(map[string]float64)

	for rows.Next() {
		var payer, debtor string
		var amount float64

		err = rows.Scan(&payer, &debtor, &amount)
		if err != nil {
			return nil, err
		}

		balances[payer] += amount  // Payer gets money
		balances[debtor] -= amount // Debtor owes money
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Optimize settlements using greedy algorithm
	optimized := optimizeSettlements(balances, userID)

	return optimized, nil
}

// optimizeSettlements uses greedy algorithm to minimize transactions
// Returns settlements for the given user
func optimizeSettlements(balances map[string]float64, userID string) []models.Settlement {
	if len(balances) == 0 {
		return []models.Settlement{}
	}

	// Separate users into creditors (positive) and debtors (negative)
	var creditors []struct {
		userID string
		amount float64
	}
	var debtors []struct {
		userID string
		amount float64
	}

	tolerance, err := strconv.ParseFloat(utils.GetEnv("SPLIT_TOLERANCE", "0.01"), 64)
	if err != nil {
		tolerance = 0.01
	}

	for uid, balance := range balances {
		if balance > tolerance {
			creditors = append(creditors, struct {
				userID string
				amount float64
			}{uid, balance})
		} else if balance < -tolerance {
			debtors = append(debtors, struct {
				userID string
				amount float64
			}{uid, -balance})
		}
	}

	// Sort by amount descending for optimal greedy matching
	sort.Slice(creditors, func(i, j int) bool {
		return creditors[i].amount > creditors[j].amount
	})
	sort.Slice(debtors, func(i, j int) bool {
		return debtors[i].amount > debtors[j].amount
	})

	// Greedy matching: pair largest debtors with largest creditors
	var settlements []models.Settlement

	for len(debtors) > 0 && len(creditors) > 0 {
		debtor := debtors[0]
		creditor := creditors[0]

		// Transfer minimum of debtor's obligation and creditor's claim
		transfer := debtor.amount
		if creditor.amount < transfer {
			transfer = creditor.amount
		}

		// Record settlement based on relationship to userID
		if debtor.userID == userID {
			// Current user owes, so negative amount
			settlements = append(settlements, models.Settlement{
				UserID: creditor.userID,
				Amount: -transfer,
			})
		} else if creditor.userID == userID {
			// Current user is owed, so positive amount
			settlements = append(settlements, models.Settlement{
				UserID: debtor.userID,
				Amount: transfer,
			})
		}

		// Update remaining balances
		debtors[0].amount -= transfer
		creditors[0].amount -= transfer

		// Remove settled users
		if debtors[0].amount < tolerance {
			debtors = debtors[1:]
		}
		if creditors[0].amount < tolerance {
			creditors = creditors[1:]
		}
	}

	return settlements
}
