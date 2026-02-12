"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
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
import { ArrowLeft } from "lucide-react"
import Link from "next/link"

export default function NewEmployeePage() {
    const router = useRouter()
    const [loading, setLoading] = useState(false)
    const [formData, setFormData] = useState({
        first_name: "",
        last_name: "",
        email: "",
        phone: "",
        role: "",
        position: "",
        department_id: "",
        hire_date: "",
    })

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        setLoading(true)

        try {
            // Validate required fields
            if (!formData.first_name || !formData.last_name || !formData.email || !formData.role) {
                toast.error("Please fill in all required fields")
                setLoading(false)
                return
            }

            // Create employee
            await api.post("/company/admin/employees", formData)

            toast.success("Employee created successfully")
            router.push("/admin/employees")
        } catch (error: any) {
            console.error("Failed to create employee:", error)

            if (error.response?.status === 409) {
                toast.error("This email ID already exists. Please contact support or use a different email.")
            } else {
                toast.error(error.response?.data?.message || "Failed to create employee")
            }
        } finally {
            setLoading(false)
        }
    }

    const handleChange = (field: string, value: string) => {
        setFormData((prev) => ({ ...prev, [field]: value }))
    }

    return (
        <div className="container mx-auto py-10">
            <div className="mb-6">
                <Link href="/admin/employees">
                    <Button variant="ghost" size="sm">
                        <ArrowLeft className="mr-2 h-4 w-4" />
                        Back to Employees
                    </Button>
                </Link>
            </div>

            <Card className="max-w-2xl mx-auto">
                <CardHeader>
                    <CardTitle>Create New Employee</CardTitle>
                    <CardDescription>
                        Add a new employee to your organization. They will receive an email with login credentials.
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <form onSubmit={handleSubmit} className="space-y-6">
                        {/* Personal Information */}
                        <div className="space-y-4">
                            <h3 className="text-lg font-semibold">Personal Information</h3>

                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <Label htmlFor="first_name">
                                        First Name <span className="text-red-500">*</span>
                                    </Label>
                                    <Input
                                        id="first_name"
                                        value={formData.first_name}
                                        onChange={(e) => handleChange("first_name", e.target.value)}
                                        placeholder="John"
                                        required
                                    />
                                </div>

                                <div className="space-y-2">
                                    <Label htmlFor="last_name">
                                        Last Name <span className="text-red-500">*</span>
                                    </Label>
                                    <Input
                                        id="last_name"
                                        value={formData.last_name}
                                        onChange={(e) => handleChange("last_name", e.target.value)}
                                        placeholder="Doe"
                                        required
                                    />
                                </div>
                            </div>

                            <div className="space-y-2">
                                <Label htmlFor="email">
                                    Email <span className="text-red-500">*</span>
                                </Label>
                                <Input
                                    id="email"
                                    type="email"
                                    value={formData.email}
                                    onChange={(e) => handleChange("email", e.target.value)}
                                    placeholder="john.doe@company.com"
                                    required
                                />
                            </div>

                            <div className="space-y-2">
                                <Label htmlFor="phone">Phone</Label>
                                <Input
                                    id="phone"
                                    type="tel"
                                    value={formData.phone}
                                    onChange={(e) => handleChange("phone", e.target.value)}
                                    placeholder="+1 (555) 000-0000"
                                />
                            </div>
                        </div>

                        {/* Role & Position */}
                        <div className="space-y-4">
                            <h3 className="text-lg font-semibold">Role & Position</h3>

                            <div className="space-y-2">
                                <Label htmlFor="role">
                                    Role <span className="text-red-500">*</span>
                                </Label>
                                <Select
                                    value={formData.role}
                                    onValueChange={(value) => handleChange("role", value)}
                                    required
                                >
                                    <SelectTrigger id="role">
                                        <SelectValue placeholder="Select a role" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="employee">Employee</SelectItem>
                                        <SelectItem value="team_lead">Team Lead</SelectItem>
                                        <SelectItem value="manager">Manager</SelectItem>
                                        <SelectItem value="hr">HR</SelectItem>
                                    </SelectContent>
                                </Select>
                                <p className="text-sm text-muted-foreground">
                                    This determines what features and data the employee can access
                                </p>
                            </div>

                            <div className="space-y-2">
                                <Label htmlFor="position">Job Title / Position</Label>
                                <Input
                                    id="position"
                                    value={formData.position}
                                    onChange={(e) => handleChange("position", e.target.value)}
                                    placeholder="Software Engineer"
                                />
                            </div>

                            <div className="space-y-2">
                                <Label htmlFor="department_id">Department</Label>
                                <Input
                                    id="department_id"
                                    value={formData.department_id}
                                    onChange={(e) => handleChange("department_id", e.target.value)}
                                    placeholder="Engineering"
                                />
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

                        {/* Submit Buttons */}
                        <div className="flex gap-4 pt-4">
                            <Button type="submit" disabled={loading}>
                                {loading ? "Creating..." : "Create Employee"}
                            </Button>
                            <Button
                                type="button"
                                variant="outline"
                                onClick={() => router.push("/admin/employees")}
                                disabled={loading}
                            >
                                Cancel
                            </Button>
                        </div>
                    </form>
                </CardContent>
            </Card>
        </div>
    )
}
