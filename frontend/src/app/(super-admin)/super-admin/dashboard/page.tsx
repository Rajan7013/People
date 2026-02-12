"use client";

import { useEffect, useState } from 'react';
import { analyticsService, PlatformStats, TenantGrowth, RevenueMetrics } from '@/services/superAdminService';
import {
    Building2,
    Users,
    DollarSign,
    TrendingUp,
    Database,
    Activity,
    ArrowUpRight,
    ArrowDownRight,
    CreditCard,
    Globe,
    Server,
    Zap,
    Loader2
} from 'lucide-react';
import {
    LineChart,
    Line,
    PieChart,
    Pie,
    Cell,
    XAxis,
    YAxis,
    CartesianGrid,
    Tooltip,
    Legend,
    ResponsiveContainer,
    AreaChart,
    Area
} from 'recharts';
import Link from 'next/link';
import { useTheme } from 'next-themes';
import { AddSuperAdminDialog } from '@/components/super-admin/AddSuperAdminDialog';

export default function SuperAdminDashboard() {
    const [stats, setStats] = useState<PlatformStats | null>(null);
    const [growth, setGrowth] = useState<TenantGrowth[]>([]);
    const [revenue, setRevenue] = useState<RevenueMetrics | null>(null);
    const [loading, setLoading] = useState(true);
    const { theme } = useTheme();

    useEffect(() => {
        loadData();
    }, []);

    const loadData = async () => {
        try {
            setLoading(true);
            const [statsRes, growthRes, revenueRes] = await Promise.all([
                analyticsService.getPlatformStats(),
                analyticsService.getTenantGrowth({ from: getDateDaysAgo(30), to: new Date().toISOString().split('T')[0] }),
                analyticsService.getRevenueMetrics(),
            ]);
            setStats(statsRes.data);
            setGrowth(Array.isArray(growthRes.data) ? growthRes.data : []);
            setRevenue(revenueRes.data);
        } catch (error) {
            console.error('Failed to load dashboard data:', error);
        } finally {
            setLoading(false);
        }
    };

    const getDateDaysAgo = (days: number) => {
        const date = new Date();
        date.setDate(date.getDate() - days);
        return date.toISOString().split('T')[0];
    };

    if (loading) {
        return (
            <div className="flex flex-col items-center justify-center min-h-[60vh]">
                <Loader2 className="w-10 h-10 text-blue-600 animate-spin mb-4" />
                <p className="text-gray-500 font-medium">Loading platform insights...</p>
            </div>
        );
    }

    const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444'];
    const isDark = theme === 'dark';

    return (
        <div className="space-y-8 animate-in fade-in duration-500">
            {/* Header */}
            <div className="flex flex-col md:flex-row md:items-end md:justify-between gap-4">
                <div>
                    <h1 className="text-3xl font-bold text-gray-900 dark:text-white tracking-tight">Platform Overview</h1>
                    <p className="text-gray-500 dark:text-gray-400 mt-2 text-lg">Real-time insights into your SaaS ecosystem.</p>
                </div>
                <div className="flex items-center gap-3">
                    <AddSuperAdminDialog onSuccess={loadData} />
                    <div className="hidden md:flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400 bg-white dark:bg-slate-900 px-3 py-1.5 rounded-full border border-gray-200 dark:border-slate-800 shadow-sm">
                        <div className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
                        System Operational
                    </div>
                </div>
            </div>

            {/* Stats Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                <StatsCard
                    title="Total Organizations"
                    value={stats?.total_tenants || 0}
                    subtitle={`${stats?.active_tenants || 0} active tenants`}
                    icon={Building2}
                    color="blue"
                />
                <StatsCard
                    title="Total Users"
                    value={(stats?.total_users || 0).toLocaleString()}
                    subtitle="Across all workspaces"
                    icon={Users}
                    color="purple"
                />
                <StatsCard
                    title="Monthly Revenue"
                    value={`$${(stats?.mrr || 0).toLocaleString()}`}
                    subtitle={`ARR: $${(stats?.arr || 0).toLocaleString()}`}
                    icon={DollarSign}
                    color="green"
                />
                <StatsCard
                    title="Active Subscriptions"
                    value={stats?.active_subscriptions || 0}
                    subtitle={`${stats?.total_subscriptions || 0} total plans`}
                    icon={Zap}
                    color="orange"
                />
            </div>

            {/* Main Charts Section */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                {/* Tenant Growth Chart (Main) */}
                <div className="lg:col-span-2 bg-white dark:bg-slate-900 p-6 rounded-2xl shadow-sm border border-gray-100 dark:border-slate-800">
                    <div className="flex items-center justify-between mb-6">
                        <div>
                            <h3 className="text-lg font-bold text-gray-900 dark:text-white">Tenant Growth</h3>
                            <p className="text-sm text-gray-500 dark:text-gray-400">New organizations over the last 30 days</p>
                        </div>
                        <div className="p-2 bg-gray-50 dark:bg-slate-800 rounded-lg">
                            <TrendingUp className="w-5 h-5 text-gray-400 dark:text-gray-500" />
                        </div>
                    </div>
                    <div className="h-[300px] w-full">
                        <ResponsiveContainer width="100%" height="100%">
                            <AreaChart data={growth}>
                                <defs>
                                    <linearGradient id="colorTenants" x1="0" y1="0" x2="0" y2="1">
                                        <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.1} />
                                        <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
                                    </linearGradient>
                                </defs>
                                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke={isDark ? '#334155' : '#f1f5f9'} />
                                <XAxis
                                    dataKey="date"
                                    axisLine={false}
                                    tickLine={false}
                                    tick={{ fill: isDark ? '#94a3b8' : '#64748b', fontSize: 12 }}
                                    dy={10}
                                    tickFormatter={(date) => new Date(date).getDate().toString()}
                                />
                                <YAxis
                                    axisLine={false}
                                    tickLine={false}
                                    tick={{ fill: isDark ? '#94a3b8' : '#64748b', fontSize: 12 }}
                                />
                                <Tooltip
                                    contentStyle={{
                                        borderRadius: '12px',
                                        border: 'none',
                                        boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)',
                                        backgroundColor: isDark ? '#1e293b' : '#fff',
                                        color: isDark ? '#fff' : '#000'
                                    }}
                                />
                                <Area
                                    type="monotone"
                                    dataKey="total_tenants"
                                    stroke="#3b82f6"
                                    strokeWidth={3}
                                    fillOpacity={1}
                                    fill="url(#colorTenants)"
                                />
                            </AreaChart>
                        </ResponsiveContainer>
                    </div>
                </div>

                {/* Revenue by Plan (Side) */}
                <div className="bg-white dark:bg-slate-900 p-6 rounded-2xl shadow-sm border border-gray-100 dark:border-slate-800">
                    <div className="flex items-center justify-between mb-6">
                        <div>
                            <h3 className="text-lg font-bold text-gray-900 dark:text-white">Revenue Mix</h3>
                            <p className="text-sm text-gray-500 dark:text-gray-400">Breakdown by plan type</p>
                        </div>
                        <div className="p-2 bg-gray-50 dark:bg-slate-800 rounded-lg">
                            <PieChart className="w-5 h-5 text-gray-400 dark:text-gray-500" />
                        </div>
                    </div>
                    <div className="h-[300px] w-full relative">
                        <ResponsiveContainer width="100%" height="100%">
                            <PieChart>
                                <Pie
                                    data={revenue?.revenue_by_plan || []}
                                    cx="50%"
                                    cy="50%"
                                    innerRadius={60}
                                    outerRadius={80}
                                    paddingAngle={5}
                                    dataKey="revenue"
                                    stroke={isDark ? '#0f172a' : '#fff'}
                                >
                                    {(revenue?.revenue_by_plan || []).map((entry, index) => (
                                        <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                                    ))}
                                </Pie>
                                <Tooltip
                                    contentStyle={{
                                        borderRadius: '8px',
                                        border: 'none',
                                        backgroundColor: isDark ? '#1e293b' : '#fff',
                                        color: isDark ? '#fff' : '#000'
                                    }}
                                />
                                <Legend
                                    verticalAlign="bottom"
                                    height={36}
                                    iconType="circle"
                                    formatter={(value, entry: any) => <span className="text-sm text-gray-600 dark:text-gray-300 ml-1">{value}</span>}
                                />
                            </PieChart>
                        </ResponsiveContainer>
                        {/* Center Text */}
                        <div className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 text-center pointer-events-none pb-8 text-gray-900 dark:text-white">
                            <p className="text-xs text-gray-500 dark:text-gray-400 font-medium">Total</p>
                            <p className="text-lg font-bold">${(revenue?.total_revenue || 0).toLocaleString()}</p>
                        </div>
                    </div>
                </div>
            </div>

            {/* Secondary Metrics Row */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                {/* Revenue Summary */}
                <div className="bg-white dark:bg-slate-900 p-6 rounded-2xl shadow-sm border border-gray-100 dark:border-slate-800 flex flex-col justify-between">
                    <div>
                        <div className="flex items-center gap-3 mb-4">
                            <div className="p-2 bg-green-50 dark:bg-green-900/20 rounded-lg text-green-600 dark:text-green-400">
                                <CreditCard className="w-5 h-5" />
                            </div>
                            <h3 className="text-lg font-bold text-gray-900 dark:text-white">Billing Status</h3>
                        </div>
                        <div className="space-y-4">
                            <div className="flex justify-between items-center p-3 bg-gray-50 dark:bg-slate-800/50 rounded-xl">
                                <span className="text-gray-600 dark:text-gray-400 font-medium">Collected</span>
                                <span className="text-lg font-bold text-green-600 dark:text-green-400">
                                    ${(revenue?.paid_revenue || 0).toLocaleString()}
                                </span>
                            </div>
                            <div className="flex justify-between items-center p-3 bg-gray-50 dark:bg-slate-800/50 rounded-xl">
                                <span className="text-gray-600 dark:text-gray-400 font-medium">Pending</span>
                                <span className="text-lg font-bold text-yellow-600 dark:text-yellow-400">
                                    ${(revenue?.pending_revenue || 0).toLocaleString()}
                                </span>
                            </div>
                        </div>
                    </div>
                    <div className="mt-6 pt-6 border-t border-gray-100 dark:border-slate-800">
                        <Link
                            href="/super-admin/invoices"
                            className="text-sm font-medium text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 flex items-center justify-center gap-1 group"
                        >
                            View All Invoices
                            <ArrowUpRight className="w-4 h-4 group-hover:translate-x-0.5 group-hover:-translate-y-0.5 transition-transform" />
                        </Link>
                    </div>
                </div>

                {/* Platform Resources */}
                <div className="bg-white dark:bg-slate-900 p-6 rounded-2xl shadow-sm border border-gray-100 dark:border-slate-800">
                    <div className="flex items-center gap-3 mb-6">
                        <div className="p-2 bg-blue-50 dark:bg-blue-900/20 rounded-lg text-blue-600 dark:text-blue-400">
                            <Server className="w-5 h-5" />
                        </div>
                        <h3 className="text-lg font-bold text-gray-900 dark:text-white">System Resources</h3>
                    </div>
                    <div className="space-y-6">
                        <div>
                            <div className="flex justify-between items-end mb-2">
                                <span className="text-sm font-medium text-gray-600 dark:text-gray-400">Storage Usage</span>
                                <span className="text-sm font-bold text-gray-900 dark:text-white">{((stats?.total_storage_gb || 0) / 1024).toFixed(2)} GB</span>
                            </div>
                            <div className="w-full bg-gray-100 dark:bg-slate-800 rounded-full h-2">
                                <div className="bg-blue-600 h-2 rounded-full" style={{ width: '15%' }}></div>
                            </div>
                            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1.5 text-right">15% of total allocated capacity</p>
                        </div>
                        <div>
                            <div className="flex justify-between items-end mb-2">
                                <span className="text-sm font-medium text-gray-600 dark:text-gray-400">API Load</span>
                                <span className="text-sm font-bold text-gray-900 dark:text-white">1.2M Req</span>
                            </div>
                            <div className="w-full bg-gray-100 dark:bg-slate-800 rounded-full h-2">
                                <div className="bg-purple-600 h-2 rounded-full" style={{ width: '45%' }}></div>
                            </div>
                            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1.5 text-right">45% of monthly threshold</p>
                        </div>
                    </div>
                </div>

                {/* Quick Actions */}
                <div className="bg-linear-to-br from-blue-600 to-indigo-700 dark:from-blue-700 dark:to-indigo-800 rounded-2xl p-6 text-white shadow-xl shadow-blue-500/20 dark:shadow-none">
                    <h3 className="text-xl font-bold mb-2">Quick Actions</h3>
                    <p className="text-blue-100 text-sm mb-6">Manage your platform efficiently.</p>

                    <div className="space-y-3">
                        <Link href="/super-admin/organizations/new">
                            <button className="w-full py-3 px-4 bg-white/10 hover:bg-white/20 backdrop-blur-md border border-white/20 rounded-xl text-left flex items-center gap-3 transition-all group">
                                <div className="p-1.5 bg-white/20 rounded-lg group-hover:bg-white/30 transition-colors">
                                    <Building2 className="w-4 h-4" />
                                </div>
                                <span className="font-medium">Onboard Organization</span>
                            </button>
                        </Link>

                        <Link href="/super-admin/plans">
                            <button className="w-full py-3 px-4 bg-white/10 hover:bg-white/20 backdrop-blur-md border border-white/20 rounded-xl text-left flex items-center gap-3 transition-all group">
                                <div className="p-1.5 bg-white/20 rounded-lg group-hover:bg-white/30 transition-colors">
                                    <Zap className="w-4 h-4" />
                                </div>
                                <span className="font-medium">Update Pricing Plans</span>
                            </button>
                        </Link>
                    </div>
                </div>
            </div>
        </div>
    );
}

