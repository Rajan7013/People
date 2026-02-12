-- Drop Soft Delete Rules to allow Hard Delete
-- We are moving to a model where "Delete" means "Hard Delete" for cleanup,
-- and "Suspend" handles the soft-state.

-- Drop rules for Tenants
DROP RULE IF EXISTS soft_delete_tenants ON tenants;

-- Drop rules for Users
DROP RULE IF EXISTS soft_delete_users ON users;

-- Drop rules for Employees
DROP RULE IF EXISTS soft_delete_employees ON employees;

-- Drop rules for Departments
DROP RULE IF EXISTS soft_delete_departments ON departments;

-- Drop the prevent_hard_delete trigger on users if it exists
DROP TRIGGER IF EXISTS prevent_hard_delete_users ON users;
