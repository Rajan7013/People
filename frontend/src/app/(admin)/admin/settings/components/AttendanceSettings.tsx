"use client"

import { useEffect, useState } from "react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Button } from "@/components/ui/button"
import { toast } from "sonner"
import api from "@/lib/api"

interface AttendancePolicy {
    working_hours_per_day: number
    grace_period_minutes: number
    overtime_threshold_minutes: number
    required_hours_per_week: number
    required_days_per_month: number
    late_fine_per_minute: number
}

export default function AttendanceSettings() {
    const [loading, setLoading] = useState(true)
    const [saving, setSaving] = useState(false)
    const [policy, setPolicy] = useState<AttendancePolicy>({
        working_hours_per_day: 8,
        grace_period_minutes: 15,
        overtime_threshold_minutes: 480,
        required_hours_per_week: 40,
        required_days_per_month: 22,
        late_fine_per_minute: 0,
    })

    useEffect(() => {
        loadPolicy()
    }, [])

    const loadPolicy = async () => {
        setLoading(true)
        try {
            const response = await api.get("/company/admin/policies/attendance")
            setPolicy(response.data)
        } catch (error) {
            console.error("Failed to load attendance policy:", error)
            toast.error("Failed to load attendance policy")
        } finally {
            setLoading(false)
        }
    }

    const handleSave = async () => {
        setSaving(true)
        try {
            await api.put("/company/admin/policies/attendance", policy)
            toast.success("Attendance policy updated successfully")
        } catch (error) {
            console.error("Failed to update attendance policy:", error)
            toast.error("Failed to update attendance policy")
        } finally {
            setSaving(false)
        }
    }

    const handleChange = (field: keyof AttendancePolicy, value: string) => {
        setPolicy((prev) => ({
            ...prev,
            [field]: parseFloat(value) || 0,
        }))
    }

    if (loading) {
        return <div>Loading attendance policy...</div>
    }

    return (
        <Card>
            <CardHeader>
                <CardTitle>Attendance Policy</CardTitle>
                <CardDescription>
                    Configure working hours, grace periods, and overtime thresholds for your organization
                </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
                <div className="grid grid-cols-2 gap-6">
                    <div className="space-y-2">
                        <Label htmlFor="working_hours">Working Hours Per Day</Label>
                        <Input
                            id="working_hours"
                            type="number"
                            step="0.5"
                            value={policy.working_hours_per_day}
                            onChange={(e) => handleChange("working_hours_per_day", e.target.value)}
                        />
                        <p className="text-sm text-muted-foreground">
                            Standard working hours per day (e.g., 8 hours)
                        </p>
                    </div>

                    <div className="space-y-2">
                        <Label htmlFor="grace_period">Grace Period (minutes)</Label>
                        <Input
                            id="grace_period"
                            type="number"
                            value={policy.grace_period_minutes}
                            onChange={(e) => handleChange("grace_period_minutes", e.target.value)}
                        />
                        <p className="text-sm text-muted-foreground">
                            Late arrival tolerance before marking as late
                        </p>
                    </div>

                    <div className="space-y-2">
                        <Label htmlFor="overtime_threshold">Overtime Threshold (minutes)</Label>
                        <Input
                            id="overtime_threshold"
                            type="number"
                            value={policy.overtime_threshold_minutes}
                            onChange={(e) => handleChange("overtime_threshold_minutes", e.target.value)}
                        />
                        <p className="text-sm text-muted-foreground">
                            Minutes after which overtime is calculated (e.g., 480 = 8 hours)
                        </p>
                    </div>

                    <div className="space-y-2">
                        <Label htmlFor="required_hours">Required Hours Per Week</Label>
                        <Input
                            id="required_hours"
                            type="number"
                            step="0.5"
                            value={policy.required_hours_per_week}
                            onChange={(e) => handleChange("required_hours_per_week", e.target.value)}
                        />
                        <p className="text-sm text-muted-foreground">
                            Minimum hours employees must work per week
                        </p>
                    </div>

                    <div className="space-y-2">
                        <Label htmlFor="required_days">Required Days Per Month</Label>
                        <Input
                            id="required_days"
                            type="number"
                            value={policy.required_days_per_month}
                            onChange={(e) => handleChange("required_days_per_month", e.target.value)}
                        />
                        <p className="text-sm text-muted-foreground">
                            Minimum days employees must attend per month
                        </p>
                    </div>

                    <div className="space-y-2">
                        <Label htmlFor="late_fine">Late Fine Per Minute ($)</Label>
                        <Input
                            id="late_fine"
                            type="number"
                            step="0.01"
                            value={policy.late_fine_per_minute}
                            onChange={(e) => handleChange("late_fine_per_minute", e.target.value)}
                        />
                        <p className="text-sm text-muted-foreground">
                            Fine deducted per minute late (after grace period)
                        </p>
                    </div>
                </div>

                <div className="flex gap-4 pt-4">
                    <Button onClick={handleSave} disabled={saving}>
                        {saving ? "Saving..." : "Save Policy"}
                    </Button>
                    <Button variant="outline" onClick={loadPolicy} disabled={saving}>
                        Reset
                    </Button>
                </div>
            </CardContent>
        </Card>
    )
}
