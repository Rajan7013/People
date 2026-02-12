"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Users, CalendarCheck, FileText } from "lucide-react";
import { employeeService } from "@/services/employeeService";
import { attendanceService } from "@/services/attendanceService";
import { leaveService } from "@/services/leaveService";
import { toast } from "sonner";

export default function HRDashboardPage() {
    const [stats, setStats] = useState({
        totalEmployees: 0,
        onLeaveToday: 0,
        attendanceRate: 0,
        pendingApprovals: 0,
    });
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchData = async () => {
            try {
                // Fetch basic counts
                const employees = await employeeService.getAll('hr');

                // Fetch Daily Stats
                let onLeaveToday = 0;
                let attendanceRate = 0;
                try {
                    const dailyStats = await attendanceService.getDailyStats('hr');
                    onLeaveToday = dailyStats.data.on_leave || 0;
                    attendanceRate = dailyStats.data.attendance_rate || 0;
                } catch (e) {
                    console.warn("Failed to fetch attendance stats", e);
                }

                // Fetch Pending Leaves
                let pendingApprovals = 0;
                try {
                    const leaves = await leaveService.getAllLeaves('hr', { status: 'pending' });
                    pendingApprovals = leaves.data.length;
                } catch (e) {
                    console.warn("Failed to fetch pending leaves", e);
                }

                setStats({
                    totalEmployees: employees.data.length,
                    onLeaveToday,
                    attendanceRate,
                    pendingApprovals,
                });
            } catch (error) {
                console.error("Dashboard fetch error:", error);
                toast.error("Failed to load dashboard statistics");
            } finally {
                setLoading(false);
            }
        };

        fetchData();
    }, []);

    return (
        <div className="space-y-6">
            <h1 className="text-3xl font-bold text-gray-900 dark:text-white">HR Dashboard</h1>

            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Total Employees</CardTitle>
                        <Users className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">
                            {loading ? "..." : stats.totalEmployees}
                        </div>
                        <p className="text-xs text-muted-foreground">Active staff</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">On Leave Today</CardTitle>
                        <FileText className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">
                            {loading ? "..." : stats.onLeaveToday}
                        </div>
                        <p className="text-xs text-muted-foreground">Employees on leave</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Attendance Rate</CardTitle>
                        <CalendarCheck className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">
                            {loading ? "..." : `${stats.attendanceRate}%`}
                        </div>
                        <p className="text-xs text-muted-foreground">Company-wide</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Pending Approvals</CardTitle>
                        <FileText className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">
                            {loading ? "..." : stats.pendingApprovals}
                        </div>
                        <p className="text-xs text-muted-foreground">Leave requests</p>
                    </CardContent>
                </Card>
            </div>

            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-7">
                <Card className="col-span-4">
                    <CardHeader>
                        <CardTitle>Attendance Overview</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="h-[200px] flex items-center justify-center text-muted-foreground">
                            {/* TODO: Implement Chart using Recharts or similar */}
                            Chart Area (Coming Soon)
                        </div>
                    </CardContent>
                </Card>
                <Card className="col-span-3">
                    <CardHeader>
                        <CardTitle>Upcoming Leaves</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-4">
                            <p className="text-sm text-muted-foreground">No upcoming leaves data available.</p>
                        </div>
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
