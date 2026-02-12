"use client"

import { useState, useEffect } from "react"
import { format, subMonths } from "date-fns"
import { Download, Plus, FileText, Loader2, Search } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select"
import { DataTable } from "@/components/ui/data-table"
import { Badge } from "@/components/ui/badge"
import { ColumnDef } from "@tanstack/react-table"
import api from "@/lib/api"
import { toast } from "sonner"
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog"
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card"

export type Payslip = {
    id: string
    employee_id: string
    employee_name?: string // May need to map this if backend doesn't send it directly
    pay_period_start: string
    pay_period_end: string
    net_salary: number
    gross_salary: number
    status: string
    payment_date?: string
    employee?: {
        first_name: string
        last_name: string
    }
}

export default function HRPayrollPage() {
    const [data, setData] = useState<Payslip[]>([])
    const [loading, setLoading] = useState(true)
    const [filterMonth, setFilterMonth] = useState<string>(String(new Date().getMonth() + 1))
    const [filterYear, setFilterYear] = useState<string>(String(new Date().getFullYear()))

    // Create Modal State
    const [createOpen, setCreateOpen] = useState(false)
    const [creating, setCreating] = useState(false)
    const [employees, setEmployees] = useState<{ id: string, name: string }[]>([])
    const [newPayslip, setNewPayslip] = useState({
        employee_id: "",
        start_date: format(subMonths(new Date(), 1), "yyyy-MM-01"), // Default to last month
        end_date: format(subMonths(new Date(), 1), "yyyy-MM-30"), // Approx end
        payment_date: format(new Date(), "yyyy-MM-dd")
    })

    // View Details State
    const [selectedPayslip, setSelectedPayslip] = useState<any | null>(null)
    const [detailsOpen, setDetailsOpen] = useState(false)
    const [detailsLoading, setDetailsLoading] = useState(false)

    const fetchPayslips = async () => {
        setLoading(true)
        try {
            // Fetch payslips with month/year filter
            const response = await api.get(`/company/hr/payslips?month=${filterMonth}&year=${filterYear}&limit=100`)
            const payslips = response.data.payslips || []

            // Normalize data (extract employee name if nested)
            const normalized = payslips.map((p: any) => ({
                ...p,
                employee_name: p.employee ? `${p.employee.first_name} ${p.employee.last_name}` : "Unknown"
            }))
            setData(normalized)
        } catch (error) {
            console.error("Failed to fetch payslips:", error)
            toast.error("Failed to load payslips")
        } finally {
            setLoading(false)
        }
    }

    const fetchEmployees = async () => {
        try {
            const res = await api.get("/company/hr/employees")
            const emps = (res.data || []).map((e: any) => ({
                id: e.id,
                name: `${e.first_name} ${e.last_name}`
            }))
            setEmployees(emps)
        } catch (error) {
            console.error("Failed to fetch employees", error)
        }
    }

    useEffect(() => {
        fetchPayslips()
        fetchEmployees()
    }, [filterMonth, filterYear])

    const handleCreate = async () => {
        if (!newPayslip.employee_id) {
            toast.error("Please select an employee")
            return
        }
        setCreating(true)
        try {
            await api.post("/company/hr/payslips", {
                employee_id: newPayslip.employee_id,
                pay_period_start: new Date(newPayslip.start_date).toISOString(),
                pay_period_end: new Date(newPayslip.end_date).toISOString(),
                payment_date: new Date(newPayslip.payment_date).toISOString()
            })
            toast.success("Payslip created successfully")
            setCreateOpen(false)
            fetchPayslips()
        } catch (error: any) {
            console.error("Failed to create payslip:", error)
            toast.error(error.response?.data?.message || "Failed to create payslip")
        } finally {
            setCreating(false)
        }
    }

    const handleViewDetails = async (id: string) => {
        setDetailsOpen(true)
        setDetailsLoading(true)
        try {
            const res = await api.get(`/company/hr/payslips/${id}`)
            setSelectedPayslip(res.data)
        } catch (error) {
            toast.error("Failed to load details")
            setDetailsOpen(false)
        } finally {
            setDetailsLoading(false)
        }
    }

    const handleExport = () => {
        if (!data.length) {
            toast.error("No data to export")
            return
        }
        const headers = ["Employee", "Period Start", "Period End", "Gross Salary", "Net Salary", "Status", "Payment Date"]
        const csvContent = [
            headers.join(","),
            ...data.map(row => [
                `"${row.employee_name}"`,
                format(new Date(row.pay_period_start), "yyyy-MM-dd"),
                format(new Date(row.pay_period_end), "yyyy-MM-dd"),
                row.gross_salary,
                row.net_salary,
                row.status,
                row.payment_date ? format(new Date(row.payment_date), "yyyy-MM-dd") : ""
            ].join(","))
        ].join("\n")

        const blob = new Blob([csvContent], { type: "text/csv;charset=utf-8;" })
        const url = URL.createObjectURL(blob)
        const link = document.createElement("a")
        link.setAttribute("href", url)
        link.setAttribute("download", `payroll_${filterYear}_${filterMonth}.csv`)
        link.style.visibility = "hidden"
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
    }

    const columns: ColumnDef<Payslip>[] = [
        {
            accessorKey: "employee_name",
            header: "Employee",
            cell: ({ row }) => <span className="font-medium">{row.original.employee_name}</span>
        },
        {
            accessorKey: "pay_period_start",
            header: "Pay Period",
            cell: ({ row }) => (
                <div className="text-sm">
                    {format(new Date(row.getValue("pay_period_start")), "MMM dd")} - {format(new Date(row.original.pay_period_end), "MMM dd, yyyy")}
                </div>
            )
        },
        {
            accessorKey: "gross_salary",
            header: "Gross",
            cell: ({ row }) => {
                const val = row.getValue("gross_salary") as number
                return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(val)
            }
        },
        {
            accessorKey: "net_salary",
            header: "Net Pay",
            cell: ({ row }) => {
                const val = row.getValue("net_salary") as number
                return <span className="font-medium">{new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(val)}</span>
            }
        },
        {
            accessorKey: "status",
            header: "Status",
            cell: ({ row }) => {
                const status = row.getValue("status") as string
                let variant: "default" | "secondary" | "outline" | "destructive" = "outline"
                if (status === 'paid') variant = "default"
                if (status === 'draft') variant = "secondary"

                return <Badge variant={variant} className="capitalize">{status}</Badge>
            }
        },
        {
            id: "actions",
            cell: ({ row }) => (
                <Button variant="ghost" size="sm" onClick={() => handleViewDetails(row.original.id)}>
                    View
                </Button>
            )
        }
    ]

    return (
        <div className="container mx-auto py-10">
            {/* Creates Payslip Dialog */}
            <Dialog open={createOpen} onOpenChange={setCreateOpen}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Create Payslip</DialogTitle>
                        <DialogDescription>Generate a new payslip for an employee. Salary will be calculated automatically based on defined structure.</DialogDescription>
                    </DialogHeader>
                    <div className="grid gap-4 py-4">
                        <div className="grid gap-2">
                            <Label>Employee</Label>
                            <Select
                                value={newPayslip.employee_id}
                                onValueChange={(val) => setNewPayslip({ ...newPayslip, employee_id: val })}
                            >
                                <SelectTrigger>
                                    <SelectValue placeholder="Select Employee" />
                                </SelectTrigger>
                                <SelectContent>
                                    {employees.map(e => (
                                        <SelectItem key={e.id} value={e.id}>{e.name}</SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                        <div className="grid grid-cols-2 gap-4">
                            <div className="grid gap-2">
                                <Label>Start Date</Label>
                                <Input type="date" value={newPayslip.start_date} onChange={(e) => setNewPayslip({ ...newPayslip, start_date: e.target.value })} />
                            </div>
                            <div className="grid gap-2">
                                <Label>End Date</Label>
                                <Input type="date" value={newPayslip.end_date} onChange={(e) => setNewPayslip({ ...newPayslip, end_date: e.target.value })} />
                            </div>
                        </div>
                        <div className="grid gap-2">
                            <Label>Payment Date</Label>
                            <Input type="date" value={newPayslip.payment_date} onChange={(e) => setNewPayslip({ ...newPayslip, payment_date: e.target.value })} />
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setCreateOpen(false)}>Cancel</Button>
                        <Button onClick={handleCreate} disabled={creating}>
                            {creating && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                            Generate
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            {/* View Details Dialog */}
            <Dialog open={detailsOpen} onOpenChange={setDetailsOpen}>
                <DialogContent className="max-w-2xl">
                    <DialogHeader>
                        <DialogTitle>Payslip Details</DialogTitle>
                    </DialogHeader>
                    {detailsLoading || !selectedPayslip ? (
                        <div className="flex justify-center p-8"><Loader2 className="animate-spin" /></div>
                    ) : (
                        <div className="space-y-6">
                            <div className="flex justify-between border-b pb-4">
                                <div>
                                    <h3 className="font-semibold text-lg">
                                        {selectedPayslip.employee?.first_name} {selectedPayslip.employee?.last_name}
                                    </h3>
                                    <p className="text-sm text-muted-foreground">{selectedPayslip.employee?.email}</p>
                                </div>
                                <div className="text-right">
                                    <Badge variant="outline">{selectedPayslip.status.toUpperCase()}</Badge>
                                    <p className="text-sm text-muted-foreground mt-1">
                                        Pay Period: {format(new Date(selectedPayslip.pay_period_start), "MMM dd")} - {format(new Date(selectedPayslip.pay_period_end), "MMM dd, yyyy")}
                                    </p>
                                </div>
                            </div>

                            <div className="grid grid-cols-2 gap-8">
                                <div>
                                    <h4 className="font-medium mb-3 text-green-600">Earnings</h4>
                                    <div className="space-y-2">
                                        <div className="flex justify-between text-sm">
                                            <span>Basic Salary</span>
                                            <span>{new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(selectedPayslip.basic_salary)}</span>
                                        </div>
                                        {selectedPayslip.components?.filter((c: any) => c.component_type === 'earning').map((c: any) => (
                                            <div key={c.id} className="flex justify-between text-sm">
                                                <span>{c.component_name}</span>
                                                <span>{new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(c.amount)}</span>
                                            </div>
                                        ))}
                                        <div className="flex justify-between font-semibold border-t pt-2 mt-2">
                                            <span>Gross Salary</span>
                                            <span>{new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(selectedPayslip.gross_salary)}</span>
                                        </div>
                                    </div>
                                </div>
                                <div>
                                    <h4 className="font-medium mb-3 text-red-600">Deductions</h4>
                                    <div className="space-y-2">
                                        {selectedPayslip.components?.filter((c: any) => c.component_type === 'deduction').map((c: any) => (
                                            <div key={c.id} className="flex justify-between text-sm">
                                                <span>{c.component_name}</span>
                                                <span>{new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(c.amount)}</span>
                                            </div>
                                        ))}
                                        <div className="flex justify-between font-semibold border-t pt-2 mt-2">
                                            <span>Total Deductions</span>
                                            <span>{new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(selectedPayslip.total_deductions)}</span>
                                        </div>
                                    </div>
                                </div>
                            </div>

                            <div className="bg-muted p-4 rounded-lg flex justify-between items-center">
                                <span className="font-bold text-lg">Net Payable</span>
                                <span className="font-bold text-xl">{new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(selectedPayslip.net_salary)}</span>
                            </div>
                        </div>
                    )}
                </DialogContent>
            </Dialog>

            <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 mb-8">
                <div>
                    <h2 className="text-3xl font-bold tracking-tight">Payroll Management</h2>
                    <p className="text-muted-foreground">Manage employee payslips and salary history.</p>
                </div>
                <div className="flex items-center gap-2">
                    <Button onClick={() => setCreateOpen(true)}>
                        <Plus className="mr-2 h-4 w-4" />
                        New Payslip
                    </Button>
                    <Button variant="outline" onClick={handleExport}>
                        <Download className="mr-2 h-4 w-4" />
                        Export
                    </Button>
                </div>
            </div>

            <div className="flex gap-4 mb-6">
                <Select value={filterMonth} onValueChange={setFilterMonth}>
                    <SelectTrigger className="w-[150px]">
                        <SelectValue placeholder="Month" />
                    </SelectTrigger>
                    <SelectContent>
                        {Array.from({ length: 12 }, (_, i) => i + 1).map(m => (
                            <SelectItem key={m} value={String(m)}>{format(new Date(2000, m - 1, 1), "MMMM")}</SelectItem>
                        ))}
                    </SelectContent>
                </Select>
                <Select value={filterYear} onValueChange={setFilterYear}>
                    <SelectTrigger className="w-[120px]">
                        <SelectValue placeholder="Year" />
                    </SelectTrigger>
                    <SelectContent>
                        {[2024, 2025, 2026, 2027].map(y => (
                            <SelectItem key={y} value={String(y)}>{y}</SelectItem>
                        ))}
                    </SelectContent>
                </Select>
            </div>

            <div className="bg-white dark:bg-zinc-950 rounded-md border">
                {loading ? (
                    <div className="p-8 text-center">Loading payroll data...</div>
                ) : (
                    <DataTable columns={columns} data={data} />
                )}
            </div>
        </div>
    )
}
