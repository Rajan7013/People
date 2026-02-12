"use client";

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import {
    Building2,
    User,
    CreditCard,
    CheckCircle2,
    Loader2,
    Globe,
    MapPin,
    Smartphone,
    Mail,
    Lock,
    ArrowLeft,
    ArrowRight,
    Briefcase,
    Rocket
} from 'lucide-react';
import { organizationService, planService, SubscriptionPlan } from '@/services/superAdminService';
import { toast } from 'sonner';

// Validation Schema
const organizationSchema = z.object({
    // Organization Details
    name: z.string().min(2, "Name must be at least 2 characters"),
    subdomain: z.string()
        .min(2, "Subdomain must be at least 2 characters")
        .regex(/^[a-z0-9-]+$/, "Subdomain must contain only lowercase letters, numbers, and hyphens"),
    domain: z.string().optional(),
    country: z.string().min(1, "Country is required"),
    address_line1: z.string().optional(),
    address_line2: z.string().optional(),
    city: z.string().optional(),
    state: z.string().optional(),
    postal_code: z.string().optional(),

    // Admin User
    admin_first_name: z.string().min(2, "First name is required"),
    admin_last_name: z.string().min(2, "Last name is required"),
    admin_email: z.string().email("Invalid email address"),
    admin_password: z.string().min(8, "Password must be at least 8 characters"),
    admin_contact_number: z.string().optional(),

    // Subscription
    plan_id: z.string().min(1, "Please select a subscription plan"),
    billing_cycle: z.enum(['monthly', 'yearly']),
    trial_days: z.coerce.number().min(0).max(365).optional(),

    // Settings
    currency: z.string().default('USD'),
    timezone: z.string().default('UTC'),
});

type FormData = z.infer<typeof organizationSchema>;

