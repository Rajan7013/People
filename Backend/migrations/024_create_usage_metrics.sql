-- Migration: Add Super Admin SaaS Infrastructure - Usage Metrics
-- This migration creates the usage_metrics table for tracking daily usage per tenant

CREATE TABLE IF NOT EXISTS usage_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    metric_date DATE NOT NULL,
    
    -- User metrics
    total_users INTEGER DEFAULT 0,
    active_users INTEGER DEFAULT 0,
    new_users INTEGER DEFAULT 0,
    
    -- Storage metrics (in MB)
    storage_used_mb BIGINT DEFAULT 0,
    storage_documents_mb BIGINT DEFAULT 0,
    storage_attachments_mb BIGINT DEFAULT 0,
    
    -- API metrics
    api_requests_count INTEGER DEFAULT 0,
    api_requests_success INTEGER DEFAULT 0,
    api_requests_failed INTEGER DEFAULT 0,
    
    -- Activity metrics
    logins_count INTEGER DEFAULT 0,
    unique_logins_count INTEGER DEFAULT 0,
    
    -- Feature usage
    attendance_records_count INTEGER DEFAULT 0,
    leave_requests_count INTEGER DEFAULT 0,
    employees_added_count INTEGER DEFAULT 0,
    
    -- Additional metrics (JSONB for flexibility)
    additional_metrics JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- One record per tenant per day
    UNIQUE(tenant_id, metric_date)
);

-- Create indexes for usage_metrics
CREATE INDEX idx_usage_metrics_tenant_id ON usage_metrics(tenant_id);
CREATE INDEX idx_usage_metrics_date ON usage_metrics(metric_date);
CREATE INDEX idx_usage_metrics_tenant_date ON usage_metrics(tenant_id, metric_date);

-- Add trigger for updated_at
CREATE TRIGGER update_usage_metrics_updated_at 
    BEFORE UPDATE ON usage_metrics 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Function to aggregate usage metrics for a date range
CREATE OR REPLACE FUNCTION get_usage_summary(
    p_tenant_id UUID,
    p_start_date DATE,
    p_end_date DATE
)
RETURNS TABLE (
    total_api_requests BIGINT,
    avg_daily_requests NUMERIC,
    total_storage_mb BIGINT,
    avg_active_users NUMERIC,
    total_logins BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        SUM(api_requests_count)::BIGINT as total_api_requests,
        AVG(api_requests_count)::NUMERIC(10,2) as avg_daily_requests,
        MAX(storage_used_mb)::BIGINT as total_storage_mb,
        AVG(active_users)::NUMERIC(10,2) as avg_active_users,
        SUM(logins_count)::BIGINT as total_logins
    FROM usage_metrics
    WHERE tenant_id = p_tenant_id
    AND metric_date BETWEEN p_start_date AND p_end_date;
END;
$$ LANGUAGE plpgsql;

-- Create a view for current month usage
CREATE OR REPLACE VIEW current_month_usage AS
SELECT 
    tenant_id,
    DATE_TRUNC('month', metric_date) as month,
    SUM(api_requests_count) as total_api_requests,
    MAX(storage_used_mb) as max_storage_mb,
    AVG(active_users) as avg_active_users,
    SUM(logins_count) as total_logins
FROM usage_metrics
WHERE metric_date >= DATE_TRUNC('month', CURRENT_DATE)
GROUP BY tenant_id, DATE_TRUNC('month', metric_date);
