"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Users, CalendarCheck } from "lucide-react";
import { employeeService } from "@/services/employeeService";
import { attendanceService } from "@/services/attendanceService";
import { toast } from "sonner";
import { format } from "date-fns";

export default function TeamLeadDashboardPage() {
    const [stats, setStats] = useState({
        myTeamCount: 0,
        presentToday: 0,
    });
    const [teamMembers, setTeamMembers] = useState<any[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchData = async () => {
            try {
                // Fetch My Team
                const team = await employeeService.getMyTeam('team-lead');
                const teamCount = team.data.length;
                setTeamMembers(team.data);

                // Fetch Team Attendance Today
                const today = format(new Date(), 'yyyy-MM-dd');
                let presentCount = 0;
                try {
                    const attendance = await attendanceService.getTeamAttendance('team-lead', { date: today });
                    presentCount = attendance.data.filter((r: any) => r.status === 'present').length;
                } catch (e) {
                    console.warn("Failed to fetch team attendance", e);
                }

                setStats({
                    myTeamCount: teamCount,
                    presentToday: presentCount,
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
            <h1 className="text-3xl font-bold text-gray-900 dark:text-white">Team Lead Dashboard</h1>

            <div className="grid gap-6 md:grid-cols-2">
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">My Team</CardTitle>
                        <Users className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">
                            {loading ? "..." : stats.myTeamCount}
                        </div>
                        <p className="text-xs text-muted-foreground">Team members</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Team Attendance</CardTitle>
                        <CalendarCheck className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">
                            {loading ? "..." : `${attendancePercentage}%`}
                        </div>
                        <p className="text-xs text-muted-foreground">{stats.presentToday} Present today</p>
                    </CardContent>
                </Card>
            </div>

            <Card>
                <CardHeader>
                    <CardTitle>Team Overview</CardTitle>
                </CardHeader>
                <CardContent>
                    {loading ? <p>Loading...</p> : (
                        teamMembers.length > 0 ? (
                            <ul className="space-y-2">
                                {teamMembers.map(member => (
                                    <li key={member.id} className="flex justify-between items-center text-sm border-b pb-2 last:border-0">
                                        <span>{member.first_name} {member.last_name}</span>
                                        <div className="flex items-center gap-2">
                                            <span className="text-muted-foreground">{member.position || 'Team Member'}</span>
                                        </div>
                                    </li>
                                ))}
                            </ul>
                        ) : (
                            <p className="text-sm text-muted-foreground">No team members found.</p>
                        )
                    )}
                </CardContent>
            </Card>
        </div>
    );
}