export default function CreateOrganizationPage() {
    const router = useRouter();
    const [isLoading, setIsLoading] = useState(false);
    const [plans, setPlans] = useState<SubscriptionPlan[]>([]);
    const [loadingPlans, setLoadingPlans] = useState(true);

    const form = useForm<FormData>({
        resolver: zodResolver(organizationSchema) as any,
        defaultValues: {
            country: 'US',
            currency: 'USD',
            timezone: 'UTC',
            billing_cycle: 'monthly',
            trial_days: 14,
        }
    });

    useEffect(() => {
        loadPlans();
    }, []);

    const loadPlans = async () => {
        try {
            setLoadingPlans(true);
            const response = await planService.getAll();
            const activePlans = (response.data.plans || []).filter(p => p.is_active);
            setPlans(activePlans);
        } catch (error) {
            console.error("Failed to load plans", error);
            toast.error("Failed to load subscription plans");
        } finally {
            setLoadingPlans(false);
        }
    };

    const onSubmit = async (data: FormData) => {
        setIsLoading(true);
        try {
            await organizationService.create({
                name: data.name,
                subdomain: data.subdomain,
                domain: data.domain,
                country: data.country,
                admin_email: data.admin_email,
                admin_name: `${data.admin_first_name} ${data.admin_last_name}`,
                admin_password: data.admin_password,
                plan_id: data.plan_id,
                billing_cycle: data.billing_cycle,
                address_line1: data.address_line1,
                address_line2: data.address_line2 || "",
                city: data.city || "",
                state: data.state || "",
                postal_code: data.postal_code || "",
                contact_number: data.admin_contact_number || "",
                currency: data.currency,
                timezone: data.timezone,
                trial_days: data.trial_days
            });

            toast.success("Organization created successfully! 🎉");
            router.push('/super-admin/organizations');
        } catch (error: any) {
            console.error(error);
            const message = error.response?.data?.error || "Failed to create organization";
            toast.error(message);
        } finally {
            setIsLoading(false);
        }
    };

    // Auto-generate subdomain from name
    const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const name = e.target.value;
        const subdomain = name.toLowerCase().replace(/[^a-z0-9]/g, '-').replace(/-+/g, '-');
        form.setValue('name', name);
        if (!form.getValues('subdomain') || form.getValues('subdomain') === '') {
            form.setValue('subdomain', subdomain);
        }
    };

    return (
        <div className="max-w-5xl mx-auto pb-12 animate-in fade-in slide-in-from-bottom-4 duration-500">
            {/* Header */}
            <div className="flex items-center gap-4 mb-8">
                <button
                    onClick={() => router.back()}
                    className="p-2 hover:bg-gray-100 dark:hover:bg-slate-800 rounded-lg transition-colors"
                >
                    <ArrowLeft className="w-5 h-5 text-gray-500 dark:text-gray-400" />
                </button>
                <div>
                    <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Onboard Organization</h1>
                    <p className="text-gray-500 dark:text-gray-400">Create a new tenant workspace and admin account.</p>
                </div>
            </div>

            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
                {/* Organization Details */}
                <div className="bg-white dark:bg-slate-900 rounded-2xl p-6 md:p-8 shadow-sm border border-gray-100 dark:border-slate-800 relative overflow-hidden">
                    <div className="absolute top-0 right-0 p-6 opacity-5 dark:opacity-10">
                        <Building2 className="w-32 h-32 text-blue-600" />
                    </div>

                    <div className="flex items-center gap-3 mb-6 relative z-10">
                        <div className="p-2 bg-blue-50 dark:bg-blue-900/20 rounded-lg text-blue-600 dark:text-blue-400">
                            <Building2 className="w-5 h-5" />
                        </div>
                        <h2 className="text-lg font-bold text-gray-900 dark:text-white">Organization Details</h2>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6 relative z-10">
                        <div className="space-y-2">
                            <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Organization Name *</label>
                            <div className="relative">
                                <Briefcase className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
                                <input
                                    {...form.register('name')}
                                    onChange={(e) => {
                                        handleNameChange(e);
                                        form.trigger('name');
                                    }}
                                    className="w-full pl-10 pr-4 py-2.5 bg-gray-50 dark:bg-slate-800 border border-gray-200 dark:border-slate-700 rounded-xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 text-gray-900 dark:text-white transition-all outline-none"
                                    placeholder="Acme Corp"
                                />
                            </div>
                            {form.formState.errors.name && (
                                <p className="text-xs text-red-500 font-medium ml-1">{form.formState.errors.name.message}</p>
                            )}
                        </div>

                        <div className="space-y-2">
                            <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Subdomain *</label>
                            <div className="flex rounded-xl shadow-sm border border-gray-200 dark:border-slate-700 bg-gray-50 dark:bg-slate-800 overflow-hidden focus-within:ring-2 focus-within:ring-blue-500/20 focus-within:border-blue-500 transition-all">
                                <div className="flex items-center pl-3">
                                    <Globe className="h-4 w-4 text-gray-400" />
                                </div>
                                <input
                                    {...form.register('subdomain')}
                                    className="flex-1 pl-3 pr-2 py-2.5 bg-transparent border-none focus:ring-0 text-gray-900 dark:text-white placeholder:text-gray-400 outline-none min-w-0"
                                    placeholder="acme"
                                />
                                <span className="flex items-center pr-4 text-gray-500 dark:text-gray-400 bg-gray-100 dark:bg-slate-700 border-l border-gray-200 dark:border-slate-600 px-3 text-sm font-medium">
                                    .peopleos.com
                                </span>
                            </div>
                            {form.formState.errors.subdomain && (
                                <p className="text-xs text-red-500 font-medium ml-1">{form.formState.errors.subdomain.message}</p>
                            )}
                        </div>

                        <div className="space-y-2">
                            <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Custom Domain (Optional)</label>
                            <div className="relative">
                                <Globe className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
                                <input
                                    {...form.register('domain')}
                                    className="w-full pl-10 pr-4 py-2.5 bg-gray-50 dark:bg-slate-800 border border-gray-200 dark:border-slate-700 rounded-xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 text-gray-900 dark:text-white transition-all outline-none"
                                    placeholder="app.acmecorp.com"
                                />
                            </div>
                        </div>

                        <div className="space-y-2">
                            <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Country *</label>
                            <div className="relative">
                                <MapPin className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
                                <select
                                    {...form.register('country')}
                                    className="w-full pl-10 pr-4 py-2.5 bg-gray-50 dark:bg-slate-800 border border-gray-200 dark:border-slate-700 rounded-xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 text-gray-900 dark:text-white transition-all outline-none appearance-none cursor-pointer"
                                >
                                    <option value="US">United States</option>
                                    <option value="CA">Canada</option>
                                    <option value="UK">United Kingdom</option>
                                    <option value="AU">Australia</option>
                                    <option value="IN">India</option>
                                    <option value="SG">Singapore</option>
                                    <option value="DE">Germany</option>
                                    <option value="FR">France</option>
                                </select>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Admin User */}
                <div className="bg-white dark:bg-slate-900 rounded-2xl p-6 md:p-8 shadow-sm border border-gray-100 dark:border-slate-800 relative overflow-hidden">
                    <div className="absolute top-0 right-0 p-6 opacity-5 dark:opacity-10">
                        <User className="w-32 h-32 text-purple-600" />
                    </div>

                    <div className="flex items-center gap-3 mb-6 relative z-10">
                        <div className="p-2 bg-purple-50 dark:bg-purple-900/20 rounded-lg text-purple-600 dark:text-purple-400">
                            <User className="w-5 h-5" />
                        </div>
                        <h2 className="text-lg font-bold text-gray-900 dark:text-white">Admin Account</h2>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6 relative z-10">
                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <label className="text-sm font-medium text-gray-700 dark:text-gray-300">First Name *</label>
                                <div className="relative">
                                    <User className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
                                    <input
                                        {...form.register('admin_first_name')}
                                        className="w-full pl-10 pr-4 py-2.5 bg-gray-50 dark:bg-slate-800 border border-gray-200 dark:border-slate-700 rounded-xl focus:ring-2 focus:ring-purple-500/20 focus:border-purple-500 text-gray-900 dark:text-white transition-all outline-none"
                                        placeholder="John"
                                    />
                                </div>
                                {form.formState.errors.admin_first_name && (
                                    <p className="text-xs text-red-500 font-medium">{form.formState.errors.admin_first_name.message}</p>
                                )}
                            </div>
                            <div className="space-y-2">
                                <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Last Name *</label>
                                <div className="relative">
                                    <User className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
                                    <input
                                        {...form.register('admin_last_name')}
                                        className="w-full pl-10 pr-4 py-2.5 bg-gray-50 dark:bg-slate-800 border border-gray-200 dark:border-slate-700 rounded-xl focus:ring-2 focus:ring-purple-500/20 focus:border-purple-500 text-gray-900 dark:text-white transition-all outline-none"
                                        placeholder="Doe"
                                    />
                                </div>
                                {form.formState.errors.admin_last_name && (
                                    <p className="text-xs text-red-500 font-medium">{form.formState.errors.admin_last_name.message}</p>
                                )}
                            </div>
                        </div>

                        <div className="space-y-2">
                            <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Email Address *</label>
                            <div className="relative">
                                <Mail className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
                                <input
                                    {...form.register('admin_email')}
                                    type="email"
                                    className="w-full pl-10 pr-4 py-2.5 bg-gray-50 dark:bg-slate-800 border border-gray-200 dark:border-slate-700 rounded-xl focus:ring-2 focus:ring-purple-500/20 focus:border-purple-500 text-gray-900 dark:text-white transition-all outline-none"
                                    placeholder="john@acme.com"
                                />
                            </div>
                            {form.formState.errors.admin_email && (
                                <p className="text-xs text-red-500 font-medium ml-1">{form.formState.errors.admin_email.message}</p>
                            )}
                        </div>

                        <div className="space-y-2">
                            <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Password *</label>
                            <div className="relative">
                                <Lock className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
                                <input
                                    {...form.register('admin_password')}
                                    type="password"
                                    className="w-full pl-10 pr-4 py-2.5 bg-gray-50 dark:bg-slate-800 border border-gray-200 dark:border-slate-700 rounded-xl focus:ring-2 focus:ring-purple-500/20 focus:border-purple-500 text-gray-900 dark:text-white transition-all outline-none"
                                    placeholder="••••••••"
                                />
                            </div>
                            {form.formState.errors.admin_password && (
                                <p className="text-xs text-red-500 font-medium ml-1">{form.formState.errors.admin_password.message}</p>
                            )}
                        </div>

                        <div className="space-y-2">
                            <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Contact Number</label>
                            <div className="relative">
                                <Smartphone className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
                                <input
                                    {...form.register('admin_contact_number')}
                                    className="w-full pl-10 pr-4 py-2.5 bg-gray-50 dark:bg-slate-800 border border-gray-200 dark:border-slate-700 rounded-xl focus:ring-2 focus:ring-purple-500/20 focus:border-purple-500 text-gray-900 dark:text-white transition-all outline-none"
                                    placeholder="+1 (555) 000-0000"
                                />
                            </div>
                        </div>
                    </div>
                </div>

                {/* Subscription Plan */}
                <div className="bg-white dark:bg-slate-900 rounded-2xl p-6 md:p-8 shadow-sm border border-gray-100 dark:border-slate-800 relative overflow-hidden">
                    <div className="absolute top-0 right-0 p-6 opacity-5 dark:opacity-10">
                        <CreditCard className="w-32 h-32 text-green-600" />
                    </div>

                    <div className="flex items-center gap-3 mb-6 relative z-10">
                        <div className="p-2 bg-green-50 dark:bg-green-900/20 rounded-lg text-green-600 dark:text-green-400">
                            <CreditCard className="w-5 h-5" />
                        </div>
                        <h2 className="text-lg font-bold text-gray-900 dark:text-white">Subscription Plan</h2>
                    </div>

                    <div className="relative z-10">
                        {/* Billing Cycle Toggle */}
                        <div className="flex justify-center mb-8">
                            <div className="bg-gray-100 dark:bg-slate-800 p-1 rounded-xl inline-flex">
                                <button
                                    type="button"
                                    onClick={() => form.setValue('billing_cycle', 'monthly')}
                                    className={`px-6 py-2 rounded-lg text-sm font-medium transition-all ${form.watch('billing_cycle') === 'monthly'
                                        ? 'bg-white dark:bg-slate-700 text-gray-900 dark:text-white shadow-sm ring-1 ring-black/5 dark:ring-white/10'
                                        : 'text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200'
                                        }`}
                                >
                                    Monthly
                                </button>
                                <button
                                    type="button"
                                    onClick={() => form.setValue('billing_cycle', 'yearly')}
                                    className={`px-6 py-2 rounded-lg text-sm font-medium transition-all ${form.watch('billing_cycle') === 'yearly'
                                        ? 'bg-white dark:bg-slate-700 text-gray-900 dark:text-white shadow-sm ring-1 ring-black/5 dark:ring-white/10'
                                        : 'text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200'
                                        }`}
                                >
                                    Yearly <span className="text-green-600 dark:text-green-400 text-xs ml-1 font-bold">-20%</span>
                                </button>
                            </div>
                        </div>

                        {loadingPlans ? (
                            <div className="flex justify-center py-12">
                                <Loader2 className="w-8 h-8 text-blue-600 animate-spin" />
                            </div>
                        ) : (
                            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                                {plans.map((plan) => (
                                    <div
                                        key={plan.id}
                                        onClick={() => form.setValue('plan_id', plan.id)}
                                        className={`
                                            cursor-pointer relative p-6 rounded-2xl border-2 transition-all duration-200
                                            ${form.watch('plan_id') === plan.id
                                                ? 'border-blue-600 bg-blue-50/50 dark:bg-blue-900/20 shadow-md ring-1 ring-blue-600'
                                                : 'border-gray-200 dark:border-slate-800 hover:border-blue-200 dark:hover:border-blue-800 bg-white dark:bg-slate-900'
                                            }
                                        `}
                                    >
                                        <div className="flex flex-col h-full">
                                            <div className="mb-4">
                                                <h3 className="font-bold text-gray-900 dark:text-white">{plan.display_name}</h3>
                                                <div className="flex items-baseline gap-1 mt-2">
                                                    <span className="text-2xl font-bold text-gray-900 dark:text-white">
                                                        ${form.watch('billing_cycle') === 'monthly' ? plan.price_monthly : plan.price_yearly}
                                                    </span>
                                                    <span className="text-sm text-gray-500 dark:text-gray-400">
                                                        /{form.watch('billing_cycle') === 'monthly' ? 'mo' : 'yr'}
                                                    </span>
                                                </div>
                                            </div>

                                            <ul className="space-y-3 mb-6 flex-1">
                                                <li className="flex items-center text-sm text-gray-600 dark:text-gray-300">
                                                    <CheckCircle2 className="w-4 h-4 text-green-500 mr-2 shrink-0" />
                                                    {plan.max_users === 0 ? "Unlimited" : plan.max_users} Users
                                                </li>
                                                <li className="flex items-center text-sm text-gray-600 dark:text-gray-300">
                                                    <CheckCircle2 className="w-4 h-4 text-green-500 mr-2 shrink-0" />
                                                    {plan.max_storage_gb} GB Storage
                                                </li>
                                            </ul>

                                            <div className={`
                                                w-full py-2 rounded-lg text-sm font-medium text-center transition-colors
                                                ${form.watch('plan_id') === plan.id
                                                    ? 'bg-blue-600 text-white'
                                                    : 'bg-gray-100 dark:bg-slate-800 text-gray-900 dark:text-white group-hover:bg-blue-600 group-hover:text-white'
                                                }
                                            `}>
                                                {form.watch('plan_id') === plan.id ? 'Selected' : 'Select Plan'}
                                            </div>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}
                        {form.formState.errors.plan_id && (
                            <p className="text-center text-red-500 text-sm mt-4 font-medium">{form.formState.errors.plan_id.message}</p>
                        )}
                    </div>
                </div>

                {/* Submit Actions */}
                <div className="flex justify-end gap-4 pt-4">
                    <button
                        type="button"
                        onClick={() => router.back()}
                        className="px-6 py-3 rounded-xl border border-gray-200 dark:border-slate-800 text-gray-600 dark:text-gray-300 font-medium hover:bg-gray-50 dark:hover:bg-slate-800 transition-colors"
                        disabled={isLoading}
                    >
                        Cancel
                    </button>
                    <button
                        type="submit"
                        disabled={isLoading}
                        className="px-8 py-3 rounded-xl bg-blue-600 hover:bg-blue-700 text-white font-medium shadow-lg shadow-blue-600/20 transition-all flex items-center gap-2 disabled:opacity-70 disabled:cursor-not-allowed group"
                    >
                        {isLoading ? (
                            <>
                                <Loader2 className="w-5 h-5 animate-spin" />
                                Creating Organization...
                            </>
                        ) : (
                            <>
                                Create Organization
                                <Rocket className="w-5 h-5 group-hover:translate-x-1 group-hover:-translate-y-1 transition-transform" />
                            </>
                        )}
                    </button>
                </div>
            </form>
        </div>
    );
}
