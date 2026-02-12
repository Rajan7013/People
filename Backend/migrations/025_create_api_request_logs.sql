-- Migration: Add Super Admin SaaS Infrastructure - API Request Logs
-- This migration creates the api_request_logs table for detailed request tracking

CREATE TABLE IF NOT EXISTS api_request_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    
    -- Request details
    method VARCHAR(10) NOT NULL,
    endpoint VARCHAR(500) NOT NULL,
    status_code INTEGER,
    response_time_ms INTEGER,
    
    -- Request/Response size
    request_size_bytes INTEGER,
    response_size_bytes INTEGER,
    
    -- Metadata
    ip_address INET,
    user_agent TEXT,
    
    -- Error tracking
    error_message TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for api_request_logs
CREATE INDEX idx_api_logs_tenant_id ON api_request_logs(tenant_id);
CREATE INDEX idx_api_logs_user_id ON api_request_logs(user_id);
CREATE INDEX idx_api_logs_created_at ON api_request_logs(created_at);
CREATE INDEX idx_api_logs_tenant_date ON api_request_logs(tenant_id, created_at);
CREATE INDEX idx_api_logs_status_code ON api_request_logs(status_code);
CREATE INDEX idx_api_logs_endpoint ON api_request_logs(endpoint);

-- Partition by month for better performance (optional, can be enabled later)
-- This is commented out for now but can be uncommented when needed
-- CREATE TABLE api_request_logs_y2026m01 PARTITION OF api_request_logs
--     FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

-- Function to clean up old logs (keep last 90 days)
CREATE OR REPLACE FUNCTION cleanup_old_api_logs()
RETURNS void AS $$
BEGIN
    DELETE FROM api_request_logs
    WHERE created_at < CURRENT_DATE - INTERVAL '90 days';
END;
$$ LANGUAGE plpgsql;

-- Create a view for API usage statistics
CREATE OR REPLACE VIEW api_usage_stats AS
SELECT 
    tenant_id,
    DATE(created_at) as request_date,
    COUNT(*) as total_requests,
    COUNT(*) FILTER (WHERE status_code >= 200 AND status_code < 300) as successful_requests,
    COUNT(*) FILTER (WHERE status_code >= 400) as failed_requests,
    AVG(response_time_ms) as avg_response_time_ms,
    MAX(response_time_ms) as max_response_time_ms,
    COUNT(DISTINCT user_id) as unique_users
FROM api_request_logs
GROUP BY tenant_id, DATE(created_at);

-- Create a view for endpoint popularity
CREATE OR REPLACE VIEW popular_endpoints AS
SELECT 
    tenant_id,
    endpoint,
    COUNT(*) as request_count,
    AVG(response_time_ms) as avg_response_time_ms,
    COUNT(*) FILTER (WHERE status_code >= 400) as error_count
FROM api_request_logs
WHERE created_at >= CURRENT_DATE - INTERVAL '7 days'
GROUP BY tenant_id, endpoint
ORDER BY request_count DESC;
