-- Update employees table columns to support encryption
-- Change salary from DECIMAL to TEXT (will store encrypted value)
-- Change phone from VARCHAR(20) to TEXT (will store encrypted value)
-- Change bank_account_number from VARCHAR(255) to TEXT (for consistency)
-- Change national_id from VARCHAR(255) to TEXT (for consistency)

ALTER TABLE employees 
ALTER COLUMN salary TYPE TEXT,
ALTER COLUMN phone TYPE TEXT,
ALTER COLUMN bank_account_number TYPE TEXT,
ALTER COLUMN national_id TYPE TEXT;
