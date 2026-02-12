# PeopleOS API Testing

## Test the server startup

```bash
# Start the development environment
make dev

# Test health endpoint
curl http://localhost:8080/health

# Test readiness endpoint  
curl http://localhost:8080/ready

# Test login with sample data
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@acme.com",
    "password": "admin123"
  }'

# Test protected endpoint (use token from login response)
curl -X GET http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN_HERE"

# Test tenant-scoped endpoint
curl -X GET http://localhost:8080/api/v1/550e8400-e29b-41d4-a716-446655440001/employees \
  -H "Authorization: Bearer YOUR_JWT_TOKEN_HERE"
```

## Sample Users for Testing

- **Admin**: admin@acme.com / admin123
- **HR**: hr@acme.com / hr123  
- **Manager**: manager@acme.com / manager123
- **Employee**: employee@acme.com / employee123

## Expected Responses

### Health Check
```json
{"status":"ok","service":"peopleos-api"}
```

### Login Success
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-11-21T10:30:00Z",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440002",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440001",
    "email": "admin@acme.com",
    "role": "admin",
    "first_name": "John",
    "last_name": "Admin"
  }
}
```

### Profile Response
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440002",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440001", 
  "email": "admin@acme.com",
  "role": "admin"
}
```