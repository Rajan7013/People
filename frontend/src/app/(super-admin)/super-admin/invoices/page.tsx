"use client";

import { useEffect, useState } from 'react';
import {
    Download,
    Filter,
    Search,
    Eye,
    CheckCircle2,
    XCircle,
    Clock,
    FileText,
    MoreVertical,
    Calendar,
    ArrowUpRight,
    Loader2,
    SearchX,
    CreditCard,
    Plus,
    Trash2,
    Edit,
    AlertTriangle
} from 'lucide-react';
import { invoiceService, Invoice, organizationService, Organization } from '@/services/superAdminService';
import { toast } from 'sonner';
import { format } from 'date-fns';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";

export default function InvoicesPage() {
    const [invoices, setInvoices] = useState<Invoice[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchTerm, setSearchTerm] = useState('');
    const [statusFilter, setStatusFilter] = useState<'all' | 'paid' | 'pending' | 'overdue' | 'void'>('all');

    useEffect(() => {
        loadInvoices();
    }, []);

    const loadInvoices = async () => {
        try {
            setLoading(true);
            const response = await invoiceService.getAll();
            setInvoices(response.data.invoices || []);
        } catch (error) {
            console.error('Failed to load invoices:', error);
            toast.error("Failed to load invoices");
        } finally {
            setLoading(false);
        }
    };

    const [viewInvoiceId, setViewInvoiceId] = useState<string | null>(null);
    const [pdfUrl, setPdfUrl] = useState<string | null>(null);

    // Edit/Delete State
    const [editInvoice, setEditInvoice] = useState<Invoice | null>(null);
    const [editLineItems, setEditLineItems] = useState<{ description: string; quantity: number; unit_price: number }[]>([]);
    const [deleteInvoiceId, setDeleteInvoiceId] = useState<string | null>(null);

    // Create Invoice State
    const [isCreateOpen, setIsCreateOpen] = useState(false);
    const [tenants, setTenants] = useState<Organization[]>([]);
    const [newLineItems, setNewLineItems] = useState<{ description: string; quantity: number; unit_price: number }[]>([
        { description: '', quantity: 1, unit_price: 0 }
    ]);
    const [newInvoiceData, setNewInvoiceData] = useState({
        tenant_id: '',
        currency: 'USD',
        issue_date: format(new Date(), 'yyyy-MM-dd'),
        due_date: format(new Date(Date.now() + 7 * 24 * 60 * 60 * 1000), 'yyyy-MM-dd'), // +7 days
        notes: ''
    });

    useEffect(() => {
        if (isCreateOpen) {
            loadTenants();
        }
    }, [isCreateOpen]);

    const loadTenants = async () => {
        try {
            const response = await organizationService.getAll();
            setTenants(response.data.organizations || []);
        } catch (error) {
            console.error('Failed to load tenants:', error);
            toast.error('Failed to load organizations');
        }
    };

    const handleAddLineItem = () => {
        setNewLineItems([...newLineItems, { description: '', quantity: 1, unit_price: 0 }]);
    };

    const handleRemoveLineItem = (index: number) => {
        setNewLineItems(newLineItems.filter((_, i) => i !== index));
    };

    const handleLineItemChange = (index: number, field: string, value: any) => {
        const updated = [...newLineItems];
        updated[index] = { ...updated[index], [field]: value };
        setNewLineItems(updated);
    };

    const calculateTotal = () => {
        return newLineItems.reduce((sum, item) => sum + (item.quantity * item.unit_price), 0);
    };

    const handleCreateInvoice = async () => {
        if (!newInvoiceData.tenant_id) {
            toast.error('Please select an organization');
            return;
        }
        if (newLineItems.some(item => !item.description || item.quantity <= 0 || item.unit_price < 0)) {
            toast.error('Please fill in valid line items');
            return;
        }

        try {
            const loadingToast = toast.loading('Creating invoice...');
            const payload = {
                ...newInvoiceData,
                subtotal: calculateTotal(),
                tax_amount: 0,
                total_amount: calculateTotal(),
                line_items: newLineItems,
                status: 'pending' as const // Explicitly type as Invoice status
            };
            console.log('Creating invoice with payload:', payload);
            await invoiceService.generate(payload);

            toast.dismiss(loadingToast);
            toast.success('Invoice created successfully');
            setIsCreateOpen(false);
            loadInvoices(); // Refresh list
            // Reset form
            setNewLineItems([{ description: '', quantity: 1, unit_price: 0 }]);
            setNewInvoiceData({
                tenant_id: '',
                currency: 'USD',
                issue_date: format(new Date(), 'yyyy-MM-dd'),
                due_date: format(new Date(Date.now() + 7 * 24 * 60 * 60 * 1000), 'yyyy-MM-dd'),
                notes: ''
            });
        } catch (error) {
            console.error(error);
            toast.error('Failed to create invoice');
        }
    };

    // ... (existing helper functions) ...

    const handleDownload = async (id: string, number: string) => {
        try {
            const loadingToast = toast.loading('Generating invoice PDF...');
            const response = await invoiceService.download(id);

            // Create Blob URL
            const blob = new Blob([response.data], { type: 'application/pdf' });
            const url = window.URL.createObjectURL(blob);

            // Trigger Download
            const link = document.createElement('a');
            link.href = url;
            link.setAttribute('download', `Invoice-${number}.pdf`);
            document.body.appendChild(link);
            link.click();
            link.parentNode?.removeChild(link);
            window.URL.revokeObjectURL(url);

            toast.dismiss(loadingToast);
            toast.success('Invoice downloaded successfully');
        } catch (error) {
            console.error(error);
            toast.error('Failed to download invoice');
        }
    };

    const handleView = async (id: string) => {
        try {
            const loadingToast = toast.loading('Loading invoice...');
            const response = await invoiceService.download(id);

            const blob = new Blob([response.data], { type: 'application/pdf' });
            const url = window.URL.createObjectURL(blob);

            setPdfUrl(url);
            setViewInvoiceId(id);
            toast.dismiss(loadingToast);
        } catch (error) {
            console.error(error);
            toast.error('Failed to load invoice PDF');
        }
    };

    const handleEdit = (invoice: Invoice) => {
        if (invoice.status === 'paid') {
            toast.error('Cannot edit paid invoices');
            return;
        }
        setEditInvoice(invoice);
        // Populate edit line items from invoice
        setEditLineItems(invoice.line_items || [{ description: '', quantity: 1, unit_price: 0 }]);
    };

    const handleDelete = async () => {
        if (!deleteInvoiceId) return;

        try {
            await invoiceService.delete(deleteInvoiceId);
            toast.success('Invoice deleted successfully');
            setDeleteInvoiceId(null);
            loadInvoices();
        } catch (error: any) {
            console.error('Failed to delete invoice:', error);
            toast.error(error.response?.data?.error || 'Failed to delete invoice');
        }
    };

    // Clean up object URL when modal closes
    useEffect(() => {
        if (!viewInvoiceId && pdfUrl) {
            window.URL.revokeObjectURL(pdfUrl);
            setPdfUrl(null);
        }
    }, [viewInvoiceId]);

    // Helper functions
    const filteredInvoices = invoices.filter(inv => {
        const matchesSearch =
            inv.invoice_number.toLowerCase().includes(searchTerm.toLowerCase()) ||
            inv.tenant_name?.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesStatus = statusFilter === 'all' || inv.status === statusFilter;
        return matchesSearch && matchesStatus;
    });

    const getStatusConfig = (status: string) => {
        switch (status) {
            case 'paid':
                return {
                    color: 'text-green-700 dark:text-green-400 bg-green-50 dark:bg-green-900/30 border-green-200 dark:border-green-900/50',
                    icon: CheckCircle2,
                    label: 'Paid'
                };
            case 'pending':
                return {
                    color: 'text-yellow-700 dark:text-yellow-400 bg-yellow-50 dark:bg-yellow-900/30 border-yellow-200 dark:border-yellow-900/50',
                    icon: Clock,
                    label: 'Pending'
                };
            case 'overdue':
                return {
                    color: 'text-red-700 dark:text-red-400 bg-red-50 dark:bg-red-900/30 border-red-200 dark:border-red-900/50',
                    icon: XCircle,
                    label: 'Overdue'
                };
            default:
                return {
                    color: 'text-gray-700 dark:text-gray-400 bg-gray-50 dark:bg-slate-800 border-gray-200 dark:border-slate-700',
                    icon: FileText,
                    label: status
                };
        }
    };

    if (loading) {
        return (
            <div className="flex flex-col items-center justify-center min-h-[60vh]">
                <Loader2 className="w-10 h-10 text-blue-600 animate-spin mb-4" />
                <p className="text-gray-500 font-medium">Loading invoices...</p>
            </div>
        );
    }

    // ... (render logic) ...

    return (
        <div className="space-y-8 animate-in fade-in duration-500">
            {/* Header */}
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div>
                    <h1 className="text-3xl font-bold text-gray-900 dark:text-white tracking-tight">Invoices & Billing</h1>
                    <p className="text-gray-500 dark:text-gray-400 mt-2 text-lg">Track payments and manage financial records.</p>
                </div>
                <button
                    onClick={() => setIsCreateOpen(true)}
                    className="px-4 py-2.5 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-xl transition-colors flex items-center gap-2 shadow-sm"
                >
                    <Plus className="w-5 h-5" />
                    Create Invoice
                </button>
            </div>

            {/* ... (Stats, Filters) ... */}

            {/* Stats Overview */}
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                <div className="bg-white dark:bg-slate-900 p-4 rounded-2xl border border-gray-100 dark:border-slate-800 shadow-sm flex items-center gap-4">
                    <div className="p-3 bg-green-50 dark:bg-green-900/20 rounded-xl text-green-600 dark:text-green-400">
                        <CheckCircle2 className="w-6 h-6" />
                    </div>
                    <div>
                        <p className="text-sm text-gray-500 dark:text-gray-400 font-medium">Total Collected</p>
                        <p className="text-xl font-bold text-gray-900 dark:text-white">
                            ${invoices.filter(i => i.status === 'paid').reduce((acc, curr) => acc + curr.total_amount, 0).toLocaleString()}
                        </p>
                    </div>
                </div>
                <div className="bg-white dark:bg-slate-900 p-4 rounded-2xl border border-gray-100 dark:border-slate-800 shadow-sm flex items-center gap-4">
                    <div className="p-3 bg-yellow-50 dark:bg-yellow-900/20 rounded-xl text-yellow-600 dark:text-yellow-400">
                        <Clock className="w-6 h-6" />
                    </div>
                    <div>
                        <p className="text-sm text-gray-500 dark:text-gray-400 font-medium">Pending Payments</p>
                        <p className="text-xl font-bold text-gray-900 dark:text-white">
                            ${invoices.filter(i => i.status === 'pending').reduce((acc, curr) => acc + curr.total_amount, 0).toLocaleString()}
                        </p>
                    </div>
                </div>
                <div className="bg-white dark:bg-slate-900 p-4 rounded-2xl border border-gray-100 dark:border-slate-800 shadow-sm flex items-center gap-4">
                    <div className="p-3 bg-red-50 dark:bg-red-900/20 rounded-xl text-red-600 dark:text-red-400">
                        <XCircle className="w-6 h-6" />
                    </div>
                    <div>
                        <p className="text-sm text-gray-500 dark:text-gray-400 font-medium">Overdue</p>
                        <p className="text-xl font-bold text-gray-900 dark:text-white">
                            ${invoices.filter(i => i.status === 'overdue').reduce((acc, curr) => acc + curr.total_amount, 0).toLocaleString()}
                        </p>
                    </div>
                </div>
            </div>

            {/* Filters & Search */}
            <div className="flex flex-col sm:flex-row gap-4">
                <div className="relative flex-1">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-gray-400 dark:text-gray-500" />
                    <input
                        type="text"
                        placeholder="Search by invoice # or organization..."
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                        className="w-full pl-10 pr-4 py-3 rounded-xl border border-gray-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all placeholder:text-gray-400"
                    />
                </div>
                <div className="flex items-center gap-2">
                    <div className="relative">
                        <Filter className="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-gray-400 dark:text-gray-500" />
                        <select
                            value={statusFilter}
                            onChange={(e) => setStatusFilter(e.target.value as any)}
                            className="pl-10 pr-8 py-3 rounded-xl border border-gray-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none appearance-none cursor-pointer"
                        >
                            <option value="all">All Status</option>
                            <option value="paid">Paid</option>
                            <option value="pending">Pending</option>
                            <option value="overdue">Overdue</option>
                        </select>
                    </div>
                </div>
            </div>

            {/* Invoices Table */}
            <div className="bg-white dark:bg-slate-900 rounded-2xl border border-gray-100 dark:border-slate-800 shadow-sm overflow-hidden">
                {filteredInvoices.length === 0 ? (
                    <div className="p-12 text-center">
                        <div className="w-16 h-16 bg-gray-50 dark:bg-slate-800 rounded-full flex items-center justify-center mx-auto mb-4">
                            <SearchX className="w-8 h-8 text-gray-400 dark:text-gray-500" />
                        </div>
                        <h3 className="text-lg font-bold text-gray-900 dark:text-white mb-2">No invoices found</h3>
                        <p className="text-gray-500 dark:text-gray-400 max-w-md mx-auto">
                            We couldn't find any invoices matching your search. Try changing the filters or search terms.
                        </p>
                    </div>
                ) : (
                    <div className="overflow-x-auto">
                        <table className="w-full text-left">
                            <thead className="bg-gray-50/50 dark:bg-slate-800/50 border-b border-gray-100 dark:border-slate-800">
                                <tr>
                                    <th className="px-6 py-4 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Invoice</th>
                                    <th className="px-6 py-4 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Organization</th>
                                    <th className="px-6 py-4 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Amount</th>
                                    <th className="px-6 py-4 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Date Issued</th>
                                    <th className="px-6 py-4 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Status</th>
                                    <th className="px-6 py-4 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider text-right">Actions</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-gray-100 dark:divide-slate-800">
                                {filteredInvoices.map((inv) => {
                                    const statusConfig = getStatusConfig(inv.status);
                                    const StatusIcon = statusConfig.icon;

                                    return (
                                        <tr
                                            key={inv.id}
                                            className="group hover:bg-gray-50 dark:hover:bg-slate-800/50 transition-colors"
                                        >
                                            <td className="px-6 py-4">
                                                <div className="flex items-center gap-3">
                                                    <div className="p-2 bg-gray-100 dark:bg-slate-800 rounded-lg text-gray-500 dark:text-gray-400 group-hover:bg-white dark:group-hover:bg-slate-700 group-hover:text-blue-600 dark:group-hover:text-blue-400 group-hover:shadow-sm transition-all duration-200">
                                                        <FileText className="w-5 h-5" />
                                                    </div>
                                                    <div>
                                                        <span className="font-bold text-gray-900 dark:text-white text-sm block">{inv.invoice_number}</span>
                                                        <span className="text-xs text-gray-500 dark:text-gray-400">Inv #{inv.invoice_number.slice(-4)}</span>
                                                    </div>
                                                </div>
                                            </td>
                                            <td className="px-6 py-4">
                                                <div className="flex items-center gap-2">
                                                    <div className="w-6 h-6 rounded-full bg-linear-to-br from-blue-500 to-indigo-600 flex items-center justify-center text-white text-xs font-bold">
                                                        {inv.tenant_name?.substring(0, 1).toUpperCase()}
                                                    </div>
                                                    <span className="text-sm font-medium text-gray-700 dark:text-gray-300">{inv.tenant_name}</span>
                                                </div>
                                            </td>
                                            <td className="px-6 py-4">
                                                <span className="text-sm font-bold text-gray-900 dark:text-white">
                                                    ${inv.total_amount.toFixed(2)}
                                                </span>
                                            </td>
                                            <td className="px-6 py-4">
                                                <div className="flex items-center gap-1.5 text-sm text-gray-500 dark:text-gray-400">
                                                    <Calendar className="w-4 h-4" />
                                                    {format(new Date(inv.issue_date), 'MMM d, yyyy')}
                                                </div>
                                            </td>
                                            <td className="px-6 py-4">
                                                <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold border ${statusConfig.color}`}>
                                                    <StatusIcon className="w-3.5 h-3.5" />
                                                    {statusConfig.label}
                                                </span>
                                            </td>
                                            <td className="px-6 py-4 text-right">
                                                <div className="flex items-center justify-end gap-2">
                                                    <button
                                                        onClick={() => handleView(inv.id)}
                                                        className="p-2 text-gray-400 hover:text-blue-600 dark:hover:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/30 rounded-lg transition-colors"
                                                        title="View Invoice"
                                                    >
                                                        <Eye className="w-4 h-4" />
                                                    </button>
                                                    <button
                                                        onClick={() => handleDownload(inv.id, inv.invoice_number)}
                                                        className="p-2 text-gray-400 hover:text-blue-600 dark:hover:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/30 rounded-lg transition-colors"
                                                        title="Download PDF"
                                                    >
                                                        <Download className="w-4 h-4" />
                                                    </button>

                                                    <DropdownMenu>
                                                        <DropdownMenuTrigger asChild>
                                                            <button className="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-slate-800 rounded-lg transition-colors">
                                                                <MoreVertical className="w-4 h-4" />
                                                            </button>
                                                        </DropdownMenuTrigger>
                                                        <DropdownMenuContent align="end">
                                                            <DropdownMenuItem onClick={() => handleView(inv.id)}>
                                                                <Eye className="w-4 h-4 mr-2" /> View Details
                                                            </DropdownMenuItem>
                                                            <DropdownMenuItem onClick={() => handleDownload(inv.id, inv.invoice_number)}>
                                                                <Download className="w-4 h-4 mr-2" /> Download PDF
                                                            </DropdownMenuItem>
                                                            <DropdownMenuItem
                                                                onClick={() => handleEdit(inv)}
                                                                disabled={inv.status === 'paid'}
                                                            >
                                                                <Edit className="w-4 h-4 mr-2" /> Edit Invoice
                                                            </DropdownMenuItem>
                                                            <DropdownMenuItem
                                                                onClick={() => setDeleteInvoiceId(inv.id)}
                                                                disabled={inv.status === 'paid'}
                                                                className="text-red-600 dark:text-red-400"
                                                            >
                                                                <Trash2 className="w-4 h-4 mr-2" /> Delete Invoice
                                                            </DropdownMenuItem>
                                                        </DropdownMenuContent>
                                                    </DropdownMenu>
                                                </div>
                                            </td>
                                        </tr>
                                    );
                                })}
                            </tbody>
                        </table>
                    </div>
                )}
            </div>

            {/* View PDF Modal */}
            <Dialog open={!!viewInvoiceId} onOpenChange={(open) => !open && setViewInvoiceId(null)}>
                <DialogContent className="max-w-4xl h-[85vh] flex flex-col p-0 gap-0 bg-gray-100 dark:bg-slate-900">
                    <DialogHeader className="p-4 bg-white dark:bg-slate-800 border-b border-gray-200 dark:border-slate-700 flex flex-row items-center justify-between">
                        <DialogTitle>View Invoice</DialogTitle>
                        <div className="flex gap-2">
                            <button
                                onClick={() => {
                                    const iframe = document.getElementById('pdf-frame') as HTMLIFrameElement;
                                    iframe?.contentWindow?.print();
                                }}
                                className="px-3 py-1.5 text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-slate-700 rounded-lg transition-colors flex items-center gap-2"
                            >
                                <div className="w-4 h-4"><FileText className="w-4 h-4" /></div> Print
                            </button>
                            <button
                                onClick={() => {
                                    // Re-use download logic or just download via link
                                    const link = document.createElement('a');
                                    link.href = pdfUrl!;
                                    link.download = "invoice.pdf";
                                    link.click();
                                }}
                                className="px-3 py-1.5 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-lg transition-colors flex items-center gap-2"
                            >
                                <Download className="w-4 h-4" /> Download
                            </button>
                        </div>
                    </DialogHeader>
                    <div className="flex-1 bg-gray-200 dark:bg-slate-950 p-4">
                        {pdfUrl && (
                            <iframe
                                id="pdf-frame"
                                src={pdfUrl}
                                className="w-full h-full rounded-lg shadow-lg border-none"
                                title="Invoice PDF"
                            />
                        )}
                    </div>
                </DialogContent>
            </Dialog>

            {/* Create Invoice Modal */}
            <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
                <DialogContent className="max-w-3xl max-h-[90vh] overflow-y-auto">
                    <DialogHeader>
                        <DialogTitle>Create New Invoice</DialogTitle>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        {/* Tenant Selection */}
                        <div>
                            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                Organization *
                            </label>
                            <select
                                value={newInvoiceData.tenant_id}
                                onChange={(e) => setNewInvoiceData({ ...newInvoiceData, tenant_id: e.target.value })}
                                className="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none"
                            >
                                <option value="">Select organization...</option>
                                {tenants.map((tenant) => (
                                    <option key={tenant.id} value={tenant.id}>
                                        {tenant.name}
                                    </option>
                                ))}
                            </select>
                        </div>

                        {/* Dates */}
                        <div className="grid grid-cols-2 gap-4">
                            <div>
                                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                    Issue Date
                                </label>
                                <input
                                    type="date"
                                    value={newInvoiceData.issue_date}
                                    onChange={(e) => setNewInvoiceData({ ...newInvoiceData, issue_date: e.target.value })}
                                    className="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none"
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                    Due Date
                                </label>
                                <input
                                    type="date"
                                    value={newInvoiceData.due_date}
                                    onChange={(e) => setNewInvoiceData({ ...newInvoiceData, due_date: e.target.value })}
                                    className="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none"
                                />
                            </div>
                        </div>

                        {/* Currency */}
                        <div>
                            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                Currency
                            </label>
                            <select
                                value={newInvoiceData.currency}
                                onChange={(e) => setNewInvoiceData({ ...newInvoiceData, currency: e.target.value })}
                                className="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none"
                            >
                                <option value="USD">USD ($)</option>
                                <option value="INR">INR (₹)</option>
                                <option value="EUR">EUR (€)</option>
                                <option value="GBP">GBP (£)</option>
                            </select>
                        </div>

                        {/* Line Items */}
                        <div>
                            <div className="flex items-center justify-between mb-3">
                                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                                    Line Items *
                                </label>
                                <button
                                    onClick={handleAddLineItem}
                                    className="text-sm text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300 font-medium flex items-center gap-1"
                                >
                                    <Plus className="w-4 h-4" /> Add Item
                                </button>
                            </div>
                            <div className="space-y-3">
                                {newLineItems.map((item, index) => (
                                    <div key={index} className="flex gap-2 items-start">
                                        <input
                                            type="text"
                                            placeholder="Description"
                                            value={item.description}
                                            onChange={(e) => handleLineItemChange(index, 'description', e.target.value)}
                                            className="flex-1 px-3 py-2 rounded-lg border border-gray-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none text-sm"
                                        />
                                        <input
                                            type="number"
                                            placeholder="Qty"
                                            value={item.quantity}
                                            onChange={(e) => handleLineItemChange(index, 'quantity', parseInt(e.target.value) || 0)}
                                            className="w-20 px-3 py-2 rounded-lg border border-gray-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none text-sm"
                                            min="1"
                                        />
                                        <input
                                            type="number"
                                            placeholder="Price"
                                            value={item.unit_price}
                                            onChange={(e) => handleLineItemChange(index, 'unit_price', parseFloat(e.target.value) || 0)}
                                            className="w-28 px-3 py-2 rounded-lg border border-gray-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none text-sm"
                                            min="0"
                                            step="0.01"
                                        />
                                        <button
                                            onClick={() => handleRemoveLineItem(index)}
                                            disabled={newLineItems.length === 1}
                                            className="p-2 text-red-600 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300 disabled:opacity-30 disabled:cursor-not-allowed"
                                        >
                                            <Trash2 className="w-4 h-4" />
                                        </button>
                                    </div>
                                ))}
                            </div>
                        </div>

                        {/* Total */}
                        <div className="flex justify-end items-center gap-2 pt-2 border-t border-gray-200 dark:border-slate-700">
                            <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Total:</span>
                            <span className="text-lg font-bold text-gray-900 dark:text-white">
                                {newInvoiceData.currency} {calculateTotal().toFixed(2)}
                            </span>
                        </div>

                        {/* Notes */}
                        <div>
                            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                Notes (Optional)
                            </label>
                            <textarea
                                value={newInvoiceData.notes}
                                onChange={(e) => setNewInvoiceData({ ...newInvoiceData, notes: e.target.value })}
                                placeholder="Add any additional notes..."
                                rows={3}
                                className="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none resize-none"
                            />
                        </div>

                        {/* Actions */}
                        <div className="flex justify-end gap-3 pt-4">
                            <button
                                onClick={() => setIsCreateOpen(false)}
                                className="px-4 py-2 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-slate-700 rounded-lg transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={handleCreateInvoice}
                                className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition-colors"
                            >
                                Create Invoice
                            </button>
                        </div>
                    </div>
                </DialogContent>
            </Dialog>

            {/* Edit Invoice Dialog */}
            <Dialog open={!!editInvoice} onOpenChange={() => { setEditInvoice(null); setEditLineItems([]); }}>
                <DialogContent className="sm:max-w-3xl max-h-[90vh] overflow-y-auto">
                    <DialogHeader>
                        <DialogTitle>Edit Invoice - {editInvoice?.invoice_number}</DialogTitle>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        {editInvoice && (
                            <>
                                {/* Organization Info (Read-only) */}
                                <div className="bg-gray-50 dark:bg-slate-800/50 p-4 rounded-lg">
                                    <div className="flex justify-between">
                                        <span className="text-sm text-gray-600 dark:text-gray-400">Organization:</span>
                                        <span className="text-sm font-semibold text-gray-900 dark:text-white">{editInvoice.tenant_name}</span>
                                    </div>
                                </div>

                                {/* Dates */}
                                <div className="grid grid-cols-2 gap-4">
                                    <div>
                                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                            Issue Date
                                        </label>
                                        <input
                                            type="date"
                                            value={editInvoice.issue_date.split('T')[0]}
                                            onChange={(e) => setEditInvoice({ ...editInvoice, issue_date: e.target.value })}
                                            className="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none"
                                        />
                                    </div>
                                    <div>
                                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                            Due Date
                                        </label>
                                        <input
                                            type="date"
                                            value={editInvoice.due_date.split('T')[0]}
                                            onChange={(e) => setEditInvoice({ ...editInvoice, due_date: e.target.value })}
                                            className="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none"
                                        />
                                    </div>
                                </div>

                                {/* Currency & Status */}
                                <div className="grid grid-cols-2 gap-4">
                                    <div>
                                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                            Currency
                                        </label>
                                        <select
                                            value={editInvoice.currency}
                                            onChange={(e) => setEditInvoice({ ...editInvoice, currency: e.target.value })}
                                            className="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none"
                                        >
                                            <option value="USD">USD ($)</option>
                                            <option value="INR">INR (₹)</option>
                                            <option value="EUR">EUR (€)</option>
                                            <option value="GBP">GBP (£)</option>
                                        </select>
                                    </div>
                                    <div>
                                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                            Status
                                        </label>
                                        <select
                                            value={editInvoice.status}
                                            onChange={(e) => setEditInvoice({ ...editInvoice, status: e.target.value as any })}
                                            className="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none"
                                        >
                                            <option value="draft">Draft</option>
                                            <option value="pending">Pending</option>
                                            <option value="paid">Paid</option>
                                            <option value="overdue">Overdue</option>
                                            <option value="cancelled">Cancelled</option>
                                        </select>
                                    </div>
                                </div>

                                {/* Line Items */}
                                <div>
                                    <div className="flex items-center justify-between mb-3">
                                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                                            Line Items
                                        </label>
                                        <button
                                            onClick={() => setEditLineItems([...editLineItems, { description: '', quantity: 1, unit_price: 0 }])}
                                            className="text-sm text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300 font-medium flex items-center gap-1"
                                        >
                                            <Plus className="w-4 h-4" /> Add Item
                                        </button>
                                    </div>
                                    <div className="space-y-3">
                                        {editLineItems.map((item, index) => (
                                            <div key={index} className="flex gap-2 items-start">
                                                <input
                                                    type="text"
                                                    placeholder="Description"
                                                    value={item.description}
                                                    onChange={(e) => {
                                                        const updated = [...editLineItems];
                                                        updated[index].description = e.target.value;
                                                        setEditLineItems(updated);
                                                    }}
                                                    className="flex-1 px-3 py-2 rounded-lg border border-gray-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none text-sm"
                                                />
                                                <input
                                                    type="number"
                                                    placeholder="Qty"
                                                    value={item.quantity}
                                                    onChange={(e) => {
                                                        const updated = [...editLineItems];
                                                        updated[index].quantity = parseInt(e.target.value) || 0;
                                                        setEditLineItems(updated);
                                                    }}
                                                    className="w-20 px-3 py-2 rounded-lg border border-gray-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none text-sm"
                                                    min="1"
                                                />
                                                <input
                                                    type="number"
                                                    placeholder="Price"
                                                    value={item.unit_price}
                                                    onChange={(e) => {
                                                        const updated = [...editLineItems];
                                                        updated[index].unit_price = parseFloat(e.target.value) || 0;
                                                        setEditLineItems(updated);
                                                    }}
                                                    className="w-28 px-3 py-2 rounded-lg border border-gray-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none text-sm"
                                                    step="0.01"
                                                    min="0"
                                                />
                                                {editLineItems.length > 1 && (
                                                    <button
                                                        onClick={() => setEditLineItems(editLineItems.filter((_, i) => i !== index))}
                                                        className="p-2 text-red-600 hover:bg-red-50 dark:hover:bg-red-900/30 rounded-lg transition-colors"
                                                    >
                                                        <Trash2 className="w-4 h-4" />
                                                    </button>
                                                )}
                                            </div>
                                        ))}
                                    </div>
                                </div>

                                {/* Total */}
                                <div className="flex justify-end items-center gap-2 pt-2 border-t border-gray-200 dark:border-slate-700">
                                    <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Total:</span>
                                    <span className="text-lg font-bold text-gray-900 dark:text-white">
                                        {editInvoice.currency} {editLineItems.reduce((sum, item) => sum + (item.quantity * item.unit_price), 0).toFixed(2)}
                                    </span>
                                </div>

                                {/* Actions */}
                                <div className="flex justify-end gap-3 pt-4">
                                    <button
                                        onClick={() => { setEditInvoice(null); setEditLineItems([]); }}
                                        className="px-4 py-2 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-slate-700 rounded-lg transition-colors"
                                    >
                                        Cancel
                                    </button>
                                    <button
                                        onClick={async () => {
                                            try {
                                                const total = editLineItems.reduce((sum, item) => sum + (item.quantity * item.unit_price), 0);

                                                // Convert dates to ISO format with time
                                                const issueDate = new Date(editInvoice.issue_date).toISOString();
                                                const dueDate = new Date(editInvoice.due_date).toISOString();

                                                await invoiceService.update(editInvoice.id, {
                                                    status: editInvoice.status,
                                                    tenant_id: editInvoice.tenant_id,
                                                    subtotal: total,
                                                    tax_rate: 0,
                                                    tax_amount: 0,
                                                    discount_amount: 0,
                                                    total_amount: total,
                                                    currency: editInvoice.currency,
                                                    issue_date: issueDate,
                                                    due_date: dueDate,
                                                    billing_details: {},
                                                    line_items: editLineItems,
                                                    notes: '',
                                                } as any);
                                                toast.success('Invoice updated successfully');
                                                setEditInvoice(null);
                                                setEditLineItems([]);
                                                loadInvoices();
                                            } catch (error: any) {
                                                console.error('Failed to update invoice:', error);
                                                console.error('Error response:', error.response?.data);
                                                toast.error(error.response?.data?.error || 'Failed to update invoice');
                                            }
                                        }}
                                        className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition-colors"
                                    >
                                        Update Invoice
                                    </button>
                                </div>
                            </>
                        )}
                    </div>
                </DialogContent>
            </Dialog>

            {/* Delete Confirmation Dialog */}
            <Dialog open={!!deleteInvoiceId} onOpenChange={() => setDeleteInvoiceId(null)}>
                <DialogContent className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle className="flex items-center gap-2 text-red-600 dark:text-red-400">
                            <AlertTriangle className="w-5 h-5" />
                            Delete Invoice
                        </DialogTitle>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        <p className="text-gray-700 dark:text-gray-300">
                            Are you sure you want to delete this invoice? This action cannot be undone.
                        </p>
                        <div className="flex justify-end gap-3 pt-4">
                            <button
                                onClick={() => setDeleteInvoiceId(null)}
                                className="px-4 py-2 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-slate-700 rounded-lg transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={handleDelete}
                                className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white font-medium rounded-lg transition-colors"
                            >
                                Delete Invoice
                            </button>
                        </div>
                    </div>
                </DialogContent>
            </Dialog>
        </div>
    );
}
