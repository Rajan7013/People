"use client"

import { ColumnDef } from "@tanstack/react-table"
import { MoreHorizontal } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Switch } from "@/components/ui/switch"
import Link from "next/link"
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { toast } from "sonner"
import api from "@/lib/api"

// Define the shape of our data.
// You can use a Zod schema here if you want.
export type Employee = {
    id: string
    first_name: string
    last_name: string
    email: string
    role: string
    position?: string
    department?: string
    status: "active" | "inactive" | "suspended" | "terminated"
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

const toggleEmployeeStatus = async (employeeId: string, currentStatus: string) => {
    try {
        const newIsActive = currentStatus !== "active"
        const newEmploymentStatus = newIsActive ? "active" : "suspended"

        await api.put(`/company/admin/employees/${employeeId}/status`, {
            is_active: newIsActive,
            employment_status: newEmploymentStatus,
        })

        toast.success(
            newIsActive
                ? "Employee activated successfully"
                : "Employee suspended successfully"
        )

        // Reload the page to reflect changes
        window.location.reload()
    } catch (error) {
        console.error("Failed to update employee status:", error)
        toast.error("Failed to update employee status")
    }
}

export const columns: ColumnDef<Employee>[] = [
    {
        accessorKey: "first_name",
        header: "First Name",
    },
    {
        accessorKey: "last_name",
        header: "Last Name",
    },
    {
        accessorKey: "email",
        header: "Email",
    },
    {
        accessorKey: "role",
        header: "Role",
        cell: ({ row }) => {
            const role = row.getValue("role") as string
            return (
                <Badge className={getRoleBadgeColor(role)} variant="secondary">
                    {role.replace("_", " ").toUpperCase()}
                </Badge>
            )
        },
    },
    {
        accessorKey: "position",
        header: "Position",
    },
    {
        accessorKey: "department",
        header: "Department",
    },
    {
        accessorKey: "status",
        header: "Status",
        cell: ({ row }) => {
            const status = row.getValue("status") as string
            const employee = row.original
            const isActive = status === "active"

            return (
                <div className="flex items-center gap-2">
                    <Switch
                        checked={isActive}
                        onCheckedChange={() => toggleEmployeeStatus(employee.id, status)}
                        aria-label="Toggle employee status"
                    />
                    <span className={`text-sm ${isActive ? 'text-green-600 font-medium' : 'text-gray-500'}`}>
                        {isActive ? "Active" : "Suspended"}
                    </span>
                </div>
            )
        },
    },
    {
        id: "actions",
        cell: ({ row }) => {
            const employee = row.original

            return (
                <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                        <Button variant="ghost" className="h-8 w-8 p-0">
                            <span className="sr-only">Open menu</span>
                            <MoreHorizontal className="h-4 w-4" />
                        </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">


                        <Link href={`/admin/employees/${employee.id}`}>
                            <DropdownMenuItem>View details</DropdownMenuItem>
                        </Link>
                        <Link href={`/admin/employees/${employee.id}/edit`}>
                            <DropdownMenuItem>Edit employee</DropdownMenuItem>
                        </Link>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                            className="text-red-600 focus:text-red-500"
                            onClick={async () => {
                                if (confirm("Are you sure you want to delete this employee? This action cannot be undone.")) {
                                    try {
                                        await api.delete(`/company/admin/employees/${employee.id}`)
                                        toast.success("Employee deleted successfully")
                                        window.location.reload()
                                    } catch (error) {
                                        console.error("Failed to delete employee:", error)
                                        toast.error("Failed to delete employee")
                                    }
                                }
                            }}
                        >
                            Delete employee
                        </DropdownMenuItem>
                    </DropdownMenuContent>
                </DropdownMenu>
            )
        },
    },
]
