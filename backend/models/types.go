package models

// Models
type User struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"` // plaintext for prototype only
	Created  int64  `json:"created"
`
}

type Group struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	OwnerID   int64   `json:"owner_id"`
	MemberIDs []int64 `json:"member_ids"`
	Created   int64   `json:"created"`
}

// Expense split can be uneven. Each split lists user and amount owed (positive number)
type Split struct {
	UserID int64   `json:"user_id"`
	Amount float64 `json:"amount"`
}

type Expense struct {
	ID          int64   `json:"id"`
	GroupID     int64   `json:"group_id"`
	Title       string  `json:"title"`
	Amount      float64 `json:"amount"`
	PaidBy      int64   `json:"paid_by"`
	Splits      []Split `json:"splits"`
	Description string  `json:"description,omitempty"`
	Created     int64   `json:"created"`
}
