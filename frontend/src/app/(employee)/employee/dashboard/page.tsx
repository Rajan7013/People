"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Clock, Calendar, FileText, UserCheck, LogOut } from "lucide-react";
import { useAuth } from "@/context/auth-context";
import { attendanceService } from "@/services/attendanceService";
import { leaveService } from "@/services/leaveService"; // Optional if we want to list leaves
import { toast } from "sonner";
import { format, differenceInMinutes } from "date-fns";

export default function EmployeeDashboardPage() {
    const { user } = useAuth();
    const [isCheckedIn, setIsCheckedIn] = useState(false);
    const [loading, setLoading] = useState(true);
    const [lastActionTime, setLastActionTime] = useState<string | null>(null);
    const [todayDuration, setTodayDuration] = useState("0h 0m");
    const [approvedLeaves, setApprovedLeaves] = useState(0);

    const fetchStatus = async () => {
        try {
            const [statusRes, leavesRes] = await Promise.all([
                attendanceService.getCurrentStatus(),
                leaveService.getMyLeaves() // Fetch leaves
            ]);

            const data = statusRes.data;
            setIsCheckedIn(data.is_checked_in);
            if (data.last_check_in) {
                setLastActionTime(data.last_check_in);
                // Simple calculation if checked in: diff from check-in to now
                if (data.is_checked_in) {
                    const checkInTime = new Date(data.last_check_in);
                    const now = new Date();
                    const diff = differenceInMinutes(now, checkInTime);
                    const hours = Math.floor(diff / 60);
                    const minutes = diff % 60;
                    setTodayDuration(`${hours}h ${minutes}m`);
                }
            }

            // Calculate approved leaves
            const leaves = leavesRes.data || [];
            const approvedCount = leaves.filter((l: any) => l.status === 'approved').length;
            setApprovedLeaves(approvedCount);

        } catch (error) {
            console.error("Failed to fetch status", error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchStatus();
    }, []);

    const handleCheckIn = async () => {
        try {
            await attendanceService.checkIn({ notes: "Web Dashboard Check-in" });
            toast.success("Checked in successfully");
            fetchStatus();
        } catch (error: any) {
            toast.error(error.response?.data?.error || "Check-in failed");
        }
    };

    const handleCheckOut = async () => {
        try {
            await attendanceService.checkOut({ notes: "Web Dashboard Check-out" });
            toast.success("Checked out successfully");
            fetchStatus();
        } catch (error: any) {
            toast.error(error.response?.data?.error || "Check-out failed");
        }
    };

    return (
        <div className="space-y-6">
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                <div>
                    <h1 className="text-3xl font-bold text-gray-900 dark:text-white">Good Morning, {user?.first_name}</h1>
                    <p className="text-muted-foreground">Here's what's happening today.</p>
                </div>
                <div className="flex gap-4">
                    {isCheckedIn ? (
                        <Button onClick={handleCheckOut} variant="destructive">
                            <LogOut className="w-4 h-4 mr-2" />
                            Clock Out
                        </Button>
                    ) : (
                        <Button onClick={handleCheckIn} className="bg-green-600 hover:bg-green-700">
                            <Clock className="w-4 h-4 mr-2" />
                            Clock In
                        </Button>
                    )}
                </div>
            </div>

            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Work Hours</CardTitle>
                        <Clock className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{todayDuration}</div>
                        <p className="text-xs text-muted-foreground">Today's active time</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Attendance</CardTitle>
                        <UserCheck className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold text-green-600">
                            {isCheckedIn ? "Present" : "Punched Out"}
                        </div>
                        <p className="text-xs text-muted-foreground">Current Status</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Approved Leaves</CardTitle>
                        <FileText className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{approvedLeaves}</div>
                        <p className="text-xs text-muted-foreground">Total approved leaves</p>
                    </CardContent>
                </Card>
            </div>

            <div className="grid gap-6 md:grid-cols-1">
                <Card>
                    <CardHeader>
                        <CardTitle>Recent Activity</CardTitle>
                    </CardHeader>
                    <CardContent>
                        {lastActionTime ? (
                            <p className="text-sm text-muted-foreground">
                                Last action: {isCheckedIn ? "Checked In" : "Checked Out"} at {format(new Date(lastActionTime), "h:mm a")}
                            </p>
                        ) : (
                            <p className="text-sm text-muted-foreground">No activity today.</p>
                        )}
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
