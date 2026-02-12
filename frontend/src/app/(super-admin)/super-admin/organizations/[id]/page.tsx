"use client";

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { organizationService, Organization, usageService, invoiceService, OrganizationUsage, Invoice } from '@/services/superAdminService';
import { ArrowLeft, Edit, Ban, CheckCircle, Trash2, RefreshCw, Building2, Users, Database, Activity, Download, FileText, Loader2, AlertCircle } from 'lucide-react';
import Link from 'next/link';
import {
    LineChart,
    Line,
    XAxis,
    YAxis,
    CartesianGrid,
    Tooltip,
    ResponsiveContainer,
    AreaChart,
    Area
} from 'recharts';
import { format } from 'date-fns';

type TabType = 'overview' | 'usage' | 'billing' | 'logs';

export default function OrganizationDetailPage() {
    const params = useParams();
    const router = useRouter();
    const organizationId = params.id as string;

    const [organization, setOrganization] = useState<Organization | null>(null);
    const [loading, setLoading] = useState(true);
    const [activeTab, setActiveTab] = useState<TabType>('overview');

    useEffect(() => {
        if (organizationId) {
            loadOrganization();
        }
    }, [organizationId]);

    const loadOrganization = async () => {
        try {
            setLoading(true);
            const response = await organizationService.getById(organizationId);
            setOrganization(response.data);
        } catch (error) {
            console.error('Failed to load organization:', error);
            alert('Failed to load organization');
        } finally {
            setLoading(false);
        }
    };

    const handleBlock = async () => {
        if (!confirm('Are you sure you want to block this organization?')) return;
        try {
            await organizationService.block(organizationId);
            loadOrganization();
        } catch (error) {
            console.error('Failed to block organization:', error);
            alert('Failed to block organization');
        }
    };

    const handleUnblock = async () => {
        try {
            await organizationService.unblock(organizationId);
            loadOrganization();
        } catch (error) {
            console.error('Failed to unblock organization:', error);
            alert('Failed to unblock organization');
        }
    };

    const handleDelete = async () => {
        if (!confirm('Are you sure you want to delete this organization? This action cannot be undone.')) return;
        try {
            await organizationService.delete(organizationId);
            router.push('/super-admin/organizations');
        } catch (error: any) {
            // If the organization is not found (404), it's already deleted. Treat as success.
            if (error.response && error.response.status === 404) {
                router.push('/super-admin/organizations');
                return;
            }
            console.error('Failed to delete organization:', error);
            alert('Failed to delete organization');
        }
    };

    const handleRenew = async () => {
        if (!confirm('Renew subscription for this organization?')) return;

        // For now, we'll just use a simple renew without selecting plan
        // In a full implementation, you'd show a modal to select plan and billing cycle
        if (!organization?.subscription) {
            alert('No active subscription to renew');
            return;
        }

        try {
            // This is a placeholder - the backend renew endpoint needs plan_id and billing_cycle
            // For now, we'll just show a message
            alert('Renew functionality requires plan selection. This will be implemented in Phase 2.');
            // await organizationService.renew(organizationId, {
            //     plan_id: organization.subscription.plan_id,
            //     billing_cycle: organization.subscription.billing_cycle
            // });
            // loadOrganization();
        } catch (error) {
            console.error('Failed to renew subscription:', error);
            alert('Failed to renew subscription');
        }
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center h-full min-h-[400px]">
                <div className="text-center">
                    <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
                    <p className="mt-4 text-gray-600 dark:text-gray-400">Loading organization...</p>
                </div>
            </div>
        );
    }

    if (!organization) {
        return (
            <div className="p-8">
                <div className="text-center">
                    <h2 className="text-2xl font-bold text-gray-900 dark:text-white">Organization Not Found</h2>
                    <Link href="/super-admin/organizations" className="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 mt-4 inline-block">
                        Back to Organizations
                    </Link>
                </div>
            </div>
        );
    }

    const tabs = [
        { id: 'overview' as TabType, name: 'Overview' },
        { id: 'usage' as TabType, name: 'Usage & Metrics' },
        { id: 'billing' as TabType, name: 'Billing' },
        { id: 'logs' as TabType, name: 'Activity Logs' },
    ];

    return (
        <div className="p-8 space-y-8 animate-in fade-in duration-500">
            {/* Header */}
            <div className="flex flex-col lg:flex-row items-start lg:items-center justify-between gap-4">
                <div className="flex items-center gap-4">
                    <Link
                        href="/super-admin/organizations"
                        className="flex items-center gap-2 text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors"
                    >
                        <ArrowLeft className="h-5 w-5" />
                        Back
                    </Link>
                    <div>
                        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">{organization.name}</h1>
                        <p className="text-gray-600 dark:text-gray-400 mt-1">{organization.subdomain}.peopleos.com</p>
                    </div>
                </div>
                <div className="flex flex-wrap items-center gap-2">
                    <Link
                        href={`/super-admin/organizations/${organizationId}/edit`}
                        className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors flex items-center gap-2 shadow-sm shadow-blue-600/20"
                    >
                        <Edit className="h-4 w-4" />
                        Edit
                    </Link>
                    {organization.status === 'active' ? (
                        <button
                            onClick={handleBlock}
                            className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors flex items-center gap-2 shadow-sm shadow-red-600/20"
                        >
                            <Ban className="h-4 w-4" />
                            Block
                        </button>
                    ) : (
                        <button
                            onClick={handleUnblock}
                            className="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors flex items-center gap-2 shadow-sm shadow-green-600/20"
                        >
                            <CheckCircle className="h-4 w-4" />
                            Unblock
                        </button>
                    )}
                    <button
                        onClick={handleRenew}
                        className="px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700 transition-colors flex items-center gap-2 shadow-sm shadow-purple-600/20"
                    >
                        <RefreshCw className="h-4 w-4" />
                        Renew Plan
                    </button>
                    <button
                        onClick={handleDelete}
                        className="px-4 py-2 bg-gray-100 dark:bg-slate-800 text-red-600 dark:text-red-400 rounded-lg hover:bg-gray-200 dark:hover:bg-slate-700 transition-colors flex items-center gap-2"
                    >
                        <Trash2 className="h-4 w-4" />
                        Delete
                    </button>
                </div>
            </div>

            {/* Status Badge */}
            <div>
                <span className={`px-3 py-1 text-sm font-semibold rounded-full ${organization.status === 'active' ? 'bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-400' :
                    organization.status === 'suspended' ? 'bg-red-100 dark:bg-red-900/30 text-red-800 dark:text-red-400' :
                        'bg-gray-100 dark:bg-gray-800 text-gray-800 dark:text-gray-400'
                    }`}>
                    {organization.status.charAt(0).toUpperCase() + organization.status.slice(1)}
                </span>
            </div>

            {/* Tabs */}
            <div className="border-b border-gray-200 dark:border-slate-800">
                <nav className="-mb-px flex space-x-8">
                    {tabs.map((tab) => (
                        <button
                            key={tab.id}
                            onClick={() => setActiveTab(tab.id)}
                            className={`
                                py-4 px-1 border-b-2 font-medium text-sm transition-colors
                                ${activeTab === tab.id
                                    ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                                    : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 hover:border-gray-300 dark:hover:border-slate-700'
                                }
                            `}
                        >
                            {tab.name}
                        </button>
                    ))}
                </nav>
            </div>

            {/* Tab Content */}
            <div className="mt-6">
                {activeTab === 'overview' && <OverviewTab organization={organization} />}
                {activeTab === 'usage' && <UsageTab organization={organization} />}
                {activeTab === 'billing' && <BillingTab organization={organization} />}
                {activeTab === 'logs' && <LogsTab organization={organization} />}
            </div>
        </div>
    );
}

