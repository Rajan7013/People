"use client";

import { useState } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import * as z from "zod";
import { toast } from "sonner";
import { Check, Copy, Plus, X } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { superAdminManagementService, type SuperAdminResponse } from "@/services/superAdminService";

const formSchema = z.object({
    email: z.string().email("Invalid email address"),
    first_name: z.string().min(2, "First name must be at least 2 characters"),
    last_name: z.string().min(2, "Last name must be at least 2 characters"),
});

type FormData = z.infer<typeof formSchema>;

interface AddSuperAdminDialogProps {
    onSuccess?: () => void;
}

export function AddSuperAdminDialog({ onSuccess }: AddSuperAdminDialogProps) {
    const [open, setOpen] = useState(false);
    const [loading, setLoading] = useState(false);
    const [createdAdmin, setCreatedAdmin] = useState<SuperAdminResponse | null>(null);
    const [passwordCopied, setPasswordCopied] = useState(false);

    const form = useForm<FormData>({
        resolver: zodResolver(formSchema),
        defaultValues: {
            email: "",
            first_name: "",
            last_name: "",
        },
    });

    const handleClose = () => {
        setOpen(false);
        setCreatedAdmin(null);
        setPasswordCopied(false);
        form.reset();
    };

    const onSubmit = async (data: FormData) => {
        try {
            setLoading(true);
            const response = await superAdminManagementService.createSuperAdmin(data);

            setCreatedAdmin(response);
            toast.success("Congratulations! Super Admin added successfully", {
                description: `${response.first_name} ${response.last_name} has been created.`,
            });

            if (onSuccess) {
                onSuccess();
            }
        } catch (error: any) {
            console.error("Failed to create super admin:", error);

            if (error.response?.status === 409) {
                toast.error("Email already exists", {
                    description: "This email is already registered in the system.",
                });
            } else {
                toast.error("Failed to create super admin", {
                    description: error.response?.data?.message || "An error occurred. Please try again.",
                });
            }
        } finally {
            setLoading(false);
        }
    };

    const copyPassword = () => {
        if (createdAdmin?.temporary_password) {
            navigator.clipboard.writeText(createdAdmin.temporary_password);
            setPasswordCopied(true);
            toast.success("Password copied to clipboard");

            setTimeout(() => setPasswordCopied(false), 2000);
        }
    };

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger asChild>
                <Button>
                    <Plus className="mr-2 h-4 w-4" />
                    Add Super Admin
                </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-[500px]">
                {!createdAdmin ? (
                    <>
                        <DialogHeader>
                            <DialogTitle>Add New Super Admin</DialogTitle>
                            <DialogDescription>
                                Create a new super admin account. A secure password will be generated automatically.
                            </DialogDescription>
                        </DialogHeader>

                        <Form {...form}>
                            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                                <FormField
                                    control={form.control}
                                    name="email"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>Email Address</FormLabel>
                                            <FormControl>
                                                <Input
                                                    type="email"
                                                    placeholder="admin@example.com"
                                                    {...field}
                                                    disabled={loading}
                                                />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />

                                <FormField
                                    control={form.control}
                                    name="first_name"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>First Name</FormLabel>
                                            <FormControl>
                                                <Input
                                                    placeholder="John"
                                                    {...field}
                                                    disabled={loading}
                                                />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />

                                <FormField
                                    control={form.control}
                                    name="last_name"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>Last Name</FormLabel>
                                            <FormControl>
                                                <Input
                                                    placeholder="Doe"
                                                    {...field}
                                                    disabled={loading}
                                                />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />

                                <DialogFooter>
                                    <Button
                                        type="button"
                                        variant="outline"
                                        onClick={handleClose}
                                        disabled={loading}
                                    >
                                        Cancel
                                    </Button>
                                    <Button type="submit" disabled={loading}>
                                        {loading ? (
                                            <>
                                                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                                                Creating...
                                            </>
                                        ) : (
                                            "Add Super Admin"
                                        )}
                                    </Button>
                                </DialogFooter>
                            </form>
                        </Form>
                    </>
                ) : (
                    <>
                        <DialogHeader>
                            <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-green-100 dark:bg-green-900">
                                <Check className="h-6 w-6 text-green-600 dark:text-green-400" />
                            </div>
                            <DialogTitle className="text-center">
                                Congratulations! Super Admin Added Successfully
                            </DialogTitle>
                            <DialogDescription className="text-center">
                                {createdAdmin.first_name} {createdAdmin.last_name} has been created as a super admin.
                            </DialogDescription>
                        </DialogHeader>

                        <div className="space-y-4 py-4">
                            <div className="rounded-lg border bg-muted/50 p-4 space-y-2">
                                <div className="flex justify-between items-center">
                                    <span className="text-sm font-medium">Email:</span>
                                    <span className="text-sm">{createdAdmin.email}</span>
                                </div>
                                <div className="flex justify-between items-center">
                                    <span className="text-sm font-medium">Name:</span>
                                    <span className="text-sm">
                                        {createdAdmin.first_name} {createdAdmin.last_name}
                                    </span>
                                </div>
                                <div className="flex justify-between items-center">
                                    <span className="text-sm font-medium">Role:</span>
                                    <span className="text-sm font-semibold text-primary">Super Admin</span>
                                </div>
                            </div>

                            <div className="rounded-lg border border-yellow-200 bg-yellow-50 dark:bg-yellow-900/20 dark:border-yellow-800 p-4 space-y-3">
                                <div className="flex items-start gap-2">
                                    <div className="rounded-full bg-yellow-100 dark:bg-yellow-900 p-1">
                                        <svg
                                            className="h-4 w-4 text-yellow-600 dark:text-yellow-400"
                                            fill="none"
                                            viewBox="0 0 24 24"
                                            stroke="currentColor"
                                        >
                                            <path
                                                strokeLinecap="round"
                                                strokeLinejoin="round"
                                                strokeWidth={2}
                                                d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
                                            />
                                        </svg>
                                    </div>
                                    <div className="flex-1">
                                        <p className="text-sm font-semibold text-yellow-800 dark:text-yellow-200">
                                            Temporary Password
                                        </p>
                                        <p className="text-xs text-yellow-700 dark:text-yellow-300 mt-1">
                                            Save this password - it won't be shown again!
                                        </p>
                                    </div>
                                </div>

                                <div className="flex items-center gap-2">
                                    <code className="flex-1 rounded bg-white dark:bg-zinc-950 px-3 py-2 text-sm font-mono border">
                                        {createdAdmin.temporary_password}
                                    </code>
                                    <Button
                                        size="sm"
                                        variant="outline"
                                        onClick={copyPassword}
                                        className="shrink-0"
                                    >
                                        {passwordCopied ? (
                                            <>
                                                <Check className="h-4 w-4 mr-1" />
                                                Copied
                                            </>
                                        ) : (
                                            <>
                                                <Copy className="h-4 w-4 mr-1" />
                                                Copy
                                            </>
                                        )}
                                    </Button>
                                </div>
                            </div>
                        </div>

                        <DialogFooter>
                            <Button onClick={handleClose} className="w-full">
                                Done
                            </Button>
                        </DialogFooter>
                    </>
                )}
            </DialogContent>
        </Dialog>
    );
}
