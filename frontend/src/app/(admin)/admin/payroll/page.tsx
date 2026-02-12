"use client";

import { useState, useEffect } from "react";
import { format } from "date-fns";
import { Download, Trash2, FileText, Filter, Plus } from "lucide-react";
import { toast } from "sonner";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import api from "@/lib/api";

interface Payslip {
    id: string;
    employee_id: string;
    employee?: {
        first_name: string;
        last_name: string;
        email: string;
    };
    role?: string;
    pay_period_start: string;
    pay_period_end: string;
    payment_date?: string;
    net_salary: number;
    status: string;
}

export default function AdminPayrollPage() {
    const [payslips, setPayslips] = useState<Payslip[]>([]);
    const [loading, setLoading] = useState(true);
    const [month, setMonth] = useState<string>(String(new Date().getMonth() + 1));
    const [year, setYear] = useState<string>(String(new Date().getFullYear()));

    const fetchPayslips = async () => {
        try {
            setLoading(true);
            const queryParams = new URLSearchParams();
            if (month) queryParams.append("month", month);
            if (year) queryParams.append("year", year);

            const response = await api.get(`/company/hr/payslips?${queryParams.toString()}`);
            setPayslips(response.data.payslips || []);
        } catch (error) {
            console.error("Failed to fetch payslips:", error);
            toast.error("Failed to load payslips");
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchPayslips();
    }, [month, year]);

    const handleDelete = async (id: string) => {
        if (!confirm("Are you sure you want to delete this payslip?")) return;
        try {
            await api.delete(`/company/hr/payslips/${id}`);
            toast.success("Payslip deleted successfully");
            fetchPayslips();
        } catch (error) {
            console.error("Failed to delete payslip:", error);
            toast.error("Failed to delete payslip");
        }
    };

    const getStatusBadge = (status: string) => {
        switch (status) {
            case "paid":
                return <Badge className="bg-green-100 text-green-800 hover:bg-green-100">Paid</Badge>;
            case "generated":
            case "approved":
                return <Badge className="bg-blue-100 text-blue-800 hover:bg-blue-100">Generated</Badge>;
            case "draft":
                return <Badge className="bg-yellow-100 text-yellow-800 hover:bg-yellow-100">Draft</Badge>;
            default:
                return <Badge variant="outline">{status}</Badge>;
        }
    };

    const formatCurrency = (amount: number) => {
        // Ideally we fetch tenant currency setting, but defaulting to USD/locale for now
        return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(amount);
    };

    return (
        <div className="space-y-6">
            <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
                <div>
                    <h2 className="text-3xl font-bold tracking-tight">Payroll</h2>
                    <p className="text-muted-foreground">
                        Manage employee payslips and salary history.
                    </p>
                </div>
                <div className="flex gap-2">
                    <Button>
                        <Plus className="mr-2 h-4 w-4" />
                        Generate Payroll
                    </Button>
                </div>
            </div>

            {/* Filters */}
            <div className="flex gap-4 items-center bg-white dark:bg-slate-900 p-4 rounded-lg border border-gray-200 dark:border-slate-800">
                <div className="flex items-center gap-2">
                    <span className="text-sm font-medium">Period:</span>
                    <Select value={month} onValueChange={setMonth}>
                        <SelectTrigger className="w-[140px]">
                            <SelectValue placeholder="Month" />
                        </SelectTrigger>
                        <SelectContent>
                            {Array.from({ length: 12 }, (_, i) => i + 1).map((m) => (
                                <SelectItem key={m} value={String(m)}>
                                    {format(new Date(2000, m - 1, 1), "MMMM")}
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                    <Select value={year} onValueChange={setYear}>
                        <SelectTrigger className="w-[100px]">
                            <SelectValue placeholder="Year" />
                        </SelectTrigger>
                        <SelectContent>
                            {Array.from({ length: 5 }, (_, i) => new Date().getFullYear() - 2 + i).map((y) => (
                                <SelectItem key={y} value={String(y)}>
                                    {y}
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                </div>
            </div>

            {/* Table */}
            <div className="border rounded-md bg-white dark:bg-slate-900">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead>Employee</TableHead>
                            <TableHead>Role</TableHead>
                            <TableHead>Pay Period</TableHead>
                            <TableHead>Payment Date</TableHead>
                            <TableHead>Net Salary</TableHead>
                            <TableHead>Status</TableHead>
                            <TableHead className="text-right">Actions</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {loading ? (
                            <TableRow>
                                <TableCell colSpan={6} className="text-center h-24">
                                    Loading...
                                </TableCell>
                            </TableRow>
                        ) : payslips.length === 0 ? (
                            <TableRow>
                                <TableCell colSpan={6} className="text-center h-24 text-muted-foreground">
                                    No payslips found for this period.
                                </TableCell>
                            </TableRow>
                        ) : (
                            payslips.map((payslip) => (
                                <TableRow key={payslip.id}>
                                    <TableCell className="font-medium">
                                        {payslip.employee ? `${payslip.employee.first_name} ${payslip.employee.last_name}` : "Unknown"}
                                    </TableCell>
                                    <TableCell className="capitalize">
                                        {payslip.role || "-"}
                                    </TableCell>
                                    <TableCell>
                                        <div className="text-sm">
                                            {format(new Date(payslip.pay_period_start), "MMM d")} - {format(new Date(payslip.pay_period_end), "MMM d")}
                                        </div>
                                    </TableCell>
                                    <TableCell>
                                        {payslip.payment_date ? format(new Date(payslip.payment_date), "MMM d, yyyy") : "-"}
                                    </TableCell>
                                    <TableCell className="font-medium">
                                        {formatCurrency(payslip.net_salary)}
                                    </TableCell>
                                    <TableCell>
                                        {getStatusBadge(payslip.status)}
                                    </TableCell>
                                    <TableCell className="text-right">
                                        <div className="flex justify-end gap-2">
                                            <Button variant="ghost" size="sm" title="View/Download">
                                                <Download className="h-4 w-4 text-blue-600" />
                                            </Button>
                                            <Button
                                                variant="ghost"
                                                size="sm"
                                                className="text-red-600 hover:text-red-700 hover:bg-red-50"
                                                onClick={() => handleDelete(payslip.id)}
                                            >
                                                <Trash2 className="h-4 w-4" />
                                            </Button>
                                        </div>
                                    </TableCell>
                                </TableRow>
                            ))
                        )}
                    </TableBody>
                </Table>
            </div>
        </div>
    );
}
