-- Migration: Add Super Admin SaaS Infrastructure - Subscriptions
-- This migration creates the subscriptions table linking tenants to plans

CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES subscription_plans(id),
    
    -- Billing
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('trial', 'active', 'past_due', 'cancelled', 'expired')),
    billing_cycle VARCHAR(20) DEFAULT 'monthly' CHECK (billing_cycle IN ('monthly', 'yearly')),
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    
    -- Dates
    trial_ends_at TIMESTAMP WITH TIME ZONE,
    current_period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    current_period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    cancelled_at TIMESTAMP WITH TIME ZONE,
    
    -- Auto-renewal
    auto_renew BOOLEAN DEFAULT TRUE,
    
    -- Metadata
    notes TEXT,
    metadata JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- One active subscription per tenant
    UNIQUE(tenant_id)
);

-- Create indexes for subscriptions
CREATE INDEX idx_subscriptions_tenant_id ON subscriptions(tenant_id);
CREATE INDEX idx_subscriptions_plan_id ON subscriptions(plan_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_period_end ON subscriptions(current_period_end);

-- Add trigger for updated_at
CREATE TRIGGER update_subscriptions_updated_at 
    BEFORE UPDATE ON subscriptions 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Create a view for active subscriptions with plan details
CREATE OR REPLACE VIEW active_subscriptions AS
SELECT 
    s.id,
    s.tenant_id,
    t.name as tenant_name,
    s.plan_id,
    sp.name as plan_name,
    sp.display_name as plan_display_name,
    s.status,
    s.billing_cycle,
    s.amount,
    s.currency,
    s.current_period_start,
    s.current_period_end,
    s.auto_renew,
    CASE 
        WHEN s.current_period_end < CURRENT_TIMESTAMP THEN true
        ELSE false
    END as is_expired
FROM subscriptions s
JOIN tenants t ON s.tenant_id = t.id
JOIN subscription_plans sp ON s.plan_id = sp.id
WHERE s.status IN ('trial', 'active', 'past_due');
