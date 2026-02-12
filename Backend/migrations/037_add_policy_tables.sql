-- Migration: 037_add_policy_tables.sql
-- Description: Add tables for policy configuration (attendance, salary components, leave types)

-- Drop existing tables if they exist (for clean migration)
DROP TABLE IF EXISTS attendance_policies CASCADE;
DROP TABLE IF EXISTS leave_types CASCADE;
-- Note: salary_components might be used by payslip schema, so we'll alter it instead

-- Attendance Policies Table
CREATE TABLE attendance_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    working_hours_per_day DECIMAL(4,2) NOT NULL DEFAULT 8.0,
    grace_period_minutes INTEGER NOT NULL DEFAULT 15,
    overtime_threshold_minutes INTEGER NOT NULL DEFAULT 480, -- 8 hours
    required_hours_per_week DECIMAL(5,2) NOT NULL DEFAULT 40.0,
    required_days_per_month INTEGER NOT NULL DEFAULT 22,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(tenant_id, deleted_at) -- Only one active policy per tenant
);

-- Salary Components Table (check if exists, add missing columns)
DO $$
BEGIN
    -- Create table if it doesn't exist
    IF NOT EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'salary_components') THEN
        CREATE TABLE salary_components (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
            name VARCHAR(255) NOT NULL,
            type VARCHAR(50) NOT NULL CHECK (type IN ('earning', 'deduction')),
            is_fixed BOOLEAN NOT NULL DEFAULT true,
            amount DECIMAL(12,2) DEFAULT 0.0,
            percentage DECIMAL(5,2) DEFAULT 0.0,
            description TEXT,
            is_percentage BOOLEAN NOT NULL DEFAULT false,
            is_taxable BOOLEAN NOT NULL DEFAULT true,
            is_active BOOLEAN NOT NULL DEFAULT true,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            deleted_at TIMESTAMP WITH TIME ZONE
        );
    ELSE
        -- Add missing columns if table exists
        IF NOT EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'salary_components' AND column_name = 'is_fixed') THEN
            ALTER TABLE salary_components ADD COLUMN is_fixed BOOLEAN NOT NULL DEFAULT true;
        END IF;
        IF NOT EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'salary_components' AND column_name = 'amount') THEN
            ALTER TABLE salary_components ADD COLUMN amount DECIMAL(12,2) DEFAULT 0.0;
        END IF;
        IF NOT EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'salary_components' AND column_name = 'percentage') THEN
            ALTER TABLE salary_components ADD COLUMN percentage DECIMAL(5,2) DEFAULT 0.0;
        END IF;
        IF NOT EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'salary_components' AND column_name = 'description') THEN
            ALTER TABLE salary_components ADD COLUMN description TEXT;
        END IF;
        IF NOT EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'salary_components' AND column_name = 'deleted_at') THEN
            ALTER TABLE salary_components ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE;
        END IF;
    END IF;
END $$;

-- Leave Types Table (for policy configuration)
CREATE TABLE leave_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    annual_limit INTEGER NOT NULL DEFAULT 0,
    carry_forward_limit INTEGER NOT NULL DEFAULT 0,
    is_paid BOOLEAN NOT NULL DEFAULT true,
    requires_approval BOOLEAN NOT NULL DEFAULT true,
    accrual_rate DECIMAL(5,2) NOT NULL DEFAULT 0.0,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for performance
CREATE INDEX idx_attendance_policies_tenant ON attendance_policies(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_salary_components_tenant ON salary_components(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_leave_types_tenant ON leave_types(tenant_id) WHERE deleted_at IS NULL;

-- Enable RLS
ALTER TABLE attendance_policies ENABLE ROW LEVEL SECURITY;
ALTER TABLE salary_components ENABLE ROW LEVEL SECURITY;
ALTER TABLE leave_types ENABLE ROW LEVEL SECURITY;

-- RLS Policies for attendance_policies
CREATE POLICY attendance_policies_tenant_isolation ON attendance_policies
    FOR ALL
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

-- RLS Policies for salary_components (drop if exists, then create)
DROP POLICY IF EXISTS salary_components_tenant_isolation ON salary_components;
CREATE POLICY salary_components_tenant_isolation ON salary_components
    FOR ALL
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

-- RLS Policies for leave_types
CREATE POLICY leave_types_tenant_isolation ON leave_types
    FOR ALL
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

-- Update trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop existing triggers if they exist
DROP TRIGGER IF EXISTS update_attendance_policies_updated_at ON attendance_policies;
DROP TRIGGER IF EXISTS update_salary_components_updated_at ON salary_components;
DROP TRIGGER IF EXISTS update_leave_types_updated_at ON leave_types;

-- Create triggers
CREATE TRIGGER update_attendance_policies_updated_at
    BEFORE UPDATE ON attendance_policies
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_salary_components_updated_at
    BEFORE UPDATE ON salary_components
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_leave_types_updated_at
    BEFORE UPDATE ON leave_types
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
