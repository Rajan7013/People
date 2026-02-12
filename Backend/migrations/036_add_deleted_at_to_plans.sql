-- Add deleted_at column to subscription_plans to support soft delete / revival logic
ALTER TABLE subscription_plans ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;
CREATE INDEX IF NOT EXISTS idx_subscription_plans_deleted_at ON subscription_plans(deleted_at);
