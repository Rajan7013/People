-- Add payslip-related tables
-- Migration: 002_payslip_schema.sql (Renumbered from 004)

-- Create salary components table for flexible salary structure
CREATE TABLE salary_components (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) CHECK (type IN ('earning', 'deduction')) NOT NULL,
    is_percentage BOOLEAN DEFAULT FALSE,
    is_taxable BOOLEAN DEFAULT TRUE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create employee salary structure table
CREATE TABLE employee_salary_structures (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    employee_id UUID REFERENCES employees(id) ON DELETE CASCADE,
    basic_salary DECIMAL(15,2) NOT NULL,
    effective_date DATE NOT NULL,
    end_date DATE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create salary structure components (allowances and deductions)
CREATE TABLE salary_structure_components (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    salary_structure_id UUID REFERENCES employee_salary_structures(id) ON DELETE CASCADE,
    component_id UUID REFERENCES salary_components(id) ON DELETE CASCADE,
    amount DECIMAL(15,2),
    percentage DECIMAL(5,2),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create payslips table
CREATE TABLE payslips (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    employee_id UUID REFERENCES employees(id) ON DELETE CASCADE,
    salary_structure_id UUID REFERENCES employee_salary_structures(id),
    pay_period_start DATE NOT NULL,
    pay_period_end DATE NOT NULL,
    payment_date DATE,
    basic_salary DECIMAL(15,2) NOT NULL,
    gross_salary DECIMAL(15,2) NOT NULL,
    total_deductions DECIMAL(15,2) NOT NULL DEFAULT 0,
    net_salary DECIMAL(15,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'approved', 'paid', 'cancelled')),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create payslip components table (detailed breakdown)
CREATE TABLE payslip_components (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payslip_id UUID REFERENCES payslips(id) ON DELETE CASCADE,
    component_id UUID REFERENCES salary_components(id) ON DELETE CASCADE,
    component_name VARCHAR(100) NOT NULL, -- Stored for historical accuracy
    component_type VARCHAR(20) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX idx_salary_components_tenant ON salary_components(tenant_id);
CREATE INDEX idx_salary_components_type ON salary_components(type);

CREATE INDEX idx_employee_salary_structures_tenant ON employee_salary_structures(tenant_id);
CREATE INDEX idx_employee_salary_structures_employee ON employee_salary_structures(employee_id);
CREATE INDEX idx_employee_salary_structures_active ON employee_salary_structures(is_active);
CREATE INDEX idx_employee_salary_structures_dates ON employee_salary_structures(effective_date, end_date);

CREATE INDEX idx_payslips_tenant ON payslips(tenant_id);
CREATE INDEX idx_payslips_employee ON payslips(employee_id);
CREATE INDEX idx_payslips_period ON payslips(pay_period_start, pay_period_end);
CREATE INDEX idx_payslips_status ON payslips(status);
CREATE INDEX idx_payslips_payment_date ON payslips(payment_date);

CREATE INDEX idx_payslip_components_payslip ON payslip_components(payslip_id);
CREATE INDEX idx_payslip_components_type ON payslip_components(component_type);

-- Add triggers for updated_at
CREATE TRIGGER update_salary_components_updated_at 
    BEFORE UPDATE ON salary_components 
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER update_employee_salary_structures_updated_at 
    BEFORE UPDATE ON employee_salary_structures 
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER update_payslips_updated_at 
    BEFORE UPDATE ON payslips 
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
