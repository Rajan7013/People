"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/context/auth-context";
import { Loader2 } from "lucide-react";

export default function DashboardRedirect() {
    const { user, isLoading } = useAuth();
    const router = useRouter();

    useEffect(() => {
        if (isLoading) return;

        if (!user) {
            router.push("/login");
            return;
        }

        // Redirect based on role
        switch (user.role) {
            case "super_admin":
                router.push("/super-admin/dashboard");
                break;
            case "admin":
                router.push("/admin/dashboard");
                break;
            case "hr":
                router.push("/hr/dashboard");
                break;
            case "manager":
                router.push("/manager/dashboard");
                break;
            case "team_lead":
                router.push("/team-lead/dashboard");
                break;
            case "employee":
                router.push("/employee/dashboard");
                break;
            default:
                router.push("/login");
        }
    }, [user, isLoading, router]);

    return (
        <div className="flex items-center justify-center min-h-screen">
            <div className="text-center">
                <Loader2 className="w-8 h-8 animate-spin mx-auto mb-4 text-blue-600" />
                <p className="text-gray-600">Redirecting to your dashboard...</p>
            </div>
        </div>
    );
}
