"use client"

import { useState, useEffect } from "react"
import { format } from "date-fns"
import { Check, X, Download, Filter, Loader2 } from "lucide-react"

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
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"

export type LeaveRequest = {
    id: string
    leave_type_name: string
    start_date: string
    end_date: string
    reason: string
    status: string
    employee_name?: string
    days_count: number
    created_at: string
}

export default function HRLeavePage() {
    const [data, setData] = useState<LeaveRequest[]>([])
    const [loading, setLoading] = useState(true)
    const [statusFilter, setStatusFilter] = useState("all")
    const [processingId, setProcessingId] = useState<string | null>(null)
    const [rejectReason, setRejectReason] = useState("")
    const [rejectDialogOpen, setRejectDialogOpen] = useState(false)
    const [selectedRequest, setSelectedRequest] = useState<LeaveRequest | null>(null)

    const fetchData = async () => {
        setLoading(true)
        try {
            // Fetch all leaves for HR
            const response = await api.get(`/company/hr/leaves`)
            const requests = Array.isArray(response.data) ? response.data : (response.data.data || [])
            setData(requests)
        } catch (error) {
            console.error("Failed to fetch leaves:", error)
            toast.error("Failed to fetch leave requests")
        } finally {
            setLoading(false)
        }
    }

    useEffect(() => {
        fetchData()
    }, [])

    const handleApprove = async (id: string) => {
        if (!confirm("Are you sure you want to approve this request?")) return;

        setProcessingId(id)
        try {
            await api.put(`/company/hr/leaves/${id}/approve`)
            toast.success("Leave request approved")
            fetchData() // Refresh list
        } catch (error: any) {
            console.error("Failed to approve:", error)
            toast.error(error.response?.data?.message || "Failed to approve request")
        } finally {
            setProcessingId(null)
        }
    }

    const handleReject = async () => {
        if (!selectedRequest || !rejectReason) return;

        setProcessingId(selectedRequest.id)
        try {
            await api.put(`/company/hr/leaves/${selectedRequest.id}/reject`, { rejection_reason: rejectReason })
            toast.success("Leave request rejected")
            setRejectDialogOpen(false)
            setRejectReason("")
            setSelectedRequest(null)
            fetchData() // Refresh list
        } catch (error: any) {
            console.error("Failed to reject:", error)
            toast.error(error.response?.data?.message || "Failed to reject request")
        } finally {
            setProcessingId(null)
        }
    }

    const filteredData = data.filter(item => {
        if (statusFilter !== "all" && item.status !== statusFilter) return false
        return true
    })

    const handleExport = () => {
        if (!filteredData.length) {
            toast.error("No data to export")
            return
        }

        const headers = ["Employee", "Type", "Start Date", "End Date", "Days", "Reason", "Status", "Requested On"]
        const csvContent = [
            headers.join(","),
            ...filteredData.map(row => [
                `"${row.employee_name || 'N/A'}"`,
                `"${row.leave_type_name}"`,
                format(new Date(row.start_date), "yyyy-MM-dd"),
                format(new Date(row.end_date), "yyyy-MM-dd"),
                row.days_count,
                `"${row.reason}"`,
                row.status,
                format(new Date(row.created_at), "yyyy-MM-dd HH:mm")
            ].join(","))
        ].join("\n")

        const blob = new Blob([csvContent], { type: "text/csv;charset=utf-8;" })
        const url = URL.createObjectURL(blob)
        const link = document.createElement("a")
        link.setAttribute("href", url)
        link.setAttribute("download", `leave_requests_${format(new Date(), "yyyy-MM-dd")}.csv`)
        link.style.visibility = "hidden"
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
    }

    const columns: ColumnDef<LeaveRequest>[] = [
        {
            accessorKey: "employee_name",
            header: "Employee",
            cell: ({ row }) => <span className="font-medium">{row.getValue("employee_name") || "Unknown"}</span>
        },
        {
            accessorKey: "leave_type_name",
            header: "Type",
        },
        {
            accessorKey: "start_date",
            header: "Duration",
            cell: ({ row }) => (
                <div className="text-sm">
                    <div>{format(new Date(row.getValue("start_date")), "MMM dd")} - {format(new Date(row.original.end_date), "MMM dd, yyyy")}</div>
                    <div className="text-muted-foreground text-xs">{row.original.days_count} days</div>
                </div>
            ),
        },
        {
            accessorKey: "reason",
            header: "Reason",
            cell: ({ row }) => <div className="max-w-[200px] truncate" title={row.getValue("reason")}>{row.getValue("reason")}</div>
        },
        {
            accessorKey: "status",
            header: "Status",
            cell: ({ row }) => {
                const status = row.getValue("status") as string
                let variant: "default" | "secondary" | "destructive" | "outline" = "default"

                switch (status) {
                    case 'approved': variant = 'default'; break; // green-ish usually
                    case 'rejected': variant = 'destructive'; break;
                    case 'pending': variant = 'secondary'; break; // yellow/gray
                    default: variant = 'outline';
                }

                // Custom coloring for badges if generic variants aren't enough
                let className = "capitalize"
                if (status === 'approved') className += " bg-green-100 text-green-800 hover:bg-green-100 border-green-200"
                if (status === 'pending') className += " bg-yellow-100 text-yellow-800 hover:bg-yellow-100 border-yellow-200"
                if (status === 'rejected') className += " bg-red-100 text-red-800 hover:bg-red-100 border-red-200"

                return (
                    <Badge variant="outline" className={className}>
                        {status}
                    </Badge>
                )
            },
        },
        {
            id: "actions",
            header: "Actions",
            cell: ({ row }) => {
                const request = row.original
                if (request.status !== "pending") return null

                return (
                    <div className="flex items-center gap-2">
                        <Button
                            size="sm"
                            variant="outline"
                            className="h-8 w-8 p-0 text-green-600 hover:text-green-700 hover:bg-green-50 border-green-200"
                            onClick={() => handleApprove(request.id)}
                            disabled={processingId === request.id}
                            title="Approve"
                        >
                            {processingId === request.id ? <Loader2 className="h-4 w-4 animate-spin" /> : <Check className="h-4 w-4" />}
                        </Button>
                        <Button
                            size="sm"
                            variant="outline"
                            className="h-8 w-8 p-0 text-red-600 hover:text-red-700 hover:bg-red-50 border-red-200"
                            onClick={() => {
                                setSelectedRequest(request)
                                setRejectDialogOpen(true)
                            }}
                            disabled={processingId === request.id}
                            title="Reject"
                        >
                            <X className="h-4 w-4" />
                        </Button>
                    </div>
                )
            }
        }
    ]

    return (
        <div className="container mx-auto py-10">
            {/* Reject Dialog */}
            <Dialog open={rejectDialogOpen} onOpenChange={setRejectDialogOpen}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Reject Leave Request</DialogTitle>
                        <DialogDescription>
                            Please provide a reason for rejecting this leave request.
                        </DialogDescription>
                    </DialogHeader>
                    <div className="grid gap-4 py-4">
                        <div className="grid gap-2">
                            <Label htmlFor="reason">Reason</Label>
                            <Textarea
                                id="reason"
                                placeholder="E.g., High workload during this period..."
                                value={rejectReason}
                                onChange={(e) => setRejectReason(e.target.value)}
                            />
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setRejectDialogOpen(false)}>Cancel</Button>
                        <Button variant="destructive" onClick={handleReject} disabled={!rejectReason || !!processingId}>
                            {processingId ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
                            Reject Request
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 mb-8">
                <div>
                    <h2 className="text-3xl font-bold tracking-tight">Leave Management</h2>
                    <p className="text-muted-foreground">Review and manage employee leave requests.</p>
                </div>
                <div className="flex items-center gap-2">
                    <Button variant="outline" onClick={handleExport}>
                        <Download className="mr-2 h-4 w-4" />
                        Export CSV
                    </Button>
                </div>
            </div>

            <div className="flex gap-4 mb-6">
                <Select value={statusFilter} onValueChange={setStatusFilter}>
                    <SelectTrigger className="w-[180px]">
                        <SelectValue placeholder="Filter by status" />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="all">All Statuses</SelectItem>
                        <SelectItem value="pending">Pending</SelectItem>
                        <SelectItem value="approved">Approved</SelectItem>
                        <SelectItem value="rejected">Rejected</SelectItem>
                    </SelectContent>
                </Select>
            </div>

            <div className="bg-white dark:bg-zinc-950 rounded-md border">
                {loading ? (
                    <div className="p-8 text-center">Loading leave requests...</div>
                ) : (
                    <DataTable columns={columns} data={filteredData} />
                )}
            </div>
        </div>
    )
}
