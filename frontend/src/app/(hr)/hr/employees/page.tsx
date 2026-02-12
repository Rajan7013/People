"use client";

import { useEffect, useState } from "react";
import { Employee, columns } from "./columns";
import { DataTable } from "@/components/ui/data-table";
import api from "@/lib/api";
import { useAuth } from "@/context/auth-context";

export default function HREmployeesPage() {
    const [data, setData] = useState<Employee[]>([]);
    const [filteredData, setFilteredData] = useState<Employee[]>([]);
    const [loading, setLoading] = useState(true);
    const { user } = useAuth();

    // Filtering state
    const [departments, setDepartments] = useState<{ id: string, name: string }[]>([]);
    const [selectedDept, setSelectedDept] = useState<string>("all");
    const [selectedRole, setSelectedRole] = useState<string>("all");
    const [selectedStatus, setSelectedStatus] = useState<string>("all");

    useEffect(() => {
        if (!user?.tenant_id) return;

        const loadData = async () => {
            setLoading(true);
            try {
                // Fetch Employees (HR Endpoint)
                const empResponse = await api.get(`/company/hr/employees`);
                const employees = (empResponse.data || []).map((emp: any) => ({
                    id: emp.id,
                    first_name: emp.first_name,
                    last_name: emp.last_name,
                    email: emp.email,
                    role: emp.role,
                    position: emp.position || "N/A",
                    department: emp.department || "N/A",
                    status: emp.status || "active",
                }));
                setData(employees);
                setFilteredData(employees);

                // Fetch Departments (HR Endpoint)
                const deptResponse = await api.get(`/company/hr/departments`);
                setDepartments(deptResponse.data || []);

            } catch (error) {
                console.error("Failed to fetch data", error);
            } finally {
                setLoading(false);
            }
        };

        loadData();
    }, [user]);

    // Handle filter changes
    useEffect(() => {
        let filtered = data;

        // Filter by department
        if (selectedDept !== "all") {
            const dept = departments.find(d => d.id === selectedDept);
            if (dept) {
                filtered = filtered.filter(emp => emp.department === dept.name);
            }
        }

        // Filter by role
        if (selectedRole !== "all") {
            filtered = filtered.filter(emp => emp.role === selectedRole);
        }

        // Filter by status
        if (selectedStatus !== "all") {
            filtered = filtered.filter(emp => emp.status === selectedStatus);
        }

        setFilteredData(filtered);
    }, [selectedDept, selectedRole, selectedStatus, data, departments]);

    return (
        <div className="container mx-auto py-10">
            <div className="flex justify-between items-center mb-6">
                <div className="space-y-1">
                    <h2 className="text-3xl font-bold tracking-tight">Employee Directory</h2>
                    <p className="text-muted-foreground">
                        Manage employee records and details.
                    </p>
                </div>

                {/* HR cannot add employees, so no Add button here */}
            </div>

            <div className="flex gap-4 mb-6 flex-wrap">
                {/* Department Filter */}
                <select
                    className="h-10 w-[180px] rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
                    value={selectedDept}
                    onChange={(e) => setSelectedDept(e.target.value)}
                >
                    <option value="all">All Departments</option>
                    {departments.map(d => (
                        <option key={d.id} value={d.id}>{d.name}</option>
                    ))}
                </select>

                {/* Role Filter */}
                <select
                    className="h-10 w-[150px] rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
                    value={selectedRole}
                    onChange={(e) => setSelectedRole(e.target.value)}
                >
                    <option value="all">All Roles</option>
                    <option value="manager">Manager</option>
                    <option value="team_lead">Team Lead</option>
                    <option value="employee">Employee</option>
                </select>

                {/* Status Filter */}
                <select
                    className="h-10 w-[150px] rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
                    value={selectedStatus}
                    onChange={(e) => setSelectedStatus(e.target.value)}
                >
                    <option value="all">All Statuses</option>
                    <option value="active">Active</option>
                    <option value="suspended">Suspended</option>
                    <option value="inactive">Inactive</option>
                    <option value="terminated">Terminated</option>
                </select>
            </div>

            {loading ? (
                <div className="flex justify-center p-8">Loading directory...</div>
            ) : (
                <div className="bg-white dark:bg-zinc-950 rounded-md border">
                    <DataTable columns={columns} data={filteredData} />
                </div>
            )}
        </div>
    );
}
