export type LeaveStatus = "pending" | "approved" | "rejected" | "cancelled";
export type LeaveType = "annual" | "sick" | "casual" | "maternity" | "paternity" | "unpaid";

export interface LeaveRequest {
    id: string;
    employee_id: string;
    leave_type: LeaveType;
    start_date: string;
    end_date: string;
    reason: string;
    status: LeaveStatus;
    rejection_reason?: string;
    created_at: string;
    employee?: {
        first_name: string;
        last_name: string;
    }
}

export interface CreateLeaveRequest {
    leave_type: LeaveType;
    start_date: string;
    end_date: string;
    reason: string;
}