function OverviewTab({ organization }: { organization: Organization }) {
    return (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Organization Info */}
            <div className="bg-white dark:bg-slate-900 p-6 rounded-2xl shadow-sm border border-gray-100 dark:border-slate-800 space-y-4">
                <h3 className="text-lg font-bold text-gray-900 dark:text-white">Organization Information</h3>
                <div className="space-y-3">
                    <InfoRow label="Name" value={organization.name} />
                    <InfoRow label="Subdomain" value={`${organization.subdomain}.peopleos.com`} />
                    <InfoRow label="Admin Email" value={organization.admin_email || 'N/A'} />
                    <InfoRow label="Country" value={organization.country || 'N/A'} />
                    <InfoRow label="Status" value={organization.status} />
                    <InfoRow label="Created" value={new Date(organization.created_at).toLocaleDateString()} />
                </div>
            </div>

            {/* Subscription Info */}
            <div className="bg-white dark:bg-slate-900 p-6 rounded-2xl shadow-sm border border-gray-100 dark:border-slate-800 space-y-4">
                <h3 className="text-lg font-bold text-gray-900 dark:text-white">Subscription</h3>
                {organization.subscription ? (
                    <div className="space-y-3">
                        <InfoRow label="Plan" value={organization.subscription.plan?.display_name || 'N/A'} />
                        <InfoRow label="Status" value={organization.subscription.status} />
                        {organization.subscription.start_date && <InfoRow label="Start Date" value={new Date(organization.subscription.start_date).toLocaleDateString()} />}
                        {organization.subscription.end_date && <InfoRow label="End Date" value={new Date(organization.subscription.end_date).toLocaleDateString()} />}
                        {organization.subscription.auto_renew !== undefined && <InfoRow label="Auto Renew" value={organization.subscription.auto_renew ? 'Yes' : 'No'} />}
                    </div>
                ) : (
                    <p className="text-gray-500 dark:text-gray-400">No active subscription</p>
                )}
            </div>

            {/* Quick Stats */}
            <div className="bg-white dark:bg-slate-900 p-6 rounded-2xl shadow-sm border border-gray-100 dark:border-slate-800 lg:col-span-2">
                <h3 className="text-lg font-bold text-gray-900 dark:text-white mb-4">Quick Stats</h3>
                <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                    <StatCard
                        icon={Users}
                        label="Total Users"
                        value={organization.total_users || 0}
                        color="blue"
                    />
                    <StatCard
                        icon={Database}
                        label="Storage Used"
                        value={`${((organization.storage_used_mb || 0) / 1024).toFixed(2)} GB`}
                        color="green"
                    />
                    <StatCard
                        icon={Activity}
                        label="API Requests"
                        value={organization.api_requests_count || 0}
                        color="purple"
                    />
                    <StatCard
                        icon={Building2}
                        label="Active Since"
                        value={new Date(organization.created_at).toLocaleDateString('en-US', { month: 'short', year: 'numeric' })}
                        color="yellow"
                    />
                </div>
            </div>
        </div>
    );
}

