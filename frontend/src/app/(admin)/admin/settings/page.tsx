"use client"

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import AttendanceSettings from "./components/AttendanceSettings"
import PayrollSettings from "./components/PayrollSettings"
import LeaveSettings from "./components/LeaveSettings"

export default function SettingsPage() {
    return (
        <div className="container mx-auto py-10">
            <div className="mb-6">
                <h2 className="text-3xl font-bold tracking-tight">Company Settings</h2>
                <p className="text-muted-foreground">
                    Configure your organization's policies and settings
                </p>
            </div>

            <Tabs defaultValue="attendance" className="space-y-4">
                <TabsList>
                    <TabsTrigger value="attendance">Attendance</TabsTrigger>
                    <TabsTrigger value="payroll">Salary & Payroll</TabsTrigger>
                    <TabsTrigger value="leaves">Leaves</TabsTrigger>
                </TabsList>

                <TabsContent value="attendance">
                    <AttendanceSettings />
                </TabsContent>

                <TabsContent value="payroll">
                    <PayrollSettings />
                </TabsContent>

                <TabsContent value="leaves">
                    <LeaveSettings />
                </TabsContent>
            </Tabs>
        </div>
    )
}
