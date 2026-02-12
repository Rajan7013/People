"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Users, Building2, CalendarCheck, FileText, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import Link from "next/link";
import { employeeService } from "@/services/employeeService";
import { departmentService } from "@/services/departmentService";
import { attendanceService } from "@/services/attendanceService";
import { leaveService } from "@/services/leaveService";
import { toast } from "sonner";

export default function AdminDashboardPage() {
    const [stats, setStats] = useState({
        totalEmployees: 0,
        totalDepartments: 0,
        attendanceRate: 0,
        pendingLeaves: 0,
    });
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchData = async () => {
            try {
                // Fetch basic counts
                const [employees, departments] = await Promise.all([
                    employeeService.getAll('admin'),
                    departmentService.getAll(),
                ]);

                // Try to fetch operational stats (might fail if Admin lacks HR permissions, but we assume Admin > HR or has HR role)
                let attendanceRate = 0;
                let pendingLeaves = 0;

                try {
                    // For Attendance, we use 'hr' role route as it has the stats endpoint. 
                    const attendanceStats = await attendanceService.getDailyStats('hr');
                    attendanceRate = attendanceStats.data.attendance_rate || 0;
                } catch (e) {
                    console.warn("Failed to fetch attendance stats", e);
                }

                try {
                    const leaves = await leaveService.getAllLeaves('hr', { status: 'pending' });
                    pendingLeaves = leaves.data.length;
                } catch (e) {
                    console.warn("Failed to fetch leave stats", e);
                }

                setStats({
                    totalEmployees: employees?.data?.length || 0,
                    totalDepartments: departments?.data?.length || 0,
                    attendanceRate,
                    pendingLeaves,
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
            <h1 className="text-3xl font-bold text-gray-900 dark:text-white">Admin Dashboard</h1>

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
                        <p className="text-xs text-muted-foreground">Active employees</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Departments</CardTitle>
                        <Building2 className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">
                            {loading ? "..." : stats.totalDepartments}
                        </div>
                        <p className="text-xs text-muted-foreground">Total departments</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Attendance Today</CardTitle>
                        <CalendarCheck className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">
                            {loading ? "..." : `${stats.attendanceRate}%`}
                        </div>
                        <p className="text-xs text-muted-foreground">Present today</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Pending Leaves</CardTitle>
                        <FileText className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">
                            {loading ? "..." : stats.pendingLeaves}
                        </div>
                        <p className="text-xs text-muted-foreground">Requires approval</p>
                    </CardContent>
                </Card>
            </div>

            <div className="grid gap-6 md:grid-cols-2">
                <Card>
                    <CardHeader>
                        <CardTitle>Recent Activity</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <p className="text-sm text-muted-foreground">No recent activity detected.</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader>
                        <CardTitle>Quick Actions</CardTitle>
                    </CardHeader>
                    <CardContent className="flex flex-col gap-4">
                        <Button asChild className="w-full justify-start" variant="outline">
                            <Link href="/admin/employees/new">
                                <Plus className="mr-2 h-4 w-4" /> Add New Employee
                            </Link>
                        </Button>
                        <Button asChild className="w-full justify-start" variant="outline">
                            <Link href="/admin/departments">
                                <Plus className="mr-2 h-4 w-4" /> Manage Departments
                            </Link>
                        </Button>
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
