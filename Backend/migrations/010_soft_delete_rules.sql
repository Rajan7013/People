-- Soft Delete Policies and Rules
-- Version: 010 (Renumbered from 016)
-- "Military Grade" security means we NEVER lose data. 
-- DELETE statements should be intercepted and converted to updates of 'deleted_at'.

-- 1. Create a generic Soft Delete function
CREATE OR REPLACE FUNCTION soft_delete_row()
RETURNS TRIGGER AS $$
BEGIN
    NEW.deleted_at = NOW();
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- 2. Create Soft Delete Rules for all Major Tables
-- We rewrite "DELETE FROM table WHERE id = X" to "UPDATE table SET deleted_at = NOW() WHERE id = X"

-- Users
CREATE RULE soft_delete_users AS ON DELETE TO users
    DO INSTEAD UPDATE users SET deleted_at = NOW() WHERE id = OLD.id;

-- Employees
CREATE RULE soft_delete_employees AS ON DELETE TO employees
    DO INSTEAD UPDATE employees SET deleted_at = NOW() WHERE id = OLD.id;

-- Departments
CREATE RULE soft_delete_departments AS ON DELETE TO departments
    DO INSTEAD UPDATE departments SET deleted_at = NOW() WHERE id = OLD.id;

-- Tenants (Critical!)
CREATE RULE soft_delete_tenants AS ON DELETE TO tenants
    DO INSTEAD UPDATE tenants SET deleted_at = NOW() WHERE id = OLD.id;

-- 3. Prevent Hard Deletes via Trigger (Defense in Depth)
CREATE OR REPLACE FUNCTION prevent_hard_delete()
RETURNS TRIGGER AS $$
BEGIN
    IF current_setting('app.allow_hard_delete', TRUE) IS NULL THEN
        RAISE EXCEPTION 'Hard delete is not allowed. Use Soft Delete or set app.allow_hard_delete.';
    END IF;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Apply to critical tables
CREATE TRIGGER prevent_hard_delete_users BEFORE DELETE ON users FOR EACH ROW EXECUTE FUNCTION prevent_hard_delete();
