-- After logging in with Google, run this to make yourself super_admin
UPDATE users 
SET role = 'super_admin'
WHERE email = 'rajanprasaila@gmail.com';

-- Verify
SELECT id, email, role, google_id FROM users WHERE email = 'rajanprasaila@gmail.com';
