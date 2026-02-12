-- Migration: Enable Row-Level Security (RLS) for Multi-Tenant Data Isolation
-- This migration adds RLS policies to all tenant-scoped tables to prevent
-- unauthorized cross-tenant data access

-- ============================================================================
-- ENABLE RLS ON ALL TENANT-SCOPED TABLES
-- ============================================================================

-- Core tables
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE departments ENABLE ROW LEVEL SECURITY;
ALTER TABLE employees ENABLE ROW LEVEL SECURITY;
ALTER TABLE attendance_policies ENABLE ROW LEVEL SECURITY;
ALTER TABLE attendance_records ENABLE ROW LEVEL SECURITY;
ALTER TABLE leave_types ENABLE ROW LEVEL SECURITY;
ALTER TABLE leave_applications ENABLE ROW LEVEL SECURITY;

-- SaaS tables
ALTER TABLE subscriptions ENABLE ROW LEVEL SECURITY;
ALTER TABLE invoices ENABLE ROW LEVEL SECURITY;
ALTER TABLE organization_details ENABLE ROW LEVEL SECURITY;
ALTER TABLE usage_metrics ENABLE ROW LEVEL SECURITY;
ALTER TABLE api_request_logs ENABLE ROW LEVEL SECURITY;

-- Additional tables (if they exist)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'biometric_devices') THEN
        ALTER TABLE biometric_devices ENABLE ROW LEVEL SECURITY;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'biometric_logs') THEN
        ALTER TABLE biometric_logs ENABLE ROW LEVEL SECURITY;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'payslips') THEN
        ALTER TABLE payslips ENABLE ROW LEVEL SECURITY;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'system_settings') THEN
        ALTER TABLE system_settings ENABLE ROW LEVEL SECURITY;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'user_preferences') THEN
        ALTER TABLE user_preferences ENABLE ROW LEVEL SECURITY;
    END IF;
END $$;

-- ============================================================================
-- CREATE HELPER FUNCTION TO GET CURRENT USER'S TENANT
-- ============================================================================

CREATE OR REPLACE FUNCTION current_tenant_id() RETURNS UUID AS $$
BEGIN
    -- Get tenant_id from current session variable
    -- This will be set by the application layer
    RETURN current_setting('app.current_tenant_id', TRUE)::UUID;
EXCEPTION
    WHEN OTHERS THEN
        RETURN NULL;
END;
$$ LANGUAGE plpgsql STABLE;

CREATE OR REPLACE FUNCTION current_user_role() RETURNS TEXT AS $$
BEGIN
    -- Get user role from current session variable
    RETURN current_setting('app.current_user_role', TRUE);
EXCEPTION
    WHEN OTHERS THEN
        RETURN NULL;
END;
$$ LANGUAGE plpgsql STABLE;

-- ============================================================================
-- RLS POLICIES FOR CORE TABLES
-- ============================================================================

-- USERS TABLE
-- Super admins can see all users, others only see users in their tenant
CREATE POLICY users_tenant_isolation ON users
    FOR ALL
    USING (
        current_user_role() = 'super_admin' OR
        tenant_id = current_tenant_id()
    );

-- DEPARTMENTS TABLE
CREATE POLICY departments_tenant_isolation ON departments
    FOR ALL
    USING (tenant_id = current_tenant_id());

-- EMPLOYEES TABLE
CREATE POLICY employees_tenant_isolation ON employees
    FOR ALL
    USING (tenant_id = current_tenant_id());

-- ATTENDANCE POLICIES TABLE
CREATE POLICY attendance_policies_tenant_isolation ON attendance_policies
    FOR ALL
    USING (tenant_id = current_tenant_id());

-- ATTENDANCE RECORDS TABLE
CREATE POLICY attendance_records_tenant_isolation ON attendance_records
    FOR ALL
    USING (tenant_id = current_tenant_id());

-- LEAVE TYPES TABLE
CREATE POLICY leave_types_tenant_isolation ON leave_types
    FOR ALL
    USING (tenant_id = current_tenant_id());

-- LEAVE APPLICATIONS TABLE
CREATE POLICY leave_applications_tenant_isolation ON leave_applications
    FOR ALL
    USING (tenant_id = current_tenant_id());

-- ============================================================================
-- RLS POLICIES FOR SAAS TABLES
-- ============================================================================

-- SUBSCRIPTIONS TABLE
-- Super admins can see all, tenants can only see their own
CREATE POLICY subscriptions_access ON subscriptions
    FOR ALL
    USING (
        current_user_role() = 'super_admin' OR
        tenant_id = current_tenant_id()
    );

-- INVOICES TABLE
-- Super admins can see all, tenants can only see their own
CREATE POLICY invoices_access ON invoices
    FOR ALL
    USING (
        current_user_role() = 'super_admin' OR
        tenant_id = current_tenant_id()
    );

-- ORGANIZATION DETAILS TABLE
-- Super admins can see all, tenants can only see their own
CREATE POLICY organization_details_access ON organization_details
    FOR ALL
    USING (
        current_user_role() = 'super_admin' OR
        tenant_id = current_tenant_id()
    );

