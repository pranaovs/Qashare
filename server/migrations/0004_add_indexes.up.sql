-- Indexes for expenses table
CREATE INDEX idx_expenses_group_created ON expenses (group_id, created_at DESC);

-- Indexes for expense_splits table
CREATE INDEX idx_expense_splits_expense ON expense_splits (expense_id);
CREATE INDEX idx_expense_splits_user_paid_expense ON expense_splits (user_id, is_paid, expense_id);

-- Indexes for group_members table
CREATE INDEX idx_group_members_group_joined ON group_members (group_id, joined_at);
CREATE INDEX idx_group_members_user_group ON group_members (user_id, group_id);

-- Indexes for groups table
CREATE INDEX idx_groups_created_by_created_at ON groups (created_by, created_at DESC);

-- Indexes for guests table
CREATE INDEX idx_guests_added_by ON guests (added_by);
