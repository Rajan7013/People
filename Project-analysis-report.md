PeopleOS - Complete 360° Project Analysis Report
Developer: Ranjan Prasaila Yadav
Contact: 
rajanprasaila@gmail.com
 | +91 7013146154
Version: 0.1.0 (Prototype/Testing Phase)
Date: February 12, 2026

CAUTION

Important Disclaimer: This is a PROTOTYPE/TESTING VERSION and is NOT production-ready. This version may contain security vulnerabilities, incomplete features, and is NOT optimized for mobile devices. The system is deployed for testing, feedback, and demonstration purposes only. We continuously update security measures, add features, and improve functionality based on user feedback and security best practices.

Table of Contents
Executive Summary
Project Overview
Technology Stack
Architecture & Design
Security Implementation
RBAC System Analysis
Database Design & RLS
Authentication & Authorization
Business Logic & Calculations
Completion Status
Competitive Analysis
Target User Personas
Future Roadmap
Executive Summary
PeopleOS is a multi-tenant SaaS Human Resource Management System (HRMS) designed to streamline employee management, attendance tracking, leave management, and payroll processing for small to medium-sized organizations. Built with modern web technologies and enterprise-grade security practices, PeopleOS implements role-based access control (RBAC) with PostgreSQL Row-Level Security (RLS) for complete data isolation.

Key Highlights
Multi-Tenant Architecture with complete data isolation
6-Tier RBAC System (Super Admin → Admin → HR → Manager → Team Lead → Employee)
PostgreSQL RLS with 42+ database migrations
Argon2id Password Hashing with pepper for enhanced security
JWT-based Authentication with refresh tokens
RESTful API architecture
Real-time Attendance Tracking with mathematical precision
Automated Payroll Calculations with configurable components
Project Overview
Project Name
PeopleOS - People Operations System

Project Aim
To provide small and medium-sized businesses with an affordable, secure, and scalable HRMS solution that eliminates manual HR processes and provides real-time insights into workforce management.

