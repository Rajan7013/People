-- Check all users with rajanprasaila@gmail.com
SELECT id, email, google_id, role, is_active, password_hash IS NOT NULL as has_password, created_at
FROM users 
WHERE email LIKE '%rajanprasaila%'
ORDER BY created_at DESC;

-- Check if there are multiple entries
SELECT email, COUNT(*) as count
FROM users
WHERE email LIKE '%rajanprasaila%'
GROUP BY email;
