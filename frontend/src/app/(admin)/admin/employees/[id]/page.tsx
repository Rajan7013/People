"use client"

import { useState, useEffect } from "react"
import { useRouter, useParams } from "next/navigation"
import { Button } from "@/components/ui/button"
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { ArrowLeft, Edit, Mail, Phone, Briefcase, Calendar, MapPin, Building, CreditCard, ShieldCheck } from "lucide-react"
import Link from "next/link"
import api from "@/lib/api"
import { toast } from "sonner"
import { Separator } from "@/components/ui/separator"

interface Employee {
    id: string
    first_name: string
    last_name: string
    email: string
    phone?: string
    role: string
    position?: string
    department?: string
    hire_date?: string
    status: string
    created_at: string
    national_id?: string
    bank_account_number?: string
}

export default function ViewEmployeePage() {
    const router = useRouter()
    const params = useParams()
    const id = params?.id as string
    const [employee, setEmployee] = useState<Employee | null>(null)
    const [loading, setLoading] = useState(true)

    useEffect(() => {
        const fetchEmployee = async () => {
            try {
                const response = await api.get(`/company/admin/employees/${id}`)
                setEmployee(response.data)
            } catch (error: any) {
                console.error("Failed to fetch employee:", error)
                toast.error("Failed to load employee details")
                router.push("/admin/employees")
            } finally {
                setLoading(false)
            }
        }

        if (id) {
            fetchEmployee()
        }
    }, [id, router])

    if (loading) {
        return (
            <div className="flex items-center justify-center h-full">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
            </div>
        )
    }

    if (!employee) {
        return (
            <div className="container mx-auto py-10 text-center">
                <h2 className="text-xl font-semibold mb-4">Employee not found</h2>
                <Link href="/admin/employees">
                    <Button>Back to Employees</Button>
                </Link>
            </div>
        )
    }

    const getRoleBadgeColor = (role: string) => {
        switch (role.toLowerCase()) {
            case "admin":
                return "bg-purple-100 text-purple-800 hover:bg-purple-200"
            case "hr":
                return "bg-blue-100 text-blue-800 hover:bg-blue-200"
            case "manager":
                return "bg-green-100 text-green-800 hover:bg-green-200"
            case "team_lead":
                return "bg-yellow-100 text-yellow-800 hover:bg-yellow-200"
            default:
                return "bg-gray-100 text-gray-800 hover:bg-gray-200"
        }
    }

    return (
        <div className="container mx-auto py-8 max-w-5xl">
            <div className="mb-6 flex items-center justify-between">
                <div className="flex items-center gap-4">
                    <Link href="/admin/employees">
                        <Button variant="ghost" size="sm">
                            <ArrowLeft className="mr-2 h-4 w-4" />
                            Back
                        </Button>
                    </Link>
                    <h1 className="text-2xl font-bold tracking-tight">Employee Details</h1>
                </div>
                <div className="flex gap-2">
                    <Link href={`/admin/employees/${employee.id}/edit`}>
                        <Button>
                            <Edit className="mr-2 h-4 w-4" />
                            Edit Employee
                        </Button>
                    </Link>
                </div>
            </div>

            <div className="grid gap-6 md:grid-cols-3">
                {/* Main Profile Card */}
                <Card className="md:col-span-1">
                    <CardHeader className="text-center">
                        <div className="w-24 h-24 bg-primary/10 rounded-full flex items-center justify-center mx-auto mb-4 text-2xl font-bold text-primary">
                            {employee.first_name[0]}{employee.last_name[0]}
                        </div>
                        <CardTitle className="text-xl">{employee.first_name} {employee.last_name}</CardTitle>
                        <CardDescription>{employee.email}</CardDescription>
                        <div className="mt-4 flex flex-wrap gap-2 justify-center">
                            <Badge className={getRoleBadgeColor(employee.role)} variant="secondary">
                                {employee.role.replace("_", " ").toUpperCase()}
                            </Badge>
                            <Badge variant={employee.status === 'active' ? 'default' : 'destructive'}>
                                {employee.status.toUpperCase()}
                            </Badge>
                        </div>
                    </CardHeader>
                    <CardContent>
                        <Separator className="my-4" />
                        <div className="space-y-4 text-sm">
                            <div className="flex items-center gap-3">
                                <Mail className="h-4 w-4 text-muted-foreground" />
                                <span>{employee.email}</span>
                            </div>
                            {employee.phone && (
                                <div className="flex items-center gap-3">
                                    <Phone className="h-4 w-4 text-muted-foreground" />
                                    <span>{employee.phone}</span>
                                </div>
                            )}
                            <div className="flex items-center gap-3">
                                <Calendar className="h-4 w-4 text-muted-foreground" />
                                <span>Joined {new Date(employee.created_at).toLocaleDateString()}</span>
                            </div>
                        </div>
                    </CardContent>
                </Card>

                {/* Details Section */}
                <div className="md:col-span-2 space-y-6">
                    <Card>
                        <CardHeader>
                            <CardTitle className="text-lg flex items-center gap-2">
                                <Briefcase className="h-5 w-5 text-primary" />
                                Employment Details
                            </CardTitle>
                        </CardHeader>
                        <CardContent className="grid gap-6 sm:grid-cols-2">
                            <div className="space-y-1">
                                <span className="text-sm text-muted-foreground">Job Title / Position</span>
                                <p className="font-medium">{employee.position || "Not assigned"}</p>
                            </div>
                            <div className="space-y-1">
                                <span className="text-sm text-muted-foreground">Department</span>
                                <p className="font-medium">{employee.department || "Not assigned"}</p>
                            </div>
                            <div className="space-y-1">
                                <span className="text-sm text-muted-foreground">Hire Date</span>
                                <p className="font-medium">
                                    {employee.hire_date ? new Date(employee.hire_date).toLocaleDateString() : "Not specified"}
                                </p>
                            </div>

                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle className="text-lg flex items-center gap-2">
                                <ShieldCheck className="h-5 w-5 text-primary" />
                                Personal & Sensitive Info
                            </CardTitle>
                            <CardDescription>
                                Protected information visible only to authorized admins.
                            </CardDescription>
                        </CardHeader>
                        <CardContent className="grid gap-6 sm:grid-cols-2">
                            <div className="space-y-1">
                                <span className="text-sm text-muted-foreground">National ID</span>
                                <p className="font-medium">{employee.national_id || "Not provided"}</p>
                            </div>
                            <div className="space-y-1">
                                <span className="text-sm text-muted-foreground">Bank Account</span>
                                <div className="flex items-center gap-2">
                                    <CreditCard className="h-4 w-4 text-muted-foreground" />
                                    <p className="font-medium">{employee.bank_account_number || "Not provided"}</p>
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    )
}