Project Goals
Automate HR Operations - Reduce manual work in attendance, leave, and payroll management
Data Security - Implement enterprise-grade security with multi-layered protection
Role-Based Access - Ensure users only see and modify data relevant to their role
Multi-Tenant Isolation - Complete separation of data between different organizations
Scalability - Support growing organizations with flexible architecture
Compliance - Maintain audit trails and ensure data privacy compliance
Technology Stack
Backend Stack
Go 1.24
Chi Router v5
PostgreSQL 15+
JWT Auth
Argon2id
Row-Level Security
42 Migrations
Technology	Version	Purpose	Why Chosen
Go	1.24.0	Backend Language	High performance, strong typing, excellent concurrency, compiled binary
Chi Router	v5.0.10	HTTP Router	Lightweight, idiomatic, middleware support, fast routing
PostgreSQL	15+	Database	Advanced RLS, ACID compliance, JSON support, enterprise-grade
pgx	v5.8.0	Database Driver	Native Go driver, better performance than lib/pq
JWT	v5.2.0	Authentication	Stateless auth, scalable, industry standard
Argon2id	Latest	Password Hashing	Winner of Password Hashing Competition, resistant to GPU attacks
Zerolog	v1.31.0	Logging	Zero-allocation JSON logger, high performance
CORS	v1.2.1	Cross-Origin	Secure API access from frontend
Frontend Stack
Technology	Version	Purpose	Why Chosen
Next.js	16.1.6	React Framework	SSR, routing, optimization, production-ready
React	19.2.3	UI Library	Component-based, virtual DOM, large ecosystem
TypeScript	5.x	Type Safety	Catch errors at compile-time, better IDE support
Tailwind CSS	4.x	Styling	Utility-first, responsive, fast development
shadcn/ui	Latest	UI Components	Accessible, customizable, modern design
Axios	1.13.5	HTTP Client	Interceptors, request/response transformation
TanStack Table	8.21.3	Data Tables	Powerful, headless, sortable tables
date-fns	3.6.0	Date Manipulation	Lightweight, immutable, tree-shakeable
Recharts	3.7.0	Charts	Composable, responsive charts
Zod	4.3.6	Validation	Type-safe schema validation
Why This Stack is Best
Performance: Go's compiled nature + PostgreSQL's query optimization = Fast response times
Type Safety: Go + TypeScript = Fewer runtime errors
Scalability: Stateless JWT + PostgreSQL connection pooling = Horizontal scaling
Security: Built-in security features in all layers
Developer Experience: Modern tooling, hot reload, strong typing
Cost-Effective: Open-source technologies, no licensing fees
Community Support: Large communities for troubleshooting
Architecture & Design
System Architecture
Database Layer - PostgreSQL
Backend Layer - Go
Frontend Layer - Next.js
Client Layer
HTTPS
REST API
JWT Token
Web Browser
React Components
API Client - Axios
State Management
Cookie Storage
Chi Router
Auth Middleware
RBAC Middleware
Status Check Middleware
Handlers
Services
Models
PostgreSQL
Row-Level Security
Audit Logs
Migrations
Request Flow
Database
Service
Handler
Middleware
Frontend
User
Database
Service
Handler
Middleware
Frontend
User
Login Request
POST /api/v1/auth/login
Validate Request
Authenticate User
Query User + Verify Password
User Data
Generate JWT Token
Token + User Info
LoginResponse
Store Token in Cookie
Redirect to Dashboard
Access Protected Resource
GET /api/v1/admin/employees
Extract JWT from Cookie
Validate Token
Check User Status (DB Query)
Check RBAC Permissions
Authorized Request
Get Employees
Set Session Context (RLS)
Query with RLS Applied
Filtered Data
Employee List
JSON Response
Display Data
Security Implementation
Current Security Layers (8 Layers)
Layer 1: Transport Security
HTTPS Only - All communication encrypted in transit
CORS Configuration - Whitelist allowed origins
Secure Headers - CSP, X-Frame-Options, X-Content-Type-Options
Layer 2: Authentication Security
Argon2id Password Hashing - Memory-hard, GPU-resistant algorithm
Pepper Secret - Additional secret added to all passwords before hashing
JWT Tokens - Signed tokens with expiration
Refresh Tokens - Separate long-lived tokens for token renewal
HttpOnly Cookies - Tokens stored in HttpOnly cookies (not accessible via JavaScript)
Layer 3: Authorization Security
Role-Based Access Control (RBAC) - 6-tier role hierarchy
Middleware Enforcement - Every request checked for permissions
Real-time Status Checks - User active status verified on each request
Layer 4: Database Security
Row-Level Security (RLS) - PostgreSQL native data isolation
Tenant Isolation - Complete separation of organization data
Session Context - User context set for every database query
Prepared Statements - Protection against SQL injection
Layer 5: Data Protection
Soft Deletes - Data marked as deleted, not physically removed
Audit Logging - All critical operations logged with user, timestamp, action
Encrypted Sensitive Fields - PII data encrypted at rest (planned)
No Plain Text Passwords - Passwords never stored in plain text
Layer 6: Input Validation
Schema Validation - Zod schemas on frontend, struct validation on backend
Type Safety - TypeScript + Go strong typing
Sanitization - Input sanitized before processing
Layer 7: Session Management
Token Expiration - Access tokens expire (configurable, default 60 min)
Refresh Token Rotation - Refresh tokens expire (configurable, default 7 days)
Logout Invalidation - Tokens cleared on logout
Concurrent Session Control - Track active sessions
Layer 8: Monitoring & Logging
Structured Logging - JSON logs with context (Zerolog)
Audit Trail - Database triggers log all modifications
Error Tracking - Errors logged with stack traces
Access Logs - All API requests logged
Future Security Enhancements (10 Layers Total)
Layer 9: Advanced Threat Protection (Planned)
Rate Limiting - Prevent brute force attacks
IP Whitelisting - Restrict access by IP for sensitive operations
2FA/MFA - Two-factor authentication for admin accounts
Session Anomaly Detection - Detect unusual login patterns
Layer 10: Compliance & Advanced Encryption (Planned)
Data Encryption at Rest - Encrypt sensitive database columns
GDPR Compliance - Data export, right to be forgotten
SOC 2 Compliance - Security controls and audits
Penetration Testing - Regular security assessments
Security Principles Followed
Defense in Depth - Multiple layers of security
Least Privilege - Users have minimum necessary permissions
Fail Secure - System defaults to deny access on errors
Zero Trust - Verify every request, never assume trust
Separation of Concerns - Security logic separated from business logic
Audit Everything - Comprehensive logging for accountability
Secure by Default - Security features enabled by default
Security Cost-Benefit Analysis
NOTE

