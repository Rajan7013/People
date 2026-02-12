"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Users, CalendarCheck, FileText } from "lucide-react";
import { employeeService } from "@/services/employeeService";
import { attendanceService } from "@/services/attendanceService";
import { leaveService } from "@/services/leaveService";
import { toast } from "sonner";
import { format } from "date-fns";

export default function ManagerDashboardPage() {
    const [stats, setStats] = useState({
        myTeamCount: 0,
        presentToday: 0,
        pendingLeaves: 0,
    });
    const [teamMembers, setTeamMembers] = useState<any[]>([]); // To store team details for "Team Status"
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchData = async () => {
            try {
                // Fetch My Team
                const team = await employeeService.getMyTeam('manager');
                const teamCount = team.data.length;
                setTeamMembers(team.data.slice(0, 5)); // Keep first 5 for the list

                // Fetch Team Attendance Today
                const today = format(new Date(), 'yyyy-MM-dd');
                let presentCount = 0;
                try {
                    const attendance = await attendanceService.getTeamAttendance('manager', { date: today });
                    presentCount = attendance.data.filter((r: any) => r.status === 'present').length;
                } catch (e) {
                    console.warn("Failed to fetch team attendance", e);
                }

                // Fetch Department Leaves (Pending)
                let pendingCount = 0;
                try {
                    const leaves = await leaveService.getDepartmentLeaves();
                    pendingCount = leaves.data.filter((l: any) => l.status === 'pending').length;
                } catch (e) {
                    console.warn("Failed to fetch department leaves", e);
                }

                setStats({
                    myTeamCount: teamCount,
                    presentToday: presentCount,
                    pendingLeaves: pendingCount,
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

    // Calculate generic attendance % if team count > 0
    const attendancePercentage = stats.myTeamCount > 0
        ? Math.round((stats.presentToday / stats.myTeamCount) * 100)
        : 0;

    return (
        <div className="space-y-6">
            <h1 className="text-3xl font-bold text-gray-900 dark:text-white">Manager Dashboard</h1>

            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">My Team</CardTitle>
                        <Users className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">
                            {loading ? "..." : stats.myTeamCount}
                        </div>
                        <p className="text-xs text-muted-foreground">Direct reports</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Department Attendance</CardTitle>
                        <CalendarCheck className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">
                            {loading ? "..." : `${attendancePercentage}%`}
                        </div>
                        <p className="text-xs text-muted-foreground">{stats.presentToday} Present today</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Leave Requests</CardTitle>
                        <FileText className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">
                            {loading ? "..." : stats.pendingLeaves}
                        </div>
                        <p className="text-xs text-muted-foreground">Pending your approval</p>
                    </CardContent>
                </Card>
            </div>

            <div className="grid gap-6 md:grid-cols-2">
                <Card>
                    <CardHeader>
                        <CardTitle>Team Status (Preview)</CardTitle>
                    </CardHeader>
                    <CardContent>
                        {loading ? <p>Loading...</p> : (
                            teamMembers.length > 0 ? (
                                <ul className="space-y-2">
                                    {teamMembers.map(member => (
                                        <li key={member.id} className="flex justify-between items-center text-sm border-b pb-2 last:border-0">
                                            <span>{member.first_name} {member.last_name}</span>
                                            <span className="text-muted-foreground">{member.position || 'Employee'}</span>
                                        </li>
                                    ))}
                                </ul>
                            ) : (
                                <p className="text-sm text-muted-foreground">No team members found.</p>
                            )
                        )}
                        {stats.myTeamCount > 5 && <p className="text-xs text-muted-foreground mt-2">and {stats.myTeamCount - 5} more...</p>}
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader>
                        <CardTitle>Direct Reports Leaves</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-4">
                            <p className="text-sm text-muted-foreground">Upcoming leaves (Coming Soon).</p>
                        </div>
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
