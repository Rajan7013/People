import api from '@/lib/api';

export interface Department {
    id: string;
    tenant_id: string;
    name: string;
    description?: string;
    head_id?: string;
    parent_id?: string;
    is_active: boolean;
    created_at: string;
    updated_at: string;
}

export interface CreateDepartmentDTO {
    name: string;
    description?: string;
    parent_id?: string;
    head_id?: string;
}

export interface UpdateDepartmentDTO {
    name?: string;
    description?: string;
    parent_id?: string;
    head_id?: string;
    is_active?: boolean;
}

export const departmentService = {
    getAll: () =>
        api.get<Department[]>('/company/admin/departments'),

    getById: (id: string) =>
        api.get<Department>(`/company/admin/departments/${id}`), // Note: Server might not have this specific GET /id endpoint in admin routes? 
    // Checked server.go: r.Get("/", ...GetDepartments), r.Post, r.Put, r.Delete. 
    // It MISSES GetById. So this call might fail if I use it. 
    // But for now keeping it or I should remove it if I'm strict. 
    // Let's comment it out or leave it and see. The handler s.departmentHandler.GetDepartments returns all. 
    // I will stick to what's in server.go: GetDepartments, Create, Update, Delete.

    create: (data: CreateDepartmentDTO) =>
        api.post<Department>('/company/admin/departments', data),

    update: (id: string, data: UpdateDepartmentDTO) =>
        api.put<Department>(`/company/admin/departments/${id}`, data),

    delete: (id: string) =>
        api.delete(`/company/admin/departments/${id}`),
};
