export interface AttendanceRecord {
    id: string;
    employee_id: string;
    date: string;
    check_in_time?: string;
    check_out_time?: string;
    status: 'present' | 'absent' | 'partial' | 'late' | 'holiday';
    total_hours?: number;
}

export interface AttendanceStats {
    total_present: number;
    total_absent: number;
    total_late: number;
    total_on_leave: number;
}