Security Investment Philosophy: We invest significant resources (X budget) in security measures. For an attacker to successfully breach our system, they would need to invest 2-3X our budget due to multiple security layers. This makes us an economically unattractive target for most threat actors.

We do NOT claim:

❌ 100% hacking proof
❌ Completely unhackable
❌ Zero vulnerabilities
We acknowledge:

✅ Zero-day vulnerabilities exist in all software
✅ Security is an ongoing process, not a destination
✅ We continuously update based on threat landscape
✅ We follow industry best practices and standards
RBAC System Analysis
Role Hierarchy
Super Admin
Organization Admin
HR Manager
Department Manager
Team Lead
Employee
Role Permissions Matrix
Feature	Super Admin	Org Admin	HR	Manager	Team Lead	Employee
View All Tenants	✅	❌	❌	❌	❌	❌
Manage Subscriptions	✅	✅	❌	❌	❌	❌
Create Employees	✅	✅	✅	❌	❌	❌
View All Employees	✅	✅	✅	❌	❌	❌
View Dept Employees	✅	✅	✅	✅	❌	❌
View Team Members	✅	✅	✅	✅	✅	❌
View Self	✅	✅	✅	✅	✅	✅
Approve Leaves	✅	✅	✅	✅	❌	❌
Generate Payroll	✅	✅	✅	❌	❌	❌
Configure Policies	✅	✅	❌	❌	❌	❌
Mark Attendance	✅	✅	✅	✅	✅	✅
View Own Payslips	✅	✅	✅	✅	✅	✅
RBAC Implementation Details
Backend RBAC (Middleware)
File: 
internal/middleware/rbac.go