interface StatsCardProps {
    title: string;
    value: string | number;
    subtitle: string;
    icon: any;
    color: 'blue' | 'green' | 'yellow' | 'purple' | 'orange';
    trend?: 'up' | 'down' | 'neutral';
    trendValue?: string;
}

function StatsCard({ title, value, subtitle, icon: Icon, color, trend, trendValue }: StatsCardProps) {
    const colorConfigs = {
        blue: 'bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400 ring-blue-600/20 dark:ring-blue-400/20',
        green: 'bg-green-50 dark:bg-green-900/20 text-green-600 dark:text-green-400 ring-green-600/20 dark:ring-green-400/20',
        yellow: 'bg-yellow-50 dark:bg-yellow-900/20 text-yellow-600 dark:text-yellow-400 ring-yellow-600/20 dark:ring-yellow-400/20',
        purple: 'bg-purple-50 dark:bg-purple-900/20 text-purple-600 dark:text-purple-400 ring-purple-600/20 dark:ring-purple-400/20',
        orange: 'bg-orange-50 dark:bg-orange-900/20 text-orange-600 dark:text-orange-400 ring-orange-600/20 dark:ring-orange-400/20',
    };

    return (
        <div className="bg-white dark:bg-slate-900 p-6 rounded-2xl shadow-sm border border-gray-100 dark:border-slate-800 hover:shadow-md transition-shadow duration-200 group">
            <div className="flex items-start justify-between mb-4">
                <div className={`p-3 rounded-xl ${colorConfigs[color]} ring-1 ring-inset transition-colors`}>
                    <Icon className="h-6 w-6" />
                </div>
                {trend && (
                    <div className={`flex items-center gap-1 text-xs font-semibold px-2 py-1 rounded-full ${trend === 'up' ? 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400' :
                        trend === 'down' ? 'bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400' :
                            'bg-gray-100 dark:bg-slate-800 text-gray-700 dark:text-gray-400'
                        }`}>
                        {trend === 'up' ? <ArrowUpRight className="w-3 h-3" /> :
                            trend === 'down' ? <ArrowDownRight className="w-3 h-3" /> :
                                null}
                        {trendValue}
                    </div>
                )}
            </div>
            <div>
                <h3 className="text-3xl font-bold text-gray-900 dark:text-white tracking-tight group-hover:scale-[1.02] transition-transform origin-left">{value}</h3>
                <p className="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider mt-1 mb-0.5">{title}</p>
                <p className="text-sm text-gray-500 dark:text-gray-400">{subtitle}</p>
            </div>
        </div>
    );
}
