"use client";

import { useEffect, useState } from 'react';
import { planService, SubscriptionPlan } from '@/services/superAdminService';
import {
    Check,
    Plus,
    Edit2,
    Trash2,
    X,
    Loader2,
    Save,
    Shield,
    Zap,
    Users,
    HardDrive,
    Server,
    MoreHorizontal
} from 'lucide-react';
import { toast } from 'sonner';

import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
    DialogFooter,
    DialogClose
} from "@/components/ui/dialog";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

export default function PlansPage() {
    const [plans, setPlans] = useState<SubscriptionPlan[]>([]);
    const [loading, setLoading] = useState(true);
    const [isEditing, setIsEditing] = useState<string | null>(null);
    const [editForm, setEditForm] = useState<Partial<SubscriptionPlan>>({});

    // New Plan State
    const [isCreateOpen, setIsCreateOpen] = useState(false);
    const [deleteId, setDeleteId] = useState<string | null>(null);

    // ... existing newPlan state ...

    const handleDelete = async () => {
        if (!deleteId) return;
        try {
            await planService.delete(deleteId);
            toast.success("Plan deleted successfully");
            setDeleteId(null);
            loadPlans();
        } catch (error: any) {
            // Check if error message indicates dependency
            if (error.response?.status === 500 && error.response?.data?.error?.includes("violates foreign key constraint")) {
                toast.error("Cannot delete plan: It is currently assigned to one or more organizations.");
            } else {
                toast.error(error.response?.data?.error || "Failed to delete plan");
            }
        }
    };
    const [newPlan, setNewPlan] = useState<Partial<SubscriptionPlan>>({
        name: '',
        display_name: '',
        description: '',
        price_monthly: 0,
        price_yearly: 0,
        currency: 'USD',
        max_users: 0,
        max_storage_gb: 0,
        features: {},
        is_active: true,
        is_visible: true
    });

    const currencies = [
        { code: 'USD', symbol: '$', name: 'US Dollar' },
        { code: 'INR', symbol: '₹', name: 'Indian Rupee' },
        { code: 'EUR', symbol: '€', name: 'Euro' },
        { code: 'GBP', symbol: '£', name: 'British Pound' },
        { code: 'AUD', symbol: 'A$', name: 'Australian Dollar' },
        { code: 'CAD', symbol: 'C$', name: 'Canadian Dollar' },
        { code: 'SGD', symbol: 'S$', name: 'Singapore Dollar' },
    ];

    const getCurrencySymbol = (code: string) => {
        return currencies.find(c => c.code === code)?.symbol || code;
    };

    useEffect(() => {
        loadPlans();
    }, []);

    const loadPlans = async () => {
        try {
            setLoading(true);
            const response = await planService.getAll();
            setPlans(response.data.plans || []);
        } catch (error) {
            toast.error("Failed to load plans");
        } finally {
            setLoading(false);
        }
    };

    const handleEditClick = (plan: SubscriptionPlan) => {
        setIsEditing(plan.id);
        setEditForm(plan);
    };

    const handleSave = async () => {
        if (!isEditing || !editForm) return;
        try {
            await planService.update(isEditing, editForm);
            toast.success("Plan updated successfully");
            setIsEditing(null);
            loadPlans();
        } catch (error) {
            toast.error("Failed to update plan");
        }
    };

    const handleCreate = async () => {
        try {
            await planService.create(newPlan);
            toast.success("Plan created successfully");
            setIsCreateOpen(false);
            loadPlans();
            // Reset form
            setNewPlan({
                name: '',
                display_name: '',
                description: '',
                price_monthly: 0,
                price_yearly: 0,
                currency: 'USD',
                max_users: 0,
                max_storage_gb: 0,
                features: {},
                is_active: true,
                is_visible: true
            });
        } catch (error) {
            toast.error("Failed to create plan");
        }
    };

    const togglePlanStatus = async (id: string, currentStatus: boolean) => {
        try {
            await planService.update(id, { is_active: !currentStatus });
            toast.success(`Plan ${!currentStatus ? 'activated' : 'deactivated'}`);
            loadPlans();
        } catch (error) {
            toast.error("Failed to update status");
        }
    };

    if (loading) {
        return (
            <div className="flex flex-col items-center justify-center min-h-[60vh]">
                <Loader2 className="w-10 h-10 text-blue-600 animate-spin mb-4" />
                <p className="text-gray-500 font-medium">Loading subscription plans...</p>
            </div>
        );
    }

    return (
        <div className="space-y-8 animate-in fade-in duration-500">
            {/* Header */}
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div>
                    <h1 className="text-3xl font-bold text-gray-900 dark:text-white tracking-tight">Subscription Plans</h1>
                    <p className="text-gray-500 dark:text-gray-400 mt-2 text-lg">Manage pricing tiers and features.</p>
                </div>

                <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
                    <DialogTrigger asChild>
                        <button className="inline-flex items-center justify-center gap-2 bg-blue-600 hover:bg-blue-700 text-white px-5 py-3 rounded-xl font-medium transition-all shadow-lg shadow-blue-600/20 hover:shadow-blue-600/30 active:scale-95">
                            <Plus className="w-5 h-5" />
                            Create New Plan
                        </button>
                    </DialogTrigger>
                    <DialogContent className="sm:max-w-[600px] bg-white dark:bg-slate-900 border-gray-100 dark:border-slate-800">
                        <DialogHeader>
                            <DialogTitle className="text-xl font-bold text-gray-900 dark:text-white">Create Subscription Plan</DialogTitle>
                            <DialogDescription className="text-gray-500 dark:text-gray-400">
                                Define a new pricing tier for your customers.
                            </DialogDescription>
                        </DialogHeader>
                        <div className="grid gap-6 py-4">
                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Internal Name</label>
                                    <input
                                        className="w-full px-3 py-2 bg-gray-50 dark:bg-slate-800 border border-gray-200 dark:border-slate-700 rounded-lg focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all"
                                        placeholder="e.g. pro_monthly"
                                        value={newPlan.name}
                                        onChange={(e) => setNewPlan({ ...newPlan, name: e.target.value })}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Display Name</label>
                                    <input
                                        className="w-full px-3 py-2 bg-gray-50 dark:bg-slate-800 border border-gray-200 dark:border-slate-700 rounded-lg focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all"
                                        placeholder="e.g. Pro Plan"
                                        value={newPlan.display_name}
                                        onChange={(e) => setNewPlan({ ...newPlan, display_name: e.target.value })}
                                    />
                                </div>
                            </div>

                            <div className="space-y-2">
                                <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Description</label>
                                <textarea
                                    className="w-full px-3 py-2 bg-gray-50 dark:bg-slate-800 border border-gray-200 dark:border-slate-700 rounded-lg focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all resize-none h-20"
                                    placeholder="Brief description of the plan..."
                                    value={newPlan.description}
                                    onChange={(e) => setNewPlan({ ...newPlan, description: e.target.value })}
                                />
                            </div>

                            <div className="space-y-2">
                                <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Currency</label>
                                <select
                                    className="w-full px-3 py-2 bg-gray-50 dark:bg-slate-800 border border-gray-200 dark:border-slate-700 rounded-lg focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all"
                                    value={newPlan.currency}
                                    onChange={(e) => setNewPlan({ ...newPlan, currency: e.target.value })}
                                >
                                    {currencies.map(c => (
                                        <option key={c.code} value={c.code}>{c.code} - {c.name} ({c.symbol})</option>
                                    ))}
                                </select>
                            </div>

                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Monthly Price</label>
                                    <div className="relative">
                                        <span className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500">
                                            {getCurrencySymbol(newPlan.currency || 'USD')}
                                        </span>
                                        <input
                                            type="number"
                                            className="w-full pl-8 pr-3 py-2 bg-gray-50 dark:bg-slate-800 border border-gray-200 dark:border-slate-700 rounded-lg focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all"
                                            value={newPlan.price_monthly}
                                            onChange={(e) => setNewPlan({ ...newPlan, price_monthly: Number(e.target.value) })}
                                        />
                                    </div>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Yearly Price</label>
                                    <div className="relative">
                                        <span className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500">
                                            {getCurrencySymbol(newPlan.currency || 'USD')}
                                        </span>
                                        <input
                                            type="number"
                                            className="w-full pl-8 pr-3 py-2 bg-gray-50 dark:bg-slate-800 border border-gray-200 dark:border-slate-700 rounded-lg focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all"
                                            value={newPlan.price_yearly}
                                            onChange={(e) => setNewPlan({ ...newPlan, price_yearly: Number(e.target.value) })}
                                        />
                                    </div>
                                </div>
                            </div>

                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Max Users</label>
                                    <input
                                        type="number"
                                        className="w-full px-3 py-2 bg-gray-50 dark:bg-slate-800 border border-gray-200 dark:border-slate-700 rounded-lg focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all"
                                        placeholder="0 for unlimited"
                                        value={newPlan.max_users}
                                        onChange={(e) => setNewPlan({ ...newPlan, max_users: Number(e.target.value) })}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Storage (GB)</label>
                                    <input
                                        type="number"
                                        className="w-full px-3 py-2 bg-gray-50 dark:bg-slate-800 border border-gray-200 dark:border-slate-700 rounded-lg focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all"
                                        value={newPlan.max_storage_gb}
                                        onChange={(e) => setNewPlan({ ...newPlan, max_storage_gb: Number(e.target.value) })}
                                    />
                                </div>
                            </div>
                        </div>
                        <DialogFooter>
                            <DialogClose asChild>
                                <button className="px-4 py-2 text-sm font-medium text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-slate-800 rounded-lg transition-colors">
                                    Cancel
                                </button>
                            </DialogClose>
                            <button
                                onClick={handleCreate}
                                className="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-lg shadow-lg shadow-blue-600/20 transition-all"
                            >
                                Create Plan
                            </button>
                        </DialogFooter>
                    </DialogContent>
                </Dialog>
            </div>

            {/* Plans Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
                {plans.map((plan) => (
                    <div
                        key={plan.id}
                        className={`
                            relative flex flex-col bg-white dark:bg-slate-900 rounded-2xl border transition-all duration-300
                            ${!plan.is_active ? 'opacity-75 grayscale bg-gray-50 dark:bg-slate-800/50' : 'hover:-translate-y-1 hover:shadow-xl shadow-sm'}
                            ${plan.name === 'enterprise' ? 'border-purple-200 dark:border-purple-900/50 shadow-purple-100 dark:shadow-purple-900/10' : 'border-gray-200 dark:border-slate-800'}
                        `}
                    >
                        {/* Status Badge */}
                        <div className="absolute top-4 right-4 flex items-center gap-2 z-10">
                            <span className={`px-2 py-0.5 rounded-full text-xs font-semibold ${plan.is_active
                                ? 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400'
                                : 'bg-gray-200 dark:bg-slate-700 text-gray-600 dark:text-gray-400'
                                }`}>
                                {plan.is_active ? 'Active' : 'Inactive'}
                            </span>
                            <DropdownMenu>
                                <DropdownMenuTrigger asChild>
                                    <button className="p-1 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-slate-800 rounded transition-colors">
                                        <MoreHorizontal className="w-4 h-4" />
                                    </button>
                                </DropdownMenuTrigger>
                                <DropdownMenuContent align="end">
                                    <DropdownMenuItem onClick={() => handleEditClick(plan)}>
                                        <Edit2 className="w-4 h-4 mr-2" /> Edit Plan
                                    </DropdownMenuItem>
                                    <DropdownMenuItem onClick={() => togglePlanStatus(plan.id, plan.is_active)}>
                                        {plan.is_active ? <X className="w-4 h-4 mr-2" /> : <Check className="w-4 h-4 mr-2" />}
                                        {plan.is_active ? 'Deactivate' : 'Activate'}
                                    </DropdownMenuItem>
                                    <DropdownMenuItem
                                        onClick={() => setDeleteId(plan.id)}
                                        className="text-red-600 dark:text-red-400 focus:text-red-600 dark:focus:text-red-400 focus:bg-red-50 dark:focus:bg-red-900/10"
                                    >
                                        <Trash2 className="w-4 h-4 mr-2" /> Delete Plan
                                    </DropdownMenuItem>
                                </DropdownMenuContent>
                            </DropdownMenu>
                        </div>

                        {/* Plan Content */}
                        <div className="p-6 flex-1 flex flex-col">
                            {/* Icon & Title */}
                            <div className="mb-4">
                                <div className={`w-12 h-12 rounded-xl flex items-center justify-center mb-4 ${plan.name === 'enterprise'
                                    ? 'bg-purple-100 dark:bg-purple-900/20 text-purple-600 dark:text-purple-400'
                                    : plan.name === 'pro'
                                        ? 'bg-blue-100 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400'
                                        : 'bg-gray-100 dark:bg-slate-800 text-gray-600 dark:text-gray-400'
                                    }`}>
                                    {plan.name === 'enterprise' ? <Server className="w-6 h-6" /> :
                                        plan.name === 'pro' ? <Zap className="w-6 h-6" /> :
                                            <Shield className="w-6 h-6" />}
                                </div>
                                {isEditing === plan.id ? (
                                    <>
                                        <input
                                            className="w-full mb-1 px-2 py-1 bg-white dark:bg-slate-800 border border-gray-200 dark:border-slate-700 rounded text-lg font-bold text-gray-900 dark:text-white"
                                            value={editForm.display_name}
                                            onChange={(e) => setEditForm({ ...editForm, display_name: e.target.value })}
                                        />
                                        <select
                                            className="w-full mt-1 px-2 py-1 text-xs bg-white dark:bg-slate-800 border border-gray-200 dark:border-slate-700 rounded text-gray-900 dark:text-white"
                                            value={editForm.currency}
                                            onChange={(e) => setEditForm({ ...editForm, currency: e.target.value })}
                                        >
                                            {currencies.map(c => (
                                                <option key={c.code} value={c.code}>{c.code}</option>
                                            ))}
                                        </select>
                                    </>
                                ) : (
                                    <h3 className="text-xl font-bold text-gray-900 dark:text-white">{plan.display_name}</h3>
                                )}
                                {isEditing === plan.id ? (
                                    <textarea
                                        className="w-full mt-2 px-2 py-1 bg-white dark:bg-slate-800 border border-gray-200 dark:border-slate-700 rounded text-sm text-gray-500 dark:text-gray-400 resize-none"
                                        value={editForm.description}
                                        onChange={(e) => setEditForm({ ...editForm, description: e.target.value })}
                                    />
                                ) : (
                                    <p className="text-sm text-gray-500 dark:text-gray-400 mt-2 line-clamp-2">{plan.description}</p>
                                )}
                            </div>

                            {/* Pricing */}
                            <div className="mb-6 pb-6 border-b border-gray-100 dark:border-slate-800">
                                {isEditing === plan.id ? (
                                    <div className="flex items-end gap-2">
                                        <div className="flex-1">
                                            <label className="text-xs text-gray-400">Monthly</label>
                                            <input
                                                type="number"
                                                className="w-full px-2 py-1 bg-white dark:bg-slate-800 border border-gray-200 dark:border-slate-700 rounded text-xl font-bold"
                                                value={editForm.price_monthly}
                                                onChange={(e) => setEditForm({ ...editForm, price_monthly: Number(e.target.value) })}
                                            />
                                        </div>
                                        <div className="flex-1">
                                            <label className="text-xs text-gray-400">Yearly</label>
                                            <input
                                                type="number"
                                                className="w-full px-2 py-1 bg-white dark:bg-slate-800 border border-gray-200 dark:border-slate-700 rounded text-xl font-bold"
                                                value={editForm.price_yearly}
                                                onChange={(e) => setEditForm({ ...editForm, price_yearly: Number(e.target.value) })}
                                            />
                                        </div>
                                    </div>
                                ) : (
                                    <div className="flex items-baseline gap-1">
                                        <span className="text-3xl font-extrabold text-gray-900 dark:text-white">
                                            {getCurrencySymbol(plan.currency)}
                                            {plan.price_monthly}
                                        </span>
                                        <span className="text-sm text-gray-500 dark:text-gray-400 font-medium">/month</span>
                                    </div>
                                )}
                                {!isEditing && (
                                    <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">
                                        or {getCurrencySymbol(plan.currency)}{plan.price_yearly}/year (save {Math.round((1 - (plan.price_yearly / (plan.price_monthly * 12 || 1))) * 100)}%)
                                    </p>
                                )}
                            </div>

                            {/* Limits */}
                            <ul className="space-y-3 mb-6 flex-1">
                                <li className="flex items-center text-sm text-gray-600 dark:text-gray-300">
                                    <Users className="w-4 h-4 text-blue-500 mr-2 shrink-0" />
                                    {isEditing === plan.id ? (
                                        <input
                                            type="number"
                                            className="w-20 px-1 py-0.5 bg-white dark:bg-slate-800 border border-gray-200 dark:border-slate-700 rounded"
                                            value={editForm.max_users}
                                            onChange={(e) => setEditForm({ ...editForm, max_users: Number(e.target.value) })}
                                        />
                                    ) : (
                                        <span>{plan.max_users === 0 ? "Unlimited" : plan.max_users} Users</span>
                                    )}
                                </li>
                                <li className="flex items-center text-sm text-gray-600 dark:text-gray-300">
                                    <HardDrive className="w-4 h-4 text-purple-500 mr-2 shrink-0" />
                                    {isEditing === plan.id ? (
                                        <input
                                            type="number"
                                            className="w-20 px-1 py-0.5 bg-white dark:bg-slate-800 border border-gray-200 dark:border-slate-700 rounded"
                                            value={editForm.max_storage_gb}
                                            onChange={(e) => setEditForm({ ...editForm, max_storage_gb: Number(e.target.value) })}
                                        />
                                    ) : (
                                        <span>{plan.max_storage_gb} GB Storage</span>
                                    )}
                                </li>
                            </ul>

                            {/* Editing Actions */}
                            {isEditing === plan.id && (
                                <div className="flex gap-2 mt-auto">
                                    <button
                                        onClick={handleSave}
                                        className="flex-1 bg-blue-600 hover:bg-blue-700 text-white px-3 py-2 rounded-lg text-sm font-medium transition-colors flex items-center justify-center gap-1"
                                    >
                                        <Save className="w-3 h-3" /> Save
                                    </button>
                                    <button
                                        onClick={() => setIsEditing(null)}
                                        className="flex-1 bg-gray-100 hover:bg-gray-200 text-gray-700 px-3 py-2 rounded-lg text-sm font-medium transition-colors"
                                    >
                                        Cancel
                                    </button>
                                </div>
                            )}

                            {/* Activate Button for Inactive Plans */}
                            {!plan.is_active && !isEditing && (
                                <div className="absolute inset-0 flex items-center justify-center bg-white/50 dark:bg-slate-900/50 backdrop-blur-[1px] rounded-2xl z-0 transition-opacity opacity-0 hover:opacity-100">
                                    <button
                                        onClick={(e) => {
                                            e.stopPropagation();
                                            togglePlanStatus(plan.id, plan.is_active);
                                        }}
                                        className="bg-green-600 hover:bg-green-700 text-white px-6 py-3 rounded-xl font-bold shadow-xl transform hover:scale-105 transition-all flex items-center gap-2"
                                    >
                                        <Check className="w-5 h-5" /> Activate Plan
                                    </button>
                                </div>
                            )}
                        </div>
                    </div>
                ))}
            </div>
            {/* Delete Confirmation Dialog */}
            <Dialog open={!!deleteId} onOpenChange={(open) => !open && setDeleteId(null)}>
                <DialogContent className="sm:max-w-[425px] bg-white dark:bg-slate-900 border-gray-100 dark:border-slate-800">
                    <DialogHeader>
                        <DialogTitle className="text-xl font-bold text-gray-900 dark:text-white flex items-center gap-2">
                            <Trash2 className="w-5 h-5 text-red-600" />
                            Delete Plan
                        </DialogTitle>
                        <DialogDescription className="text-gray-500 dark:text-gray-400 pt-2">
                            Are you sure you want to delete this plan? This action cannot be undone.
                            <br /><br />
                            <span className="text-sm text-red-500 bg-red-50 dark:bg-red-900/10 px-2 py-1 rounded">
                                Note: Plans with active subscriptions cannot be deleted.
                            </span>
                        </DialogDescription>
                    </DialogHeader>
                    <DialogFooter className="gap-2 mt-4">
                        <button
                            onClick={() => setDeleteId(null)}
                            className="px-4 py-2 text-sm font-medium text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-slate-800 rounded-lg transition-colors"
                        >
                            Cancel
                        </button>
                        <button
                            onClick={handleDelete}
                            className="px-4 py-2 text-sm font-medium text-white bg-red-600 hover:bg-red-700 rounded-lg shadow-lg shadow-red-600/20 transition-all flex items-center gap-2"
                        >
                            <Trash2 className="w-4 h-4" /> Delete Plan
                        </button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}