go
// RequireRole middleware checks if user has required role
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims, ok := auth.GetClaimsFromContext(r.Context())
            if !ok {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            
            allowed := false
            for _, role := range allowedRoles {
                if claims.Role == role {
                    allowed = true
                    break
                }
            }
            
            if !allowed {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
Database RLS (PostgreSQL)
File: 
migrations/032_granular_rls_policies.sql

Example policy for employees table:

sql
-- Manager sees only their department
CREATE POLICY employees_manager_department ON employees
    FOR SELECT
    USING (
        current_user_role() = 'manager' AND
        tenant_id = current_tenant_id() AND
        department_id = current_user_department()
    );
Frontend RBAC (Route Protection)
Routes are protected by layout-level authentication checks and role-based redirects.

Database Design & RLS
Database Schema Overview
Total Tables: 25+
Total Migrations: 42
RLS Enabled: Yes (27 tables)

Core Tables
contains
contains
is employee
has
submits
receives
contains
contains
categorizes
defines
TENANTS
uuid
id
PK
string
name
string
subdomain
boolean
is_active
timestamp
created_at
USERS
uuid
id
PK
uuid
tenant_id
FK
string
email
string
password_hash
string
role
uuid
department_id
FK
uuid
team_id
FK
boolean
is_active
EMPLOYEES
uuid
id
PK
uuid
tenant_id
FK
uuid
user_id
FK
string
employee_code
uuid
department_id
FK
decimal
salary
string
employment_status
ATTENDANCE_RECORDS
uuid
id
PK
uuid
tenant_id
FK
uuid
employee_id
FK
date
date
timestamp
check_in_time
timestamp
check_out_time
decimal
total_hours
string
status
LEAVE_APPLICATIONS
PAYSLIPS
DEPARTMENTS
TEAMS
LEAVE_TYPES
SALARY_COMPONENTS
PAYSLIP_COMPONENTS
Row-Level Security (RLS) Implementation
How RLS Works
Session Context Set: On every authenticated request, backend sets PostgreSQL session variables:

sql
SELECT set_session_context(
    'user-id'::UUID,
    'tenant-id'::UUID,
    'role'::TEXT,
    'department-id'::UUID,
    'team-id'::UUID
);
RLS Policies Applied: PostgreSQL automatically filters queries based on policies:

sql
-- Employee sees only their own records
CREATE POLICY attendance_employee_self ON attendance_records
    FOR ALL
    USING (
        current_user_role() = 'employee' AND
        employee_id IN (
            SELECT id FROM employees WHERE user_id = current_user_id()
        )
    );
Data Isolation: Each query returns only data the user is authorized to see

RLS Security Benefits
Database-Level Enforcement: Even if application code has bugs, database enforces access control
Multi-Tenant Isolation: Impossible for one tenant to see another tenant's data
Audit Trail: All policies logged and auditable
Performance: PostgreSQL optimizes RLS queries efficiently
Database Security Features
Soft Deletes: Records marked with deleted_at timestamp, not physically deleted
Audit Logging: Triggers log all INSERT/UPDATE/DELETE operations
Encryption Ready: Schema supports encrypted columns (implementation in progress)
Constraints: Foreign keys, check constraints, unique constraints enforce data integrity
Indexes: Optimized indexes on frequently queried columns
Authentication & Authorization
Password Storage
Method: Argon2id + Pepper

go
// Password hashing with Argon2id
func HashPassword(password, pepper string) (string, error) {
    // Combine password with pepper
    saltedPassword := password + pepper
    
    // Generate salt
    salt := make([]byte, 16)
    rand.Read(salt)
    
    // Hash with Argon2id (memory=64MB, iterations=3, parallelism=2)
    hash := argon2.IDKey(
        []byte(saltedPassword),
        salt,
        3,      // iterations
        64*1024, // memory (64 MB)
        2,      // parallelism
        32,     // key length
    )
    
    // Encode as: $argon2id$v=19$m=65536,t=3,p=2$salt$hash
    return encodeHash(hash, salt), nil
}
Why Argon2id?

Winner of Password Hashing Competition (2015)
Resistant to GPU/ASIC attacks (memory-hard)
Configurable memory, time, and parallelism
Industry standard (recommended by OWASP)
What is Pepper?

Additional secret added to all passwords
Stored in environment variable (not in database)
Even if database is compromised, passwords cannot be cracked without pepper
Adds extra layer of security beyond salt
Storage:

❌ Plain text passwords: NEVER stored
❌ Reversible encryption: NOT used
✅ One-way hash: Argon2id with salt and pepper
✅ Timestamps: All login attempts logged
JWT Token Management
Token Structure:

json
{
  "user_id": "uuid",
  "tenant_id": "uuid",
  "email": "user@example.com",
  "role": "employee",
  "department_id": "uuid",
  "team_id": "uuid",
  "exp": 1234567890,
  "iat": 1234567890
}
Token Lifecycle:

Backend
Frontend
User
Backend
Frontend
User
After 55 minutes
Login
POST /auth/login
Verify Password
Generate Access Token (60 min)
Generate Refresh Token (7 days)
Tokens in HttpOnly Cookies
Redirect to Dashboard
GET /api/resource (Token expires soon)
401 Unauthorized
POST /auth/refresh (with Refresh Token)
Validate Refresh Token
Generate New Access Token
New Access Token
Retry GET /api/resource
200 OK + Data
Cookie Configuration:

HttpOnly: true (prevents XSS attacks)
Secure: true (HTTPS only)
SameSite: Strict (prevents CSRF)
Path: /
Max-Age: 3600 seconds (access token), 604800 seconds (refresh token)
Login System Workflow
User enters email + password
Frontend validates input (Zod schema)
Frontend sends POST to /api/v1/auth/login
Backend queries database for user by email
Backend verifies password (Argon2id + pepper)
Backend checks user status (is_active, deleted_at)
Backend generates JWT tokens
Backend sets HttpOnly cookies
Backend returns user data (without password hash)
Frontend stores user in state
Frontend redirects based on role
Login Restrictions:

✅ Only active users can login
✅ Deleted users cannot login
✅ Suspended accounts immediately blocked
✅ Email must be verified (future enhancement)
✅ Rate limiting on login endpoint (future enhancement)
Business Logic & Calculations
Attendance Tracking
Attendance Calculation Logic
Formula:

Total Hours = (Check-Out Time - Check-In Time) - Break Time
Status = Determine based on check-in time and policy
Status Determination:

go
func DetermineAttendanceStatus(checkInTime time.Time, policy AttendancePolicy) string {
    workStartTime := policy.WorkStartTime
    gracePeriod := policy.GracePeriodMinutes
    
    lateThreshold := workStartTime.Add(time.Duration(gracePeriod) * time.Minute)
    
    if checkInTime.IsZero() {
        return "absent"
    } else if checkInTime.After(lateThreshold) {
        return "late"
    } else {
        return "present"
    }
}
Database Storage:

sql
CREATE TABLE attendance_records (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    employee_id UUID NOT NULL,
    date DATE NOT NULL,
    check_in_time TIMESTAMP,
    check_out_time TIMESTAMP,
    total_hours DECIMAL(5,2),
    status VARCHAR(20), -- 'present', 'late', 'absent', 'half_day'
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(employee_id, date)
);
Math & Precision:

Time stored as TIMESTAMP (microsecond precision)
Hours calculated as DECIMAL(5,2) (e.g., 8.50 hours)
Timezone-aware calculations
Rounding to 2 decimal places
Attendance Tracking Flow
Database
Backend
Frontend
Employee
Database
Backend
Frontend
Employee
8 hours later
Click "Check In"
POST /api/v1/employee/attendance/check-in
Get Attendance Policy
Policy (work start time, grace period)
Calculate Status (present/late)
INSERT attendance_record
Record Created
Success + Current Status
"Checked In at 09:05 AM (Late)"
Click "Check Out"
POST /api/v1/employee/attendance/check-out
UPDATE attendance_record
Calculate Total Hours
Save Total Hours
Updated
Success + Total Hours
"Checked Out. Total: 8.25 hours"
Payroll Calculations
Payroll Formula
Gross Salary = Basic Salary + Sum(Earnings Components)
Total Deductions = Sum(Deduction Components) + Late Fines + Absent Deductions
Net Salary = Gross Salary - Total Deductions
Component Types:

Earnings: HRA, DA, Bonus, Overtime, Allowances
Deductions: PF, ESI, Tax, Loan Repayment, Late Fine, Absent Deduction
Payroll Calculation Implementation
go
func CalculatePayslip(employee Employee, month time.Time, policy PayrollPolicy) Payslip {
    // 1. Get base salary
    basicSalary := employee.Salary
    
    // 2. Calculate earnings
    earnings := []Component{}
    for _, comp := range policy.EarningsComponents {
        amount := calculateComponent(comp, basicSalary)
        earnings = append(earnings, Component{
            Name: comp.Name,
            Type: "earning",
            Amount: amount,
        })
    }
    grossSalary := basicSalary + sumComponents(earnings)
    
    // 3. Calculate attendance-based deductions
    attendanceRecords := getAttendanceForMonth(employee.ID, month)
    lateDays := countStatus(attendanceRecords, "late")
    absentDays := countStatus(attendanceRecords, "absent")
    
    lateFine := lateDays * policy.LateFinePerDay
    absentDeduction := absentDays * (basicSalary / 30) // Per-day salary
    
    // 4. Calculate standard deductions
    deductions := []Component{}
    for _, comp := range policy.DeductionComponents {
        amount := calculateComponent(comp, basicSalary)
        deductions = append(deductions, Component{
            Name: comp.Name,
            Type: "deduction",
            Amount: amount,
        })
    }
    
    // 5. Add attendance deductions
    deductions = append(deductions, 
        Component{Name: "Late Fine", Type: "deduction", Amount: lateFine},
        Component{Name: "Absent Deduction", Type: "deduction", Amount: absentDeduction},
    )
    
    totalDeductions := sumComponents(deductions)
    
    // 6. Calculate net salary
    netSalary := grossSalary - totalDeductions
    
    return Payslip{
        EmployeeID: employee.ID,
        Month: month,
        BasicSalary: basicSalary,
        GrossSalary: grossSalary,
        TotalDeductions: totalDeductions,
        NetSalary: netSalary,
        Components: append(earnings, deductions...),
    }
}
Database Storage:

sql
CREATE TABLE payslips (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    employee_id UUID NOT NULL,
    pay_period_start DATE NOT NULL,
    pay_period_end DATE NOT NULL,
    payment_date DATE,
    basic_salary DECIMAL(12,2) NOT NULL,
    gross_salary DECIMAL(12,2) NOT NULL,
    total_deductions DECIMAL(12,2) NOT NULL,
    net_salary DECIMAL(12,2) NOT NULL,
    status VARCHAR(20), -- 'draft', 'generated', 'paid'
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE TABLE payslip_components (
    id UUID PRIMARY KEY,
    payslip_id UUID NOT NULL,
    component_name VARCHAR(100) NOT NULL,
    component_type VARCHAR(20) NOT NULL, -- 'earning' or 'deduction'
    amount DECIMAL(12,2) NOT NULL
);
Math & Precision:

Currency stored as DECIMAL(12,2) (e.g., 50000.00)
All calculations use decimal arithmetic (no floating point errors)
Rounding to 2 decimal places for currency
Performance Tracking
Metrics Calculated:

Attendance Rate: 
(Present Days / Total Working Days) * 100
Punctuality Rate: 
((Present Days - Late Days) / Total Working Days) * 100
Leave Utilization: 
(Leaves Taken / Leaves Allowed) * 100
Average Hours: 
Sum(Total Hours) / Number of Days
Storage:

Real-time calculations (no caching currently)
Future: Materialized views for performance
Future: Daily aggregation jobs
Completion Status
Overall Project Completion: 75%
75%
15%
10%
Project Completion Status
Completed
In Progress
Pending
Portal-Wise Completion
Portal	Completion	Status	Details
Super Admin	90%	✅ Complete	Tenant management, subscriptions, platform overview
Organization Admin	85%	✅ Complete	Employee CRUD, settings, reports (basic)
HR Portal	80%	✅ Complete	Employee management, attendance, leaves, payroll
Manager Portal	70%	🟡 Functional	Department view, leave approvals, team attendance
Team Lead Portal	65%	🟡 Functional	Team view, attendance tracking
Employee Portal	75%	✅ Complete	My attendance, leaves, payslips (just completed)
Feature-Wise Completion
✅ Completed Features (100%)
Authentication System

Login with email/password ✅
JWT token generation ✅
Refresh token mechanism ✅
Logout functionality ✅
Password hashing (Argon2id) ✅
User Management

Create users ✅
Update user profiles ✅
Activate/Deactivate users ✅
Role assignment ✅
Soft delete ✅
Employee Management

Create employees ✅
View employee list ✅
Edit employee details ✅
Department assignment ✅
Team assignment ✅
Employment status tracking ✅
Attendance System

Check-in/Check-out ✅
Attendance history ✅
Status calculation (present/late/absent) ✅
Date range filtering ✅
Export to CSV ✅
Leave Management

Leave types configuration ✅
Leave request submission ✅
Leave approval/rejection ✅
Leave balance tracking ✅
Leave history ✅
Payroll System

Salary components configuration ✅
Payslip generation ✅
Earnings calculation ✅
Deductions calculation ✅
Payslip viewing ✅
RBAC System

6-tier role hierarchy ✅
Middleware enforcement ✅
RLS policies (32 policies) ✅
Session context management ✅
Database

42 migrations ✅
RLS enabled on 27 tables ✅
Audit logging ✅
Soft delete support ✅
🟡 In Progress Features (70-90%)
Reports & Analytics (70%)

Basic dashboard stats ✅
Attendance reports ✅
Payroll reports ✅
Advanced analytics ⏳
Custom report builder ❌
Settings & Configuration (80%)

Attendance policies ✅
Leave types ✅
Salary components ✅
Company profile ⏳
Email templates ❌
UI/UX (75%)

Responsive design (desktop) ✅
Dark mode ✅
Loading states ✅
Error handling ✅
Mobile optimization ❌
Accessibility (WCAG) ⏳
❌ Pending Features (0-50%)
Advanced Security (40%)

2FA/MFA ❌
Rate limiting ❌
IP whitelisting ❌
Session anomaly detection ❌
Data encryption at rest ⏳
Notifications (30%)

Email notifications ⏳
In-app notifications ❌
SMS notifications ❌
Push notifications ❌
Integrations (10%)

Biometric device integration ⏳
Google Calendar sync ❌
Slack integration ❌
Zapier integration ❌
Advanced Features (20%)

Document management ❌
Performance reviews ❌
Training management ❌
Recruitment module ❌
Asset management ❌
Technical Debt & Refactoring Needed
Frontend

Mobile responsive design needed
Component library standardization
State management optimization
Bundle size optimization
Backend

API response caching
Database query optimization
Background job processing
Websocket for real-time updates
Database

Materialized views for reports
Partitioning for large tables
Index optimization
Query performance tuning
DevOps

CI/CD pipeline
Automated testing
Docker containerization
Kubernetes deployment
Competitive Analysis
Market Landscape
Based on research, the HRMS market in 2026 is dominated by:

Enterprise Solutions:

Workday (HCM suite for large enterprises)
SAP SuccessFactors (ERP-integrated HR)
Oracle HCM Cloud (Analytics-heavy)
ADP Workforce Now (Payroll-focused)
Mid-Market Solutions:

HiBob (Employee engagement focus)
Personio (European SMB focus)
BambooHR (User-friendly HR)
Rippling (HR + IT platform)
SMB Solutions:

Gusto (Payroll + HR for small teams)
Zoho People (Cost-effective, modular)
Deel (Global payroll & compliance)
PeopleOS Competitive Advantages
Feature	PeopleOS	Competitors	Advantage
Pricing	Affordable SaaS	$5-50/user/month	Lower cost for SMBs
Security	Multi-layer + RLS	Varies	Database-level isolation
Customization	Open architecture	Limited	Highly customizable
Multi-Tenancy	Native support	Some lack it	True SaaS architecture
RBAC Granularity	6-tier + RLS	Basic roles	Fine-grained control
Technology	Modern stack	Legacy systems	Better performance
Deployment	Cloud-native	Hybrid/On-prem	Easier scaling
What PeopleOS Can Improve
Enterprise Features: Lack advanced analytics, AI-powered insights
Integrations: Limited third-party integrations compared to Workday/SAP
Mobile App: No native mobile app (competitors have iOS/Android apps)
Global Compliance: Limited multi-country payroll support (vs Deel)
Brand Recognition: New product vs established players
Support: Limited support compared to enterprise vendors
Market Positioning
Target Market: Small to Medium Businesses (10-500 employees)
Value Proposition: Affordable, secure, and easy-to-use HRMS with enterprise-grade features
Differentiation: Database-level security (RLS) + Modern tech stack + Transparent pricing

Target User Personas
Persona 1: Growing Tech Startup
Company Profile:

Size: 50-100 employees
Industry: Software/Technology
Growth: Rapid (20% YoY)
Tech Savvy: High
Budget: $5,000-10,000/year for HRMS
Pain Points:

Manual attendance tracking in spreadsheets
No centralized leave management
Payroll errors due to manual calculations
Need for role-based access as team grows
Why PeopleOS:

Modern UI familiar to tech employees
API-first architecture for integrations
Affordable pricing for growing team
Scalable multi-tenant architecture
Persona 2: Traditional Manufacturing Company
Company Profile:

Size: 200-500 employees
Industry: Manufacturing
Growth: Stable (5% YoY)
Tech Savvy: Medium
Budget: $15,000-25,000/year for HRMS
Pain Points:

Complex shift management
Biometric attendance integration needed
Department-wise reporting required
Compliance with labor laws
Why PeopleOS:

Flexible attendance policies
Department and team hierarchy support
Audit trail for compliance
Customizable reports
Persona 3: Multi-Location Retail Chain
Company Profile:

Size: 100-300 employees across 10 locations
Industry: Retail
Growth: Expanding (15% YoY)
Tech Savvy: Low to Medium
Budget: $10,000-20,000/year for HRMS
Pain Points:

Managing employees across multiple locations
Different managers for different stores
Centralized payroll processing
Leave approval workflow
Why PeopleOS:

Multi-tenant architecture (each location can be a department)
Manager-level access for store managers
Centralized HR and payroll
Simple, intuitive interface
Persona 4: Professional Services Firm
Company Profile:

Size: 30-80 employees
Industry: Consulting/Legal/Accounting
Growth: Moderate (10% YoY)
Tech Savvy: Medium to High
Budget: $3,000-8,000/year for HRMS
Pain Points:

Project-based time tracking
Billable hours calculation
Client-wise employee allocation
Performance tracking
Why PeopleOS:

Flexible team structure (teams = projects)
Detailed attendance tracking
Customizable salary components
Role-based reporting
Future Roadmap
Phase 1: Core Enhancements (Q1-Q2 2026)
 Mobile responsive design
 Advanced reporting & analytics
 Email notifications
 2FA/MFA for admin accounts
 Rate limiting & DDoS protection
 Performance optimization
Phase 2: Advanced Features (Q3-Q4 2026)
 Native mobile apps (iOS/Android)
 Biometric device integration
 Document management
 Performance review module
 Training & development tracking
 Advanced analytics with AI insights
Phase 3: Enterprise Features (2027)
 Multi-country payroll support
 Advanced compliance tools
 Custom workflow builder
 API marketplace
 White-label solution
 On-premise deployment option
Appendix
A. Technology Versions
Backend:

Go: 1.24.0
PostgreSQL: 15+
Chi Router: 5.0.10
JWT: 5.2.0
pgx: 5.8.0
Frontend:

Next.js: 16.1.6
React: 19.2.3
TypeScript: 5.x
Tailwind CSS: 4.x
B. Environment Variables
Backend:

env
DATABASE_URL=postgresql://user:pass@localhost:5432/peopleos
JWT_SECRET=your-secret-key
PEPPER_SECRET=your-pepper-secret
ACCESS_TOKEN_TTL=60  # minutes
REFRESH_TOKEN_TTL=10080  # minutes (7 days)
PORT=8080
CORS_ORIGIN=http://localhost:3000
Frontend:

env
NEXT_PUBLIC_API_URL=http://localhost:8080
C. API Endpoints Summary
Total Endpoints: 50+

Authentication:

POST /api/v1/auth/login
POST /api/v1/auth/logout
POST /api/v1/auth/refresh
GET /api/v1/auth/me
Employees:

GET /api/v1/admin/employees
POST /api/v1/admin/employees
GET /api/v1/admin/employees/:id
PUT /api/v1/admin/employees/:id
DELETE /api/v1/admin/employees/:id
Attendance:

GET /api/v1/admin/attendance
POST /api/v1/employee/attendance/check-in
POST /api/v1/employee/attendance/check-out
GET /api/v1/employee/attendance
Leaves:

GET /api/v1/admin/leaves
POST /api/v1/employee/leaves
PUT /api/v1/admin/leaves/:id/approve
PUT /api/v1/admin/leaves/:id/reject
Payroll:

GET /api/v1/admin/payroll
POST /api/v1/admin/payroll/generate
GET /api/v1/employee/payslips
GET /api/v1/employee/payslips/:id
D. Database Migrations List
001_initial_schema.sql
 - Core tables
002_payslip_schema.sql
 - Payroll tables
003_biometric_support.sql
 - Biometric integration
004_leave_management.sql
 - Leave tables
005_system_management.sql
 - System settings
008_enable_rls.sql
 - Enable RLS
027_enable_row_level_security.sql
 - RLS policies
032_granular_rls_policies.sql
 - Granular RBAC
037_add_policy_tables.sql
 - Policy configuration
040_add_team_lead_role.sql
 - Team lead support
... (42 total migrations)

Conclusion
PeopleOS is a modern, secure, and scalable HRMS solution built with enterprise-grade technologies and security practices. While currently in prototype phase (75% complete), it demonstrates strong fundamentals in architecture, security, and functionality.

Key Strengths:

✅ Multi-layered security (8 layers implemented, 10 planned)
✅ Database-level access control (RLS)
✅ Modern technology stack
✅ Scalable multi-tenant architecture
✅ Comprehensive RBAC system
Areas for Improvement:

⏳ Mobile optimization
⏳ Advanced analytics
⏳ Third-party integrations
⏳ Enterprise features
Security Philosophy: We acknowledge that no system is 100% secure. Our approach is to make breaching our system economically unfeasible for most attackers through multiple security layers, continuous updates, and adherence to industry best practices.

Document Version: 1.0
Last Updated: February 12, 2026
Author: Ranjan Prasaila Yadav
Contact: 
rajanprasaila@gmail.com
 | +91 7013146154

NOTE

This is a living document and will be updated as the project evolves. For the latest information, please refer to the project repository and documentation.