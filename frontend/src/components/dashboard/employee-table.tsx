"use client"

import { useEffect, useState } from "react"
import { ColumnDef } from "@tanstack/react-table"
import { Badge } from "@/components/ui/badge"
import { DataTable } from "@/components/ui/data-table"
import api from "@/lib/api"
import { format } from "date-fns"

// Define the shape of our data (simplified for shared view)
export type Employee = {
    id: string
    first_name: string
    last_name: string
    email: string
    role: string
    position?: string
    department?: string
    status: string
    phone?: string
    hire_date?: string
}

interface EmployeeTableProps {
    role: "admin" | "hr" | "manager" | "team_lead" | "employee"
}

export function EmployeeTable({ role }: EmployeeTableProps) {
    const [data, setData] = useState<Employee[]>([])
    const [loading, setLoading] = useState(true)

    useEffect(() => {
        const fetchData = async () => {
            setLoading(true)
            try {
                let endpoint = ""
                // Determine endpoint based on role
                switch (role) {
                    case "admin":
                        endpoint = "/company/admin/employees"
                        break
                    case "hr":
                        endpoint = "/company/hr/employees"
                        break
                    case "manager":
                        endpoint = "/company/manager/team"
                        break
                    case "team_lead":
                        endpoint = "/company/team-lead/team"
                        break
                    default:
                        endpoint = "/company/employee/team" // Fallback? Or empty
                        break
                }

                if (endpoint) {
                    const response = await api.get(endpoint)
                    // Handle both array response and wrapped response if any
                    const employees = Array.isArray(response.data) ? response.data : []

                    // normalize data if needed
                    const normalized = employees.map((emp: any) => ({
                        id: emp.id,
                        first_name: emp.first_name,
                        last_name: emp.last_name,
                        email: emp.email,
                        role: emp.role,
                        position: emp.position || emp.job_title || "N/A", // API might return job_title
                        department: emp.department || "N/A",
                        status: emp.status || emp.employment_status || "active",
                        phone: emp.phone,
                        hire_date: emp.hire_date || emp.date_of_joining
                    }))
                    setData(normalized)
                }
            } catch (error) {
                console.error("Failed to fetch employees:", error)
            } finally {
                setLoading(false)
            }
        }

        fetchData()
    }, [role])

    // Define columns dynamically or statically with conditions
    const columns: ColumnDef<Employee>[] = [
        {
            accessorKey: "first_name",
            header: "Name",
            cell: ({ row }) => `${row.original.first_name} ${row.original.last_name}`,
        },
        {
            accessorKey: "email",
            header: "Email",
        },
        {
            accessorKey: "role",
            header: "Role",
            cell: ({ row }) => {
                const roleStr = row.getValue("role") as string
                return (
                    <Badge variant="outline" className="capitalize">
                        {roleStr.replace("_", " ")}
                    </Badge>
                )
            },
        },
        {
            accessorKey: "department",
            header: "Department",
        },
        {
            accessorKey: "position",
            header: "Position",
        },
        {
            accessorKey: "status",
            header: "Status",
            cell: ({ row }) => {
                const status = row.getValue("status") as string
                return (
                    <Badge
                        variant={status === "active" ? "default" : "secondary"}
                        className={status === "active" ? "bg-green-600 hover:bg-green-700" : ""}
                    >
                        {status}
                    </Badge>
                )
            }
        },
    ]

    // Add extra columns for HR/Admin if needed (e.g. actions)
    // For now, keeping it read-only for the shared component as requested for Manager/Top views

    if (loading) {
        return <div>Loading team data...</div>
    }

    return (
        <div className="bg-white rounded-md border shadow-sm">
            <div className="p-4 border-b">
                <h3 className="font-semibold text-lg">
                    {role === 'manager' || role === 'team_lead' ? 'My Team' : 'Employee Directory'}
                </h3>
                <p className="text-sm text-muted-foreground">
                    {data.length} records found
                </p>
            </div>
            <DataTable columns={columns} data={data} />
        </div>
    )
}
