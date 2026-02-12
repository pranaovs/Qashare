-- Add is_settlement column to expenses
ALTER TABLE expenses ADD COLUMN is_settlement BOOLEAN NOT NULL DEFAULT FALSE;
