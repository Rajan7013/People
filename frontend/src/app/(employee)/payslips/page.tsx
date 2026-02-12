"use client"

import { useState, useEffect } from "react"
import { format } from "date-fns"
import { Eye, DollarSign, Loader2 } from "lucide-react"

import { Button } from "@/components/ui/button"
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
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog"

export type Payslip = {
    id: string
    pay_period_start: string
    pay_period_end: string
    payment_date: string
    net_salary: number
    gross_salary: number
    status: string
    components?: any[]
    total_deductions?: number
    basic_salary?: number
}

export default function EmployeePayslipPage() {
    const [data, setData] = useState<Payslip[]>([])
    const [loading, setLoading] = useState(true)
    const [yearFilter, setYearFilter] = useState(String(new Date().getFullYear()))

    // View Details State
    const [selectedPayslip, setSelectedPayslip] = useState<Payslip | null>(null)
    const [detailsOpen, setDetailsOpen] = useState(false)
    const [detailsLoading, setDetailsLoading] = useState(false)

    const fetchData = async () => {
        setLoading(true)
        try {
            // Employee endpoint: /company/employee/payslips
            const response = await api.get(`/company/employee/payslips?year=${yearFilter}`)
            const payslips = response.data.payslips || []
            setData(payslips)
        } catch (error) {
            console.error("Failed to fetch payslips:", error)
            toast.error("Failed to fetch payslips")
            setData([]) // Set empty array on error
        } finally {
            setLoading(false)
        }
    }

    useEffect(() => {
        fetchData()
    }, [yearFilter])

    const handleViewDetails = async (id: string) => {
        setDetailsOpen(true)
        setDetailsLoading(true)
        try {
            const res = await api.get(`/company/employee/payslips/${id}`)
            setSelectedPayslip(res.data)
        } catch (error) {
            toast.error("Failed to load details")
            setDetailsOpen(false)
        } finally {
            setDetailsLoading(false)
        }
    }

    const columns: ColumnDef<Payslip>[] = [
        {
            accessorKey: "pay_period_start",
            header: "Month",
            cell: ({ row }) => format(new Date(row.getValue("pay_period_start")), "MMMM yyyy"),
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
                return <span className="font-medium text-green-600 dark:text-green-400">{new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(val)}</span>
            }
        },
        {
            accessorKey: "payment_date",
            header: "Paid On",
            cell: ({ row }) => {
                const val = row.getValue("payment_date") as string
                return val ? format(new Date(val), "MMM dd, yyyy") : "-"
            }
        },
        {
            accessorKey: "status",
            header: "Status",
            cell: ({ row }) => {
                const status = row.getValue("status") as string
                let className = "capitalize"
                if (status === 'paid') className += " bg-green-100 text-green-800 hover:bg-green-100 border-green-200"
                else if (status === 'generated') className += " bg-blue-100 text-blue-800 hover:bg-blue-100 border-blue-200"
                else className += " bg-gray-100 text-gray-800 hover:bg-gray-100 border-gray-200"

                return <Badge variant="outline" className={className}>{status}</Badge>
            }
        },
        {
            id: "actions",
            cell: ({ row }) => (
                <Button variant="ghost" size="sm" onClick={() => handleViewDetails(row.original.id)}>
                    <Eye className="h-4 w-4 mr-2" /> View
                </Button>
            )
        }
    ]

    return (
        <div className="container mx-auto py-8">
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
                                        {format(new Date(selectedPayslip.pay_period_start), "MMMM yyyy")}
                                    </h3>
                                    <p className="text-sm text-muted-foreground">
                                        Period: {format(new Date(selectedPayslip.pay_period_start), "MMM dd")} - {format(new Date(selectedPayslip.pay_period_end), "MMM dd, yyyy")}
                                    </p>
                                </div>
                                <div className="text-right">
                                    <Badge variant="outline">{selectedPayslip.status.toUpperCase()}</Badge>
                                </div>
                            </div>

                            <div className="grid grid-cols-2 gap-8">
                                <div>
                                    <h4 className="font-medium mb-3 text-green-600">Earnings</h4>
                                    <div className="space-y-2">
                                        <div className="flex justify-between text-sm">
                                            <span>Basic Salary</span>
                                            <span>{new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(selectedPayslip.basic_salary || 0)}</span>
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
                                        {(!selectedPayslip.components || selectedPayslip.components.filter((c: any) => c.component_type === 'deduction').length === 0) && (
                                            <div className="text-sm text-muted-foreground">No deductions</div>
                                        )}
                                        <div className="flex justify-between font-semibold border-t pt-2 mt-2">
                                            <span>Total Deductions</span>
                                            <span>{new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(selectedPayslip.total_deductions || 0)}</span>
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

            <div className="flex justify-between items-center mb-8">
                <div>
                    <h2 className="text-3xl font-bold tracking-tight">My Payslips</h2>
                    <p className="text-muted-foreground">View your salary history and payment details.</p>
                </div>
                <Select value={yearFilter} onValueChange={setYearFilter}>
                    <SelectTrigger className="w-[120px]">
                        <SelectValue placeholder="Year" />
                    </SelectTrigger>
                    <SelectContent>
                        {[2024, 2025, 2026].map(y => (
                            <SelectItem key={y} value={String(y)}>{y}</SelectItem>
                        ))}
                    </SelectContent>
                </Select>
            </div>

            <div className="bg-white dark:bg-zinc-950 rounded-md border">
                {loading ? (
                    <div className="p-8 text-center">
                        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 dark:border-gray-100 mx-auto mb-4"></div>
                        <p className="text-muted-foreground">Loading payslips...</p>
                    </div>
                ) : data.length === 0 ? (
                    <div className="p-12 text-center">
                        <DollarSign className="mx-auto h-12 w-12 text-gray-400 mb-4" />
                        <h3 className="text-lg font-semibold mb-2">No Payslips Available</h3>
                        <p className="text-muted-foreground mb-4">
                            There are no payslips for the selected year.
                        </p>
                        <p className="text-sm text-muted-foreground">
                            Payslips will appear here once they are generated by your organization.
                        </p>
                    </div>
                ) : (
                    <DataTable columns={columns} data={data} />
                )}
            </div>
        </div>
    )
}
