"use client";

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { organizationService, Organization, planService, SubscriptionPlan } from '@/services/superAdminService';
import { ArrowLeft, Check, CreditCard, Calendar, Building2 } from 'lucide-react';
import Link from 'next/link';
import { COUNTRIES } from '@/lib/constants';

export default function EditOrganizationPage() {
    const params = useParams();
    const router = useRouter();
    const organizationId = params.id as string;

    const [organization, setOrganization] = useState<Organization | null>(null);
    const [plans, setPlans] = useState<SubscriptionPlan[]>([]);
    const [loading, setLoading] = useState(true);
    const [submitting, setSubmitting] = useState(false);
    const [formData, setFormData] = useState<{
        name: string;
        subdomain: string;
        country: string;
        admin_email: string;
        status: 'active' | 'suspended' | 'inactive';
        plan_id: string;
        billing_cycle: 'monthly' | 'yearly';
    }>({
        name: '',
        subdomain: '',
        country: '',
        admin_email: '',
        status: 'active',
        plan_id: '',
        billing_cycle: 'monthly',
    });

    useEffect(() => {
        if (organizationId) {
            loadData();
        }
    }, [organizationId]);

    const loadData = async () => {
        try {
            setLoading(true);
            const [orgResponse, plansResponse] = await Promise.all([
                organizationService.getById(organizationId),
                planService.getAll()
            ]);

            setOrganization(orgResponse.data);
            setPlans(plansResponse.data.plans || []);

            setFormData({
                name: orgResponse.data.name,
                subdomain: orgResponse.data.subdomain,
                country: orgResponse.data.country || '',
                admin_email: orgResponse.data.admin_email || '',
                status: orgResponse.data.status,
                plan_id: orgResponse.data.subscription?.plan_id || '',
                billing_cycle: (orgResponse.data.subscription?.billing_cycle as 'monthly' | 'yearly') || 'monthly',
            });
        } catch (error) {
            console.error('Failed to load data:', error);
            alert('Failed to load organization data');
        } finally {
            setLoading(false);
        }
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!formData.name || !formData.subdomain || !formData.country) {
            alert('Please fill in all required fields');
            return;
        }

        try {
            setSubmitting(true);

            // 1. Update basic details
            await organizationService.update(organizationId, {
                name: formData.name,
                subdomain: formData.subdomain,
                country: formData.country,
                admin_email: formData.admin_email,
                status: formData.status
            });

            // 2. If plan changed, trigger renewal/upgrade logic
            // Note: In real app, changing plan might be a separate API call (e.g. /change-plan)
            // For now, we'll assume the update endpoint might handle it OR we use the renew endpoint as a proxy for "Force Plan Change"
            if (formData.plan_id && formData.plan_id !== organization?.subscription?.plan_id) {
                await organizationService.renew(organizationId, {
                    plan_id: formData.plan_id,
                    billing_cycle: formData.billing_cycle
                });
            }

            alert('Organization updated successfully');
            router.push(`/super-admin/organizations/${organizationId}`);
        } catch (error: any) {
            console.error('Failed to update organization:', error);
            alert(error.response?.data?.error || 'Failed to update organization');
        } finally {
            setSubmitting(false);
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

    if (!organization) return null;

    const currentPlan = plans.find(p => p.id === organization.subscription?.plan_id);

    return (
        <div className="p-8 space-y-8 animate-in fade-in duration-500">
            {/* Header */}
            <div className="flex items-center gap-4">
                <Link
                    href={`/super-admin/organizations/${organizationId}`}
                    className="flex items-center gap-2 text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors"
                >
                    <ArrowLeft className="h-5 w-5" />
                    Back
                </Link>
                <div>
                    <h1 className="text-3xl font-bold text-gray-900 dark:text-white">Edit Organization</h1>
                    <p className="text-gray-600 dark:text-gray-400 mt-1">Update organization details and subscription</p>
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                {/* Main Form */}
                <div className="lg:col-span-2 space-y-6">
                    <div className="bg-white dark:bg-slate-900 p-8 rounded-2xl shadow-sm border border-gray-100 dark:border-slate-800">
                        <h2 className="text-xl font-bold text-gray-900 dark:text-white mb-6 flex items-center gap-2">
                            <Building2 className="w-5 h-5 text-gray-400" />
                            Organization Details
                        </h2>
                        <form id="edit-form" onSubmit={handleSubmit} className="space-y-6">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                {/* Organization Name */}
                                <div>
                                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                        Organization Name *
                                    </label>
                                    <input
                                        type="text"
                                        value={formData.name}
                                        onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                        className="w-full px-4 py-3 bg-gray-50 dark:bg-slate-800 border border-gray-200 dark:border-slate-700 rounded-xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 text-gray-900 dark:text-white transition-all outline-none"
                                        required
                                    />
                                </div>

                                {/* Subdomain */}
                                <div>
                                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                        Subdomain *
                                    </label>
                                    <div className="flex rounded-xl shadow-sm border border-gray-200 dark:border-slate-700 bg-gray-50 dark:bg-slate-800 overflow-hidden focus-within:ring-2 focus-within:ring-blue-500/20 focus-within:border-blue-500 transition-all">
                                        <input
                                            type="text"
                                            value={formData.subdomain}
                                            onChange={(e) => setFormData({ ...formData, subdomain: e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, '') })}
                                            className="flex-1 pl-4 pr-2 py-3 bg-transparent border-none focus:ring-0 text-gray-900 dark:text-white placeholder:text-gray-400 outline-none min-w-0"
                                            required
                                        />
                                        <span className="flex items-center pr-4 text-gray-500 dark:text-gray-400 bg-gray-100 dark:bg-slate-700 border-l border-gray-200 dark:border-slate-600 px-3 text-sm font-medium whitespace-nowrap">
                                            .peopleos.com
                                        </span>
                                    </div>
                                </div>

                                {/* Country */}
                                <div>
                                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                        Country *
                                    </label>
                                    <select
                                        value={formData.country}
                                        onChange={(e) => setFormData({ ...formData, country: e.target.value })}
                                        className="w-full px-4 py-3 bg-gray-50 dark:bg-slate-800 border border-gray-200 dark:border-slate-700 rounded-xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 text-gray-900 dark:text-white transition-all outline-none appearance-none"
                                        required
                                    >
                                        <option value="">Select Country</option>
                                        {COUNTRIES.map(country => (
                                            <option key={country} value={country}>{country}</option>
                                        ))}
                                    </select>
                                </div>

                                {/* Admin Email */}
                                <div>
                                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                        Admin Email
                                    </label>
                                    <input
                                        type="email"
                                        value={formData.admin_email}
                                        onChange={(e) => setFormData({ ...formData, admin_email: e.target.value })}
                                        className="w-full px-4 py-3 bg-gray-50 dark:bg-slate-800 border border-gray-200 dark:border-slate-700 rounded-xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 text-gray-900 dark:text-white transition-all outline-none"
                                    />
                                </div>

                                {/* Status */}
                                <div>
                                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                        Status *
                                    </label>
                                    <select
                                        value={formData.status}
                                        onChange={(e) => setFormData({ ...formData, status: e.target.value as 'active' | 'suspended' | 'inactive' })}
                                        className="w-full px-4 py-3 bg-gray-50 dark:bg-slate-800 border border-gray-200 dark:border-slate-700 rounded-xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 text-gray-900 dark:text-white transition-all outline-none appearance-none"
                                        required
                                    >
                                        <option value="active">Active</option>
                                        <option value="suspended">Suspended</option>
                                        <option value="inactive">Inactive</option>
                                    </select>
                                </div>
                            </div>
                        </form>
                    </div>

                    {/* Subscription Section */}
                    <div className="bg-white dark:bg-slate-900 p-8 rounded-2xl shadow-sm border border-gray-100 dark:border-slate-800">
                        <h2 className="text-xl font-bold text-gray-900 dark:text-white mb-6 flex items-center gap-2">
                            <CreditCard className="w-5 h-5 text-gray-400" />
                            Subscription Management
                        </h2>

                        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                            <div className="md:col-span-2 p-4 bg-blue-50 dark:bg-blue-900/20 rounded-xl border border-blue-100 dark:border-blue-800 flex items-start gap-4">
                                <div className="p-2 bg-blue-100 dark:bg-blue-900/40 rounded-lg text-blue-600 dark:text-blue-400">
                                    <Check className="w-5 h-5" />
                                </div>
                                <div>
                                    <h3 className="font-bold text-gray-900 dark:text-white">Current Plan: {organization.subscription?.plan?.display_name || 'None'}</h3>
                                    <p className="text-sm text-gray-600 dark:text-gray-300 mt-1">
                                        Status: <span className="font-semibold capitalize">{organization.subscription?.status || 'N/A'}</span> •
                                        Auto Renew: <span className="font-semibold">{organization.subscription?.auto_renew ? 'Yes' : 'No'}</span>
                                    </p>
                                    {organization.subscription?.end_date && (
                                        <p className="text-xs text-gray-500 dark:text-gray-400 mt-2">
                                            Valid until: {new Date(organization.subscription.end_date).toLocaleDateString()}
                                        </p>
                                    )}
                                </div>
                            </div>

                            <div>
                                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                    Change Plan
                                </label>
                                <select
                                    value={formData.plan_id}
                                    onChange={(e) => setFormData({ ...formData, plan_id: e.target.value })}
                                    className="w-full px-4 py-3 bg-gray-50 dark:bg-slate-800 border border-gray-200 dark:border-slate-700 rounded-xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 text-gray-900 dark:text-white transition-all outline-none appearance-none"
                                >
                                    <option value="">Select Plan</option>
                                    {plans.map(plan => (
                                        <option key={plan.id} value={plan.id}>
                                            {plan.display_name} ({plan.currency} {formData.billing_cycle === 'monthly' ? plan.price_monthly : plan.price_yearly}/{formData.billing_cycle === 'monthly' ? 'mo' : 'yr'})
                                        </option>
                                    ))}
                                </select>
                            </div>

                            <div>
                                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                    Billing Cycle
                                </label>
                                <div className="flex bg-gray-100 dark:bg-slate-800 p-1 rounded-xl">
                                    <button
                                        type="button"
                                        onClick={() => setFormData({ ...formData, billing_cycle: 'monthly' })}
                                        className={`flex-1 py-2 px-4 rounded-lg text-sm font-medium transition-all ${formData.billing_cycle === 'monthly'
                                            ? 'bg-white dark:bg-slate-600 text-gray-900 dark:text-white shadow-sm'
                                            : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'
                                            }`}
                                    >
                                        Monthly
                                    </button>
                                    <button
                                        type="button"
                                        onClick={() => setFormData({ ...formData, billing_cycle: 'yearly' })}
                                        className={`flex-1 py-2 px-4 rounded-lg text-sm font-medium transition-all ${formData.billing_cycle === 'yearly'
                                            ? 'bg-white dark:bg-slate-600 text-gray-900 dark:text-white shadow-sm'
                                            : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'
                                            }`}
                                    >
                                        Yearly
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Sidebar / Actions */}
                <div className="space-y-6">
                    <div className="bg-white dark:bg-slate-900 p-6 rounded-2xl shadow-sm border border-gray-100 dark:border-slate-800 sticky top-6">
                        <h3 className="font-bold text-gray-900 dark:text-white mb-4">Actions</h3>
                        <p className="text-sm text-gray-500 dark:text-gray-400 mb-6">
                            Review your changes before updating. Changing the plan may result in immediate charges or prorated billing adjustments.
                        </p>

                        <div className="space-y-3">
                            <button
                                onClick={handleSubmit}
                                disabled={submitting}
                                className="w-full px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-xl shadow-lg shadow-blue-600/20 transition-all disabled:opacity-70 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                            >
                                {submitting ? (
                                    <>
                                        <div className="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                                        Updating...
                                    </>
                                ) : (
                                    'Save Changes'
                                )}
                            </button>
                            <Link
                                href={`/super-admin/organizations/${organizationId}`}
                                className="w-full px-6 py-3 bg-gray-50 dark:bg-slate-800 hover:bg-gray-100 dark:hover:bg-slate-700 text-gray-700 dark:text-gray-300 font-medium rounded-xl transition-all flex items-center justify-center"
                            >
                                Cancel
                            </Link>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}
