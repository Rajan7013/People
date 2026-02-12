-- Run this in Supabase SQL Editor to fix the audit trigger
-- This replaces the broken trigger function with the corrected version

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
    v_tenant_id := NULLIF(current_setting('app.current_tenant', TRUE), '')::uuid;
    v_user_id := NULLIF(current_setting('app.current_user_id', TRUE), '')::uuid;

    -- If no tenant is set, try to extract it from the record itself
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
        NULLIF(current_setting('app.client_ip', TRUE), '')::inet,
        NULLIF(current_setting('app.user_agent', TRUE), ''),
        NOW()
    );

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
