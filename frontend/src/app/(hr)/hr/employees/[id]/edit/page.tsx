"use client"

import { useState, useEffect } from "react"
import { useRouter, useParams } from "next/navigation"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select"
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card"
import { toast } from "sonner"
import api from "@/lib/api"
import { ArrowLeft, Loader2 } from "lucide-react"
import Link from "next/link"

export default function HREditEmployeePage() {
    const router = useRouter()
    const params = useParams()
    const id = params?.id as string
    const [loading, setLoading] = useState(true)
    const [saving, setSaving] = useState(false)
    const [departments, setDepartments] = useState<any[]>([])

    const [formData, setFormData] = useState({
        first_name: "",
        last_name: "",
        email: "",
        phone: "",
        role: "",
        position: "",
        department_id: "",
        hire_date: "",
        // Read-only fields for display
        status: "",
        created_at: ""
    })

    // Fetch employee data and departments
    useEffect(() => {
        const loadData = async () => {
            try {
                // Use HR endpoints
                const [empRes, deptRes] = await Promise.all([
                    api.get(`/company/hr/employees/${id}`),
                    api.get("/company/hr/departments")
                ])

                const emp = empRes.data
                setFormData({
                    first_name: emp.first_name,
                    last_name: emp.last_name,
                    email: emp.email,
                    phone: emp.phone || "",
                    role: emp.role,
                    position: emp.position || "",
                    department_id: emp.department ?
                        deptRes.data.find((d: any) => d.name === emp.department)?.id || ""
                        : "",
                    hire_date: emp.hire_date || "",
                    status: emp.status,
                    created_at: emp.created_at
                })

                setDepartments(deptRes.data || [])
            } catch (error) {
                console.error("Failed to load data:", error)
                toast.error("Failed to load employee details")
                router.push("/hr/employees")
            } finally {
                setLoading(false)
            }
        }

        if (id) {
            loadData()
        }
    }, [id, router])

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        setSaving(true)

        try {
            // Validate required fields
            if (!formData.first_name || !formData.last_name) {
                toast.error("First name and last name are required")
                setSaving(false)
                return
            }

            // Update employee
            const updatePayload = {
                first_name: formData.first_name,
                last_name: formData.last_name,
                phone: formData.phone,
                role: formData.role,
                position: formData.position,
                department_id: formData.department_id,
                hire_date: formData.hire_date
            }

            await api.put(`/company/hr/employees/${id}`, updatePayload)

            toast.success("Employee updated successfully")
            router.push("/hr/employees")
        } catch (error: any) {
            console.error("Failed to update employee:", error)
            toast.error(error.response?.data?.message || "Failed to update employee")
        } finally {
            setSaving(false)
        }
    }

    const handleChange = (field: string, value: string) => {
        setFormData((prev) => ({ ...prev, [field]: value }))
    }

    if (loading) {
        return (
            <div className="flex items-center justify-center h-full min-h-[50vh]">
                <Loader2 className="h-8 w-8 animate-spin text-primary" />
            </div>
        )
    }

    return (
        <div className="container mx-auto py-10 max-w-2xl">
            <div className="mb-6">
                <Link href="/hr/employees">
                    <Button variant="ghost" size="sm">
                        <ArrowLeft className="mr-2 h-4 w-4" />
                        Back to Employees
                    </Button>
                </Link>
            </div>

            <Card>
                <CardHeader>
                    <CardTitle>Edit Employee</CardTitle>
                    <CardDescription>
                        Update details for {formData.first_name} {formData.last_name}
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <form onSubmit={handleSubmit} className="space-y-6">
                        {/* Personal Information */}
                        <div className="space-y-4">
                            <h3 className="text-lg font-medium">Personal Information</h3>
                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <Label htmlFor="first_name">First Name *</Label>
                                    <Input
                                        id="first_name"
                                        placeholder="John"
                                        value={formData.first_name}
                                        onChange={(e) => handleChange("first_name", e.target.value)}
                                        required
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="last_name">Last Name *</Label>
                                    <Input
                                        id="last_name"
                                        placeholder="Doe"
                                        value={formData.last_name}
                                        onChange={(e) => handleChange("last_name", e.target.value)}
                                        required
                                    />
                                </div>
                            </div>

                            <div className="space-y-2">
                                <Label htmlFor="email">Email Address</Label>
                                <Input
                                    id="email"
                                    type="email"
                                    value={formData.email}
                                    disabled
                                    className="bg-muted"
                                    title="Email cannot be changed directly"
                                />
                                <p className="text-xs text-muted-foreground">Email cannot be changed directly.</p>
                            </div>

                            <div className="space-y-2">
                                <Label htmlFor="phone">Phone Number</Label>
                                <Input
                                    id="phone"
                                    type="tel"
                                    placeholder="+1 (555) 000-0000"
                                    value={formData.phone}
                                    onChange={(e) => handleChange("phone", e.target.value)}
                                />
                            </div>
                        </div>

                        {/* Role & Department */}
                        <div className="space-y-4">
                            <h3 className="text-lg font-medium">Role & Position</h3>

                            <div className="space-y-2">
                                <Label htmlFor="role">Role *</Label>
                                <Select
                                    value={formData.role}
                                    onValueChange={(val) => handleChange("role", val)}
                                >
                                    <SelectTrigger>
                                        <SelectValue placeholder="Select a role" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="manager">Manager</SelectItem>
                                        <SelectItem value="team_lead">Team Lead</SelectItem>
                                        <SelectItem value="employee">Employee</SelectItem>
                                    </SelectContent>
                                </Select>
                                <p className="text-xs text-muted-foreground">
                                    Determines access permissions.
                                </p>
                            </div>

                            <div className="space-y-2">
                                <Label htmlFor="position">Job Title / Position</Label>
                                <Input
                                    id="position"
                                    placeholder="Software Engineer"
                                    value={formData.position}
                                    onChange={(e) => handleChange("position", e.target.value)}
                                />
                            </div>

                            <div className="space-y-2">
                                <Label htmlFor="department">Department</Label>
                                <Select
                                    value={formData.department_id}
                                    onValueChange={(val) => handleChange("department_id", val)}
                                >
                                    <SelectTrigger>
                                        <SelectValue placeholder="Select a department" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {departments.map((dept) => (
                                            <SelectItem key={dept.id} value={dept.id}>
                                                {dept.name}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                            </div>

                            <div className="space-y-2">
                                <Label htmlFor="hire_date">Hire Date</Label>
                                <Input
                                    id="hire_date"
                                    type="date"
                                    value={formData.hire_date}
                                    onChange={(e) => handleChange("hire_date", e.target.value)}
                                />
                            </div>
                        </div>

                        <div className="flex justify-end gap-4 pt-4">
                            <Link href="/hr/employees">
                                <Button variant="outline" type="button">Cancel</Button>
                            </Link>
                            <Button type="submit" disabled={saving}>
                                {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                                Save Changes
                            </Button>
                        </div>
                    </form>
                </CardContent>
            </Card>
        </div>
    )
}
