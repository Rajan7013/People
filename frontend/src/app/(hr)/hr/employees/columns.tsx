"use client"

import { ColumnDef } from "@tanstack/react-table"
import { MoreHorizontal } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import Link from "next/link"
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

// Define the shape of our data.
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
            const isActive = status === "active"

            return (
                <div className="flex items-center gap-2">
                    <span className={`text-sm ${isActive ? 'text-green-600 font-medium' : 'text-gray-500'}`}>
                        {isActive ? "Active" : status.charAt(0).toUpperCase() + status.slice(1)}
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
                        <Link href={`/hr/employees/${employee.id}`}>
                            <DropdownMenuItem>View details</DropdownMenuItem>
                        </Link>
                        <Link href={`/hr/employees/${employee.id}/edit`}>
                            <DropdownMenuItem>Edit employee</DropdownMenuItem>
                        </Link>
                    </DropdownMenuContent>
                </DropdownMenu>
            )
        },
    },
]
