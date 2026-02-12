"use client";

export default function SettingsPage() {
    return (
        <div className="space-y-8 animate-in fade-in duration-500">
            {/* Header */}
            <div>
                <h1 className="text-3xl font-bold text-gray-900 dark:text-white tracking-tight">Platform Settings</h1>
                <p className="text-gray-500 dark:text-gray-400 mt-2 text-lg">Configure platform-wide settings</p>
            </div>

            {/* Coming Soon */}
            <div className="bg-white dark:bg-slate-900 p-12 rounded-2xl border border-gray-100 dark:border-slate-800 shadow-sm text-center">
                <div className="max-w-md mx-auto">
                    <div className="text-6xl mb-6">⚙️</div>
                    <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-3">Settings Coming Soon</h2>
                    <p className="text-gray-500 dark:text-gray-400 text-lg">
                        Platform settings and configuration options will be available in a future update.
                    </p>
                </div>
            </div>
        </div>
    );
}
