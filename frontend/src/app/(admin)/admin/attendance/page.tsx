"use client";

import { useState, useEffect } from "react";
import { format } from "date-fns";
import { Calendar as CalendarIcon, Search, Filter, Download } from "lucide-react";
import { toast } from "sonner";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
    Popover,
    PopoverContent,
    PopoverTrigger,
} from "@/components/ui/popover";
import { Calendar } from "@/components/ui/calendar";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import api from "@/lib/api";

interface AttendanceRecord {
    id: string;
    employee_id: string;
    employee_name?: string; // Backend should enrich this ideally, or we join
    role?: string;
    check_in_time: string;
    check_out_time?: string;
    status: string;
    total_hours?: number;
    date: string;
}

export default function AdminAttendancePage() {
    const [records, setRecords] = useState<AttendanceRecord[]>([]);
    const [loading, setLoading] = useState(true);
    const [date, setDate] = useState<Date | undefined>(new Date());
    const [filterStatus, setFilterStatus] = useState("all");
    const [searchQuery, setSearchQuery] = useState("");

    // Fetch attendance data
    const fetchAttendance = async () => {
        try {
            setLoading(true);
            const queryParams = new URLSearchParams();
            if (date) {
                const dateStr = format(date, "yyyy-MM-dd");
                queryParams.append("start_date", dateStr);
                queryParams.append("end_date", dateStr); // Single day view for now
            }

            if (filterStatus && filterStatus !== "all") {
                queryParams.append("status", filterStatus);
            }

            if (searchQuery) {
                queryParams.append("search", searchQuery);
            }

            const response = await api.get(`/company/hr/attendance?${queryParams.toString()}`);
            setRecords(response.data.records || []);
        } catch (error) {
            console.error("Failed to fetch attendance:", error);
            toast.error("Failed to load attendance records");
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        // Debounce search
        const timer = setTimeout(() => {
            fetchAttendance();
        }, 500);

        return () => clearTimeout(timer);
    }, [date, filterStatus, searchQuery]);

    const getStatusBadge = (status: string) => {
        switch (status.toLowerCase()) {
            case "present":
                return <Badge className="bg-green-100 text-green-800 hover:bg-green-100">Present</Badge>;
            case "absent":
                return <Badge className="bg-red-100 text-red-800 hover:bg-red-100">Absent</Badge>;
            case "late":
                return <Badge className="bg-yellow-100 text-yellow-800 hover:bg-yellow-100">Late</Badge>;
            case "half_day":
                return <Badge className="bg-orange-100 text-orange-800 hover:bg-orange-100">Half Day</Badge>;
            default:
                return <Badge variant="outline">{status}</Badge>;
        }
    };

    return (
        <div className="space-y-6">
            <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
                <div>
                    <h2 className="text-3xl font-bold tracking-tight">Attendance</h2>
                    <p className="text-muted-foreground">
                        Monitor employee attendance and work hours.
                    </p>
                </div>
                <div className="flex gap-2">
                    <Button variant="outline" size="sm">
                        <Download className="mr-2 h-4 w-4" />
                        Export
                    </Button>
                </div>
            </div>

            {/* Filters */}
            <div className="flex flex-col sm:flex-row gap-4 items-center justify-between bg-white dark:bg-slate-900 p-4 rounded-lg border border-gray-200 dark:border-slate-800 shadow-sm">
                <div className="flex items-center gap-4 w-full flex-wrap">
                    {/* Search Input */}
                    <div className="relative w-full sm:w-auto">
                        <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                        <Input
                            type="search"
                            placeholder="Search employee..."
                            className="pl-9 w-full sm:w-[250px]"
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                        />
                    </div>

                    <div className="flex items-center gap-2">
                        <Popover>
                            <PopoverTrigger asChild>
                                <Button
                                    variant={"outline"}
                                    className={`w-[240px] justify-start text-left font-normal ${!date && "text-muted-foreground"}`}
                                >
                                    <CalendarIcon className="mr-2 h-4 w-4" />
                                    {date ? format(date, "PPP") : <span>Pick a date</span>}
                                </Button>
                            </PopoverTrigger>
                            <PopoverContent className="w-auto p-0" align="start">
                                <Calendar
                                    mode="single"
                                    selected={date}
                                    onSelect={setDate}
                                    initialFocus
                                />
                            </PopoverContent>
                        </Popover>
                    </div>

                    <div className="flex items-center gap-2">
                        <Select value={filterStatus} onValueChange={setFilterStatus}>
                            <SelectTrigger className="w-[180px]">
                                <SelectValue placeholder="Filter by status" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="all">All Status</SelectItem>
                                <SelectItem value="present">Present</SelectItem>
                                <SelectItem value="absent">Absent</SelectItem>
                                <SelectItem value="late">Late</SelectItem>
                            </SelectContent>
                        </Select>
                    </div>
                </div>
            </div>

            {/* Table */}
            <div className="border rounded-md">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead>Employee</TableHead>
                            <TableHead>Role</TableHead>
                            <TableHead>Date</TableHead>
                            <TableHead>Check In</TableHead>
                            <TableHead>Check Out</TableHead>
                            <TableHead>Work Hours</TableHead>
                            <TableHead>Status</TableHead>
                            {/* <TableHead className="text-right">Actions</TableHead> */}
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {loading ? (
                            <TableRow>
                                <TableCell colSpan={6} className="text-center h-24">
                                    Loading...
                                </TableCell>
                            </TableRow>
                        ) : records.length === 0 ? (
                            <TableRow>
                                <TableCell colSpan={6} className="text-center h-24 text-muted-foreground">
                                    No attendance records found for this date.
                                </TableCell>
                            </TableRow>
                        ) : (
                            records.map((record) => (
                                <TableRow key={record.id}>
                                    <TableCell className="font-medium">
                                        {record.employee_name || "Unknown Employee"}
                                        {/* We might need to fetch employee details if name isn't in record */}
                                    </TableCell>
                                    <TableCell className="capitalize">{record.role || "-"}</TableCell>
                                    <TableCell>{format(new Date(record.date || record.check_in_time), "MMM dd, yyyy")}</TableCell>
                                    <TableCell>
                                        {record.check_in_time ? format(new Date(record.check_in_time), "hh:mm a") : "-"}
                                    </TableCell>
                                    <TableCell>
                                        {record.check_out_time ? format(new Date(record.check_out_time), "hh:mm a") : "-"}
                                    </TableCell>
                                    <TableCell>
                                        {record.total_hours ? `${record.total_hours.toFixed(1)} hrs` : "-"}
                                    </TableCell>
                                    <TableCell>
                                        {getStatusBadge(record.status)}
                                    </TableCell>
                                    {/* <TableCell className="text-right">
                                        <Button variant="ghost" size="sm">Edit</Button>
                                    </TableCell> */}
                                </TableRow>
                            ))
                        )}
                    </TableBody>
                </Table>
            </div>
        </div>
    );
}
