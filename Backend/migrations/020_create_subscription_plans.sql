-- Migration: Add Super Admin SaaS Infrastructure - Subscription Plans
-- This migration creates the subscription_plans table for defining pricing tiers

CREATE TABLE IF NOT EXISTS subscription_plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Pricing
    price_monthly DECIMAL(10,2) NOT NULL DEFAULT 0,
    price_yearly DECIMAL(10,2) NOT NULL DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'USD',
    
    -- Limits
    max_users INTEGER,
    max_storage_gb INTEGER,
    max_api_requests_monthly INTEGER,
    max_departments INTEGER,
    
    -- Features (JSONB for flexibility)
    -- Example: {"biometric_support": true, "advanced_reports": true, "api_access": true}
    features JSONB DEFAULT '{}',
    
    -- Status
    is_active BOOLEAN DEFAULT TRUE,
    is_visible BOOLEAN DEFAULT TRUE,
    sort_order INTEGER DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for subscription_plans
CREATE INDEX idx_subscription_plans_name ON subscription_plans(name);
CREATE INDEX idx_subscription_plans_active ON subscription_plans(is_active);
CREATE INDEX idx_subscription_plans_visible ON subscription_plans(is_visible, sort_order);

-- Add trigger for updated_at
CREATE TRIGGER update_subscription_plans_updated_at 
    BEFORE UPDATE ON subscription_plans 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Insert default plans
INSERT INTO subscription_plans (name, display_name, description, price_monthly, price_yearly, max_users, max_storage_gb, max_api_requests_monthly, max_departments, features, sort_order) VALUES
('free', 'Free', 'Perfect for trying out PeopleOS', 0, 0, 5, 1, 1000, 2, '{"basic_hr": true, "attendance": true, "leave_management": true}', 1),
('basic', 'Basic', 'For small teams getting started', 29, 290, 25, 10, 10000, 5, '{"basic_hr": true, "attendance": true, "leave_management": true, "reports": true}', 2),
('pro', 'Professional', 'For growing organizations', 99, 990, 100, 50, 50000, 20, '{"basic_hr": true, "attendance": true, "leave_management": true, "reports": true, "biometric_support": true, "advanced_analytics": true}', 3),
('enterprise', 'Enterprise', 'For large organizations with custom needs', 299, 2990, NULL, NULL, NULL, NULL, '{"basic_hr": true, "attendance": true, "leave_management": true, "reports": true, "biometric_support": true, "advanced_analytics": true, "api_access": true, "custom_integrations": true, "dedicated_support": true}', 4);
