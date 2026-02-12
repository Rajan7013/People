-- Migration to add specific subscription plans requested by user

-- First, deactivate existing plans to avoid confusion (optional, but safer)
UPDATE subscription_plans SET is_active = false WHERE is_active = true AND name NOT IN ('free', 'individual', 'basic', 'startup', 'pro', 'business', 'enterprise');

-- Insert new plans (idempotent - skip if already exists)
INSERT INTO subscription_plans (
    id, name, display_name, description, price_monthly, price_yearly, currency,
    max_users, max_storage_gb, max_api_requests_monthly, max_departments,
    features, is_active, is_visible, sort_order, created_at, updated_at
) VALUES 
-- Free Plan
(
    gen_random_uuid(), 'free', 'Free', 'Forever free plan for individuals.', 
    0, 0, 'USD',
    1, 1, 1000, 1,
    '{"sso": false, "audit_logs": false, "custom_domain": false}'::jsonb,
    true, true, 1, NOW(), NOW()
),
-- Individual Plan
(
    gen_random_uuid(), 'individual', 'Individual', 'Perfect for freelancers and solo entrepreneurs.', 
    9, 90, 'USD',
    1, 10, 5000, 1,
    '{"sso": false, "audit_logs": false, "custom_domain": true}'::jsonb,
    true, true, 2, NOW(), NOW()
),
-- Basic Plan
(
    gen_random_uuid(), 'basic', 'Basic', 'Great for small teams just getting started.', 
    29, 290, 'USD',
    5, 50, 20000, 3,
    '{"sso": false, "audit_logs": false, "custom_domain": true}'::jsonb,
    true, true, 3, NOW(), NOW()
),
-- Startup Plan
(
    gen_random_uuid(), 'startup', 'Startup', 'For growing startups needing more resources.', 
    49, 490, 'USD',
    10, 100, 50000, 5,
    '{"sso": false, "audit_logs": true, "custom_domain": true}'::jsonb,
    true, true, 4, NOW(), NOW()
),
-- Pro Plan
(
    gen_random_uuid(), 'pro', 'Pro', 'Advanced features for scaling companies.', 
    99, 990, 'USD',
    25, 500, 200000, 10,
    '{"sso": true, "audit_logs": true, "custom_domain": true}'::jsonb,
    true, true, 5, NOW(), NOW()
),
-- Business Plan
(
    gen_random_uuid(), 'business', 'Business', 'Everything needed for a mid-sized organization.', 
    199, 1990, 'USD',
    50, 1000, 500000, 20,
    '{"sso": true, "audit_logs": true, "custom_domain": true, "priority_support": true}'::jsonb,
    true, true, 6, NOW(), NOW()
),
-- Enterprise Plan
(
    gen_random_uuid(), 'enterprise', 'Enterprise', 'Unlimited possibilities for large enterprises.', 
    499, 4990, 'USD',
    10000, 10000, 10000000, 100,
    '{"sso": true, "audit_logs": true, "custom_domain": true, "priority_support": true, "dedicated_manager": true}'::jsonb,
    true, true, 7, NOW(), NOW()
)
ON CONFLICT (name) DO NOTHING;
