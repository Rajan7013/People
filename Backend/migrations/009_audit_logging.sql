-- Enable Automatic Audit Logging for ALL Tables
-- Version: 009 (Renumbered from 012)
-- This migration creates a generic trigger function and applies it to every table
-- to record all INSERT, UPDATE, and DELETE operations into the `audit_logs` table.

-- 1. Create the Generic Audit Trigger Function
CREATE OR REPLACE FUNCTION log_audit_event()
RETURNS TRIGGER AS $$
DECLARE
    v_tenant_id UUID;
    v_user_id UUID;
    v_old_data JSONB;
    v_new_data JSONB;
    v_resource_id UUID;
BEGIN
    -- Get current tenant and user from session context
    -- (The app must set these: app.current_tenant, app.current_user)
    v_tenant_id := NULLIF(current_setting('app.current_tenant', TRUE), '')::uuid;
    v_user_id := NULLIF(current_setting('app.current_user_id', TRUE), '')::uuid;

    -- If no tenant is set, try to extract it from the record itself if possible, 
    -- otherwise it might remain NULL (system action)
    IF v_tenant_id IS NULL THEN
        IF (TG_OP = 'DELETE' OR TG_OP = 'UPDATE') AND (to_jsonb(OLD) ? 'tenant_id') THEN
            v_tenant_id := (to_jsonb(OLD)->>'tenant_id')::uuid;
        ELSIF (TG_OP = 'INSERT' OR TG_OP = 'UPDATE') AND (to_jsonb(NEW) ? 'tenant_id') THEN
            v_tenant_id := (to_jsonb(NEW)->>'tenant_id')::uuid;
        END IF;
    END IF;

    -- Set Old/New Data
    IF (TG_OP = 'INSERT') THEN
        v_old_data := NULL;
        v_new_data := to_jsonb(NEW);
        IF (to_jsonb(NEW) ? 'id') THEN v_resource_id := (to_jsonb(NEW)->>'id')::uuid; END IF;
    ELSIF (TG_OP = 'UPDATE') THEN
        v_old_data := to_jsonb(OLD);
        v_new_data := to_jsonb(NEW);
        IF (to_jsonb(NEW) ? 'id') THEN v_resource_id := (to_jsonb(NEW)->>'id')::uuid; END IF;
    ELSIF (TG_OP = 'DELETE') THEN
        v_old_data := to_jsonb(OLD);
        v_new_data := NULL;
        IF (to_jsonb(OLD) ? 'id') THEN v_resource_id := (to_jsonb(OLD)->>'id')::uuid; END IF;
    END IF;

    -- Insert into Audit Log
    -- We bypass RLS for this insertion to ensure the log is always written, 
    -- but usually the policy on audit_logs allows INSERTs.
    INSERT INTO audit_logs (
        id,
        tenant_id,
        user_id,
        action,
        resource_type,
        resource_id,
        old_values,
        new_values,
        ip_address,
        user_agent,
        created_at
    ) VALUES (
        uuid_generate_v4(),
        v_tenant_id,
        v_user_id,
        TG_OP, 
        TG_TABLE_NAME::text,
        v_resource_id,
        v_old_data,
        v_new_data,
        NULLIF(current_setting('app.client_ip', TRUE), '')::inet, -- App needs to set this
        NULLIF(current_setting('app.user_agent', TRUE), ''),      -- App needs to set this
        NOW()
    );

    RETURN NULL; -- Result is ignored for AFTER triggers
END;
$$ LANGUAGE plpgsql;


-- 2. Apply Trigger to ALL Tables (Dynamic SQL to loop through tables)
-- This block dynamically finds all tables in the 'public' schema and attaches the trigger.
DO $$
DECLARE
    t text;
BEGIN
    FOR t IN 
        SELECT table_name 
        FROM information_schema.tables 
        WHERE table_schema = 'public' 
          AND table_type = 'BASE TABLE'
          AND table_name NOT IN ('audit_logs', 'schema_migrations') -- Exclude audit_logs itself to prevent loops!
    LOOP
        EXECUTE format('DROP TRIGGER IF EXISTS audit_trigger_%I ON %I', t, t);
        EXECUTE format('CREATE TRIGGER audit_trigger_%I AFTER INSERT OR UPDATE OR DELETE ON %I FOR EACH ROW EXECUTE FUNCTION log_audit_event()', t, t);
    END LOOP;
END;
$$;
