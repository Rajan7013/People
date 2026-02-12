-- Update employment_status check constraint to include 'suspended'

ALTER TABLE employees DROP CONSTRAINT employees_employment_status_check;

ALTER TABLE employees ADD CONSTRAINT employees_employment_status_check 
    CHECK (employment_status IN ('active', 'inactive', 'terminated', 'suspended'));
