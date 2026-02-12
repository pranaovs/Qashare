-- Change monetary columns from DOUBLE PRECISION to NUMERIC for exact decimal arithmetic.
-- This prevents floating-point rounding errors in settlement calculations
-- (e.g. proportional debt distribution involving division and multiplication).
-- NUMERIC(19,4) supports up to 15 integer digits with 4 decimal places.

ALTER TABLE expenses ALTER COLUMN amount TYPE NUMERIC(19,4);
ALTER TABLE expense_splits ALTER COLUMN amount TYPE NUMERIC(19,4);
