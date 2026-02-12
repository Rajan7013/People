-- User Settings Tables
-- Version: 006 (Renumbered from 005)
-- This migration adds tables for user profile settings and preferences

-- Create user_profiles table for extended user information
CREATE TABLE user_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    phone VARCHAR(20),
    address TEXT,
    date_of_birth DATE,
    emergency_contact_name VARCHAR(100),
    emergency_contact_phone VARCHAR(20),
    profile_picture_url TEXT,
    bio TEXT,
    job_title VARCHAR(100),
    department_name VARCHAR(100),
    manager_name VARCHAR(100),
    hire_date DATE,
    work_location VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create user_preferences table for user-specific settings
CREATE TABLE user_preferences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    preference_key VARCHAR(100) NOT NULL,
    preference_value TEXT,
    preference_type VARCHAR(20) DEFAULT 'string' CHECK (preference_type IN ('string', 'integer', 'boolean', 'json')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create security_settings table for user security preferences
CREATE TABLE security_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    two_factor_enabled BOOLEAN DEFAULT FALSE,
    session_timeout INTEGER DEFAULT 3600, -- in seconds
    login_notifications BOOLEAN DEFAULT TRUE,
    device_tracking BOOLEAN DEFAULT TRUE,
    ip_restrictions TEXT[], -- Array of allowed IP addresses
    last_password_change TIMESTAMP WITH TIME ZONE,
    password_expiry_days INTEGER DEFAULT 90,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create user_themes table for UI customization
CREATE TABLE user_themes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    theme_name VARCHAR(50) DEFAULT 'light',
    primary_color VARCHAR(7) DEFAULT '#1976d2',
    secondary_color VARCHAR(7) DEFAULT '#dc004e',
    font_size VARCHAR(10) DEFAULT 'medium',
    compact_mode BOOLEAN DEFAULT FALSE,
    sidebar_collapsed BOOLEAN DEFAULT FALSE,
    language VARCHAR(5) DEFAULT 'en',
    timezone VARCHAR(50) DEFAULT 'UTC',
    date_format VARCHAR(20) DEFAULT 'YYYY-MM-DD',
    time_format VARCHAR(10) DEFAULT '24h',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE UNIQUE INDEX idx_user_profiles_user_id ON user_profiles(user_id);
CREATE INDEX idx_user_profiles_tenant_id ON user_profiles(tenant_id);

CREATE UNIQUE INDEX idx_user_preferences_user_key ON user_preferences(user_id, preference_key);
CREATE INDEX idx_user_preferences_tenant_id ON user_preferences(tenant_id);

CREATE UNIQUE INDEX idx_security_settings_user_id ON security_settings(user_id);
CREATE INDEX idx_security_settings_tenant_id ON security_settings(tenant_id);

CREATE UNIQUE INDEX idx_user_themes_user_id ON user_themes(user_id);
CREATE INDEX idx_user_themes_tenant_id ON user_themes(tenant_id);

-- Insert default preferences for existing users
INSERT INTO user_preferences (user_id, tenant_id, preference_key, preference_value, preference_type)
SELECT 
    u.id,
    u.tenant_id,
    key_name,
    default_value,
    'boolean'
FROM users u
CROSS JOIN (
    VALUES 
        ('email_notifications', 'true'),
        ('push_notifications', 'false'),
        ('leave_approval_notifications', 'true'),
        ('attendance_notifications', 'false'),
        ('payslip_notifications', 'true'),
        ('system_maintenance_notifications', 'false'),
        ('marketing_emails', 'false')
) AS defaults(key_name, default_value);

-- Insert default security settings for existing users
INSERT INTO security_settings (user_id, tenant_id, two_factor_enabled, session_timeout, login_notifications)
SELECT 
    id,
    tenant_id,
    FALSE,
    3600,
    TRUE
FROM users;

-- Insert default themes for existing users
INSERT INTO user_themes (user_id, tenant_id, theme_name, primary_color, secondary_color)
SELECT 
    id,
    tenant_id,
    'light',
    '#1976d2',
    '#dc004e'
FROM users;

-- Create triggers for updated_at
CREATE TRIGGER update_user_profiles_updated_at BEFORE UPDATE ON user_profiles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_preferences_updated_at BEFORE UPDATE ON user_preferences FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_security_settings_updated_at BEFORE UPDATE ON security_settings FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_themes_updated_at BEFORE UPDATE ON user_themes FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
