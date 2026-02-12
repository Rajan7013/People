"use client"

import { useEffect, useState } from "react"
import { ColumnDef } from "@tanstack/react-table"
import { Badge } from "@/components/ui/badge"
import { DataTable } from "@/components/ui/data-table"
import { Button } from "@/components/ui/button" // Assuming we might add approve actions later
import api from "@/lib/api"
import { format } from "date-fns"

export type LeaveRequest = {
    id: string
    leave_type_name: string
    start_date: string
    end_date: string
    reason: string
    status: string
    employee_name?: string
    days_count: number
}

interface LeaveRequestsTableProps {
    role: "admin" | "hr" | "manager" | "team_lead" | "employee"
}

export function LeaveRequestsTable({ role }: LeaveRequestsTableProps) {
    const [data, setData] = useState<LeaveRequest[]>([])
    const [loading, setLoading] = useState(true)

    useEffect(() => {
        const fetchData = async () => {
            setLoading(true)
            try {
                let endpoint = ""
                switch (role) {
                    case "admin":
                    case "hr":
                        endpoint = "/company/hr/leaves"
                        break
                    case "manager":
                        endpoint = "/company/manager/leaves"
                        break
                    case "employee":
                        endpoint = "/company/employee/leaves"
                        break
                    default:
                        // Team Lead not supported in basic plan for leaves yet, or fallback to employee view
                        endpoint = "/company/employee/leaves"
                }

                if (endpoint) {
                    const response = await api.get(endpoint)
                    // Adjust based on actual API response (array vs object)
                    const requests = Array.isArray(response.data) ? response.data : (response.data.data || [])
                    setData(requests)
                }
            } catch (error) {
                console.error("Failed to fetch leaves:", error)
            } finally {
                setLoading(false)
            }
        }

        fetchData()
    }, [role])

    const columns: ColumnDef<LeaveRequest>[] = [
        // Show Employee Name for Approvers
        ...((role === "admin" || role === "hr" || role === "manager") ? [{
            accessorKey: "employee_name",
            header: "Employee",
        }] : []),
        {
            accessorKey: "leave_type_name",
            header: "Type",
        },
        {
            accessorKey: "start_date",
            header: "Start Date",
            cell: ({ row }) => format(new Date(row.getValue("start_date")), "MMM dd, yyyy"),
        },
        {
            accessorKey: "end_date",
            header: "End Date",
            cell: ({ row }) => format(new Date(row.getValue("end_date")), "MMM dd, yyyy"),
        },
        {
            accessorKey: "days_count",
            header: "Days",
        },
        {
            accessorKey: "status",
            header: "Status",
            cell: ({ row }) => {
                const status = row.getValue("status") as string
                let variant: "default" | "secondary" | "destructive" | "outline" = "default"

                switch (status) {
                    case 'approved': variant = 'default'; break; // green-ish in default theme usually
                    case 'rejected': variant = 'destructive'; break;
                    case 'pending': variant = 'secondary'; break;
                    default: variant = 'outline';
                }

                return (
                    <Badge variant={variant} className="capitalize">
                        {status}
                    </Badge>
                )
            },
        },
    ]

    if (loading) {
        return <div>Loading leave requests...</div>
    }

    return (
        <div className="bg-white rounded-md border shadow-sm">
            <div className="p-4 border-b">
                <h3 className="font-semibold text-lg">
                    {role === 'employee' ? 'My Leave Requests' : 'Leave Requests'}
                </h3>
            </div>
            <DataTable columns={columns} data={data} />
        </div>
    )
}
