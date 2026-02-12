-- Add sensitive columns for encryption
ALTER TABLE employees 
ADD COLUMN bank_account_number VARCHAR(255), -- Encrypted
ADD COLUMN national_id VARCHAR(255),         -- Encrypted
ADD COLUMN national_id_hash VARCHAR(64);     -- Blind Index (HMAC-SHA256)

-- Index for blind search
CREATE INDEX idx_employees_national_id_hash ON employees(national_id_hash);
