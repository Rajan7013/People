-- Migration 032: Granular RLS Policies for Role-Based Access Control
-- This migration replaces broad tenant-isolation policies with granular role-based policies

-- ============================================================================
-- DROP EXISTING BROAD POLICIES
-- ============================================================================

DROP POLICY IF EXISTS users_tenant_isolation ON users;
DROP POLICY IF EXISTS employees_tenant_isolation ON employees;
DROP POLICY IF EXISTS attendance_records_tenant_isolation ON attendance_records;
DROP POLICY IF EXISTS leave_applications_tenant_isolation ON leave_applications;

-- ============================================================================
-- USERS TABLE - Granular Policies
-- ============================================================================

-- Policy 1: Super Admin sees all users
CREATE POLICY users_super_admin ON users
    FOR ALL
    USING (current_user_role() = 'super_admin');

-- Policy 2: Org Admin sees all users in their tenant
CREATE POLICY users_org_admin ON users
    FOR ALL
    USING (
        current_user_role() = 'admin' AND
        tenant_id = current_tenant_id()
    );

-- Policy 3: Manager sees only their department
CREATE POLICY users_manager_department ON users
    FOR SELECT
    USING (
        current_user_role() = 'manager' AND
        tenant_id = current_tenant_id() AND
        (
            -- Users in manager's department
            id IN (
                SELECT e.user_id 
                FROM employees e 
                WHERE e.department_id = current_user_department()
            )
            OR
            -- Manager can see themselves
            id = current_user_id()
        )
    );

-- Policy 4: HR sees employees (not managers/admins/other HR)
CREATE POLICY users_hr_employees ON users
    FOR SELECT
    USING (
        current_user_role() = 'hr' AND
        tenant_id = current_tenant_id() AND
        role IN ('employee', 'team_lead')
    );

-- Policy 5: HR can insert employees
CREATE POLICY users_hr_insert ON users
    FOR INSERT
    WITH CHECK (
        current_user_role() IN ('hr', 'admin') AND
        tenant_id = current_tenant_id() AND
        role IN ('employee', 'team_lead', 'manager')
    );

-- Policy 6: HR can update employees
CREATE POLICY users_hr_update ON users
    FOR UPDATE
    USING (
        current_user_role() IN ('hr', 'admin') AND
        tenant_id = current_tenant_id()
    );

-- Policy 7: Team Lead sees only their team
CREATE POLICY users_team_lead_team ON users
    FOR SELECT
    USING (
        current_user_role() = 'team_lead' AND
        tenant_id = current_tenant_id() AND
        (team_id = current_user_team() OR id = current_user_id())
    );

-- Policy 8: Employee sees only themselves
CREATE POLICY users_employee_self ON users
    FOR SELECT
    USING (
        current_user_role() = 'employee' AND
        id = current_user_id()
    );

-- Policy 9: Employee can update own profile (limited)
CREATE POLICY users_employee_update_self ON users
    FOR UPDATE
    USING (id = current_user_id());

-- ============================================================================
-- EMPLOYEES TABLE - Granular Policies
-- ============================================================================

-- Super Admin sees all
CREATE POLICY employees_super_admin ON employees
    FOR ALL
    USING (current_user_role() = 'super_admin');

-- Org Admin sees all in tenant
CREATE POLICY employees_org_admin ON employees
    FOR ALL
    USING (
        current_user_role() = 'admin' AND
        tenant_id = current_tenant_id()
    );

-- Manager sees only their department
CREATE POLICY employees_manager_department ON employees
    FOR SELECT
    USING (
        current_user_role() = 'manager' AND
        tenant_id = current_tenant_id() AND
        department_id = current_user_department()
    );

-- HR sees all employees and can manage them
CREATE POLICY employees_hr_all ON employees
    FOR ALL
    USING (
        current_user_role() = 'hr' AND
        tenant_id = current_tenant_id()
    );

-- Team Lead sees their team
CREATE POLICY employees_team_lead_team ON employees
    FOR SELECT
    USING (
        current_user_role() = 'team_lead' AND
        tenant_id = current_tenant_id() AND
        user_id IN (
            SELECT id FROM users WHERE team_id = current_user_team()
        )
    );

-- Employee sees themselves
CREATE POLICY employees_employee_self ON employees
    FOR SELECT
    USING (
        current_user_role() = 'employee' AND
        user_id = current_user_id()
    );

-- ============================================================================
-- ATTENDANCE RECORDS - Granular Policies
-- ============================================================================

