"use client"

import { useState, useEffect } from "react"
import { format } from "date-fns"
import { Plus, Download, Loader2, Calendar as CalendarIcon } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
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

export type LeaveRequest = {
    id: string
    leave_type_name: string
    start_date: string
    end_date: string
    reason: string
    status: string
    days_count: number
    created_at: string
}

export type LeaveType = {
    id: string
    name: string
    days_allowed: number
}

const columns: ColumnDef<LeaveRequest>[] = [
    {
        accessorKey: "leave_type_name",
        header: "Type",
    },
    {
        accessorKey: "start_date",
        header: "Date Range",
        cell: ({ row }) => (
            <div className="text-sm">
                <span>{format(new Date(row.getValue("start_date")), "MMM dd, yyyy")}</span>
                <span className="mx-2">-</span>
                <span>{format(new Date(row.original.end_date), "MMM dd, yyyy")}</span>
            </div>
        ),
    },
    {
        accessorKey: "days_count",
        header: "Days",
        cell: ({ row }) => <span className="text-center block w-10">{row.getValue("days_count")}</span>
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
]

export default function EmployeeLeavePage() {
    const [data, setData] = useState<LeaveRequest[]>([])
    const [loading, setLoading] = useState(true)

    // Create Modal State
    const [createOpen, setCreateOpen] = useState(false)
    const [creating, setCreating] = useState(false)
    const [leaveTypes, setLeaveTypes] = useState<LeaveType[]>([])
    const [newRequest, setNewRequest] = useState({
        leave_type_id: "",
        start_date: "",
        end_date: "",
        reason: ""
    })

    const fetchData = async () => {
        setLoading(true)
        try {
            const response = await api.get(`/company/employee/leaves`)
            const requests = Array.isArray(response.data) ? response.data : (response.data.data || [])
            setData(requests)
        } catch (error) {
            console.error("Failed to fetch leaves:", error)
            toast.error("Failed to fetch leave history")
        } finally {
            setLoading(false)
        }
    }

    const fetchLeaveTypes = async () => {
        try {
            const res = await api.get("/company/employee/leaves/types")
            setLeaveTypes(res.data || [])
        } catch (error) {
            console.error("Failed to fetch leave types", error)
        }
    }

    useEffect(() => {
        fetchData()
        fetchLeaveTypes()
    }, [])

    const handleCreate = async () => {
        if (!newRequest.leave_type_id || !newRequest.start_date || !newRequest.end_date) {
            toast.error("Please fill in all required fields")
            return
        }

        setCreating(true)
        try {
            await api.post("/company/employee/leaves", {
                leave_type_id: newRequest.leave_type_id,
                start_date: new Date(newRequest.start_date).toISOString(),
                end_date: new Date(newRequest.end_date).toISOString(),
                reason: newRequest.reason
            })
            toast.success("Leave requested successfully")
            setCreateOpen(false)
            setNewRequest({ leave_type_id: "", start_date: "", end_date: "", reason: "" })
            fetchData()
        } catch (error: any) {
            console.error("Failed to create request:", error)
            toast.error(error.response?.data?.message || "Failed to submit request")
        } finally {
            setCreating(false)
        }
    }

    return (
        <div className="container mx-auto py-8">
            <Dialog open={createOpen} onOpenChange={setCreateOpen}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Request Leave</DialogTitle>
                        <DialogDescription>Submit a new leave request for approval.</DialogDescription>
                    </DialogHeader>
                    <div className="grid gap-4 py-4">
                        <div className="grid gap-2">
                            <Label>Leave Type</Label>
                            <Select
                                value={newRequest.leave_type_id}
                                onValueChange={(val) => setNewRequest({ ...newRequest, leave_type_id: val })}
                            >
                                <SelectTrigger>
                                    <SelectValue placeholder="Select Type" />
                                </SelectTrigger>
                                <SelectContent>
                                    {leaveTypes.map(lt => (
                                        <SelectItem key={lt.id} value={lt.id}>{lt.name}</SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                        <div className="grid grid-cols-2 gap-4">
                            <div className="grid gap-2">
                                <Label>Start Date</Label>
                                <Input type="date" value={newRequest.start_date} onChange={(e) => setNewRequest({ ...newRequest, start_date: e.target.value })} />
                            </div>
                            <div className="grid gap-2">
                                <Label>End Date</Label>
                                <Input type="date" value={newRequest.end_date} onChange={(e) => setNewRequest({ ...newRequest, end_date: e.target.value })} />
                            </div>
                        </div>
                        <div className="grid gap-2">
                            <Label>Reason</Label>
                            <Textarea
                                placeholder="Optional reason for leave..."
                                value={newRequest.reason}
                                onChange={(e) => setNewRequest({ ...newRequest, reason: e.target.value })}
                            />
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setCreateOpen(false)}>Cancel</Button>
                        <Button onClick={handleCreate} disabled={creating}>
                            {creating && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                            Submit Request
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 mb-8">
                <div>
                    <h2 className="text-3xl font-bold tracking-tight">My Leaves</h2>
                    <p className="text-muted-foreground">View your leave history and track status.</p>
                </div>
                <Button onClick={() => setCreateOpen(true)}>
                    <Plus className="mr-2 h-4 w-4" />
                    Request Leave
                </Button>
            </div>

            <div className="bg-white dark:bg-zinc-950 rounded-md border">
                {loading ? (
                    <div className="p-8 text-center">Loading leave history...</div>
                ) : (
                    <DataTable columns={columns} data={data} />
                )}
            </div>
        </div>
    )
}
