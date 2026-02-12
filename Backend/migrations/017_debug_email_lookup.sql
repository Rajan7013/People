-- Check exact email match
SELECT id, email, google_id, role, deleted_at 
FROM users 
WHERE email = 'rajanprasaila@gmail.com';

-- Check with LOWER to see if there's a case issue
SELECT id, email, google_id, role, deleted_at 
FROM users 
WHERE LOWER(email) = LOWER('rajanprasaila@gmail.com');

-- Check if deleted_at is causing the issue
SELECT id, email, google_id, role, deleted_at, deleted_at IS NULL as is_active_user
FROM users 
WHERE email = 'rajanprasaila@gmail.com';

-- Show all columns to see what might be different
SELECT * 
FROM users 
WHERE email = 'rajanprasaila@gmail.com';
