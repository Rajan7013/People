import api from '@/lib/api';

export interface LeaveRequest {
    id: string;
    user_id: string;
    tenant_id: string;
    leave_type: 'sick' | 'casual' | 'vacation' | 'maternity' | 'paternity' | 'unpaid';
    start_date: string;
    end_date: string;
    reason: string;
    status: 'pending' | 'approved' | 'rejected' | 'cancelled';
    approved_by?: string;
    rejection_reason?: string;
    created_at: string;
    updated_at: string;
    employee_name?: string; // Joined field
    duration_days?: number;
}

export interface CreateLeaveDTO {
    leave_type: string;
    start_date: string;
    end_date: string;
    reason: string;
}

export const leaveService = {
    // Employee Actions
    apply: (data: CreateLeaveDTO) =>
        api.post('/company/employee/leaves', data),

    getMyLeaves: () =>
        api.get<LeaveRequest[]>('/company/employee/leaves'),

    // Manager Actions
    getDepartmentLeaves: () =>
        api.get<LeaveRequest[]>('/company/manager/leaves'),

    approve: (role: 'manager' | 'hr', id: string) =>
        api.put(`/company/${role}/leaves/${id}/approve`),

    reject: (role: 'manager' | 'hr', id: string, reason?: string) =>
        api.put(`/company/${role}/leaves/${id}/reject`, { reason }),

    // HR Actions
    getAllLeaves: (role: 'hr' = 'hr', params?: { status?: string }) =>
        api.get<LeaveRequest[]>(`/company/${role}/leaves`, { params }),
};
