-- First, let's see if the user exists and what their current state is
SELECT id, email, google_id, role, is_active 
FROM users 
WHERE email = 'rajanprasaila@gmail.com';

-- If the user exists without google_id, we need to link their Google account
-- You'll need to get your Google ID from the next login attempt
-- The backend logs will show: "Failed to find or create user"
-- But we can also just let the user try to login and capture the google_id from logs

-- For now, let's make sure the user is active and has super_admin role
UPDATE users 
SET role = 'super_admin', 
    is_active = true,
    password_hash = NULL  -- Remove password requirement for OAuth
WHERE email = 'rajanprasaila@gmail.com';
