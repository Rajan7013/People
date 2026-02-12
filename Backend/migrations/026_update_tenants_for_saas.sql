-- Migration: Update Tenants Table for SaaS Features
-- This migration adds missing fields to the tenants table for Super Admin management

-- Add new columns to tenants table
ALTER TABLE tenants 
    ADD COLUMN IF NOT EXISTS country VARCHAR(100),
    ADD COLUMN IF NOT EXISTS admin_email VARCHAR(255),
    ADD COLUMN IF NOT EXISTS storage_used_mb BIGINT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS api_requests_count INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS last_activity_at TIMESTAMP WITH TIME ZONE;

-- Remove the old plan column (we now use subscriptions table)
-- First, let's check if there are any tenants using the old plan field
-- and migrate them to the new subscription system

-- Create a temporary function to migrate existing plans to subscriptions
CREATE OR REPLACE FUNCTION migrate_tenant_plans()
RETURNS void AS $$
DECLARE
    tenant_record RECORD;
    free_plan_id UUID;
    basic_plan_id UUID;
    pro_plan_id UUID;
    enterprise_plan_id UUID;
BEGIN
    -- Get plan IDs
    SELECT id INTO free_plan_id FROM subscription_plans WHERE name = 'free';
    SELECT id INTO basic_plan_id FROM subscription_plans WHERE name = 'basic';
    SELECT id INTO pro_plan_id FROM subscription_plans WHERE name = 'pro';
    SELECT id INTO enterprise_plan_id FROM subscription_plans WHERE name = 'enterprise';
    
    -- Migrate existing tenants to subscriptions
    FOR tenant_record IN 
        SELECT id, plan, created_at 
        FROM tenants 
        WHERE id NOT IN (SELECT tenant_id FROM subscriptions)
    LOOP
        -- Determine which plan to assign based on old plan field
        INSERT INTO subscriptions (
            tenant_id,
            plan_id,
            status,
            billing_cycle,
            amount,
            currency,
            current_period_start,
            current_period_end,
            auto_renew
        ) VALUES (
            tenant_record.id,
            CASE 
                WHEN tenant_record.plan = 'basic' THEN basic_plan_id
                WHEN tenant_record.plan = 'pro' THEN pro_plan_id
                WHEN tenant_record.plan = 'enterprise' THEN enterprise_plan_id
                ELSE free_plan_id
            END,
            'active',
            'monthly',
            CASE 
                WHEN tenant_record.plan = 'basic' THEN 29.00
                WHEN tenant_record.plan = 'pro' THEN 99.00
                WHEN tenant_record.plan = 'enterprise' THEN 299.00
                ELSE 0.00
            END,
            'USD',
            tenant_record.created_at,
            tenant_record.created_at + INTERVAL '1 month',
            true
        );
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Execute the migration
SELECT migrate_tenant_plans();

-- Drop the migration function (no longer needed)
DROP FUNCTION migrate_tenant_plans();

-- Now we can safely drop the old plan column
ALTER TABLE tenants DROP COLUMN IF EXISTS plan;

-- Create indexes for new columns
CREATE INDEX IF NOT EXISTS idx_tenants_country ON tenants(country);
CREATE INDEX IF NOT EXISTS idx_tenants_admin_email ON tenants(admin_email);
CREATE INDEX IF NOT EXISTS idx_tenants_last_activity ON tenants(last_activity_at);

-- Add a check constraint for storage
ALTER TABLE tenants 
    ADD CONSTRAINT chk_storage_non_negative 
    CHECK (storage_used_mb >= 0);

-- Add a check constraint for API requests
ALTER TABLE tenants 
    ADD CONSTRAINT chk_api_requests_non_negative 
    CHECK (api_requests_count >= 0);

-- Create a function to update tenant last activity
CREATE OR REPLACE FUNCTION update_tenant_last_activity()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE tenants 
    SET last_activity_at = CURRENT_TIMESTAMP 
    WHERE id = NEW.tenant_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers to update last_activity_at
CREATE TRIGGER update_tenant_activity_on_user_login
    AFTER UPDATE OF last_login_at ON users
    FOR EACH ROW
    WHEN (NEW.last_login_at IS DISTINCT FROM OLD.last_login_at)
    EXECUTE FUNCTION update_tenant_last_activity();

-- Comment on new columns
COMMENT ON COLUMN tenants.country IS 'Country where the organization is located';
COMMENT ON COLUMN tenants.admin_email IS 'Primary admin contact email for the organization';
COMMENT ON COLUMN tenants.storage_used_mb IS 'Total storage used by the tenant in megabytes';
COMMENT ON COLUMN tenants.api_requests_count IS 'Total API requests made by the tenant (current period)';
COMMENT ON COLUMN tenants.last_activity_at IS 'Timestamp of last activity by any user in this tenant';
