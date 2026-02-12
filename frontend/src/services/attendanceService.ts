import api from '@/lib/api';

export interface AttendanceRecord {
    id: string;
    user_id: string;
    tenant_id: string;
    date: string;
    check_in?: string;
    check_out?: string;
    status: 'present' | 'absent' | 'late' | 'half_day' | 'on_leave';
    notes?: string;
    created_at: string;
    updated_at: string;
    employee_name?: string; // Joined field
}

export interface AttendanceStats {
    total_employees: number;
    present: number;
    absent: number;
    late: number;
    on_leave: number;
    attendance_rate: number;
}

export const attendanceService = {
    // Employee Actions
    checkIn: (data: { notes?: string; latitude?: number; longitude?: number }) =>
        api.post('/company/employee/attendance/checkin', data),

    checkOut: (data: { notes?: string; latitude?: number; longitude?: number }) =>
        api.post('/company/employee/attendance/checkout', data),

    getCurrentStatus: () =>
        api.get<{ is_checked_in: boolean; last_check_in: string | null; last_check_out: string | null }>('/company/employee/attendance/my-status'),

    getMyAttendance: (params?: { from?: string; to?: string }) =>
        api.get<AttendanceRecord[]>('/company/employee/attendance', { params }),

    // Manager/Team Lead Actions
    getTeamAttendance: (role: 'manager' | 'team-lead', params?: { date?: string }) =>
        api.get<AttendanceRecord[]>(`/company/${role}/attendance`, { params }),

    // HR/Admin Actions
    // Admin doesn't have explicit attendance routes in server.go, so we rely on HR routes logic or assume Admin can access HR.
    getAllRecords: (role: 'hr', params?: { date?: string; department_id?: string }) =>
        api.get<AttendanceRecord[]>(`/company/${role}/attendance`, { params }),

    getDailyStats: (role: 'hr', params?: { date?: string }) =>
        api.get<AttendanceStats>(`/company/${role}/attendance/stats`, { params }),

    // Individual record management
    updateRecord: (role: 'hr', id: string, data: any) =>
        api.put(`/company/${role}/attendance/records/${id}`, data),
};
