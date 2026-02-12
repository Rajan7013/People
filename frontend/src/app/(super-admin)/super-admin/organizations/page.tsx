"use client";

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import {
    Building2,
    MoreVertical,
    Search,
    Filter,
    Plus,
    Loader2,
    Shield,
    Users,
    HardDrive,
    Calendar,
    ArrowUpRight,
    SearchX
} from 'lucide-react';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { organizationService, Organization } from '@/services/superAdminService';
import { toast } from 'sonner';
import { format } from 'date-fns';

export default function OrganizationsPage() {
    const router = useRouter();
    const [organizations, setOrganizations] = useState<Organization[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchTerm, setSearchTerm] = useState('');
    const [statusFilter, setStatusFilter] = useState<'all' | 'active' | 'suspended'>('all');

    useEffect(() => {
        loadOrganizations();
    }, []);

    const loadOrganizations = async () => {
        try {
            setLoading(true);
            const response = await organizationService.getAll();
            setOrganizations(response.data.organizations || []);
        } catch (error) {
            toast.error("Failed to load organizations");
            console.error(error);
        } finally {
            setLoading(false);
        }
    };

    const handleStatusChange = async (id: string, currentStatus: string) => {
        try {
            if (currentStatus === 'active') {
                await organizationService.block(id);
                toast.success("Organization suspended successfully");
            } else {
                await organizationService.unblock(id);
                toast.success("Organization activated successfully");
            }
            loadOrganizations();
        } catch (error) {
            toast.error("Failed to update status");
        }
    };

    const handleDelete = async (id: string) => {
        if (!confirm('Are you sure you want to delete this organization? This action cannot be undone.')) return;
        try {
            await organizationService.delete(id);
            toast.success("Organization deleted successfully");
            loadOrganizations();
        } catch (error: any) {
            // If the organization is not found (404), it's already deleted. Treat as success.
            if (error.response && error.response.status === 404) {
                toast.success("Organization deleted successfully");
                loadOrganizations();
                return;
            }
            toast.error("Failed to delete organization");
            console.error(error);
        }
    };

    const filteredOrgs = organizations.filter(org => {
        const matchesSearch = org.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
            org.domain?.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesStatus = statusFilter === 'all' || org.status === statusFilter;
        return matchesSearch && matchesStatus;
    });

    if (loading) {
        return (
            <div className="flex flex-col items-center justify-center min-h-[60vh]">
                <Loader2 className="w-10 h-10 text-blue-600 animate-spin mb-4" />
                <p className="text-gray-500 font-medium">Loading organizations...</p>
            </div>
        );
    }

    return (
        <div className="space-y-8 animate-in fade-in duration-500">
            {/* Header */}
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div>
                    <h1 className="text-3xl font-bold text-gray-900 dark:text-white tracking-tight">Organizations</h1>
                    <p className="text-gray-500 dark:text-gray-400 mt-2 text-lg">Manage tenants and their subscriptions.</p>
                </div>
                <button
                    onClick={() => router.push('/super-admin/organizations/new')}
                    className="inline-flex items-center justify-center gap-2 bg-blue-600 hover:bg-blue-700 text-white px-5 py-3 rounded-xl font-medium transition-all shadow-lg shadow-blue-600/20 hover:shadow-blue-600/30 active:scale-95"
                >
                    <Plus className="w-5 h-5" />
                    New Organization
                </button>
            </div>

            {/* Filters & Search */}
            <div className="flex flex-col sm:flex-row gap-4">
                <div className="relative flex-1">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-gray-400 dark:text-gray-500" />
                    <input
                        type="text"
                        placeholder="Search organizations..."
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                        className="w-full pl-10 pr-4 py-3 rounded-xl border border-gray-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-hidden transition-all placeholder:text-gray-400"
                    />
                </div>
                <div className="flex items-center gap-2">
                    <div className="relative">
                        <Filter className="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-gray-400 dark:text-gray-500" />
                        <select
                            value={statusFilter}
                            onChange={(e) => setStatusFilter(e.target.value as any)}
                            className="pl-10 pr-8 py-3 rounded-xl border border-gray-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-hidden appearance-none cursor-pointer"
                        >
                            <option value="all">All Status</option>
                            <option value="active">Active</option>
                            <option value="suspended">Suspended</option>
                        </select>
                    </div>
                </div>
            </div>

            {/* Organizations List */}
            <div className="grid grid-cols-1 gap-4">
                {filteredOrgs.length === 0 ? (
                    <div className="bg-white dark:bg-slate-900 rounded-2xl p-12 text-center border border-gray-100 dark:border-slate-800 shadow-sm">
                        <div className="w-16 h-16 bg-gray-50 dark:bg-slate-800 rounded-full flex items-center justify-center mx-auto mb-4">
                            <SearchX className="w-8 h-8 text-gray-400 dark:text-gray-500" />
                        </div>
                        <h3 className="text-lg font-bold text-gray-900 dark:text-white mb-2">No organizations found</h3>
                        <p className="text-gray-500 dark:text-gray-400 max-w-md mx-auto">
                            We couldn't find any organizations matching your search criteria. Try adjusting your filters or create a new one.
                        </p>
                    </div>
                ) : (
                    filteredOrgs.map((org) => (
                        <div
                            key={org.id}
                            className="group bg-white dark:bg-slate-900 rounded-2xl p-5 border border-gray-100 dark:border-slate-800 shadow-sm hover:shadow-md hover:border-blue-100 dark:hover:border-blue-900/30 transition-all duration-200"
                        >
                            <div className="flex flex-col md:flex-row md:items-center justify-between gap-6">
                                {/* Org Info */}
                                <div className="flex items-start gap-4 flex-1">
                                    <div className="w-12 h-12 rounded-xl bg-linear-to-br from-blue-500 to-indigo-600 flex items-center justify-center text-white font-bold text-lg shadow-sm shrink-0">
                                        {org.name.substring(0, 2).toUpperCase()}
                                    </div>
                                    <div>
                                        <div className="flex items-center gap-2 mb-1">
                                            <h3 className="text-lg font-bold text-gray-900 dark:text-white group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-colors">
                                                {org.name}
                                            </h3>
                                            <span className={`px-2.5 py-0.5 rounded-full text-xs font-semibold ${org.status === 'active'
                                                ? 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400'
                                                : 'bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400'
                                                }`}>
                                                {org.status === 'active' ? 'Active' : 'Suspended'}
                                            </span>
                                        </div>
                                        <div className="flex flex-wrap items-center gap-x-4 gap-y-2 text-sm text-gray-500 dark:text-gray-400">
                                            <div className="flex items-center gap-1.5">
                                                <Shield className="w-4 h-4" />
                                                {org.subscription?.plan?.display_name || 'Free Plan'}
                                            </div>
                                            <div className="flex items-center gap-1.5">
                                                <Users className="w-4 h-4" />
                                                {org.total_users || 0} Users
                                            </div>
                                            <div className="flex items-center gap-1.5">
                                                <HardDrive className="w-4 h-4" />
                                                {((org.storage_used_mb || 0) / 1024).toFixed(1)} GB
                                            </div>
                                            <div className="flex items-center gap-1.5">
                                                <Calendar className="w-4 h-4" />
                                                Joined {format(new Date(org.created_at), 'MMM d, yyyy')}
                                            </div>
                                        </div>
                                    </div>
                                </div>

                                {/* Actions */}
                                <div className="flex items-center gap-3 self-end md:self-center">
                                    <button
                                        onClick={() => router.push(`/super-admin/organizations/${org.id}`)}
                                        className="hidden sm:flex items-center gap-2 px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-50 dark:bg-slate-800 hover:bg-gray-100 dark:hover:bg-slate-700 rounded-lg transition-colors"
                                    >
                                        Manage
                                        <ArrowUpRight className="w-4 h-4" />
                                    </button>

                                    <DropdownMenu>
                                        <DropdownMenuTrigger asChild>
                                            <button className="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-slate-800 rounded-lg transition-colors">
                                                <MoreVertical className="w-5 h-5" />
                                            </button>
                                        </DropdownMenuTrigger>
                                        <DropdownMenuContent align="end" className="w-48">
                                            <DropdownMenuItem onClick={() => router.push(`/super-admin/organizations/${org.id}`)}>
                                                View Details
                                            </DropdownMenuItem>
                                            <DropdownMenuItem onClick={() => router.push(`/super-admin/organizations/${org.id}/edit`)}>
                                                Edit Organization
                                            </DropdownMenuItem>
                                            <DropdownMenuItem onClick={() => handleStatusChange(org.id, org.status)}>
                                                {org.status === 'active' ? 'Suspend Organization' : 'Activate Organization'}
                                            </DropdownMenuItem>
                                            <DropdownMenuItem
                                                onClick={() => handleDelete(org.id)}
                                                className="text-red-600 dark:text-red-400 focus:text-red-600 dark:focus:text-red-400"
                                            >
                                                Delete Organization
                                            </DropdownMenuItem>
                                        </DropdownMenuContent>
                                    </DropdownMenu>
                                </div>
                            </div>
                        </div>
                    ))
                )}
            </div>
        </div>
    );
}
