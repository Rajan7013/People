-- Migration: Make password_hash nullable
-- This allows users to register/login via OAuth (e.g., Google) without a password

ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;

-- Comment
COMMENT ON COLUMN users.password_hash IS 'Hashed password for email/password auth. NULL for OAuth users.';
