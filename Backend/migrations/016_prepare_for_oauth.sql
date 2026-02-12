-- Check current state
SELECT id, email, google_id, role FROM users WHERE email = 'rajanprasaila@gmail.com';

-- The google_id will be in the backend logs when you try to login
-- Look for the error message and extract the google_id from the Google user info

-- For now, let's just ensure the account is ready to accept a google_id
-- The backend code should automatically link it on next login attempt

-- Make sure password_hash is NULL (already done based on screenshot)
UPDATE users 
SET password_hash = NULL,
    email_verified_at = COALESCE(email_verified_at, NOW())
WHERE email = 'rajanprasaila@gmail.com';

-- Verify
SELECT id, email, google_id, role, password_hash IS NULL as ready_for_oauth 
FROM users 
WHERE email = 'rajanprasaila@gmail.com';
