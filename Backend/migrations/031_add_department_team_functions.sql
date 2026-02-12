-- Migration 031: Add Helper Functions for Department, Team, and User Context
-- These functions retrieve session variables set by the application for RLS enforcement

-- Get current user's department ID
CREATE OR REPLACE FUNCTION current_user_department() RETURNS UUID AS $$
BEGIN
    RETURN current_setting('app.current_user_department', TRUE)::UUID;
EXCEPTION
    WHEN OTHERS THEN
        RETURN NULL;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- Get current user's team ID
CREATE OR REPLACE FUNCTION current_user_team() RETURNS UUID AS $$
BEGIN
    RETURN current_setting('app.current_user_team', TRUE)::UUID;
EXCEPTION
    WHEN OTHERS THEN
        RETURN NULL;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- Get current user ID
CREATE OR REPLACE FUNCTION current_user_id() RETURNS UUID AS $$
BEGIN
    RETURN current_setting('app.current_user_id', TRUE)::UUID;
EXCEPTION
    WHEN OTHERS THEN
        RETURN NULL;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- Update set_session_context to include all context variables
CREATE OR REPLACE FUNCTION set_session_context(
    p_user_id UUID,
    p_tenant_id UUID, 
    p_user_role TEXT,
    p_department_id UUID DEFAULT NULL,
    p_team_id UUID DEFAULT NULL
)
RETURNS VOID AS $$
BEGIN
    -- Set user ID
    PERFORM set_config('app.current_user_id', p_user_id::TEXT, TRUE);
    
    -- Set tenant ID (can be NULL for super_admin)
    IF p_tenant_id IS NOT NULL THEN
        PERFORM set_config('app.current_tenant_id', p_tenant_id::TEXT, TRUE);
    ELSE
        PERFORM set_config('app.current_tenant_id', '', TRUE);
    END IF;
    
    -- Set user role
    PERFORM set_config('app.current_user_role', p_user_role, TRUE);
    
    -- Set department ID (optional)
    IF p_department_id IS NOT NULL THEN
        PERFORM set_config('app.current_user_department', p_department_id::TEXT, TRUE);
    ELSE
        PERFORM set_config('app.current_user_department', '', TRUE);
    END IF;
    
    -- Set team ID (optional)
    IF p_team_id IS NOT NULL THEN
        PERFORM set_config('app.current_user_team', p_team_id::TEXT, TRUE);
    ELSE
        PERFORM set_config('app.current_user_team', '', TRUE);
    END IF;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Add comments
COMMENT ON FUNCTION current_user_id() IS 'Returns the current user ID from session variable for RLS';
COMMENT ON FUNCTION current_user_department() IS 'Returns the current user department ID from session variable for RLS';
COMMENT ON FUNCTION current_user_team() IS 'Returns the current user team ID from session variable for RLS';
COMMENT ON FUNCTION set_session_context(UUID, UUID, TEXT, UUID, UUID) IS 'Sets all session variables for RLS enforcement including user, tenant, role, department, and team';

-- Verification query
-- SELECT set_session_context(
--     'user-uuid'::UUID, 
--     'tenant-uuid'::UUID, 
--     'manager', 
--     'dept-uuid'::UUID, 
--     NULL
-- );
-- SELECT current_user_id(), current_tenant_id(), current_user_role(), current_user_department(), current_user_team();
