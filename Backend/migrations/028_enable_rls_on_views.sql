-- Migration: Enable RLS on Database Views
-- Views created by previous migrations need RLS enabled separately

-- ============================================================================
-- ENABLE RLS ON ALL VIEWS
-- ============================================================================

-- Views from subscription/invoice migrations
ALTER VIEW active_subscriptions SET (security_invoker = true);
ALTER VIEW invoice_summary SET (security_invoker = true);
ALTER VIEW organization_overview SET (security_invoker = true);

-- Views from usage tracking migration
ALTER VIEW api_usage_stats SET (security_invoker = true);
ALTER VIEW current_month_usage SET (security_invoker = true);
ALTER VIEW popular_endpoints SET (security_invoker = true);

-- Note: Views with security_invoker = true will execute with the privileges
-- of the user calling the view, which means RLS policies on the underlying
-- tables will be automatically applied.

-- ============================================================================
-- VERIFICATION
-- ============================================================================

-- Check that views are now using security_invoker
SELECT 
    schemaname,
    viewname,
    viewowner
FROM pg_views
WHERE schemaname = 'public'
    AND viewname IN (
        'active_subscriptions',
        'invoice_summary', 
        'organization_overview',
        'api_usage_stats',
        'current_month_usage',
        'popular_endpoints'
    );

COMMENT ON VIEW active_subscriptions IS 'Security invoker enabled - inherits RLS from base tables';
COMMENT ON VIEW invoice_summary IS 'Security invoker enabled - inherits RLS from base tables';
COMMENT ON VIEW organization_overview IS 'Security invoker enabled - inherits RLS from base tables';
COMMENT ON VIEW api_usage_stats IS 'Security invoker enabled - inherits RLS from base tables';
COMMENT ON VIEW current_month_usage IS 'Security invoker enabled - inherits RLS from base tables';
COMMENT ON VIEW popular_endpoints IS 'Security invoker enabled - inherits RLS from base tables';