-- Org Admin sees all
CREATE POLICY attendance_org_admin ON attendance_records
    FOR ALL
    USING (
        current_user_role() = 'admin' AND
        tenant_id = current_tenant_id()
    );

-- Manager sees department
CREATE POLICY attendance_manager_dept ON attendance_records
    FOR SELECT
    USING (
        current_user_role() = 'manager' AND
        tenant_id = current_tenant_id() AND
        employee_id IN (
            SELECT id FROM employees WHERE department_id = current_user_department()
        )
    );

-- HR sees all and can manage
CREATE POLICY attendance_hr_all ON attendance_records
    FOR ALL
    USING (
        current_user_role() = 'hr' AND
        tenant_id = current_tenant_id()
    );

-- Team Lead sees team
CREATE POLICY attendance_team_lead_team ON attendance_records
    FOR SELECT
    USING (
        current_user_role() = 'team_lead' AND
        tenant_id = current_tenant_id() AND
        employee_id IN (
            SELECT e.id FROM employees e
            JOIN users u ON e.user_id = u.id
            WHERE u.team_id = current_user_team()
        )
    );

-- Employee manages own
CREATE POLICY attendance_employee_self ON attendance_records
    FOR ALL
    USING (
        current_user_role() = 'employee' AND
        employee_id IN (
            SELECT id FROM employees WHERE user_id = current_user_id()
        )
    );

-- ============================================================================
-- LEAVE APPLICATIONS - Granular Policies
-- ============================================================================

-- Org Admin sees all
CREATE POLICY leaves_org_admin ON leave_applications
    FOR ALL
    USING (
        current_user_role() = 'admin' AND
        tenant_id = current_tenant_id()
    );

-- Manager sees department and can approve
CREATE POLICY leaves_manager_dept ON leave_applications
    FOR ALL
    USING (
        current_user_role() = 'manager' AND
        tenant_id = current_tenant_id() AND
        employee_id IN (
            SELECT id FROM employees WHERE department_id = current_user_department()
        )
    );

-- HR sees all employees (not managers) and can approve
CREATE POLICY leaves_hr_employees ON leave_applications
    FOR ALL
    USING (
        current_user_role() = 'hr' AND
        tenant_id = current_tenant_id() AND
        employee_id IN (
            SELECT e.id FROM employees e
            JOIN users u ON e.user_id = u.id
            WHERE u.role IN ('employee', 'team_lead')
        )
    );

-- Team Lead sees team (read-only)
CREATE POLICY leaves_team_lead_team ON leave_applications
    FOR SELECT
    USING (
        current_user_role() = 'team_lead' AND
        tenant_id = current_tenant_id() AND
        employee_id IN (
            SELECT e.id FROM employees e
            JOIN users u ON e.user_id = u.id
            WHERE u.team_id = current_user_team()
        )
    );

-- Employee manages own
CREATE POLICY leaves_employee_self ON leave_applications
    FOR ALL
    USING (
        current_user_role() = 'employee' AND
        employee_id IN (
            SELECT id FROM employees WHERE user_id = current_user_id()
        )
    );

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON POLICY users_manager_department ON users IS 'Managers see only their department employees';
COMMENT ON POLICY users_hr_employees ON users IS 'HR sees employees but not managers or other HR';
COMMENT ON POLICY users_team_lead_team ON users IS 'Team leads see only their team members';
COMMENT ON POLICY users_employee_self ON users IS 'Employees see only themselves';

COMMENT ON POLICY employees_manager_department ON employees IS 'Managers see only their department';
COMMENT ON POLICY employees_team_lead_team ON employees IS 'Team leads see only their team';

COMMENT ON POLICY attendance_manager_dept ON attendance_records IS 'Managers see department attendance';
COMMENT ON POLICY attendance_team_lead_team ON attendance_records IS 'Team leads see team attendance';

COMMENT ON POLICY leaves_manager_dept ON leave_applications IS 'Managers can approve department leaves';
COMMENT ON POLICY leaves_hr_employees ON leave_applications IS 'HR can approve employee leaves (not manager leaves)';

-- ============================================================================
-- VERIFICATION QUERIES
-- ============================================================================

-- To test these policies, use:
-- SELECT set_session_context('user-id'::UUID, 'tenant-id'::UUID, 'manager', 'dept-id'::UUID, NULL);
-- SELECT * FROM users; -- Should see only department users
-- SELECT * FROM employees; -- Should see only department employees
