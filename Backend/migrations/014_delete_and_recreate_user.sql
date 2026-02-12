-- SOLUTION 1: Delete the problematic user and let Google OAuth create it fresh
-- First, check what exists
SELECT id, email, google_id, role, is_active FROM users WHERE email = 'rajanprasaila@gmail.com';

-- Delete the old user (if exists)
DELETE FROM users WHERE email = 'rajanprasaila@gmail.com';

-- Verify deletion
SELECT COUNT(*) FROM users WHERE email = 'rajanprasaila@gmail.com';

-- After running this, login with Google and it will create a fresh account
-- Then run the next migration to make it super_admin
