# PeopleOS - Modern HRMS Platform

<div align="center">

![PeopleOS Logo](https://img.shields.io/badge/PeopleOS-HRMS-blue?style=for-the-badge)
[![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![Next.js](https://img.shields.io/badge/Next.js-16.1-black?style=flat-square&logo=next.js)](https://nextjs.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791?style=flat-square&logo=postgresql)](https://postgresql.org)
[![License](https://img.shields.io/badge/License-Proprietary-red?style=flat-square)](LICENSE)

**A modern, secure, and scalable Human Resource Management System for SMBs**

[Features](#features) • [Quick Start](#quick-start) • [Documentation](#documentation) • [Architecture](#architecture) • [Security](#security)

</div>

---

> [!CAUTION]
> **Prototype Version**: This is a testing/prototype version (v0.1.0) and is **NOT production-ready**. Use for evaluation and testing purposes only. See [Security Disclaimer](#security-disclaimer) for details.

---

## 📋 Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Technology Stack](#technology-stack)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Security](#security)
- [API Documentation](#api-documentation)
- [Database Schema](#database-schema)
- [RBAC System](#rbac-system)
- [Contributing](#contributing)
- [Roadmap](#roadmap)
- [License](#license)
- [Contact](#contact)

---

## 🎯 Overview

**PeopleOS** is a comprehensive Human Resource Management System (HRMS) designed for small to medium-sized businesses. Built with modern technologies and enterprise-grade security practices, PeopleOS streamlines employee management, attendance tracking, leave management, and payroll processing.

### Key Highlights

- 🏢 **Multi-Tenant Architecture** - Complete data isolation between organizations
- 🔐 **8-Layer Security** - Enterprise-grade security with PostgreSQL RLS
- 👥 **6-Tier RBAC** - Granular role-based access control
- ⚡ **High Performance** - Go backend with optimized PostgreSQL queries
- 📱 **Modern UI** - Built with Next.js 16 and React 19
- 🔄 **Real-time Updates** - Live attendance tracking and notifications
- 📊 **Analytics Dashboard** - Comprehensive HR metrics and reports
- 🌐 **RESTful API** - Well-documented API for integrations

---

## ✨ Features

### Core Modules

#### 👤 Employee Management
- ✅ Employee CRUD operations
- ✅ Department and team assignment
- ✅ Employment status tracking
- ✅ Role-based access control
- ✅ Bulk import/export (CSV)
- ✅ Advanced search and filtering

#### ⏰ Attendance System
- ✅ Check-in/Check-out tracking
- ✅ Automatic status calculation (Present/Late/Absent)
- ✅ Configurable attendance policies
- ✅ Grace period management
- ✅ Overtime tracking
- ✅ Date range reports

#### 🏖️ Leave Management
- ✅ Multiple leave types (Sick, Casual, Earned, etc.)
- ✅ Leave request workflow
- ✅ Manager/HR approval system
- ✅ Leave balance tracking
- ✅ Calendar integration
- ✅ Leave history and reports

#### 💰 Payroll System
- ✅ Automated payslip generation
- ✅ Configurable salary components
- ✅ Earnings and deductions
- ✅ Attendance-based calculations
- ✅ Late fines and absent deductions
- ✅ Payslip viewing and download

#### 📊 Reports & Analytics
- ✅ Dashboard with key metrics
- ✅ Attendance reports
- ✅ Leave reports
- ✅ Payroll reports
- ✅ Employee performance metrics
- ⏳ Custom report builder (coming soon)

#### ⚙️ Settings & Configuration
- ✅ Attendance policies
- ✅ Leave types management
- ✅ Salary components
- ✅ Company profile
- ⏳ Email templates (coming soon)
- ⏳ Notification settings (coming soon)

---

## 🛠️ Technology Stack

### Backend

| Technology | Version | Purpose |
|------------|---------|---------|
| **Go** | 1.24.0 | Backend language |
| **Chi Router** | 5.0.10 | HTTP routing |
| **PostgreSQL** | 15+ | Primary database |
| **pgx** | 5.8.0 | PostgreSQL driver |
| **JWT** | 5.2.0 | Authentication |
| **Argon2id** | Latest | Password hashing |
| **Zerolog** | 1.31.0 | Structured logging |

### Frontend

| Technology | Version | Purpose |
|------------|---------|---------|
| **Next.js** | 16.1.6 | React framework |
| **React** | 19.2.3 | UI library |
| **TypeScript** | 5.x | Type safety |
| **Tailwind CSS** | 4.x | Styling |
| **shadcn/ui** | Latest | UI components |
| **TanStack Table** | 8.21.3 | Data tables |
| **Axios** | 1.13.5 | HTTP client |
| **Zod** | 4.3.6 | Schema validation |

### Database

- **PostgreSQL 15+** with Row-Level Security (RLS)
- **42 Migrations** for schema management
- **Audit logging** with database triggers
- **Soft deletes** for data retention

---

## 🏗️ Architecture

### System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Client Layer                          │
│                      (Web Browser)                           │
└────────────────────────┬────────────────────────────────────┘
                         │ HTTPS
┌────────────────────────▼────────────────────────────────────┐
│                    Frontend Layer                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Next.js    │  │  React UI    │  │   Axios      │      │
│  │   Routing    │  │  Components  │  │   Client     │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└────────────────────────┬────────────────────────────────────┘
                         │ REST API
┌────────────────────────▼────────────────────────────────────┐
│                     Backend Layer                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Chi Router   │→ │ Middleware   │→ │  Handlers    │      │
│  │              │  │ (Auth/RBAC)  │  │              │      │
│  └──────────────┘  └──────────────┘  └──────┬───────┘      │
│                                              │               │
│  ┌──────────────┐  ┌──────────────┐        │               │
│  │  Services    │← │   Models     │←───────┘               │
│  └──────┬───────┘  └──────────────┘                        │
└─────────┼──────────────────────────────────────────────────┘
          │
┌─────────▼──────────────────────────────────────────────────┐
│                    Database Layer                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ PostgreSQL   │  │     RLS      │  │ Audit Logs   │      │
│  │   Tables     │  │   Policies   │  │              │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

### Request Flow

1. **Client** sends HTTPS request with JWT token
2. **Frontend** validates input and makes API call
3. **Middleware** validates JWT and checks RBAC permissions
4. **Handler** processes request and calls service layer
5. **Service** applies business logic and queries database
6. **Database** enforces RLS policies and returns filtered data
7. **Response** flows back through layers to client

---

## 🚀 Quick Start

### Prerequisites

- **Go** 1.24 or higher
- **Node.js** 18 or higher
- **PostgreSQL** 15 or higher
- **Git**

### Clone Repository

```bash
git clone https://github.com/rajanprasaila/PeopleOS.git
cd PeopleOS
```

### Backend Setup

```bash
cd Backend

# Install dependencies
go mod download

# Create .env file
cp .env.example .env

# Edit .env with your database credentials
nano .env

# Run migrations
go run cmd/migrate/main.go up

# Start server
go run cmd/server/main.go
```

Backend will start on `http://localhost:8080`

### Frontend Setup

```bash
cd Frontend

# Install dependencies
npm install

# Create .env.local file
cp .env.example .env.local

# Edit .env.local with API URL
nano .env.local

# Start development server
npm run dev
```

Frontend will start on `http://localhost:3000`

### Default Credentials

**Super Admin:**
- Email: `admin@peopleos.com`
- Password: `admin123` (Change immediately!)

---

## ⚙️ Configuration

### Backend Environment Variables

```env
# Database
DATABASE_URL=postgresql://user:password@localhost:5432/peopleos

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-this
PEPPER_SECRET=your-pepper-secret-for-passwords
ACCESS_TOKEN_TTL=60          # minutes
REFRESH_TOKEN_TTL=10080      # minutes (7 days)

# Server
PORT=8080
ENVIRONMENT=development

# CORS
CORS_ORIGIN=http://localhost:3000
```

### Frontend Environment Variables

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### Database Configuration

1. Create PostgreSQL database:
```sql
CREATE DATABASE peopleos;
```

2. Run migrations:
```bash
cd Backend
go run cmd/migrate/main.go up
```

3. Verify migrations:
```bash
go run cmd/migrate/main.go status
```

---

## 📖 Usage

### Creating Your First Organization

1. **Login as Super Admin**
   - Navigate to `http://localhost:3000`
   - Login with super admin credentials

2. **Create Tenant**
   - Go to Super Admin Dashboard
   - Click "Create Organization"
   - Fill in organization details

3. **Create Admin User**
   - Navigate to Users section
   - Create organization admin
   - Assign role: "admin"

4. **Configure Organization**
   - Login as organization admin
   - Go to Settings
   - Configure attendance policies
   - Add leave types
   - Set up salary components

5. **Add Employees**
   - Navigate to Employees
   - Click "Add Employee"
   - Fill in employee details
   - Assign department and role

### Role-Based Access

| Role | Access Level |
|------|-------------|
| **Super Admin** | Platform-wide access, all tenants |
| **Org Admin** | Full access within organization |
| **HR** | Employee management, payroll, leaves |
| **Manager** | Department employees, approvals |
| **Team Lead** | Team members, attendance |
| **Employee** | Own data only |

---

## 🔐 Security

### Security Layers

PeopleOS implements **8 layers of security** with 2 more planned:

1. **Transport Security** - HTTPS, CORS, Secure Headers
2. **Authentication** - Argon2id + Pepper, JWT tokens
3. **Authorization** - 6-tier RBAC, Middleware enforcement
4. **Database Security** - PostgreSQL RLS, Tenant isolation
5. **Data Protection** - Soft deletes, Audit logging
6. **Input Validation** - Zod schemas, Type safety
7. **Session Management** - Token expiration, Refresh rotation
8. **Monitoring** - Structured logging, Audit trails

### Password Security

- **Hashing Algorithm**: Argon2id (Winner of Password Hashing Competition)
- **Pepper**: Additional secret beyond salt
- **Parameters**: 64MB memory, 3 iterations, parallelism=2
- **Storage**: Never store plain text passwords

### Token Management

- **Access Token**: 60 minutes (configurable)
- **Refresh Token**: 7 days (configurable)
- **Storage**: HttpOnly cookies (XSS protection)
- **Transmission**: HTTPS only

### Row-Level Security (RLS)

PostgreSQL RLS ensures complete data isolation:

```sql
-- Example: Employees can only see their own data
CREATE POLICY employees_employee_self ON employees
    FOR SELECT
    USING (
        current_user_role() = 'employee' AND
        user_id = current_user_id()
    );
```

**32+ RLS policies** enforce access control at database level.

### Security Disclaimer

> [!WARNING]
> **This is a prototype version (v0.1.0)** and should NOT be used in production without:
> - Security audit and penetration testing
> - Implementation of additional security layers (2FA, rate limiting, etc.)
> - Regular security updates and patches
> - Proper backup and disaster recovery procedures
>
> We do NOT claim this system is 100% secure or hacking-proof. Zero-day vulnerabilities may exist. Use at your own risk.

---

## 📚 API Documentation

### Base URL

```
http://localhost:8080/api/v1
```

### Authentication

All protected endpoints require JWT token in cookie or `Authorization` header:

```
Authorization: Bearer <token>
```

### Endpoints Overview

#### Authentication
- `POST /auth/login` - User login
- `POST /auth/logout` - User logout
- `POST /auth/refresh` - Refresh access token
- `GET /auth/me` - Get current user

#### Employees
- `GET /admin/employees` - List employees
- `POST /admin/employees` - Create employee
- `GET /admin/employees/:id` - Get employee details
- `PUT /admin/employees/:id` - Update employee
- `DELETE /admin/employees/:id` - Delete employee

#### Attendance
- `GET /admin/attendance` - Get attendance records
- `POST /employee/attendance/check-in` - Check in
- `POST /employee/attendance/check-out` - Check out
- `GET /employee/attendance` - Get own attendance

#### Leaves
- `GET /admin/leaves` - List leave requests
- `POST /employee/leaves` - Submit leave request
- `PUT /admin/leaves/:id/approve` - Approve leave
- `PUT /admin/leaves/:id/reject` - Reject leave

#### Payroll
- `GET /admin/payroll` - List payslips
- `POST /admin/payroll/generate` - Generate payslips
- `GET /employee/payslips` - Get own payslips
- `GET /employee/payslips/:id` - Get payslip details

**Full API documentation**: See [API.md](docs/API.md)

---

## 🗄️ Database Schema

### Core Tables

- **tenants** - Organizations/companies
- **users** - User accounts
- **employees** - Employee records
- **departments** - Organizational departments
- **teams** - Work teams
- **attendance_records** - Daily attendance
- **leave_applications** - Leave requests
- **leave_types** - Leave categories
- **payslips** - Salary slips
- **salary_components** - Earnings/deductions
- **audit_logs** - System audit trail

### Migrations

42 database migrations manage schema evolution:

```bash
# List migrations
go run cmd/migrate/main.go status

# Apply migrations
go run cmd/migrate/main.go up

# Rollback
go run cmd/migrate/main.go down
```

**Schema documentation**: See [DATABASE.md](docs/DATABASE.md)

---

## 👥 RBAC System

### Role Hierarchy

```
Super Admin (Platform Owner)
    ↓
Organization Admin (Company Owner)
    ↓
HR Manager ←→ Department Manager
    ↓              ↓
Team Lead ←────────┘
    ↓
Employee
```

### Permission Matrix

See [RBAC.md](docs/RBAC.md) for complete permission matrix.

### Implementation

RBAC is enforced at **3 levels**:

1. **Middleware** - HTTP request level
2. **Service Layer** - Business logic level
3. **Database RLS** - Data access level

---

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

### Code Standards

- **Go**: Follow [Effective Go](https://golang.org/doc/effective_go)
- **TypeScript**: Use ESLint configuration
- **Commits**: Use [Conventional Commits](https://www.conventionalcommits.org/)
- **Testing**: Write tests for new features

---

## 🗺️ Roadmap

### Phase 1: Core Enhancements (Q1-Q2 2026)
- [ ] Mobile responsive design
- [ ] Advanced reporting & analytics
- [ ] Email notifications
- [ ] 2FA/MFA authentication
- [ ] Rate limiting & DDoS protection

### Phase 2: Advanced Features (Q3-Q4 2026)
- [ ] Native mobile apps (iOS/Android)
- [ ] Biometric device integration
- [ ] Document management
- [ ] Performance review module
- [ ] AI-powered insights

### Phase 3: Enterprise Features (2027)
- [ ] Multi-country payroll
- [ ] Advanced compliance tools
- [ ] Custom workflow builder
- [ ] API marketplace
- [ ] White-label solution

**Full roadmap**: See [ROADMAP.md](docs/ROADMAP.md)

---

## 📊 Project Status

**Current Version**: 0.1.0 (Prototype)  
**Completion**: 75%

| Module | Status | Completion |
|--------|--------|------------|
| Authentication | ✅ Complete | 100% |
| Employee Management | ✅ Complete | 100% |
| Attendance System | ✅ Complete | 100% |
| Leave Management | ✅ Complete | 100% |
| Payroll System | ✅ Complete | 100% |
| RBAC System | ✅ Complete | 100% |
| Reports & Analytics | 🟡 Functional | 70% |
| Settings | 🟡 Functional | 80% |
| Mobile UI | ❌ Pending | 0% |
| Integrations | ❌ Pending | 10% |

---

## 📄 License

**Proprietary License** - Copyright © 2026 Ranjan Prasaila Yadav

This software is proprietary and confidential. Unauthorized copying, distribution, or use is strictly prohibited.

For licensing inquiries, contact: rajanprasaila@gmail.com

---

## 📞 Contact

**Developer**: Ranjan Prasaila Yadav

- **Email**: rajanprasaila@gmail.com
- **WhatsApp**: +91 7013146154
- **GitHub**: [@rajanprasaila](https://github.com/rajanprasaila)

---

## 🙏 Acknowledgments

- **Go Community** - For excellent backend tools
- **Next.js Team** - For amazing React framework
- **PostgreSQL** - For robust database with RLS
- **shadcn/ui** - For beautiful UI components

---

## 📝 Additional Documentation

- [Installation Guide](docs/INSTALLATION.md)
- [API Reference](docs/API.md)
- [Database Schema](docs/DATABASE.md)
- [RBAC System](docs/RBAC.md)
- [Security Guide](docs/SECURITY.md)
- [Deployment Guide](docs/DEPLOYMENT.md)
- [Contributing Guidelines](CONTRIBUTING.md)
- [Changelog](CHANGELOG.md)

---

<div align="center">

**Built with ❤️ by Ranjan Prasaila Yadav**

⭐ Star this repo if you find it helpful!

</div>