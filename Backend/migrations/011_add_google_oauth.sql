-- Make password_hash nullable for Google OAuth users
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;

-- Add google_id column if not exists
ALTER TABLE users ADD COLUMN IF NOT EXISTS google_id VARCHAR(255) UNIQUE;

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_google_id ON users(google_id);

-- Add comment
COMMENT ON COLUMN users.google_id IS 'Google OAuth user ID for authentication';
COMMENT ON COLUMN users.password_hash IS 'Password hash - nullable for OAuth users';
