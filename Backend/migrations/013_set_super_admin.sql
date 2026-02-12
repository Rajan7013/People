-- Update rajanprasaila@gmail.com to super_admin role
UPDATE users 
SET role = 'super_admin',
    is_active = true
WHERE email = 'rajanprasaila@gmail.com';

-- Verify the update
SELECT id, email, role, is_active, google_id, created_at
FROM users 
WHERE email = 'rajanprasaila@gmail.com';
