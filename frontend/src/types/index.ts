export interface User {
    id: string;
    tenant_id: string;
    email: string;
    role: 'super_admin' | 'admin' | 'hr' | 'manager' | 'team_lead' | 'employee';
    first_name?: string;
    last_name?: string;
}

export interface AuthResponse {
    user: User;
    token?: string; // If using Bearer tokens manually
    expires_at?: string;
}

export interface ApiError {
    error: string;
}
