ALTER TABLE expenses ADD COLUMN transacted_at TIMESTAMPTZ;
UPDATE expenses SET transacted_at = created_at WHERE transacted_at IS NULL;
ALTER TABLE expenses ALTER COLUMN transacted_at SET NOT NULL;
ALTER TABLE expenses ALTER COLUMN transacted_at SET DEFAULT now();
