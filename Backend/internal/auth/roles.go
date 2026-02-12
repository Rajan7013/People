package auth

// Role constants
const (
	RoleSuperAdmin  = "super_admin"
	RoleTenantAdmin = "admin" // Tenant Owner
	RoleHRManager   = "hr_manager"
	RoleEmployee    = "employee"
	RoleGuest       = "guest"
)

// RoleLevel defines the hierarchy of roles
// Higher number = Higher privilege
var RoleLevels = map[string]int{
	RoleSuperAdmin:  100,
	RoleTenantAdmin: 50,
	RoleHRManager:   30,
	RoleEmployee:    10,
	RoleGuest:       0,
}

// GetRoleLevel returns the integer level of a role
func GetRoleLevel(role string) int {
	if level, ok := RoleLevels[role]; ok {
		return level
	}
	return 0 // Default to lowest level
}

// HasRole checks if the user's role satisfies the required role
// Returns true if userRoleLevel >= requiredRoleLevel
func HasRole(userRole, requiredRole string) bool {
	return GetRoleLevel(userRole) >= GetRoleLevel(requiredRole)
}
