import api from '@/lib/api';

export interface Employee {
    id: string;
    tenant_id: string;
    user_id?: string;
    first_name: string;
    last_name: string;
    email: string;
    phone?: string;
    position?: string;
    department_id?: string;
    department?: string; // department name
    manager_id?: string;
    date_of_birth?: string;
    hire_date?: string;
    role: string;
    status: 'active' | 'inactive' | 'terminated';
    address?: string;
    created_at: string;
    updated_at: string;
}

export interface CreateEmployeeDTO {
    first_name: string;
    last_name: string;
    email: string;
    role: string;
    position?: string;
    department_id?: string;
    phone?: string;
    hire_date?: string;
}

export interface UpdateEmployeeDTO {
    first_name?: string;
    last_name?: string;
    department_id?: string;
    position?: string;
    phone?: string;
    status?: string;
    role?: string;
}

export const employeeService = {
    // Admin/HR Actions
    getAll: (role: 'admin' | 'hr' = 'admin') =>
        api.get<Employee[]>(`/company/${role}/employees`),

    getById: (role: 'admin' | 'hr', id: string) =>
        api.get<Employee>(`/company/${role}/employees/${id}`),

    create: (data: CreateEmployeeDTO) =>
        api.post<Employee>('/company/admin/employees', data), // Only Admin usually creates? Or HR too? server.go allows HR?
    // server.go: HR has r.Get("/employees"), no POST. Only Admin has POST.

    update: (id: string, data: UpdateEmployeeDTO) => // Admin only for now based on server.go
        api.put<Employee>(`/company/admin/employees/${id}`, data),

    delete: (id: string) => // Admin only
        api.delete(`/company/admin/employees/${id}`),

    // For Managers and Team Leads
    getMyTeam: (role: 'manager' | 'team-lead') =>
        api.get<Employee[]>(`/company/${role}/team`),
};
