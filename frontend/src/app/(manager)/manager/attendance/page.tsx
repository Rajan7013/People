"use client"

import { useState, useEffect } from "react"
import { format, subDays } from "date-fns"
import { Calendar as CalendarIcon, Download } from "lucide-react"
import { DateRange } from "react-day-picker"

import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Calendar } from "@/components/ui/calendar"
import {
    Popover,
    PopoverContent,
    PopoverTrigger,
} from "@/components/ui/popover"
import { DataTable } from "@/components/ui/data-table"
import { Badge } from "@/components/ui/badge"
import { ColumnDef } from "@tanstack/react-table"
import api from "@/lib/api"
import { toast } from "sonner"

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

const columns: ColumnDef<AttendanceRecord>[] = [
    {
        accessorKey: "employee_name",
        header: "Team Member",
        cell: ({ row }) => <span className="font-medium">{row.getValue("employee_name")}</span>
    },
    {
        accessorKey: "date",
        header: "Date",
        cell: ({ row }) => format(new Date(row.getValue("date")), "MMM dd, yyyy"),
    },
    {
        accessorKey: "check_in_time",
        header: "In",
        cell: ({ row }) => {
            const val = row.getValue("check_in_time") as string
            return val ? format(new Date(val), "hh:mm a") : "-"
        },
    },
    {
        accessorKey: "check_out_time",
        header: "Out",
        cell: ({ row }) => {
            const val = row.getValue("check_out_time") as string
            return val ? format(new Date(val), "hh:mm a") : "-"
        },
    },
    {
        accessorKey: "total_hours",
        header: "Hrs",
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

export default function ManagerAttendancePage() {
    const [date, setDate] = useState<DateRange | undefined>({
        from: subDays(new Date(), 7),
        to: new Date(),
    })
    const [data, setData] = useState<AttendanceRecord[]>([])
    const [loading, setLoading] = useState(true)

    const [error, setError] = useState<string | null>(null)

    const fetchData = async () => {
        if (!date?.from) return

        setLoading(true)
        setError(null)
        try {
            const fromStr = format(date.from, "yyyy-MM-dd")
            const toStr = date.to ? format(date.to, "yyyy-MM-dd") : fromStr

            // Manager endpoint fetches department attendance
            const response = await api.get(`/company/manager/attendance?start_date=${fromStr}&end_date=${toStr}`)

            let records = []
            if (response.data.records) {
                records = response.data.records
            } else if (Array.isArray(response.data)) {
                records = response.data
            }
            setData(records)
        } catch (error: any) {
            console.error("Failed to fetch attendance:", error)
            // Show specific backend error message if available
            const errorMsg = error.response?.data || error.message || "Failed to fetch attendance records"

            if (typeof errorMsg === 'string' && errorMsg.includes("User does not belong to a department")) {
                setError("NoDepartment")
            } else {
                toast.error(typeof errorMsg === 'string' ? errorMsg : "Failed to fetch attendance records")
            }
        } finally {
            setLoading(false)
        }
    }

    useEffect(() => {
        fetchData()
    }, [date])

    const handleExport = () => {
        if (!data.length) {
            toast.error("No data to export")
            return
        }

        const headers = ["Employee", "Date", "Check In", "Check Out", "Total Hours", "Status"]
        const csvContent = [
            headers.join(","),
            ...data.map(row => [
                `"${row.employee_name}"`,
                format(new Date(row.date), "yyyy-MM-dd"),
                row.check_in_time ? format(new Date(row.check_in_time), "HH:mm") : "",
                row.check_out_time ? format(new Date(row.check_out_time), "HH:mm") : "",
                row.total_hours || 0,
                row.status
            ].join(","))
        ].join("\n")

        const blob = new Blob([csvContent], { type: "text/csv;charset=utf-8;" })
        const url = URL.createObjectURL(blob)
        const link = document.createElement("a")
        link.setAttribute("href", url)
        link.setAttribute("download", `department_attendance_${format(new Date(), "yyyy-MM-dd")}.csv`)
        link.style.visibility = "hidden"
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
    }

    if (error === "NoDepartment") {
        return (
            <div className="container mx-auto py-8">
                <h2 className="text-3xl font-bold tracking-tight mb-4">Department Attendance</h2>
                <div className="flex flex-col items-center justify-center p-12 border border-dashed rounded-lg bg-muted/50">
                    <div className="h-12 w-12 rounded-full bg-yellow-100 dark:bg-yellow-900/20 flex items-center justify-center mb-4">
                        <CalendarIcon className="h-6 w-6 text-yellow-600 dark:text-yellow-400" />
                    </div>
                    <h3 className="text-lg font-semibold mb-2">No Department Assigned</h3>
                    <p className="text-muted-foreground text-center max-w-sm">
                        You do not currently belong to any department. Please contact your administrator to be assigned to a department to view attendance records.
                    </p>
                </div>
            </div>
        )
    }

    return (
        <div className="container mx-auto py-8">
            <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 mb-8">
                <div>
                    <h2 className="text-3xl font-bold tracking-tight">Department Attendance</h2>
                    <p className="text-muted-foreground">Monitor attendance for your department members.</p>
                </div>
                <div className="flex items-center gap-2">
                    <Button variant="outline" onClick={handleExport}>
                        <Download className="mr-2 h-4 w-4" />
                        Export CSV
                    </Button>
                </div>
            </div>

            <div className="flex gap-4 mb-6">
                <div className="grid gap-2">
                    <Popover>
                        <PopoverTrigger asChild>
                            <Button
                                id="date"
                                variant={"outline"}
                                className={cn(
                                    "w-[300px] justify-start text-left font-normal",
                                    !date && "text-muted-foreground"
                                )}
                            >
                                <CalendarIcon className="mr-2 h-4 w-4" />
                                {date?.from ? (
                                    date.to ? (
                                        <>
                                            {format(date.from, "LLL dd, y")} -{" "}
                                            {format(date.to, "LLL dd, y")}
                                        </>
                                    ) : (
                                        format(date.from, "LLL dd, y")
                                    )
                                ) : (
                                    <span>Pick a date range</span>
                                )}
                            </Button>
                        </PopoverTrigger>
                        <PopoverContent className="w-auto p-0" align="start">
                            <Calendar
                                initialFocus
                                mode="range"
                                defaultMonth={date?.from}
                                selected={date}
                                onSelect={setDate}
                                numberOfMonths={1}
                            />
                        </PopoverContent>
                    </Popover>
                </div>
            </div>

            <div className="bg-white dark:bg-zinc-950 rounded-md border">
                {loading ? (
                    <div className="p-8 text-center">Loading attendance data...</div>
                ) : (
                    <DataTable columns={columns} data={data} />
                )}
            </div>
        </div>
    )
}
