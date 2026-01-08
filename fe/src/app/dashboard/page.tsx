"use client";

import RecentPodCard from "@/components/dashboard/RecentPodCard";
import ChatbotLogCard from "@/components/dashboard/ChatbotLogCard";
import { ProtectedRoute } from "@/components/auth";
import { useAuth } from "@/lib/auth-context";
import Link from "next/link";

/**
 * Dashboard Page
 * 
 * This is a PROTECTED page - requires authentication.
 * The ProtectedRoute component handles:
 * 1. Showing loading state while checking auth
 * 2. Redirecting to /sign-in if not authenticated
 * 3. Rendering content if authenticated
 */
export default function DashboardPage() {
    return (
        <ProtectedRoute>
            <DashboardContent />
        </ProtectedRoute>
    );
}

/**
 * Dashboard Content - only rendered when authenticated
 */
function DashboardContent() {
    // Access user data from auth context
    const { user } = useAuth();
    return (
        <div className="flex min-h-screen">
            {/* Main Content Area */}
            <div className="flex-1 p-4 sm:p-8 lg:p-8  max-w-4xl mx-auto flex flex-col gap-6 lg:gap-8 sm:max-w-6xl">
                {/* Welcome Section - Now uses real user name */}
                <div>
                    <h1 className="text-1xl md:text-2xl font-family-name:var(--font-plus-jakarta-sans) sm:text-2xl font-bold text-[#2B2D42] mb-1">
                        Welcome Back {user?.name || "User"}
                    </h1>
                    <p className="text-xs sm:text-base text-gray-500 font-family-name:var(--font-inter)">Ready to learn something new today?</p>
                </div>

                {/* Recent Pods */}
                <section>
                    <div className="flex items-center justify-between mb-4">
                        <h2 className="text-base sm:text-lg font-family-name:var(--font-plus-jakarta-sans)  font-bold text-[#2B2D42]">Recent Pods</h2>

                        <Link href={'/dashboard/pods'}>
                            <button className="text-xs text-gray-400 hover:text-[#FF8811] cursor-pointer font-bold">See all</button>
                        </Link>
                    </div>
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                        <RecentPodCard
                            title="Cara Belajar Mobil Kopling (99% bisa 1% nya hanya tuhan ....)"
                            description="Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."
                            fileCount={12}
                            date="2 Jan 25"
                        />
                        <RecentPodCard
                            title="Cara Belajar Mobil Kopling (99% bisa 1% nya hanya tuhan ....)"
                            description="Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."
                            fileCount={12}
                            date="2 Jan 25"
                        />
                    </div>
                </section>

                {/* Last Chatbot Logs */}
                <section>
                    <div className="flex items-center justify-between mb-4">
                        <h2 className="text-base sm:text-lg font-bold text-[#2B2D42]">Last Chatbot Logs</h2>
                        <Link href={'/dashboard/chatbot-logs'} className="text-xs text-gray-400 hover:text-[#FF8811] font-bold">See all</Link>
                    </div>
                    <div className="space-y-3">
                        <ChatbotLogCard
                            title="Mempelajari materi limit dengan analogi"
                            snippet="Berikan saya Trik cepat pahan limit.."
                        />
                        <ChatbotLogCard
                            title="Mempelajari materi limit dengan analogi"
                            snippet="Berikan saya Trik cepat pahan limit.."
                        />
                        <ChatbotLogCard
                            title="Mempelajari materi limit dengan analogi"
                            snippet="Berikan saya Trik cepat pahan limit.."
                        />
                    </div>
                </section>
            </div>

            {/* Right Sidebar */}
        </div>
    );
}