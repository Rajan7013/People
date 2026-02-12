-- Migration: Add Super Admin SaaS Infrastructure - Organization Details
-- This migration creates the organization_details table for extended tenant information

CREATE TABLE IF NOT EXISTS organization_details (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID UNIQUE NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Contact Information
    admin_name VARCHAR(255),
    admin_email VARCHAR(255),
    contact_number VARCHAR(20),
    website VARCHAR(255),
    
    -- Address
    address_line1 VARCHAR(255),
    address_line2 VARCHAR(255),
    city VARCHAR(100),
    state VARCHAR(100),
    country VARCHAR(100),
    postal_code VARCHAR(20),
    
    -- Business Details
    company_registration_number VARCHAR(100),
    tax_id VARCHAR(100),
    currency VARCHAR(3) DEFAULT 'USD',
    timezone VARCHAR(50) DEFAULT 'UTC',
    
    -- Logo
    logo_url VARCHAR(500),
    
    -- Additional metadata
    metadata JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for organization_details
CREATE INDEX idx_organization_details_tenant_id ON organization_details(tenant_id);
CREATE INDEX idx_organization_details_country ON organization_details(country);
CREATE INDEX idx_organization_details_admin_email ON organization_details(admin_email);

-- Add trigger for updated_at
CREATE TRIGGER update_organization_details_updated_at 
    BEFORE UPDATE ON organization_details 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Create a comprehensive view for organization management
CREATE OR REPLACE VIEW organization_overview AS
SELECT 
    t.id as tenant_id,
    t.name as organization_name,
    t.subdomain,
    t.domain,
    t.status,
    od.admin_name,
    od.admin_email,
    od.contact_number,
    od.country,
    od.city,
    od.website,
    s.plan_id,
    sp.display_name as plan_name,
    s.status as subscription_status,
    s.current_period_end as subscription_expires_at,
    t.created_at,
    (SELECT COUNT(*) FROM users WHERE tenant_id = t.id AND deleted_at IS NULL) as total_users,
    (SELECT COUNT(*) FROM employees WHERE tenant_id = t.id AND employment_status = 'active') as active_employees
FROM tenants t
LEFT JOIN organization_details od ON t.id = od.tenant_id
LEFT JOIN subscriptions s ON t.id = s.tenant_id
LEFT JOIN subscription_plans sp ON s.plan_id = sp.id
WHERE t.deleted_at IS NULL;
