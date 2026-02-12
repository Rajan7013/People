-- Fix the deleted_at issue for rajanprasaila@gmail.com
UPDATE users 
SET deleted_at = NULL
WHERE email = 'rajanprasaila@gmail.com';

-- Verify the fix
SELECT id, email, google_id, role, deleted_at, deleted_at IS NULL as is_active
FROM users 
WHERE email = 'rajanprasaila@gmail.com';
