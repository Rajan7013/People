"use client";

import { useEffect, useState } from "react";
import { useAuth } from "@/context/auth-context";
import { departmentService, Department, CreateDepartmentDTO } from "@/services/departmentService";
import { Button } from "@/components/ui/button";
import { Plus, MoreHorizontal, Pencil, Trash2 } from "lucide-react";
import { DataTable } from "@/components/ui/data-table"; // Reuse existing
import { ColumnDef } from "@tanstack/react-table";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogFooter,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { toast } from "sonner";
import { Switch } from "@/components/ui/switch";

export default function DepartmentsPage() {
    const { user } = useAuth();
    const [departments, setDepartments] = useState<Department[]>([]);
    const [loading, setLoading] = useState(true);
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [editingDept, setEditingDept] = useState<Department | null>(null);
    const [formData, setFormData] = useState<CreateDepartmentDTO>({ name: "", description: "" });
    const [isActive, setIsActive] = useState(true); // For editing

    const fetchDepartments = async () => {
        if (!user?.tenant_id) return;
        try {
            const response = await departmentService.getAll();
            setDepartments(response.data || []);
        } catch (error) {
            console.error("Failed to fetch departments", error);
            toast.error("Failed to load departments");
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        if (user) {
            fetchDepartments();
        }
    }, [user]);

    const handleCreate = () => {
        setEditingDept(null);
        setFormData({ name: "", description: "" });
        setIsActive(true);
        setIsDialogOpen(true);
    };

    const handleEdit = (dept: Department) => {
        setEditingDept(dept);
        setFormData({ name: dept.name, description: dept.description || "" });
        setIsActive(dept.is_active);
        setIsDialogOpen(true);
    };

    const handleDelete = async (id: string) => {
        if (!confirm("Are you sure you want to delete this department?")) return;
        try {
            await departmentService.delete(id);
            toast.success("Department deleted successfully");
            fetchDepartments();
        } catch (error) {
            toast.error("Failed to delete department");
        }
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            if (editingDept) {
                await departmentService.update(editingDept.id, {
                    ...formData,
                    is_active: isActive,
                });
                toast.success("Department updated successfully");
            } else {
                await departmentService.create(formData);
                toast.success("Department created successfully");
            }
            setIsDialogOpen(false);
            fetchDepartments();
        } catch (error: any) {
            toast.error(error.response?.data?.error || "Operation failed");
        }
    };

    const columns: ColumnDef<Department>[] = [
        {
            accessorKey: "name",
            header: "Name",
        },
        {
            accessorKey: "description",
            header: "Description",
        },
        {
            accessorKey: "is_active",
            header: "Status",
            cell: ({ row }) => {
                const active = row.getValue("is_active") as boolean;
                return (
                    <span className={`px-2 py-1 rounded-full text-xs font-semibold ${active ? "bg-green-100 text-green-800" : "bg-gray-100 text-gray-800"}`}>
                        {active ? "Active" : "Inactive"}
                    </span>
                );
            },
        },
        {
            id: "actions",
            cell: ({ row }) => {
                const dept = row.original;
                return (
                    <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                            <Button variant="ghost" className="h-8 w-8 p-0">
                                <span className="sr-only">Open menu</span>
                                <MoreHorizontal className="h-4 w-4" />
                            </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                            <DropdownMenuItem onClick={() => handleEdit(dept)}>
                                <Pencil className="mr-2 h-4 w-4" /> Edit
                            </DropdownMenuItem>
                            <DropdownMenuItem onClick={() => handleDelete(dept.id)} className="text-red-600">
                                <Trash2 className="mr-2 h-4 w-4" /> Delete
                            </DropdownMenuItem>
                        </DropdownMenuContent>
                    </DropdownMenu>
                );
            },
        },
    ];

    return (
        <div className="container mx-auto py-10">
            <div className="flex justify-between items-center mb-6">
                <h2 className="text-3xl font-bold tracking-tight">Departments</h2>
                <Button onClick={handleCreate}>
                    <Plus className="mr-2 h-4 w-4" /> Add Department
                </Button>
            </div>

            {loading ? (
                <div>Loading departments...</div>
            ) : (
                <DataTable columns={columns} data={departments} />
            )}

            <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>{editingDept ? "Edit Department" : "Add Department"}</DialogTitle>
                    </DialogHeader>
                    <form onSubmit={handleSubmit} className="space-y-4">
                        <div>
                            <Label htmlFor="name">Name</Label>
                            <Input
                                id="name"
                                value={formData.name}
                                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                required
                            />
                        </div>
                        <div>
                            <Label htmlFor="description">Description (Optional)</Label>
                            <Textarea
                                id="description"
                                value={formData.description}
                                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                            />
                        </div>

                        {editingDept && (
                            <div className="flex items-center gap-2">
                                <Switch
                                    id="is_active"
                                    checked={isActive}
                                    onCheckedChange={setIsActive}
                                />
                                <Label htmlFor="is_active">Active Status</Label>
                            </div>
                        )}

                        <DialogFooter>
                            <Button type="button" variant="outline" onClick={() => setIsDialogOpen(false)}>
                                Cancel
                            </Button>
                            <Button type="submit">
                                {editingDept ? "Update" : "Create"}
                            </Button>
                        </DialogFooter>
                    </form>
                </DialogContent>
            </Dialog>
        </div>
    );
}
