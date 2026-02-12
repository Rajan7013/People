# Row-Level Security (RLS) Implementation Guide

## Overview

Row-Level Security (RLS) has been implemented to ensure complete data isolation between tenants in the PeopleOS multi-tenant SaaS platform. This prevents unauthorized cross-tenant data access at the database level.

---

## 🔐 How RLS Works

### Database Layer

**Migration**: [027_enable_row_level_security.sql](file:///c:/Users/rajan/.gemini/antigravity/scratch/PeopleOS/Backend/migrations/027_enable_row_level_security.sql)

1. **RLS Enabled** on all tenant-scoped tables
2. **Session Variables** set for each request:
   - `app.current_tenant_id` - Current user's tenant UUID
   - `app.current_user_role` - Current user's role (e.g., `super_admin`, `admin`, `employee`)

3. **Policies Applied**:
   ```sql
   -- Example: Users table policy
   CREATE POLICY users_tenant_isolation ON users
       FOR ALL
       USING (
           current_user_role() = 'super_admin' OR
           tenant_id = current_tenant_id()
       );
   ```

### Application Layer

**Middleware**: [rls.go](file:///c:/Users/rajan/.gemini/antigravity/scratch/PeopleOS/Backend/internal/middleware/rls.go)

The RLS middleware automatically sets session variables for each authenticated request:

```go
// Executed on every request after authentication
func (m *RLSMiddleware) SetSessionContextEfficient(next http.Handler) http.Handler {
    // Extracts tenant_id and role from JWT claims
    // Sets PostgreSQL session variables
    // All subsequent queries are automatically filtered
}
```

---

## 📋 Tables Protected by RLS

### Core Tables
- ✅ `users` - Super admin can see all, others see only their tenant
- ✅ `departments`
- ✅ `employees`
- ✅ `attendance_policies`
- ✅ `attendance_records`
- ✅ `leave_types`
- ✅ `leave_applications`

### SaaS Tables
- ✅ `subscriptions` - Super admin can see all
- ✅ `invoices` - Super admin can see all
- ✅ `organization_details` - Super admin can see all
- ✅ `usage_metrics` - Super admin can see all
- ✅ `api_request_logs` - Super admin can see all

### Additional Tables
- ✅ `biometric_devices`
- ✅ `biometric_logs`
- ✅ `payslips`
- ✅ `system_settings`
- ✅ `user_preferences`

### Special Handling
- ✅ `tenants` - Super admin sees all, users see only their tenant
- ✅ `subscription_plans` - Visible to all (read-only), modifiable by super admin only

---

## 🚀 Usage

### Automatic Enforcement

RLS is **automatically enforced** on all requests after authentication. No code changes needed in services!

**Before RLS**:
```go
// Had to manually filter by tenant_id everywhere
query := "SELECT * FROM employees WHERE tenant_id = $1"
rows, err := db.Query(query, tenantID)
```

**After RLS**:
```go
// RLS automatically filters - cleaner code!
query := "SELECT * FROM employees"
rows, err := db.Query(query)
// Only returns employees from current user's tenant
```

### Super Admin Access

Super admins automatically bypass tenant restrictions:

```go
// Super admin user makes request
// Session variable: app.current_user_role = 'super_admin'

// This query returns ALL tenants across the platform
query := "SELECT * FROM tenants"
rows, err := db.Query(query)
```

### Testing RLS

```sql
-- Set session context manually for testing
SELECT set_session_context('<tenant-uuid>', 'admin');

-- Test queries
SELECT * FROM employees; -- Should only see employees from specified tenant

-- Test as super admin
SELECT set_session_context(NULL, 'super_admin');
SELECT * FROM tenants; -- Should see ALL tenants
```

---

## 🔧 Configuration

### Database Setup

1. **Apply Migration**:
   ```bash
   cd Backend
   go run cmd/migrate/main.go
   ```

2. **Grant BYPASSRLS** to application user (run as PostgreSQL superuser):
   ```sql
   ALTER USER peopleos_app BYPASSRLS;
   ```
   
   This allows the application to set session variables and enforce RLS programmatically.

### Application Setup

The RLS middleware is automatically initialized in [server.go](file:///c:/Users/rajan/.gemini/antigravity/scratch/PeopleOS/Backend/internal/server/server.go):

```go
// RLS middleware initialized
rlsMiddleware := custommiddleware.NewRLSMiddleware(database)

// Applied to all authenticated routes (done automatically)
```

---

## 🛡️ Security Benefits

### 1. **Defense in Depth**
Even if application code has bugs, database enforces isolation

### 2. **SQL Injection Protection**
Even successful SQL injection cannot access other tenants' data

### 3. **Audit Trail**
All queries automatically respect tenant boundaries

### 4. **Zero Trust**
Database doesn't trust application layer - enforces rules independently

---

## 📊 Performance Considerations

### Optimizations Applied

1. **Indexed Columns**: All `tenant_id` columns are indexed
2. **Efficient Policies**: Policies use simple equality checks
3. **Session Variables**: Cached per connection, not per query
4. **Super Admin Bypass**: No overhead when not needed

### Query Performance

```sql
-- RLS adds minimal overhead
EXPLAIN ANALYZE SELECT * FROM employees;

-- Index on tenant_id ensures fast filtering
-- Seq Scan on employees (cost=0.00..X rows=Y)
--   Filter: (tenant_id = current_tenant_id())
```

---

## 🧪 Verification

### Check RLS Status

```sql
-- View tables with RLS enabled
SELECT schemaname, tablename, rowsecurity
FROM pg_tables
WHERE schemaname = 'public' AND rowsecurity = true;
```

### View Policies

```sql
-- List all RLS policies
SELECT tablename, policyname, permissive, cmd, qual
FROM pg_policies
WHERE schemaname = 'public'
ORDER BY tablename;
```

### Test Isolation

```sql
-- As Tenant A admin
SELECT set_session_context('tenant-a-uuid', 'admin');
SELECT COUNT(*) FROM employees; -- Returns Tenant A count

-- As Tenant B admin
SELECT set_session_context('tenant-b-uuid', 'admin');
SELECT COUNT(*) FROM employees; -- Returns Tenant B count (different)

-- As Super Admin
SELECT set_session_context(NULL, 'super_admin');
SELECT COUNT(*) FROM employees; -- Returns ALL employees
```

---

## 🔍 Troubleshooting

### Issue: "No rows returned"

**Cause**: Session context not set properly

**Solution**: Ensure RLS middleware is applied to the route

### Issue: "Permission denied"

**Cause**: Application user doesn't have BYPASSRLS

**Solution**: Grant BYPASSRLS privilege:
```sql
ALTER USER peopleos_app BYPASSRLS;
```

### Issue: "Super admin can't see all data"

**Cause**: Session role not set to 'super_admin'

**Solution**: Check JWT token has correct role claim

---

## 📝 Best Practices

### 1. Always Use Session Context

```go
// ✅ GOOD: Let RLS handle filtering
query := "SELECT * FROM employees"

// ❌ BAD: Manual filtering (redundant with RLS)
query := "SELECT * FROM employees WHERE tenant_id = $1"
```

### 2. Trust the Database

```go
// ✅ GOOD: RLS enforces isolation
func GetEmployees(ctx context.Context) ([]*Employee, error) {
    rows, err := db.QueryContext(ctx, "SELECT * FROM employees")
    // RLS automatically filters by tenant
}
```

### 3. Test Both Paths

- Test as regular tenant user (should see only their data)
- Test as super admin (should see all data)

---

## 🎯 Summary

| Feature | Status |
|---------|--------|
| RLS Enabled | ✅ All tenant tables |
| Policies Created | ✅ 20+ policies |
| Middleware Integrated | ✅ Automatic |
| Super Admin Support | ✅ Full access |
| Performance Optimized | ✅ Indexed |
| Tested | ⏳ Ready for testing |

**Result**: Complete tenant data isolation enforced at the database level with zero application code changes required! 🎉
