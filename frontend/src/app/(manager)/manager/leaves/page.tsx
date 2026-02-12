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

export default function ManagerLeavePage() {
    const [data, setData] = useState<LeaveRequest[]>([])
    const [loading, setLoading] = useState(true)
    const [statusFilter, setStatusFilter] = useState("all")
    const [processingId, setProcessingId] = useState<string | null>(null)
    const [rejectReason, setRejectReason] = useState("")
    const [rejectDialogOpen, setRejectDialogOpen] = useState(false)
    const [selectedRequest, setSelectedRequest] = useState<LeaveRequest | null>(null)

    const [error, setError] = useState<string | null>(null)

    const fetchData = async () => {
        setLoading(true)
        setError(null)
        try {
            // Fetch department leaves for Manager
            const response = await api.get(`/company/manager/leaves`)
            const requests = Array.isArray(response.data) ? response.data : (response.data.data || [])
            setData(requests)
        } catch (error: any) {
            console.error("Failed to fetch leaves:", error)
            const errorMsg = error.response?.data || error.message || "Failed to fetch leave requests"

            if (typeof errorMsg === 'string' && errorMsg.includes("User does not belong to a department")) {
                setError("NoDepartment")
            } else {
                toast.error(typeof errorMsg === 'string' ? errorMsg : "Failed to fetch leave requests")
            }
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
            await api.put(`/company/manager/leaves/${id}/approve`)
            toast.success("Leave request approved")
            fetchData()
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
            await api.put(`/company/manager/leaves/${selectedRequest.id}/reject`, { rejection_reason: rejectReason })
            toast.success("Leave request rejected")
            setRejectDialogOpen(false)
            setRejectReason("")
            setSelectedRequest(null)
            fetchData()
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

    const columns: ColumnDef<LeaveRequest>[] = [
        {
            accessorKey: "employee_name",
            header: "Team Member",
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
                    case 'approved': variant = 'default'; break;
                    case 'rejected': variant = 'destructive'; break;
                    case 'pending': variant = 'secondary'; break;
                    default: variant = 'outline';
                }

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

    if (error === "NoDepartment") {
        return (
            <div className="container mx-auto py-10">
                <h2 className="text-3xl font-bold tracking-tight mb-4">Leave Approvals</h2>
                <div className="flex flex-col items-center justify-center p-12 border border-dashed rounded-lg bg-muted/50">
                    <div className="h-12 w-12 rounded-full bg-yellow-100 dark:bg-yellow-900/20 flex items-center justify-center mb-4">
                        <Filter className="h-6 w-6 text-yellow-600 dark:text-yellow-400" />
                    </div>
                    <h3 className="text-lg font-semibold mb-2">No Department Assigned</h3>
                    <p className="text-muted-foreground text-center max-w-sm">
                        You do not currently belong to any department. Please contact your administrator to be assigned to a department to view leave requests.
                    </p>
                </div>
            </div>
        )
    }

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
                                placeholder="E.g., High workload..."
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
                    <h2 className="text-3xl font-bold tracking-tight">Leave Approvals</h2>
                    <p className="text-muted-foreground">Review and manage accumulated leave requests.</p>
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
