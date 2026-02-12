-- Migration: Add biometric device support
-- Description: Creates tables for biometric devices and attendance logging
-- Version: 003 (Renumbered from 007)

-- Create biometric_devices table
CREATE TABLE IF NOT EXISTS biometric_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    device_name VARCHAR(255) NOT NULL,
    device_type VARCHAR(50) NOT NULL CHECK (device_type IN ('fingerprint', 'face_id', 'iris', 'voice', 'card', 'hybrid')),
    serial_number VARCHAR(255) NOT NULL,
    mac_address VARCHAR(17), -- MAC address format: XX:XX:XX:XX:XX:XX
    ip_address INET,
    port INTEGER CHECK (port > 0 AND port <= 65535),
    location VARCHAR(255),
    department VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'offline', 'error')),
    last_sync_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    UNIQUE(tenant_id, serial_number)
);

CREATE INDEX IF NOT EXISTS idx_biometric_devices_tenant_id ON biometric_devices(tenant_id);
CREATE INDEX IF NOT EXISTS idx_biometric_devices_status ON biometric_devices(status);
CREATE INDEX IF NOT EXISTS idx_biometric_devices_device_type ON biometric_devices(device_type);

-- Create biometric_attendance_logs table
CREATE TABLE IF NOT EXISTS biometric_attendance_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    device_id UUID NOT NULL REFERENCES biometric_devices(id) ON DELETE CASCADE,
    employee_code VARCHAR(50) NOT NULL,
    employee_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    event_type VARCHAR(20) NOT NULL CHECK (event_type IN ('check_in', 'check_out', 'break_start', 'break_end')),
    biometric_data TEXT, -- Encrypted biometric template or reference
    confidence DECIMAL(5,2) CHECK (confidence >= 0 AND confidence <= 100), -- Matching confidence percentage
    is_processed BOOLEAN NOT NULL DEFAULT false,
    processed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_biometric_logs_tenant_id ON biometric_attendance_logs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_biometric_logs_device_id ON biometric_attendance_logs(device_id);
CREATE INDEX IF NOT EXISTS idx_biometric_logs_employee_code ON biometric_attendance_logs(employee_code);
CREATE INDEX IF NOT EXISTS idx_biometric_logs_employee_id ON biometric_attendance_logs(employee_id);
CREATE INDEX IF NOT EXISTS idx_biometric_logs_timestamp ON biometric_attendance_logs(timestamp);
CREATE INDEX IF NOT EXISTS idx_biometric_logs_processed ON biometric_attendance_logs(is_processed);
CREATE INDEX IF NOT EXISTS idx_biometric_logs_event_type ON biometric_attendance_logs(event_type);

-- Alter attendance_records table to add biometric support
ALTER TABLE attendance_records 
ADD COLUMN IF NOT EXISTS source VARCHAR(20) NOT NULL DEFAULT 'manual' CHECK (source IN ('manual', 'biometric', 'mobile', 'web', 'api')),
ADD COLUMN IF NOT EXISTS device_id UUID REFERENCES biometric_devices(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS biometric_log_id UUID REFERENCES biometric_attendance_logs(id) ON DELETE SET NULL;

-- Create indexes for new attendance_records columns
CREATE INDEX IF NOT EXISTS idx_attendance_records_source ON attendance_records(source);
CREATE INDEX IF NOT EXISTS idx_attendance_records_device_id ON attendance_records(device_id);
CREATE INDEX IF NOT EXISTS idx_attendance_records_biometric_log_id ON attendance_records(biometric_log_id);

-- Create trigger to update updated_at timestamp for biometric_devices
CREATE OR REPLACE FUNCTION update_biometric_devices_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS trigger_biometric_devices_updated_at ON biometric_devices;
CREATE TRIGGER trigger_biometric_devices_updated_at
    BEFORE UPDATE ON biometric_devices
    FOR EACH ROW
    EXECUTE FUNCTION update_biometric_devices_updated_at();

