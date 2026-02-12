"use client";

import { useEffect, useState } from 'react';
import { analyticsService, PlatformStats, TenantGrowth } from '@/services/superAdminService';
import { AreaChart, Area, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { TrendingUp, TrendingDown, Users, Building2, DollarSign } from 'lucide-react';

export default function AnalyticsPage() {
    const [stats, setStats] = useState<PlatformStats | null>(null);
    const [growth, setGrowth] = useState<TenantGrowth[]>([]);
    const [loading, setLoading] = useState(true);
    const [dateRange, setDateRange] = useState('30');

    useEffect(() => {
        loadData();
    }, [dateRange]);

    const loadData = async () => {
        try {
            setLoading(true);
            const [statsRes, growthRes] = await Promise.all([
                analyticsService.getPlatformStats(),
                analyticsService.getTenantGrowth({
                    from: getDateDaysAgo(parseInt(dateRange)),
                    to: new Date().toISOString().split('T')[0],
                }),
            ]);
            setStats(statsRes.data);
            setGrowth(Array.isArray(growthRes.data) ? growthRes.data : []);
        } catch (error) {
            console.error('Failed to load analytics:', error);
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
            <div className="flex items-center justify-center h-full">
                <div className="text-center">
                    <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
                    <p className="mt-4 text-gray-600">Loading analytics...</p>
                </div>
            </div>
        );
    }

    const growthRate = growth.length > 1
        ? ((growth[growth.length - 1].total_tenants - growth[0].total_tenants) / growth[0].total_tenants * 100)
        : 0;

    return (
        <div className="space-y-8 animate-in fade-in duration-500">
            {/* Header */}
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div>
                    <h1 className="text-3xl font-bold text-gray-900 dark:text-white tracking-tight">Analytics</h1>
                    <p className="text-gray-500 dark:text-gray-400 mt-2 text-lg">Platform performance metrics</p>
                </div>
                <select
                    value={dateRange}
                    onChange={(e) => setDateRange(e.target.value)}
                    className="px-4 py-2 bg-white dark:bg-slate-900 border border-gray-200 dark:border-slate-800 rounded-xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 text-gray-900 dark:text-white outline-none cursor-pointer"
                >
                    <option value="7">Last 7 days</option>
                    <option value="30">Last 30 days</option>
                    <option value="90">Last 90 days</option>
                </select>
            </div>

            {/* Key Metrics */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
                <div className="bg-white dark:bg-slate-900 p-6 rounded-2xl border border-gray-100 dark:border-slate-800 shadow-sm">
                    <div className="flex items-center justify-between mb-4">
                        <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Total Tenants</p>
                        <div className="p-2 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
                            <Building2 className="h-5 w-5 text-blue-600 dark:text-blue-400" />
                        </div>
                    </div>
                    <p className="text-3xl font-bold text-gray-900 dark:text-white">{stats?.total_tenants || 0}</p>
                    <div className="flex items-center gap-1 mt-2">
                        {growthRate >= 0 ? (
                            <TrendingUp className="h-4 w-4 text-green-600 dark:text-green-400" />
                        ) : (
                            <TrendingDown className="h-4 w-4 text-red-600 dark:text-red-400" />
                        )}
                        <span className={`text-sm font-medium ${growthRate >= 0 ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}`}>
                            {Math.abs(growthRate).toFixed(1)}%
                        </span>
                        <span className="text-sm text-gray-500 dark:text-gray-400">vs previous period</span>
                    </div>
                </div>

                <div className="bg-white dark:bg-slate-900 p-6 rounded-2xl border border-gray-100 dark:border-slate-800 shadow-sm">
                    <div className="flex items-center justify-between mb-4">
                        <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Active Tenants</p>
                        <div className="p-2 bg-green-50 dark:bg-green-900/20 rounded-lg">
                            <Building2 className="h-5 w-5 text-green-600 dark:text-green-400" />
                        </div>
                    </div>
                    <p className="text-3xl font-bold text-gray-900 dark:text-white">{stats?.active_tenants || 0}</p>
                    <p className="text-sm text-gray-500 dark:text-gray-400 mt-2">
                        {((stats?.active_tenants || 0) / (stats?.total_tenants || 1) * 100).toFixed(1)}% of total
                    </p>
                </div>

                <div className="bg-white dark:bg-slate-900 p-6 rounded-2xl border border-gray-100 dark:border-slate-800 shadow-sm">
                    <div className="flex items-center justify-between mb-4">
                        <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Total Users</p>
                        <div className="p-2 bg-purple-50 dark:bg-purple-900/20 rounded-lg">
                            <Users className="h-5 w-5 text-purple-600 dark:text-purple-400" />
                        </div>
                    </div>
                    <p className="text-3xl font-bold text-gray-900 dark:text-white">{stats?.total_users || 0}</p>
                    <p className="text-sm text-gray-500 dark:text-gray-400 mt-2">
                        Avg {((stats?.total_users || 0) / (stats?.total_tenants || 1)).toFixed(1)} per tenant
                    </p>
                </div>

                <div className="bg-white dark:bg-slate-900 p-6 rounded-2xl border border-gray-100 dark:border-slate-800 shadow-sm">
                    <div className="flex items-center justify-between mb-4">
                        <p className="text-sm font-medium text-gray-500 dark:text-gray-400">MRR</p>
                        <div className="p-2 bg-yellow-50 dark:bg-yellow-900/20 rounded-lg">
                            <DollarSign className="h-5 w-5 text-yellow-600 dark:text-yellow-400" />
                        </div>
                    </div>
                    <p className="text-3xl font-bold text-gray-900 dark:text-white">
                        ${(stats?.mrr || 0).toLocaleString()}
                    </p>
                    <p className="text-sm text-gray-500 dark:text-gray-400 mt-2">
                        ARR: ${(stats?.arr || 0).toLocaleString()}
                    </p>
                </div>
            </div>

            {/* Charts */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                {/* Tenant Growth */}
                <div className="bg-white dark:bg-slate-900 p-6 rounded-2xl border border-gray-100 dark:border-slate-800 shadow-sm">
                    <h3 className="text-lg font-bold text-gray-900 dark:text-white mb-6">Tenant Growth</h3>
                    <div className="h-[300px] w-full">
                        <ResponsiveContainer width="100%" height="100%">
                            <AreaChart data={growth}>
                                <defs>
                                    <linearGradient id="colorTenants" x1="0" y1="0" x2="0" y2="1">
                                        <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.1} />
                                        <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
                                    </linearGradient>
                                </defs>
                                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#E2E8F0" className="dark:stroke-slate-800" />
                                <XAxis
                                    dataKey="date"
                                    stroke="#94a3b8"
                                    fontSize={12}
                                    tickLine={false}
                                    axisLine={false}
                                />
                                <YAxis
                                    stroke="#94a3b8"
                                    fontSize={12}
                                    tickLine={false}
                                    axisLine={false}
                                    tickFormatter={(value) => `${value}`}
                                />
                                <Tooltip
                                    contentStyle={{
                                        backgroundColor: 'rgba(255, 255, 255, 0.9)',
                                        borderRadius: '8px',
                                        border: 'none',
                                        boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)'
                                    }}
                                    itemStyle={{ color: '#1e293b' }}
                                />
                                <Area
                                    type="monotone"
                                    dataKey="total_tenants"
                                    stroke="#3b82f6"
                                    strokeWidth={2}
                                    fillOpacity={1}
                                    fill="url(#colorTenants)"
                                    name="Total Tenants"
                                />
                            </AreaChart>
                        </ResponsiveContainer>
                    </div>
                </div>

                {/* New Tenants */}
                <div className="bg-white dark:bg-slate-900 p-6 rounded-2xl border border-gray-100 dark:border-slate-800 shadow-sm">
                    <h3 className="text-lg font-bold text-gray-900 dark:text-white mb-6">New Tenants</h3>
                    <div className="h-[300px] w-full">
                        <ResponsiveContainer width="100%" height="100%">
                            <BarChart data={growth}>
                                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#E2E8F0" className="dark:stroke-slate-800" />
                                <XAxis
                                    dataKey="date"
                                    stroke="#94a3b8"
                                    fontSize={12}
                                    tickLine={false}
                                    axisLine={false}
                                />
                                <YAxis
                                    stroke="#94a3b8"
                                    fontSize={12}
                                    tickLine={false}
                                    axisLine={false}
                                />
                                <Tooltip
                                    cursor={{ fill: 'transparent' }}
                                    contentStyle={{
                                        backgroundColor: 'rgba(255, 255, 255, 0.9)',
                                        borderRadius: '8px',
                                        border: 'none',
                                        boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)'
                                    }}
                                    itemStyle={{ color: '#1e293b' }}
                                />
                                <Bar
                                    dataKey="new_tenants"
                                    fill="#10b981"
                                    radius={[4, 4, 0, 0]}
                                    name="New Tenants"
                                />
                            </BarChart>
                        </ResponsiveContainer>
                    </div>
                </div>
            </div>

            {/* Additional Metrics */}
            <div className="bg-white dark:bg-slate-900 p-6 rounded-2xl border border-gray-100 dark:border-slate-800 shadow-sm">
                <h3 className="text-lg font-bold text-gray-900 dark:text-white mb-6">Subscription Metrics</h3>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
                    <div>
                        <p className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">Active Subscriptions</p>
                        <p className="text-2xl font-bold text-gray-900 dark:text-white">{stats?.active_subscriptions || 0}</p>
                        <div className="w-full bg-gray-100 dark:bg-slate-800 rounded-full h-2 mt-3">
                            <div
                                className="bg-blue-600 h-2 rounded-full transition-all duration-500"
                                style={{ width: `${((stats?.active_subscriptions || 0) / (stats?.total_subscriptions || 1) * 100)}%` }}
                            ></div>
                        </div>
                        <p className="text-sm text-gray-500 dark:text-gray-400 mt-2">
                            of {stats?.total_subscriptions || 0} total
                        </p>
                    </div>
                    <div>
                        <p className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">Conversion Rate</p>
                        <p className="text-2xl font-bold text-gray-900 dark:text-white">
                            {((stats?.active_subscriptions || 0) / (stats?.total_tenants || 1) * 100).toFixed(1)}%
                        </p>
                        <div className="w-full bg-gray-100 dark:bg-slate-800 rounded-full h-2 mt-3">
                            <div
                                className="bg-green-500 h-2 rounded-full transition-all duration-500"
                                style={{ width: `${Math.min(100, ((stats?.active_subscriptions || 0) / (stats?.total_tenants || 1) * 100))}%` }}
                            ></div>
                        </div>
                        <p className="text-sm text-gray-500 dark:text-gray-400 mt-2">Tenants with active subs</p>
                    </div>
                    <div>
                        <p className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">Avg Revenue per Tenant</p>
                        <p className="text-2xl font-bold text-gray-900 dark:text-white">
                            ${((stats?.mrr || 0) / (stats?.active_tenants || 1)).toFixed(2)}
                        </p>
                        <div className="w-full bg-gray-100 dark:bg-slate-800 rounded-full h-2 mt-3">
                            <div className="bg-yellow-500 h-2 rounded-full" style={{ width: '100%' }}></div>
                        </div>
                        <p className="text-sm text-gray-500 dark:text-gray-400 mt-2">Monthly</p>
                    </div>
                </div>
            </div>
        </div>
    );
}
