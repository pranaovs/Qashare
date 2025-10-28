-- USERS
CREATE TABLE IF NOT EXISTS users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT,
    is_guest BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- GROUPS
CREATE TABLE IF NOT EXISTS groups (
    group_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_name TEXT NOT NULL,
    description TEXT,
    created_by UUID REFERENCES users (user_id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- GROUP MEMBERS
CREATE TABLE IF NOT EXISTS group_members (
    user_id UUID REFERENCES users (user_id) ON DELETE CASCADE,
    group_id UUID REFERENCES groups (group_id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (user_id, group_id)
);

-- EXPENSES
CREATE TABLE IF NOT EXISTS expenses (
    expense_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID REFERENCES groups (group_id) ON DELETE CASCADE,
    added_by UUID REFERENCES users (user_id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    amount DOUBLE PRECISION NOT NULL,
    is_incomplete_amount BOOLEAN DEFAULT FALSE,
    is_incomplete_split BOOLEAN DEFAULT FALSE,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION
);

-- EXPENSE SPLITS
CREATE TABLE IF NOT EXISTS expense_splits (
    expense_id UUID REFERENCES expenses (expense_id) ON DELETE CASCADE,
    user_id UUID REFERENCES users (user_id) ON DELETE CASCADE,
    amount DOUBLE PRECISION NOT NULL,
    user_role TEXT CHECK (user_role IN ('paid', 'owes')),
    PRIMARY KEY (expense_id, user_id)
);