function UsageTab({ organization }: { organization: Organization }) {
    const [usage, setUsage] = useState<OrganizationUsage | null>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        loadUsage();
    }, [organization.id]);

    const loadUsage = async () => {
        try {
            setLoading(true);
            const response = await usageService.getOrganizationUsage(organization.id);
            setUsage(response.data);
        } catch (error) {
            console.error("Failed to load usage data", error);
        } finally {
            setLoading(false);
        }
    };

    if (loading) {
        return (
            <div className="flex justify-center p-12">
                <Loader2 className="w-8 h-8 animate-spin text-blue-600" />
            </div>
        );
    }

    if (!usage) {
        return (
            <div className="text-center p-12 text-gray-500 dark:text-gray-400">
                <AlertCircle className="w-12 h-12 mx-auto mb-4 text-gray-400" />
                <p>Failed to load usage data.</p>
            </div>
        );
    }

    return (
        <div className="space-y-6 animate-in fade-in duration-500">
            <div className="bg-white dark:bg-slate-900 p-6 rounded-2xl border border-gray-100 dark:border-slate-800 shadow-sm">
                <h3 className="text-lg font-bold text-gray-900 dark:text-white mb-6">Usage Metrics</h3>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                    <div className="p-4 bg-blue-50 dark:bg-blue-900/20 rounded-xl border border-blue-100 dark:border-blue-900/50">
                        <p className="text-sm font-medium text-blue-600 dark:text-blue-400 mb-1">Total Users</p>
                        <p className="text-3xl font-bold text-gray-900 dark:text-white">{usage.total_new_users}</p>
                        <p className="text-sm text-blue-600/70 dark:text-blue-400/70 mt-1">{usage.avg_active_users.toFixed(0)} avg active</p>
                    </div>
                    <div className="p-4 bg-green-50 dark:bg-green-900/20 rounded-xl border border-green-100 dark:border-green-900/50">
                        <p className="text-sm font-medium text-green-600 dark:text-green-400 mb-1">Max Storage Used</p>
                        <p className="text-3xl font-bold text-gray-900 dark:text-white">
                            {(usage.max_storage_mb / 1024).toFixed(2)} <span className="text-lg font-medium text-gray-500 dark:text-gray-400">GB</span>
                        </p>
                        <p className="text-sm text-green-600/70 dark:text-green-400/70 mt-1">Peak monthly usage</p>
                    </div>
                    <div className="p-4 bg-purple-50 dark:bg-purple-900/20 rounded-xl border border-purple-100 dark:border-purple-900/50">
                        <p className="text-sm font-medium text-purple-600 dark:text-purple-400 mb-1">Total API Requests</p>
                        <p className="text-3xl font-bold text-gray-900 dark:text-white">{usage.total_api_requests.toLocaleString()}</p>
                        <p className="text-sm text-purple-600/70 dark:text-purple-400/70 mt-1">~{usage.avg_daily_requests.toFixed(0)} per day</p>
                    </div>
                </div>
            </div>

            {usage.daily_metrics && usage.daily_metrics.length > 0 ? (
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                    <div className="bg-white dark:bg-slate-900 p-6 rounded-2xl border border-gray-100 dark:border-slate-800 shadow-sm">
                        <h3 className="text-lg font-bold text-gray-900 dark:text-white mb-6">API Traffic</h3>
                        <div className="h-[300px] w-full">
                            <ResponsiveContainer width="100%" height="100%">
                                <AreaChart data={usage.daily_metrics}>
                                    <defs>
                                        <linearGradient id="colorApi" x1="0" y1="0" x2="0" y2="1">
                                            <stop offset="5%" stopColor="#8884d8" stopOpacity={0.8} />
                                            <stop offset="95%" stopColor="#8884d8" stopOpacity={0} />
                                        </linearGradient>
                                    </defs>
                                    <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#E5E7EB" />
                                    <XAxis
                                        dataKey="date"
                                        tickFormatter={(str) => format(new Date(str), 'MMM d')}
                                        stroke="#9CA3AF"
                                        fontSize={12}
                                        tickLine={false}
                                        axisLine={false}
                                    />
                                    <YAxis
                                        stroke="#9CA3AF"
                                        fontSize={12}
                                        tickLine={false}
                                        axisLine={false}
                                    />
                                    <Tooltip
                                        contentStyle={{ borderRadius: '8px', border: 'none', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)' }}
                                    />
                                    <Area
                                        type="monotone"
                                        dataKey="api_requests"
                                        stroke="#8884d8"
                                        fillOpacity={1}
                                        fill="url(#colorApi)"
                                        name="Requests"
                                    />
                                </AreaChart>
                            </ResponsiveContainer>
                        </div>
                    </div>

                    <div className="bg-white dark:bg-slate-900 p-6 rounded-2xl border border-gray-100 dark:border-slate-800 shadow-sm">
                        <h3 className="text-lg font-bold text-gray-900 dark:text-white mb-6">Active Users Trend</h3>
                        <div className="h-[300px] w-full">
                            <ResponsiveContainer width="100%" height="100%">
                                <LineChart data={usage.daily_metrics}>
                                    <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#E5E7EB" />
                                    <XAxis
                                        dataKey="date"
                                        tickFormatter={(str) => format(new Date(str), 'MMM d')}
                                        stroke="#9CA3AF"
                                        fontSize={12}
                                        tickLine={false}
                                        axisLine={false}
                                    />
                                    <YAxis
                                        stroke="#9CA3AF"
                                        fontSize={12}
                                        tickLine={false}
                                        axisLine={false}
                                    />
                                    <Tooltip
                                        contentStyle={{ borderRadius: '8px', border: 'none', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)' }}
                                    />
                                    <Line
                                        type="monotone"
                                        dataKey="active_users"
                                        stroke="#10B981"
                                        strokeWidth={2}
                                        dot={{ r: 4, fill: '#10B981' }}
                                        activeDot={{ r: 6 }}
                                        name="Active Users"
                                    />
                                </LineChart>
                            </ResponsiveContainer>
                        </div>
                    </div>
                </div>
            ) : (
                <div className="p-8 text-center bg-gray-50 dark:bg-slate-800/50 rounded-xl border border-dashed border-gray-200 dark:border-slate-800">
                    <p className="text-gray-500 dark:text-gray-400">Detailed usage charts will appear here once sufficient data is collected.</p>
                </div>
            )}
        </div>
    );
}

function BillingTab({ organization }: { organization: Organization }) {
    const [invoices, setInvoices] = useState<Invoice[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        loadInvoices();
    }, [organization.id]);

    const loadInvoices = async () => {
        try {
            setLoading(true);
            const response = await invoiceService.getAll({ tenant_id: organization.id });
            setInvoices(response.data.invoices || []);
        } catch (error) {
            console.error("Failed to load invoices", error);
        } finally {
            setLoading(false);
        }
    };

    const handleDownload = async (id: string, number: string) => {
        try {
            // Mock download for now as default implementation usually returns a direct link or blob
            // This would typically window.open(response.data.download_url)
            alert(`Downloading invoice ${number}...`);
        } catch (error) {
            console.error("Failed to download invoice", error);
        }
    };

    return (
        <div className="space-y-6 animate-in fade-in duration-500">
            <div className="bg-white dark:bg-slate-900 p-6 rounded-2xl border border-gray-100 dark:border-slate-800 shadow-sm">
                <h3 className="text-lg font-bold text-gray-900 dark:text-white mb-6">Current Subscription</h3>
                {organization.subscription ? (
                    <div className="flex flex-col md:flex-row items-center justify-between gap-6 p-6 bg-gray-50 dark:bg-slate-800/50 rounded-xl border border-gray-100 dark:border-slate-700">
                        <div className="flex items-center gap-4">
                            <div className="w-12 h-12 bg-blue-100 dark:bg-blue-900/30 rounded-full flex items-center justify-center text-blue-600 dark:text-blue-400">
                                <FileText className="w-6 h-6" />
                            </div>
                            <div>
                                <h4 className="text-lg font-bold text-gray-900 dark:text-white">{organization.subscription.plan?.display_name || 'N/A'} Plan</h4>
                                <p className="text-sm text-gray-500 dark:text-gray-400">
                                    {organization.subscription.billing_cycle === 'monthly' ? 'Billed Monthly' : 'Billed Yearly'}
                                    {' • '}
                                    Next billing on {new Date(organization.subscription.current_period_end).toLocaleDateString()}
                                </p>
                            </div>
                        </div>
                        <div className="flex items-center gap-3">
                            <span className={`px-3 py-1 rounded-full text-sm font-semibold ${organization.subscription.status === 'active'
                                ? 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400'
                                : 'bg-yellow-100 dark:bg-yellow-900/30 text-yellow-700 dark:text-yellow-400'
                                }`}>
                                {organization.subscription.status.toUpperCase()}
                            </span>
                        </div>
                    </div>
                ) : (
                    <div className="text-center p-8 bg-gray-50 dark:bg-slate-800 rounded-xl border border-dashed border-gray-200 dark:border-slate-700">
                        <p className="text-gray-500 dark:text-gray-400">No active subscription found.</p>
                    </div>
                )}
            </div>

            <div className="bg-white dark:bg-slate-900 rounded-2xl border border-gray-100 dark:border-slate-800 shadow-sm overflow-hidden">
                <div className="p-6 border-b border-gray-100 dark:border-slate-800 flex justify-between items-center">
                    <h3 className="text-lg font-bold text-gray-900 dark:text-white">Invoice History</h3>
                </div>

                {loading ? (
                    <div className="p-12 flex justify-center">
                        <Loader2 className="w-8 h-8 animate-spin text-blue-600" />
                    </div>
                ) : invoices.length === 0 ? (
                    <div className="p-12 text-center text-gray-500 dark:text-gray-400">
                        <FileText className="w-12 h-12 mx-auto mb-4 text-gray-300 dark:text-gray-600" />
                        <p>No invoices found for this organization.</p>
                    </div>
                ) : (
                    <div className="overflow-x-auto">
                        <table className="w-full text-left text-sm">
                            <thead className="bg-gray-50 dark:bg-slate-800/50 text-gray-500 dark:text-gray-400 font-medium">
                                <tr>
                                    <th className="px-6 py-4">Invoice</th>
                                    <th className="px-6 py-4">Date</th>
                                    <th className="px-6 py-4">Amount</th>
                                    <th className="px-6 py-4">Status</th>
                                    <th className="px-6 py-4 text-right">Actions</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-gray-100 dark:divide-slate-800">
                                {invoices.map((invoice) => (
                                    <tr key={invoice.id} className="hover:bg-gray-50 dark:hover:bg-slate-800/50 transition-colors">
                                        <td className="px-6 py-4 font-medium text-gray-900 dark:text-white">
                                            {invoice.invoice_number}
                                        </td>
                                        <td className="px-6 py-4 text-gray-500 dark:text-gray-400">
                                            {format(new Date(invoice.issue_date), 'MMM d, yyyy')}
                                        </td>
                                        <td className="px-6 py-4 font-medium text-gray-900 dark:text-white">
                                            {invoice.currency} {invoice.total_amount.toFixed(2)}
                                        </td>
                                        <td className="px-6 py-4">
                                            <span className={`px-2.5 py-0.5 rounded-full text-xs font-semibold ${invoice.status === 'paid'
                                                ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                                                : invoice.status === 'overdue'
                                                    ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
                                                    : 'bg-gray-100 text-gray-700 dark:bg-slate-800 dark:text-gray-400'
                                                }`}>
                                                {invoice.status.charAt(0).toUpperCase() + invoice.status.slice(1)}
                                            </span>
                                        </td>
                                        <td className="px-6 py-4 text-right">
                                            <button
                                                onClick={() => handleDownload(invoice.id, invoice.invoice_number)}
                                                className="p-2 text-gray-400 hover:text-blue-600 transition-colors"
                                                title="Download Invoice"
                                            >
                                                <Download className="w-4 h-4" />
                                            </button>
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                )}
            </div>
        </div>
    );
}

function LogsTab({ organization }: { organization: Organization }) {
    return (
        <div className="bg-white dark:bg-slate-900 p-6 rounded-2xl shadow-sm border border-gray-100 dark:border-slate-800">
            <h3 className="text-lg font-bold text-gray-900 dark:text-white mb-4">Activity Logs</h3>
            <p className="text-gray-500 dark:text-gray-400">Activity logs will be available soon</p>
        </div>
    );
}

function InfoRow({ label, value }: { label: string; value: string }) {
    return (
        <div className="flex justify-between items-center">
            <span className="text-sm text-gray-600 dark:text-gray-400">{label}</span>
            <span className="text-sm font-medium text-gray-900 dark:text-white">{value}</span>
        </div>
    );
}

interface StatCardProps {
    icon: any;
    label: string;
    value: string | number;
    color: 'blue' | 'green' | 'purple' | 'yellow';
}

function StatCard({ icon: Icon, label, value, color }: StatCardProps) {
    const colorClasses = {
        blue: 'bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400',
        green: 'bg-green-50 dark:bg-green-900/20 text-green-600 dark:text-green-400',
        purple: 'bg-purple-50 dark:bg-purple-900/20 text-purple-600 dark:text-purple-400',
        yellow: 'bg-yellow-50 dark:bg-yellow-900/20 text-yellow-600 dark:text-yellow-400',
    };

    return (
        <div className="flex items-center gap-3">
            <div className={`p-3 rounded-lg ${colorClasses[color]}`}>
                <Icon className="h-6 w-6" />
            </div>
            <div>
                <p className="text-sm text-gray-600 dark:text-gray-400">{label}</p>
                <p className="text-lg font-bold text-gray-900 dark:text-white">{value}</p>
            </div>
        </div>
    );
}
