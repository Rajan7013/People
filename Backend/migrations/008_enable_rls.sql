-- Enable Row Level Security (RLS) on all tenant-scoped tables
-- Version: 008 (Renumbered from 011)
-- This ensures that a user can ONLY access data belonging to their specific tenant.
-- The application must set the configuration parameter 'app.current_tenant' 
-- at the beginning of each transaction/request.

-- =====================================================================================
-- 1. Enable RLS on Tables
-- =====================================================================================

-- Core & Auth
ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE employees ENABLE ROW LEVEL SECURITY;
ALTER TABLE departments ENABLE ROW LEVEL SECURITY;

-- Attendance & Leave
ALTER TABLE attendance_policies ENABLE ROW LEVEL SECURITY;
ALTER TABLE attendance_records ENABLE ROW LEVEL SECURITY;
ALTER TABLE leave_types ENABLE ROW LEVEL SECURITY;
ALTER TABLE leave_balances ENABLE ROW LEVEL SECURITY;
ALTER TABLE leave_applications ENABLE ROW LEVEL SECURITY;
ALTER TABLE leave_requests ENABLE ROW LEVEL SECURITY; 

-- System & Settings
ALTER TABLE system_settings ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE system_backups ENABLE ROW LEVEL SECURITY;
ALTER TABLE system_metrics ENABLE ROW LEVEL SECURITY;
ALTER TABLE api_keys ENABLE ROW LEVEL SECURITY;
ALTER TABLE webhooks ENABLE ROW LEVEL SECURITY;

-- User Profile & Preferences
ALTER TABLE user_profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_preferences ENABLE ROW LEVEL SECURITY;
ALTER TABLE security_settings ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_themes ENABLE ROW LEVEL SECURITY;

-- Payslips & Payroll
ALTER TABLE salary_components ENABLE ROW LEVEL SECURITY;
ALTER TABLE employee_salary_structures ENABLE ROW LEVEL SECURITY;
ALTER TABLE salary_structure_components ENABLE ROW LEVEL SECURITY; 
ALTER TABLE payslips ENABLE ROW LEVEL SECURITY;
ALTER TABLE payslip_components ENABLE ROW LEVEL SECURITY; 

-- Biometrics
ALTER TABLE biometric_devices ENABLE ROW LEVEL SECURITY;
ALTER TABLE biometric_attendance_logs ENABLE ROW LEVEL SECURITY;


-- =====================================================================================
-- 2. Create Policies for Standard Tables (Direct tenant_id check)
-- =====================================================================================

-- Function to get current tenant safely
CREATE OR REPLACE FUNCTION current_app_tenant() RETURNS uuid AS $$
    SELECT NULLIF(current_setting('app.current_tenant', TRUE), '')::uuid;
$$ LANGUAGE sql STABLE;

-- Generic Policy Template applied to each table manually
-- (Postgres doesn't support "apply to all", so we list them)

-- Core
CREATE POLICY tenant_isolation_tenants ON tenants
    USING (id = current_app_tenant());

CREATE POLICY tenant_isolation_users ON users
    USING (tenant_id = current_app_tenant());

CREATE POLICY tenant_isolation_employees ON employees
    USING (tenant_id = current_app_tenant());

CREATE POLICY tenant_isolation_departments ON departments
    USING (tenant_id = current_app_tenant());

-- Attendance
CREATE POLICY tenant_isolation_att_policies ON attendance_policies
    USING (tenant_id = current_app_tenant());

CREATE POLICY tenant_isolation_att_records ON attendance_records
    USING (tenant_id = current_app_tenant());

-- Leave
CREATE POLICY tenant_isolation_leave_types ON leave_types
    USING (tenant_id = current_app_tenant());

CREATE POLICY tenant_isolation_leave_balances ON leave_balances
    USING (tenant_id = current_app_tenant());

CREATE POLICY tenant_isolation_leave_applications ON leave_applications
    USING (tenant_id = current_app_tenant());

CREATE POLICY tenant_isolation_leave_requests ON leave_requests
    USING (tenant_id = current_app_tenant());

-- System
CREATE POLICY tenant_isolation_sys_settings ON system_settings
    USING (tenant_id = current_app_tenant());

CREATE POLICY tenant_isolation_audit_logs ON audit_logs
    USING (tenant_id = current_app_tenant());

CREATE POLICY tenant_isolation_sys_backups ON system_backups
    USING (tenant_id = current_app_tenant());

CREATE POLICY tenant_isolation_sys_metrics ON system_metrics
    USING (tenant_id = current_app_tenant());

CREATE POLICY tenant_isolation_api_keys ON api_keys
    USING (tenant_id = current_app_tenant());

CREATE POLICY tenant_isolation_webhooks ON webhooks
    USING (tenant_id = current_app_tenant());

-- User Settings
CREATE POLICY tenant_isolation_user_profiles ON user_profiles
    USING (tenant_id = current_app_tenant());

CREATE POLICY tenant_isolation_user_prefs ON user_preferences
    USING (tenant_id = current_app_tenant());

CREATE POLICY tenant_isolation_sec_settings ON security_settings
    USING (tenant_id = current_app_tenant());

CREATE POLICY tenant_isolation_user_themes ON user_themes
    USING (tenant_id = current_app_tenant());

-- Payroll
CREATE POLICY tenant_isolation_salary_comps ON salary_components
    USING (tenant_id = current_app_tenant());

CREATE POLICY tenant_isolation_emp_salary_structs ON employee_salary_structures
    USING (tenant_id = current_app_tenant());

CREATE POLICY tenant_isolation_payslips ON payslips
    USING (tenant_id = current_app_tenant());

-- Biometrics
CREATE POLICY tenant_isolation_bio_devices ON biometric_devices
    USING (tenant_id = current_app_tenant());

CREATE POLICY tenant_isolation_bio_logs ON biometric_attendance_logs
    USING (tenant_id = current_app_tenant());

-- =====================================================================================
-- 3. Create JOIN Policies for Child Tables (No direct tenant_id)
-- =====================================================================================

-- salary_structure_components -> belongs_to employee_salary_structures(tenant_id)
CREATE POLICY tenant_isolation_sal_struct_comps ON salary_structure_components
    USING (
        salary_structure_id IN (
            SELECT id FROM employee_salary_structures 
            WHERE tenant_id = current_app_tenant()
        )
    );

-- payslip_components -> belongs_to payslips(tenant_id)
CREATE POLICY tenant_isolation_payslip_comps ON payslip_components
    USING (
        payslip_id IN (
            SELECT id FROM payslips 
            WHERE tenant_id = current_app_tenant()
        )
    );
