"use client"

import { useEffect, useState } from "react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Button } from "@/components/ui/button"
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
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select"
import { toast } from "sonner"
import api from "@/lib/api"
import { Plus } from "lucide-react"

interface SalaryComponent {
    id: string
    name: string
    type: "earning" | "deduction"
    is_fixed: boolean
    amount: number
    percentage: number
    description: string
}

export default function PayrollSettings() {
    const [loading, setLoading] = useState(true)
    const [components, setComponents] = useState<SalaryComponent[]>([])
    const [isDialogOpen, setIsDialogOpen] = useState(false)
    const [newComponent, setNewComponent] = useState({
        name: "",
        type: "earning",
        is_fixed: true,
        amount: 0,
        percentage: 0,
        description: "",
    })

    useEffect(() => {
        loadComponents()
    }, [])

    const loadComponents = async () => {
        setLoading(true)
        try {
            const response = await api.get("/company/admin/policies/salary-components")
            setComponents(response.data || [])
        } catch (error) {
            console.error("Failed to load salary components:", error)
            toast.error("Failed to load salary components")
        } finally {
            setLoading(false)
        }
    }

    const handleCreate = async () => {
        try {
            await api.post("/company/admin/policies/salary-components", newComponent)
            toast.success("Salary component created successfully")
            setIsDialogOpen(false)
            setNewComponent({
                name: "",
                type: "earning",
                is_fixed: true,
                amount: 0,
                percentage: 0,
                description: "",
            })
            loadComponents()
        } catch (error) {
            console.error("Failed to create salary component:", error)
            toast.error("Failed to create salary component")
        }
    }

    if (loading) {
        return <div>Loading salary components...</div>
    }

    return (
        <div className="space-y-6">
            <Card>
                <CardHeader>
                    <div className="flex justify-between items-center">
                        <div>
                            <CardTitle>Salary Components</CardTitle>
                            <CardDescription>
                                Define earnings and deductions for employee salaries
                            </CardDescription>
                        </div>
                        <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
                            <DialogTrigger asChild>
                                <Button>
                                    <Plus className="mr-2 h-4 w-4" />
                                    Add Component
                                </Button>
                            </DialogTrigger>
                            <DialogContent>
                                <DialogHeader>
                                    <DialogTitle>Add Salary Component</DialogTitle>
                                    <DialogDescription>
                                        Create a new earning or deduction component
                                    </DialogDescription>
                                </DialogHeader>
                                <div className="space-y-4 py-4">
                                    <div className="space-y-2">
                                        <Label htmlFor="name">Name</Label>
                                        <Input
                                            id="name"
                                            value={newComponent.name}
                                            onChange={(e) =>
                                                setNewComponent({ ...newComponent, name: e.target.value })
                                            }
                                            placeholder="e.g., Basic Salary, HRA, Tax"
                                        />
                                    </div>

                                    <div className="space-y-2">
                                        <Label htmlFor="type">Type</Label>
                                        <Select
                                            value={newComponent.type}
                                            onValueChange={(value: "earning" | "deduction") =>
                                                setNewComponent({ ...newComponent, type: value })
                                            }
                                        >
                                            <SelectTrigger>
                                                <SelectValue />
                                            </SelectTrigger>
                                            <SelectContent>
                                                <SelectItem value="earning">Earning</SelectItem>
                                                <SelectItem value="deduction">Deduction</SelectItem>
                                            </SelectContent>
                                        </Select>
                                    </div>

                                    <div className="space-y-2">
                                        <Label htmlFor="calculation">Calculation Method</Label>
                                        <Select
                                            value={newComponent.is_fixed ? "fixed" : "percentage"}
                                            onValueChange={(value) =>
                                                setNewComponent({
                                                    ...newComponent,
                                                    is_fixed: value === "fixed",
                                                })
                                            }
                                        >
                                            <SelectTrigger>
                                                <SelectValue />
                                            </SelectTrigger>
                                            <SelectContent>
                                                <SelectItem value="fixed">Fixed Amount</SelectItem>
                                                <SelectItem value="percentage">Percentage</SelectItem>
                                            </SelectContent>
                                        </Select>
                                    </div>

                                    {newComponent.is_fixed ? (
                                        <div className="space-y-2">
                                            <Label htmlFor="amount">Amount ($)</Label>
                                            <Input
                                                id="amount"
                                                type="number"
                                                step="0.01"
                                                value={newComponent.amount}
                                                onChange={(e) =>
                                                    setNewComponent({
                                                        ...newComponent,
                                                        amount: parseFloat(e.target.value) || 0,
                                                    })
                                                }
                                            />
                                        </div>
                                    ) : (
                                        <div className="space-y-2">
                                            <Label htmlFor="percentage">Percentage (%)</Label>
                                            <Input
                                                id="percentage"
                                                type="number"
                                                step="0.1"
                                                value={newComponent.percentage}
                                                onChange={(e) =>
                                                    setNewComponent({
                                                        ...newComponent,
                                                        percentage: parseFloat(e.target.value) || 0,
                                                    })
                                                }
                                            />
                                        </div>
                                    )}

                                    <div className="space-y-2">
                                        <Label htmlFor="description">Description (Optional)</Label>
                                        <Input
                                            id="description"
                                            value={newComponent.description}
                                            onChange={(e) =>
                                                setNewComponent({
                                                    ...newComponent,
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
                                    <Button onClick={handleCreate}>Create Component</Button>
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
                                <TableHead>Type</TableHead>
                                <TableHead>Value</TableHead>
                                <TableHead>Description</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {components.length === 0 ? (
                                <TableRow>
                                    <TableCell colSpan={4} className="text-center text-muted-foreground">
                                        No salary components defined. Click "Add Component" to create one.
                                    </TableCell>
                                </TableRow>
                            ) : (
                                components.map((comp) => (
                                    <TableRow key={comp.id}>
                                        <TableCell className="font-medium">{comp.name}</TableCell>
                                        <TableCell>
                                            <span
                                                className={`capitalize ${comp.type === "earning"
                                                        ? "text-green-600"
                                                        : "text-red-600"
                                                    }`}
                                            >
                                                {comp.type}
                                            </span>
                                        </TableCell>
                                        <TableCell>
                                            {comp.is_fixed
                                                ? `$${comp.amount.toFixed(2)}`
                                                : `${comp.percentage}%`}
                                        </TableCell>
                                        <TableCell className="text-muted-foreground">
                                            {comp.description || "-"}
                                        </TableCell>
                                    </TableRow>
                                ))
                            )}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>
        </div>
    )
}
