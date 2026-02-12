-- Migration 030: Add Team Support and Update Role Enum
-- This migration adds team_id to users table and updates role enum for team_lead

-- Add team_id column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS team_id UUID;

-- Create index for team_id
CREATE INDEX IF NOT EXISTS idx_users_team_id ON users(team_id);

-- Add team_lead role to enum (if not exists)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type t 
                   JOIN pg_enum e ON t.oid = e.enumtypid  
                   WHERE t.typname = 'user_role' AND e.enumlabel = 'team_lead') THEN
        ALTER TYPE user_role ADD VALUE 'team_lead';
    END IF;
END$$;

-- Allow NULL tenant_id for super_admin
ALTER TABLE users ALTER COLUMN tenant_id DROP NOT NULL;

-- Update existing super admin to have NULL tenant_id FIRST (before adding constraint)
UPDATE users 
SET tenant_id = NULL 
WHERE role = 'super_admin';

-- Add constraint: super_admin must have NULL tenant_id, others must have tenant_id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'check_super_admin_no_tenant'
    ) THEN
        ALTER TABLE users ADD CONSTRAINT check_super_admin_no_tenant 
        CHECK (
            (role = 'super_admin' AND tenant_id IS NULL) OR 
            (role != 'super_admin' AND tenant_id IS NOT NULL)
        );
    END IF;
END$$;

-- Add comments
COMMENT ON COLUMN users.team_id IS 'Team ID for team lead and team members - used for team-level access control';
COMMENT ON CONSTRAINT check_super_admin_no_tenant ON users IS 'Ensures super_admin has NULL tenant_id for platform-level access';

-- Verification query
-- SELECT email, role, tenant_id, team_id FROM users WHERE role = 'super_admin';
