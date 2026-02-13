package db

import (
	"context"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pranaovs/qashare/models"
)

// GetSettlement calculates the net balance between the current user and all other group members.
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
func GetSettlement(ctx context.Context, pool *pgxpool.Pool, userID, groupID uuid.UUID, splitTolerance float64) ([]models.Settlement, error) {
	// Validate input
	if groupID == "" {
		return nil, ErrInvalidInput.Msg("group id missing")
	}
	if userID == "" {
		return nil, ErrInvalidInput.Msg("user id missing")
	}

	// Query to calculate proportional debt distribution when multiple payers exist.
	// Accumulation is done in PostgreSQL using NUMERIC precision to avoid
	// floating-point errors that would occur if summed in Go with float64.
	query := `
	WITH expense_totals AS (
	  SELECT
	    expense_id,
	    SUM(amount) as total_paid
	  FROM expense_splits
	  WHERE is_paid = true
	  GROUP BY expense_id
	),
	proportional_debts AS (
	  SELECT
	    es_payer.user_id as payer_id,
	    es_debtor.user_id as debtor_id,
	    es_debtor.amount * (es_payer.amount / et.total_paid) as proportional_amount
	  FROM expense_splits es_payer
	  JOIN expense_splits es_debtor ON es_payer.expense_id = es_debtor.expense_id
	  JOIN expenses e ON e.expense_id = es_payer.expense_id
	  JOIN expense_totals et ON et.expense_id = es_payer.expense_id
	  WHERE e.group_id = $1
	    AND es_payer.is_paid = true
	    AND es_debtor.is_paid = false
	    AND es_payer.user_id != es_debtor.user_id
	    AND et.total_paid > 0
	)
	SELECT user_id, SUM(balance)::float8 AS net_balance
	FROM (
	  SELECT payer_id AS user_id, SUM(proportional_amount) AS balance
	  FROM proportional_debts GROUP BY payer_id
	  UNION ALL
	  SELECT debtor_id AS user_id, -SUM(proportional_amount) AS balance
	  FROM proportional_debts GROUP BY debtor_id
	) AS net
	GROUP BY user_id
	`

	rows, err := pool.Query(ctx, query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Net balances are already accumulated in NUMERIC by PostgreSQL
	balances := make(map[uuid.UUID]float64)

	for rows.Next() {
		var userID uuid.UUID
		var balance float64

		err = rows.Scan(&userID, &balance)
		if err != nil {
			return nil, err
		}

		balances[userID] = balance
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Step 3: Optimize settlements to minimize transactions
	optimized := optimizeSettlements(balances, userID, splitTolerance)

	return optimized, nil
}

// optimizeSettlements uses greedy algorithm to minimize transactions
// Returns settlements for the given user
func optimizeSettlements(balances map[uuid.UUID]float64, userID uuid.UUID, tolerance float64) []models.Settlement {
	if len(balances) == 0 {
		return []models.Settlement{}
	}

	// Separate users into creditors (positive) and debtors (negative)
	var creditors []struct {
		userID uuid.UUID
		amount float64
	}
	var debtors []struct {
		userID uuid.UUID
		amount float64
	}

	for uid, balance := range balances {
		if balance > tolerance {
			creditors = append(creditors, struct {
				userID uuid.UUID
				amount float64
			}{uid, balance})
		} else if balance < -tolerance {
			debtors = append(debtors, struct {
				userID uuid.UUID
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
	settlements := make([]models.Settlement, 0)

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

// GetSettlements retrieves all settlement expenses in a group where the
// specified user is a participant (either payer or receiver).
// Returns a slice of ExpenseDetails ordered by creation time descending.
func GetSettlements(ctx context.Context, pool *pgxpool.Pool, userID, groupID uuid.UUID) ([]models.ExpenseDetails, error) {
	if groupID == "" {
		return nil, ErrInvalidInput.Msg("group id missing")
	}
	if userID == "" {
		return nil, ErrInvalidInput.Msg("user id missing")
	}

	query := `
		SELECT e.expense_id, e.group_id, e.added_by, e.title, e.description,
			extract(epoch from e.created_at)::bigint, e.amount,
			e.is_incomplete_amount, e.is_incomplete_split, e.is_settlement,
			e.latitude, e.longitude,
			es.user_id, es.amount, es.is_paid
		FROM expenses e
		JOIN expense_splits es ON e.expense_id = es.expense_id
		WHERE e.group_id = $1
			AND e.is_settlement = true
			AND e.expense_id IN (
				SELECT expense_id FROM expense_splits WHERE user_id = $2
			)
		ORDER BY e.created_at DESC, es.is_paid DESC, es.user_id`

	rows, err := pool.Query(ctx, query, groupID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	expenseMap := make(map[uuid.UUID]*models.ExpenseDetails)
	var order []uuid.UUID

	for rows.Next() {
		var exp models.Expense
		var splitUserID *uuid.UUID
		var splitAmount *float64
		var splitIsPaid *bool

		err = rows.Scan(
			&exp.ExpenseID, &exp.GroupID, &exp.AddedBy, &exp.Title,
			&exp.Description, &exp.CreatedAt, &exp.Amount,
			&exp.IsIncompleteAmount, &exp.IsIncompleteSplit, &exp.IsSettlement,
			&exp.Latitude, &exp.Longitude,
			&splitUserID, &splitAmount, &splitIsPaid,
		)
		if err != nil {
			return nil, err
		}

		if _, exists := expenseMap[exp.ExpenseID]; !exists {
			expenseMap[exp.ExpenseID] = &models.ExpenseDetails{
				Expense: exp,
				Splits:  make([]models.ExpenseSplit, 0),
			}
			order = append(order, exp.ExpenseID)
		}

		if splitUserID != nil {
			expenseMap[exp.ExpenseID].Splits = append(expenseMap[exp.ExpenseID].Splits, models.ExpenseSplit{
				ExpenseID: exp.ExpenseID,
				UserID:    *splitUserID,
				Amount:    *splitAmount,
				IsPaid:    *splitIsPaid,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	results := make([]models.ExpenseDetails, 0, len(order))
	for _, id := range order {
		results = append(results, *expenseMap[id])
	}

	return results, nil
}
