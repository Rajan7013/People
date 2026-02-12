// Super Admin API Service
import api from '@/lib/api';

// ============================================================================
// ORGANIZATIONS
// ============================================================================

export interface Organization {
    id: string;
    name: string;
    subdomain: string;
    domain?: string;
    status: 'active' | 'suspended' | 'inactive';
    country?: string;
    admin_email?: string;
    storage_used_mb: number;
    api_requests_count: number;
    created_at: string;
    updated_at: string;
    // Joined fields
    subscription?: {
        plan?: {
            display_name: string;
        };
        plan_id?: string;
        status: string;
        billing_cycle?: string;
        start_date?: string;
        end_date?: string;
        auto_renew?: boolean;
        current_period_end: string;
    };
    total_users?: number;
    active_employees?: number;
}

export interface CreateOrganizationRequest {
    name: string;
    subdomain: string;
    domain?: string;
    country?: string;
    admin_name: string;
    admin_email: string;
    admin_password: string;
    plan_id: string;
    billing_cycle: 'monthly' | 'yearly';
    trial_days?: number;
    // Organization details
    contact_number?: string;
    website?: string;
    address_line1?: string;
    address_line2?: string;
    city?: string;
    state?: string;
    postal_code?: string;
    currency?: string;
    timezone?: string;
}

export const organizationService = {
    getAll: (params?: { status?: string; plan?: string; country?: string; search?: string }) =>
        api.get<{ organizations: Organization[]; total: number }>('/platform/organizations', { params }),

    getById: (id: string) =>
        api.get<Organization>(`/platform/organizations/${id}`),

    create: (data: CreateOrganizationRequest) =>
        api.post<Organization>('/platform/organizations', data),

    update: (id: string, data: Partial<Organization>) =>
        api.put<Organization>(`/platform/organizations/${id}`, data),

    delete: (id: string) =>
        api.delete(`/platform/organizations/${id}`),

    block: (id: string) =>
        api.post(`/platform/organizations/${id}/block`),

    unblock: (id: string) =>
        api.post(`/platform/organizations/${id}/unblock`),

    renew: (id: string, data: { plan_id: string; billing_cycle: string }) =>
        api.post(`/platform/organizations/${id}/renew`, data),
};

// ============================================================================
// SUBSCRIPTION PLANS
// ============================================================================

export interface SubscriptionPlan {
    id: string;
    name: string;
    display_name: string;
    description?: string;
    price_monthly: number;
    price_yearly: number;
    currency: string;
    max_users?: number;
    max_storage_gb?: number;
    max_api_requests_monthly?: number;
    max_departments?: number;
    features: Record<string, any>;
    is_active: boolean;
    is_visible: boolean;
    sort_order: number;
}

export const planService = {
    getAll: () =>
        api.get<{ plans: SubscriptionPlan[] }>('/platform/plans'),

    getById: (id: string) =>
        api.get<SubscriptionPlan>(`/platform/plans/${id}`),

    create: (data: Partial<SubscriptionPlan>) =>
        api.post<SubscriptionPlan>('/platform/plans', data),

    update: (id: string, data: Partial<SubscriptionPlan>) =>
        api.put<SubscriptionPlan>(`/platform/plans/${id}`, data),

    delete: (id: string) =>
        api.delete(`/platform/plans/${id}`),
};

// ============================================================================
// INVOICES
// ============================================================================

export interface Invoice {
    id: string;
    invoice_number: string;
    tenant_id: string;
    tenant_name?: string;
    subtotal: number;
    tax_amount: number;
    total_amount: number;
    currency: string;
    status: 'draft' | 'pending' | 'paid' | 'overdue' | 'cancelled';
    issue_date: string;
    due_date: string;
    paid_at?: string;
    line_items: any[];
}

export const invoiceService = {
    getAll: (params?: { status?: string; tenant_id?: string; from?: string; to?: string }) =>
        api.get<{ invoices: Invoice[]; total: number }>('/platform/invoices', { params }),

    getById: (id: string) =>
        api.get<Invoice>(`/platform/invoices/${id}`),

    markAsPaid: (id: string) =>
        api.post(`/platform/invoices/${id}/mark-paid`),

    generate: (data: Partial<Invoice>) =>
        api.post<Invoice>('/platform/invoices/generate', data),

    update: (id: string, data: Partial<Invoice>) =>
        api.put<{ invoice: Invoice }>(`/platform/invoices/${id}`, data),

    delete: (id: string) =>
        api.delete(`/platform/invoices/${id}`),

    download: (id: string) =>
        api.get(`/platform/invoices/${id}/download`, { responseType: 'blob' }),
};

// ============================================================================
// ANALYTICS
// ============================================================================

export interface PlatformStats {
    total_tenants: number;
    active_tenants: number;
    total_subscriptions: number;
    active_subscriptions: number;
    total_users: number;
    mrr: number;
    arr: number;
    total_storage_gb: number;
}

export interface TenantGrowth {
    date: string;
    new_tenants: number;
    total_tenants: number;
}

export interface RevenueMetrics {
    total_revenue: number;
    paid_revenue: number;
    pending_revenue: number;
    revenue_by_plan: Array<{
        plan_name: string;
        revenue: number;
        tenant_count: number;
    }>;
}

export const analyticsService = {
    getPlatformStats: () =>
        api.get<PlatformStats>('/platform/analytics/platform-stats'),

    getTenantGrowth: (params?: { from?: string; to?: string }) =>
        api.get<TenantGrowth[]>('/platform/analytics/tenant-growth', { params }),

    getRevenueMetrics: (params?: { from?: string; to?: string }) =>
        api.get<RevenueMetrics>('/platform/analytics/revenue', { params }),
};

// ============================================================================
// USAGE
// ============================================================================

export interface OrganizationUsage {
    total_api_requests: number;
    avg_daily_requests: number;
    max_storage_mb: number;
    avg_active_users: number;
    total_logins: number;
    total_new_users: number;
    total_attendance_records: number;
    total_leave_requests: number;
    // Optional for now as backend summary doesn't return it yet
    daily_metrics?: Array<{
        date: string;
        api_requests: number;
        active_users: number;
    }>;
}

export const usageService = {
    getOrganizationUsage: (id: string, params?: { from?: string; to?: string }) =>
        api.get<OrganizationUsage>(`/platform/usage/organizations/${id}`, { params }),
};

// ============================================================================
// SUPER ADMIN MANAGEMENT
// ============================================================================

export interface CreateSuperAdminRequest {
    email: string;
    first_name: string;
    last_name: string;
}

export interface SuperAdminResponse {
    id: string;
    email: string;
    first_name: string;
    last_name: string;
    role: string;
    temporary_password?: string;
    is_active: boolean;
    created_at: string;
    updated_at: string;
}

export interface SuperAdminListItem {
    id: string;
    email: string;
    first_name: string;
    last_name: string;
    is_active: boolean;
    created_at: string;
    last_login_at?: string;
}

export const superAdminManagementService = {
    createSuperAdmin: async (data: CreateSuperAdminRequest): Promise<SuperAdminResponse> => {
        const response = await api.post<SuperAdminResponse>('/platform/admins', data);
        return response.data;
    },

    getAllSuperAdmins: async (): Promise<{ super_admins: SuperAdminListItem[]; total: number }> => {
        const response = await api.get<{ super_admins: SuperAdminListItem[]; total: number }>('/platform/admins');
        return response.data;
    },
};
