-- Check if deleted_at is the issue
SELECT 
    id, 
    email, 
    google_id, 
    role, 
    deleted_at,
    deleted_at IS NULL as is_not_deleted,
    CASE 
        WHEN deleted_at IS NULL THEN 'User should be found'
        ELSE 'User will NOT be found (deleted_at is not NULL)'
    END as query_result
FROM users 
WHERE email = 'rajanprasaila@gmail.com';

-- Also check all users to see the deleted_at pattern
SELECT email, deleted_at, deleted_at IS NULL as is_active
FROM users
ORDER BY created_at DESC
LIMIT 5;
