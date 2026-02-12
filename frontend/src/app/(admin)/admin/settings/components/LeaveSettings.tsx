"use client"

import { useEffect, useState } from "react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Button } from "@/components/ui/button"
import { Switch } from "@/components/ui/switch"
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table"
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog"
import { toast } from "sonner"
import api from "@/lib/api"
import { Plus } from "lucide-react"

interface LeaveType {
    id: string
    name: string
    annual_limit: number
    carry_forward_limit: number
    is_paid: boolean
    requires_approval: boolean
    accrual_rate: number
    description: string
}

export default function LeaveSettings() {
    const [loading, setLoading] = useState(true)
    const [leaveTypes, setLeaveTypes] = useState<LeaveType[]>([])
    const [isDialogOpen, setIsDialogOpen] = useState(false)
    const [newLeaveType, setNewLeaveType] = useState({
        name: "",
        annual_limit: 0,
        carry_forward_limit: 0,
        is_paid: true,
        requires_approval: true,
        accrual_rate: 0,
        description: "",
    })

    useEffect(() => {
        loadLeaveTypes()
    }, [])

    const loadLeaveTypes = async () => {
        setLoading(true)
        try {
            const response = await api.get("/company/admin/policies/leave-types")
            setLeaveTypes(response.data || [])
        } catch (error) {
            console.error("Failed to load leave types:", error)
            toast.error("Failed to load leave types")
        } finally {
            setLoading(false)
        }
    }

    const handleCreate = async () => {
        try {
            await api.post("/company/admin/policies/leave-types", newLeaveType)
            toast.success("Leave type created successfully")
            setIsDialogOpen(false)
            setNewLeaveType({
                name: "",
                annual_limit: 0,
                carry_forward_limit: 0,
                is_paid: true,
                requires_approval: true,
                accrual_rate: 0,
                description: "",
            })
            loadLeaveTypes()
        } catch (error) {
            console.error("Failed to create leave type:", error)
            toast.error("Failed to create leave type")
        }
    }

    if (loading) {
        return <div>Loading leave types...</div>
    }

    return (
        <Card>
            <CardHeader>
                <div className="flex justify-between items-center">
                    <div>
                        <CardTitle>Leave Types</CardTitle>
                        <CardDescription>
                            Configure leave types and their annual limits
                        </CardDescription>
                    </div>
                    <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
                        <DialogTrigger asChild>
                            <Button>
                                <Plus className="mr-2 h-4 w-4" />
                                Add Leave Type
                            </Button>
                        </DialogTrigger>
                        <DialogContent>
                            <DialogHeader>
                                <DialogTitle>Add Leave Type</DialogTitle>
                                <DialogDescription>
                                    Create a new leave type for your organization
                                </DialogDescription>
                            </DialogHeader>
                            <div className="space-y-4 py-4">
                                <div className="space-y-2">
                                    <Label htmlFor="leave_name">Name</Label>
                                    <Input
                                        id="leave_name"
                                        value={newLeaveType.name}
                                        onChange={(e) =>
                                            setNewLeaveType({ ...newLeaveType, name: e.target.value })
                                        }
                                        placeholder="e.g., Annual Leave, Sick Leave"
                                    />
                                </div>

                                <div className="grid grid-cols-2 gap-4">
                                    <div className="space-y-2">
                                        <Label htmlFor="annual_limit">Annual Limit (days)</Label>
                                        <Input
                                            id="annual_limit"
                                            type="number"
                                            value={newLeaveType.annual_limit}
                                            onChange={(e) =>
                                                setNewLeaveType({
                                                    ...newLeaveType,
                                                    annual_limit: parseInt(e.target.value) || 0,
                                                })
                                            }
                                        />
                                    </div>

                                    <div className="space-y-2">
                                        <Label htmlFor="carry_forward">Carryforward Limit (days)</Label>
                                        <Input
                                            id="carry_forward"
                                            type="number"
                                            value={newLeaveType.carry_forward_limit}
                                            onChange={(e) =>
                                                setNewLeaveType({
                                                    ...newLeaveType,
                                                    carry_forward_limit: parseInt(e.target.value) || 0,
                                                })
                                            }
                                        />
                                    </div>
                                </div>

                                <div className="flex items-center justify-between">
                                    <Label htmlFor="is_paid">Paid Leave</Label>
                                    <Switch
                                        id="is_paid"
                                        checked={newLeaveType.is_paid}
                                        onCheckedChange={(checked) =>
                                            setNewLeaveType({ ...newLeaveType, is_paid: checked })
                                        }
                                    />
                                </div>

                                <div className="flex items-center justify-between">
                                    <Label htmlFor="requires_approval">Requires Approval</Label>
                                    <Switch
                                        id="requires_approval"
                                        checked={newLeaveType.requires_approval}
                                        onCheckedChange={(checked) =>
                                            setNewLeaveType({
                                                ...newLeaveType,
                                                requires_approval: checked,
                                            })
                                        }
                                    />
                                </div>

                                <div className="space-y-2">
                                    <Label htmlFor="description">Description (Optional)</Label>
                                    <Input
                                        id="description"
                                        value={newLeaveType.description}
                                        onChange={(e) =>
                                            setNewLeaveType({
                                                ...newLeaveType,
                                                description: e.target.value,
                                            })
                                        }
                                        placeholder="Brief description"
                                    />
                                </div>
                            </div>
                            <DialogFooter>
                                <Button variant="outline" onClick={() => setIsDialogOpen(false)}>
                                    Cancel
                                </Button>
                                <Button onClick={handleCreate}>Create Leave Type</Button>
                            </DialogFooter>
                        </DialogContent>
                    </Dialog>
                </div>
            </CardHeader>
            <CardContent>
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead>Name</TableHead>
                            <TableHead>Annual Limit</TableHead>
                            <TableHead>Carryforward</TableHead>
                            <TableHead>Paid</TableHead>
                            <TableHead>Requires Approval</TableHead>
                            <TableHead>Description</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {leaveTypes.length === 0 ? (
                            <TableRow>
                                <TableCell colSpan={6} className="text-center text-muted-foreground">
                                    No leave types defined. Click "Add Leave Type" to create one.
                                </TableCell>
                            </TableRow>
                        ) : (
                            leaveTypes.map((lt) => (
                                <TableRow key={lt.id}>
                                    <TableCell className="font-medium">{lt.name}</TableCell>
                                    <TableCell>{lt.annual_limit} days</TableCell>
                                    <TableCell>{lt.carry_forward_limit} days</TableCell>
                                    <TableCell>
                                        <span
                                            className={
                                                lt.is_paid ? "text-green-600" : "text-gray-500"
                                            }
                                        >
                                            {lt.is_paid ? "Yes" : "No"}
                                        </span>
                                    </TableCell>
                                    <TableCell>
                                        {lt.requires_approval ? "Yes" : "No"}
                                    </TableCell>
                                    <TableCell className="text-muted-foreground">
                                        {lt.description || "-"}
                                    </TableCell>
                                </TableRow>
                            ))
                        )}
                    </TableBody>
                </Table>
            </CardContent>
        </Card>
    )
}
