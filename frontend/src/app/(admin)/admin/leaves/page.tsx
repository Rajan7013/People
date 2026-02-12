"use client";

import { useState, useEffect } from "react";
import { format } from "date-fns";
import { Check, X, FileText, Filter, Search } from "lucide-react";
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
import { Input } from "@/components/ui/input";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import api from "@/lib/api";

interface LeaveRequest {
    id: string;
    employee_id: string;
    employee_name: string;
    employee_code?: string;
    role?: string;
    leave_type: string;
    start_date: string;
    end_date: string;
    days_requested: number;
    reason: string;
    status: "pending" | "approved" | "rejected" | "cancelled";
    created_at: string;
}

export default function AdminLeavesPage() {
    const [leaves, setLeaves] = useState<LeaveRequest[]>([]);
    const [loading, setLoading] = useState(true);
    const [activeTab, setActiveTab] = useState("pending");
    const [rejectReason, setRejectReason] = useState("");
    const [selectedLeaveId, setSelectedLeaveId] = useState<string | null>(null);
    const [isRejectDialogOpen, setIsRejectDialogOpen] = useState(false);

    const fetchLeaves = async () => {
        try {
            setLoading(true);
            const response = await api.get("/company/hr/leaves");
            // The handler is typically mounted at /api/v1/leaves or /api/v1/company/hr/leaves depending on role
            // Based on server.go it is /api/v1/company/hr/leaves for HR (and inherited by Admin?)
            // Actually server.go says:
            // r.Route("/leaves", func(r chi.Router) { ... h.GetLeaveRequests ... }) inside HR group
            // So it's /api/v1/company/hr/leaves
            setLeaves(response.data || []);
        } catch (error) {
            console.error("Failed to fetch leaves:", error);
            toast.error("Failed to load leave requests");
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchLeaves();
    }, []);

    const handleApprove = async (id: string) => {
        try {
            await api.put(`/company/hr/leaves/${id}/approve`);
            toast.success("Leave approved");
            fetchLeaves();
        } catch (error) {
            console.error("Failed to approve leave:", error);
            toast.error("Failed to approve leave");
        }
    };

    const handleReject = async () => {
        if (!selectedLeaveId) return;
        try {
            await api.put(`/company/hr/leaves/${selectedLeaveId}/reject`, {
                rejection_reason: rejectReason
            });
            toast.success("Leave rejected");
            setIsRejectDialogOpen(false);
            setRejectReason("");
            fetchLeaves();
        } catch (error) {
            console.error("Failed to reject leave:", error);
            toast.error("Failed to reject leave");
        }
    };

    const openRejectDialog = (id: string) => {
        setSelectedLeaveId(id);
        setIsRejectDialogOpen(true);
    };

    const [searchQuery, setSearchQuery] = useState("");

    const filteredLeaves = leaves.filter(leave => {
        const matchesStatus = activeTab === "all" ? true : leave.status === activeTab;
        const matchesSearch = searchQuery === "" ||
            leave.employee_name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
            leave.employee_code?.toLowerCase().includes(searchQuery.toLowerCase());

        return matchesStatus && matchesSearch;
    });

    const getStatusBadge = (status: string) => {
        switch (status) {
            case "approved":
                return <Badge className="bg-green-100 text-green-800 hover:bg-green-100">Approved</Badge>;
            case "rejected":
                return <Badge className="bg-red-100 text-red-800 hover:bg-red-100">Rejected</Badge>;
            case "pending":
                return <Badge className="bg-yellow-100 text-yellow-800 hover:bg-yellow-100">Pending</Badge>;
            case "cancelled":
                return <Badge variant="outline">Cancelled</Badge>;
            default:
                return <Badge variant="outline">{status}</Badge>;
        }
    };

    return (
        <div className="space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h2 className="text-3xl font-bold tracking-tight">Leaves</h2>
                    <p className="text-muted-foreground">
                        Manage employee leave requests.
                    </p>
                </div>
            </div>

            <div className="flex flex-col sm:flex-row justify-between items-center bg-white dark:bg-slate-900 p-4 rounded-lg border border-gray-200 dark:border-slate-800 shadow-sm mb-4">
                <div className="relative w-full sm:w-[300px]">
                    <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                    <Input
                        type="search"
                        placeholder="Search employee..."
                        className="pl-9"
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                    />
                </div>
            </div>

            <Tabs defaultValue="pending" value={activeTab} onValueChange={setActiveTab} className="space-y-4">
                <TabsList>
                    <TabsTrigger value="pending">Pending</TabsTrigger>
                    <TabsTrigger value="approved">Approved</TabsTrigger>
                    <TabsTrigger value="rejected">Rejected</TabsTrigger>
                    <TabsTrigger value="all">All History</TabsTrigger>
                </TabsList>

                <div className="border rounded-md bg-white dark:bg-slate-900">
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Employee</TableHead>
                                <TableHead>Role</TableHead>
                                <TableHead>Type</TableHead>
                                <TableHead>Dates</TableHead>
                                <TableHead>Days</TableHead>
                                <TableHead>Reason</TableHead>
                                <TableHead>Status</TableHead>
                                {activeTab === "pending" && <TableHead className="text-right">Actions</TableHead>}
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {loading ? (
                                <TableRow>
                                    <TableCell colSpan={7} className="text-center h-24">
                                        Loading...
                                    </TableCell>
                                </TableRow>
                            ) : filteredLeaves.length === 0 ? (
                                <TableRow>
                                    <TableCell colSpan={7} className="text-center h-24 text-muted-foreground">
                                        No authorized leave requests found.
                                    </TableCell>
                                </TableRow>
                            ) : (
                                filteredLeaves.map((leave) => (
                                    <TableRow key={leave.id}>
                                        <TableCell className="font-medium">
                                            {leave.employee_name || "Unknown"}
                                            {leave.employee_code && <span className="text-xs text-muted-foreground block">{leave.employee_code}</span>}
                                        </TableCell>
                                        <TableCell className="capitalize">{leave.role || "-"}</TableCell>
                                        <TableCell className="capitalize">{leave.leave_type}</TableCell>
                                        <TableCell>
                                            <div className="text-sm">
                                                {format(new Date(leave.start_date), "MMM d")} - {format(new Date(leave.end_date), "MMM d, yyyy")}
                                            </div>
                                        </TableCell>
                                        <TableCell>{leave.days_requested}</TableCell>
                                        <TableCell className="max-w-[200px] truncate" title={leave.reason}>
                                            {leave.reason}
                                        </TableCell>
                                        <TableCell>
                                            {getStatusBadge(leave.status)}
                                        </TableCell>
                                        {activeTab === "pending" && (
                                            <TableCell className="text-right">
                                                <div className="flex justify-end gap-2">
                                                    <Button
                                                        size="sm"
                                                        variant="outline"
                                                        className="text-green-600 hover:text-green-700 hover:bg-green-50"
                                                        onClick={() => handleApprove(leave.id)}
                                                    >
                                                        <Check className="h-4 w-4" />
                                                    </Button>
                                                    <Button
                                                        size="sm"
                                                        variant="outline"
                                                        className="text-red-600 hover:text-red-700 hover:bg-red-50"
                                                        onClick={() => openRejectDialog(leave.id)}
                                                    >
                                                        <X className="h-4 w-4" />
                                                    </Button>
                                                </div>
                                            </TableCell>
                                        )}
                                    </TableRow>
                                ))
                            )}
                        </TableBody>
                    </Table>
                </div>
            </Tabs>

            <Dialog open={isRejectDialogOpen} onOpenChange={setIsRejectDialogOpen}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Reject Leave Request</DialogTitle>
                        <DialogDescription>
                            Please provide a reason for rejecting this leave request.
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        <div className="space-y-2">
                            <Label htmlFor="reason">Rejection Reason</Label>
                            <Textarea
                                id="reason"
                                placeholder="e.g., Critical project deadline"
                                value={rejectReason}
                                onChange={(e) => setRejectReason(e.target.value)}
                            />
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setIsRejectDialogOpen(false)}>Cancel</Button>
                        <Button variant="destructive" onClick={handleReject} disabled={!rejectReason.trim()}>
                            Reject Request
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}
