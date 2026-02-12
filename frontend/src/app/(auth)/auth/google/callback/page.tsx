"use client";

import { useEffect, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useAuth } from '@/context/auth-context';
import api from '@/lib/api';

function GoogleCallbackContent() {
    const router = useRouter();
    const searchParams = useSearchParams();
    const { login } = useAuth();

    useEffect(() => {
        const code = searchParams.get('code');
        const state = searchParams.get('state');

        if (code && state) {
            const exchangeToken = async () => {
                try {
                    const { data } = await api.get(`/auth/google/callback?code=${code}&state=${state}`);
                    localStorage.setItem('token', data.token);
                    window.location.href = '/dashboard';
                } catch (error: any) {
                    window.location.replace(`/login?error=${encodeURIComponent(error.response?.data?.error || 'google_auth_failed')}`);
                }
            };

            exchangeToken();
        } else {
            router.push('/login?error=missing_params');
        }
    }, [searchParams, router]);

    return (
        <div className="flex min-h-screen items-center justify-center">
            <div className="text-center">
                <h2 className="text-2xl font-bold mb-4">Authenticating...</h2>
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
            </div>
        </div>
    );
}

export default function GoogleCallbackPage() {
    return (
        <Suspense fallback={
            <div className="flex min-h-screen items-center justify-center">
                <div className="text-center">
                    <h2 className="text-2xl font-bold mb-4">Authenticating...</h2>
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
                </div>
            </div>
        }>
            <GoogleCallbackContent />
        </Suspense>
    );
}
