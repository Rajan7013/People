"use client"

import { useEffect, useState } from "react"
import { ColumnDef } from "@tanstack/react-table"
import { Badge } from "@/components/ui/badge"
import { DataTable } from "@/components/ui/data-table"
import api from "@/lib/api"
import { format } from "date-fns"

export type AttendanceRecord = {
    id: string
    date: string
    employee_name: string
    check_in_time: string | null
    check_out_time: string | null
    total_hours: number | null
    status: string
    notes: string | null
}

interface AttendanceTableProps {
    role: "admin" | "hr" | "manager" | "team_lead" | "employee"
}

export function AttendanceTable({ role }: AttendanceTableProps) {
    const [data, setData] = useState<AttendanceRecord[]>([])
    const [loading, setLoading] = useState(true)

    useEffect(() => {
        const fetchData = async () => {
            setLoading(true)
            try {
                let endpoint = ""
                switch (role) {
                    case "admin":
                    case "hr":
                        endpoint = "/company/hr/attendance" // Use HR endpoint which returns paginated standard list
                        break
                    case "manager":
                        endpoint = "/company/manager/attendance"
                        break
                    case "team_lead":
                        endpoint = "/company/team-lead/attendance"
                        break
                    case "employee":
                        endpoint = "/company/employee/attendance"
                        break
                    default:
                        endpoint = "/company/employee/attendance"
                }

                if (endpoint) {
                    const response = await api.get(endpoint)
                    // Handle different response structures
                    let records = []
                    if (response.data.records) {
                        records = response.data.records
                    } else if (Array.isArray(response.data)) {
                        records = response.data
                    }

                    setData(records)
                }
            } catch (error) {
                console.error("Failed to fetch attendance:", error)
            } finally {
                setLoading(false)
            }
        }

        fetchData()
    }, [role])

    const columns: ColumnDef<AttendanceRecord>[] = [
        {
            accessorKey: "date",
            header: "Date",
            cell: ({ row }) => format(new Date(row.getValue("date")), "MMM dd, yyyy"),
        },
        // Only show Employee Name if not 'employee' role (i.e. viewing others)
        ...(role !== "employee" ? [{
            accessorKey: "employee_name",
            header: "Employee",
        }] : []),
        {
            accessorKey: "check_in_time",
            header: "Check In",
            cell: ({ row }) => {
                const val = row.getValue("check_in_time") as string
                return val ? format(new Date(val), "hh:mm a") : "-"
            },
        },
        {
            accessorKey: "check_out_time",
            header: "Check Out",
            cell: ({ row }) => {
                const val = row.getValue("check_out_time") as string
                return val ? format(new Date(val), "hh:mm a") : "-"
            },
        },
        {
            accessorKey: "total_hours",
            header: "Total Hrs",
            cell: ({ row }) => {
                const val = row.getValue("total_hours") as number
                return val ? val.toFixed(2) : "-"
            }
        },
        {
            accessorKey: "status",
            header: "Status",
            cell: ({ row }) => {
                const status = row.getValue("status") as string
                let variant: "default" | "secondary" | "destructive" | "outline" = "default"

                if (status === "late") variant = "destructive"
                else if (status === "present") variant = "default"
                else variant = "secondary"

                return (
                    <Badge variant={variant} className="capitalize">
                        {status}
                    </Badge>
                )
            },
        },
    ]

    if (loading) {
        return <div>Loading attendance records...</div>
    }

    return (
        <div className="bg-white rounded-md border shadow-sm">
            <div className="p-4 border-b">
                <h3 className="font-semibold text-lg">
                    {role === 'employee' ? 'My Attendance History' : 'Attendance Records'}
                </h3>
                <p className="text-sm text-muted-foreground">
                    {data.length} records in default view
                </p>
            </div>
            <DataTable columns={columns} data={data} />
        </div>
    )
}
