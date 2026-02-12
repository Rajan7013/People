-- Add global system settings (Configuration Only, No Mock Users)
-- Version: 007 (Renumbered from 006)

-- Safely insert settings for ALL existing tenants (if any)
-- This avoids FK violations on a fresh empty database
INSERT INTO system_settings (tenant_id, setting_key, setting_value, setting_type, description, updated_by)
SELECT 
    t.id, 
    d.key, 
    d.val, 
    d.type, 
    d.description,
    NULL -- updated_by (system default)
FROM tenants t
CROSS JOIN (VALUES
    ('admin.user_management_enabled', 'true', 'boolean', 'Allow admin to manage other users settings'),
    ('admin.force_password_reset', 'false', 'boolean', 'Force all users to reset passwords'),
    ('admin.password_policy', '{"min_length": 12, "require_uppercase": true, "require_lowercase": true, "require_numbers": true, "require_symbols": true}', 'json', 'Global password policy settings (Bank Grade)'),
    ('admin.session_timeout_override', 'false', 'boolean', 'Allow admin to override user session timeouts'),
    ('admin.default_theme', 'light', 'string', 'Default theme for new users'),
    ('admin.allowed_themes', '["light", "dark", "auto"]', 'json', 'Themes users can select'),
    ('admin.enforce_2fa', 'true', 'boolean', 'Force all users to enable 2FA (Bank Grade)'),
    ('admin.ip_whitelist_enabled', 'false', 'boolean', 'Enable global IP whitelisting'),
    ('admin.global_ip_whitelist', '[]', 'json', 'Global allowed IP addresses'),
    ('admin.user_profile_fields_required', '["phone", "emergency_contact_name"]', 'json', 'Required user profile fields')
) AS d(key, val, type, description)
WHERE NOT EXISTS (
    SELECT 1 FROM system_settings s 
    WHERE s.tenant_id = t.id AND s.setting_key = d.key
);
