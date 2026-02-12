"use client";

import { useAuth } from "@/context/auth-context";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { House, Users, CalendarCheck, FileText, LogOut, Settings, CreditCard, AlignJustify } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { ThemeToggle } from "@/components/theme-toggle";

export default function AdminLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    const { user, logout, isLoading, isAuthenticated } = useAuth();
    const pathname = usePathname();
    const router = useRouter();
    const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

    // Protect the route - only allow 'admin'
    useEffect(() => {
        if (!isLoading) {
            if (!isAuthenticated) {
                router.push("/login");
            } else if (user?.role !== 'admin') {
                // Redirect based on their actual role if they try to access admin
                if (user?.role === 'super_admin') router.push('/super-admin/dashboard');
                else if (user?.role === 'hr') router.push('/hr/dashboard');
                else if (user?.role === 'manager') router.push('/manager/dashboard');
                else if (user?.role === 'team_lead') router.push('/team-lead/dashboard');
                else router.push('/dashboard');
            }
        }
    }, [isLoading, isAuthenticated, user, router]);

    if (isLoading) {
        return (
            <div className="flex h-screen items-center justify-center bg-gray-50 dark:bg-slate-950">
                <div className="text-center">
                    <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
                    <p className="mt-4 text-gray-600 dark:text-gray-400">Loading...</p>
                </div>
            </div>
        );
    }

    if (!isAuthenticated || user?.role !== 'admin') {
        return null;
    }

    const navItems = [
        { name: "Dashboard", href: "/admin/dashboard", icon: House },
        { name: "Employees", href: "/admin/employees", icon: Users },
        { name: "Departments", href: "/admin/departments", icon: Users }, // Using Users icon as placeholder
        { name: "Attendance", href: "/admin/attendance", icon: CalendarCheck }, // Admin view of attendance
        { name: "Leaves", href: "/admin/leaves", icon: FileText },
        { name: "Payroll", href: "/admin/payroll", icon: CreditCard },
        { name: "Settings", href: "/admin/settings", icon: Settings },
        { name: "Organization Profile", href: "/admin/profile", icon: Users },
    ];

    return (
        <div className="flex h-screen bg-gray-50 dark:bg-slate-950 transition-colors duration-300">
            {/* Sidebar */}
            <aside className={`
                fixed lg:static inset-y-0 left-0 z-50 w-64 bg-white dark:bg-slate-900 border-r border-gray-200 dark:border-slate-800 flex flex-col transition-transform duration-300 ease-in-out
                ${isMobileMenuOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'}
            `}>
                <div className="h-16 flex items-center px-6 border-b border-gray-200 dark:border-slate-800">
                    <h1 className="text-xl font-bold text-gray-900 dark:text-white">PeopleOS <span className="text-xs font-normal text-gray-500 uppercase">Admin</span></h1>
                </div>

                <nav className="flex-1 p-4 space-y-1 overflow-y-auto">
                    {navItems.map((item) => {
                        const Icon = item.icon;
                        const isActive = pathname.startsWith(item.href);
                        return (
                            <Link
                                key={item.href}
                                href={item.href}
                                onClick={() => setIsMobileMenuOpen(false)}
                                className={`flex items-center space-x-3 px-4 py-2 rounded-lg transition-colors ${isActive
                                    ? "bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-400"
                                    : "text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-slate-800"
                                    }`}
                            >
                                <Icon className="w-5 h-5" />
                                <span>{item.name}</span>
                            </Link>
                        );
                    })}
                </nav>

                <div className="p-4 border-t border-gray-200 dark:border-slate-800">
                    <div className="flex items-center space-x-3 mb-4 px-2">
                        <div className="w-8 h-8 rounded-full bg-blue-100 dark:bg-blue-900 flex items-center justify-center text-sm font-bold text-blue-700 dark:text-blue-300">
                            {user?.first_name?.[0]}{user?.last_name?.[0]}
                        </div>
                        <div className="flex-1 min-w-0">
                            <p className="font-medium text-sm text-gray-900 dark:text-gray-100 truncate">{user?.first_name} {user?.last_name}</p>
                            <p className="text-xs text-gray-500 dark:text-gray-400 capitalize truncate">{user?.role?.replace('_', ' ')}</p>
                        </div>
                    </div>
                    <Button variant="ghost" className="w-full justify-start text-red-600 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-900/10" onClick={logout}>
                        <LogOut className="w-4 h-4 mr-2" />
                        Logout
                    </Button>
                </div>
            </aside>

            {/* Main Content */}
            <div className="flex-1 flex flex-col overflow-hidden">
                {/* Header */}
                <header className="h-16 bg-white/80 dark:bg-slate-900/80 backdrop-blur-md border-b border-gray-200 dark:border-slate-800 flex items-center justify-between px-4 lg:px-8 z-30">
                    <div className="flex items-center gap-4">
                        <Button variant="ghost" size="icon" className="lg:hidden" onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}>
                            <AlignJustify className="w-6 h-6" />
                        </Button>
                        <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
                            {navItems.find(item => pathname.startsWith(item.href))?.name || 'Dashboard'}
                        </h2>
                    </div>
                    <div className="flex items-center gap-4">
                        <ThemeToggle />
                    </div>
                </header>

                {/* Page Content */}
                <main className="flex-1 overflow-auto bg-gray-50 dark:bg-slate-950 p-4 lg:p-8">
                    {children}
                </main>
            </div>

            {/* Mobile Overlay */}
            {isMobileMenuOpen && (
                <div
                    className="fixed inset-0 bg-black/50 z-40 lg:hidden"
                    onClick={() => setIsMobileMenuOpen(false)}
                />
            )}
        </div>
    );
}
