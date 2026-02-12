-- Add team_lead to user_role enum
-- We use IF NOT EXISTS to avoid errors if it was partially applied before, 
-- though standard postgres ALTER TYPE doesn't have IF NOT EXISTS for values.
-- We'll just run the ALTER command. If it fails due to existence, it's fine.

ALTER TYPE user_role ADD VALUE IF NOT EXISTS 'team_lead';