-- USAGE METRICS TABLE
-- Super admins can see all, tenants can only see their own
CREATE POLICY usage_metrics_access ON usage_metrics
    FOR ALL
    USING (
        current_user_role() = 'super_admin' OR
        tenant_id = current_tenant_id()
    );

-- API REQUEST LOGS TABLE
-- Super admins can see all, tenants can only see their own
CREATE POLICY api_request_logs_access ON api_request_logs
    FOR ALL
    USING (
        current_user_role() = 'super_admin' OR
        tenant_id = current_tenant_id()
    );

-- ============================================================================
-- RLS POLICIES FOR ADDITIONAL TABLES (if they exist)
-- ============================================================================

DO $$
BEGIN
    -- BIOMETRIC DEVICES
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'biometric_devices') THEN
        EXECUTE 'CREATE POLICY biometric_devices_tenant_isolation ON biometric_devices
            FOR ALL USING (tenant_id = current_tenant_id())';
    END IF;

    -- BIOMETRIC LOGS
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'biometric_logs') THEN
        EXECUTE 'CREATE POLICY biometric_logs_tenant_isolation ON biometric_logs
            FOR ALL USING (tenant_id = current_tenant_id())';
    END IF;

    -- PAYSLIPS
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'payslips') THEN
        EXECUTE 'CREATE POLICY payslips_tenant_isolation ON payslips
            FOR ALL USING (tenant_id = current_tenant_id())';
    END IF;

    -- SYSTEM SETTINGS
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'system_settings') THEN
        EXECUTE 'CREATE POLICY system_settings_tenant_isolation ON system_settings
            FOR ALL USING (tenant_id = current_tenant_id())';
    END IF;

    -- USER PREFERENCES
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'user_preferences') THEN
        EXECUTE 'CREATE POLICY user_preferences_tenant_isolation ON user_preferences
            FOR ALL USING (tenant_id = current_tenant_id())';
    END IF;
END $$;

-- ============================================================================
-- TENANTS TABLE - Special handling
-- ============================================================================

-- Tenants table needs special policy since it doesn't have tenant_id
ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;

-- Super admins can see all tenants, regular users can only see their own tenant
CREATE POLICY tenants_access ON tenants
    FOR ALL
    USING (
        current_user_role() = 'super_admin' OR
        id = current_tenant_id()
    );

-- ============================================================================
-- SUBSCRIPTION PLANS TABLE - Public read access
-- ============================================================================

-- Subscription plans are visible to all authenticated users
ALTER TABLE subscription_plans ENABLE ROW LEVEL SECURITY;

CREATE POLICY subscription_plans_read_all ON subscription_plans
    FOR SELECT
    USING (is_visible = TRUE OR current_user_role() = 'super_admin');

-- Only super admins can modify plans
CREATE POLICY subscription_plans_modify_super_admin ON subscription_plans
    FOR ALL
    USING (current_user_role() = 'super_admin');

-- ============================================================================
-- GRANT BYPASS RLS TO APPLICATION USER
-- ============================================================================

-- The application database user should have BYPASSRLS privilege
-- This allows the application to set session variables and enforce RLS programmatically
-- Run this command as superuser:
-- ALTER USER peopleos_app BYPASSRLS;

-- Note: The application layer MUST set these session variables for each request:
-- SET LOCAL app.current_tenant_id = '<tenant_uuid>';
-- SET LOCAL app.current_user_role = '<user_role>';

-- ============================================================================
-- HELPER FUNCTION TO SET SESSION CONTEXT
-- ============================================================================

CREATE OR REPLACE FUNCTION set_session_context(p_tenant_id UUID, p_user_role TEXT)
RETURNS VOID AS $$
BEGIN
    PERFORM set_config('app.current_tenant_id', p_tenant_id::TEXT, TRUE);
    PERFORM set_config('app.current_user_role', p_user_role, TRUE);
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- VERIFICATION QUERIES
-- ============================================================================

-- To verify RLS is working, run these queries:

-- 1. Check which tables have RLS enabled:
-- SELECT schemaname, tablename, rowsecurity
-- FROM pg_tables
-- WHERE schemaname = 'public' AND rowsecurity = true;

-- 2. View all RLS policies:
-- SELECT schemaname, tablename, policyname, permissive, roles, cmd, qual
-- FROM pg_policies
-- WHERE schemaname = 'public';

-- 3. Test RLS as a tenant user:
-- SELECT set_session_context('<tenant_uuid>', 'admin');
-- SELECT * FROM employees; -- Should only see employees from that tenant

-- 4. Test RLS as super admin:
-- SELECT set_session_context(NULL, 'super_admin');
-- SELECT * FROM tenants; -- Should see all tenants

COMMENT ON FUNCTION current_tenant_id() IS 'Returns the current tenant ID from session variable';
COMMENT ON FUNCTION current_user_role() IS 'Returns the current user role from session variable';
COMMENT ON FUNCTION set_session_context(UUID, TEXT) IS 'Sets session variables for RLS enforcement';
