"use client";

import React, { createContext, useContext, useEffect, useState } from 'react';
import api from '@/lib/api';
import { User } from '@/types';
import { useRouter, usePathname } from 'next/navigation';

interface AuthContextType {
    user: User | null;
    isLoading: boolean;
    login: (data: any) => Promise<void>;
    logout: () => Promise<void>;
    isAuthenticated: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
    const [user, setUser] = useState<User | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const router = useRouter();
    const pathname = usePathname();

    useEffect(() => {
        checkAuth();
    }, []);

    const checkAuth = async () => {
        const token = localStorage.getItem('token');
        if (!token) {
            setIsLoading(false);
            return;
        }

        try {
            // Decode JWT token to get user info (JWT format: header.payload.signature)
            const payload = JSON.parse(atob(token.split('.')[1]));

            // Create user object from JWT claims
            const userData: User = {
                id: payload.user_id,
                tenant_id: payload.tenant_id || null,
                email: payload.email,
                role: payload.role,
                first_name: '',
                last_name: ''
            };

            setUser(userData);
        } catch (error) {
            console.error("Auth check failed:", error);
            setUser(null);
            localStorage.removeItem('token');
        } finally {
            setIsLoading(false);
        }
    };

    const login = async (credentials: any) => {
        const { data } = await api.post('/auth/login', credentials);
        localStorage.setItem('token', data.token); // Save access token
        setUser(data.user);

        // Redirect based on role
        if (data.user.role === 'super_admin') {
            router.push('/super-admin/dashboard');
        } else if (data.user.role === 'admin') {
            router.push('/admin/dashboard');
        } else if (data.user.role === 'hr') {
            router.push('/hr/dashboard');
        } else if (data.user.role === 'manager') {
            router.push('/manager/dashboard');
        } else if (data.user.role === 'team_lead') {
            router.push('/team-lead/dashboard');
        } else {
            router.push('/dashboard');
        }
    };

    const logout = async () => {
        try {
            await api.post('/auth/logout');
        } finally {
            localStorage.removeItem('token'); // Clear access token
            setUser(null);
            router.push('/login');
        }
    };

    return (
        <AuthContext.Provider
            value={{
                user,
                isLoading,
                login,
                logout,
                isAuthenticated: !!user
            }}
        >
            {children}
        </AuthContext.Provider>
    );
}

export const useAuth = () => {
    const context = useContext(AuthContext);
    if (context === undefined) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
};
