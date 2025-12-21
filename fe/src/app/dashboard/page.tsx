"use client";

import React from "react";
import RightSidebar from "@/components/dashboard/RightSidebar";
import RecentPodCard from "@/components/dashboard/RecentPodCard";
import ChatbotLogCard from "@/components/dashboard/ChatbotLogCard";

export default function DashboardPage() {
    return (
        <div className="flex min-h-full">
            {/* Main Content Area */}
            <div className="flex-1 p-8 min-w-0">
                <div className="max-w-4xl mx-auto flex flex-col gap-8">
                    {/* Welcome Section */}
                    <div>
                        <h1 className="text-3xl font-bold text-gray-900 mb-1">
                            Welcome Back Tom Lembong
                        </h1>
                        <p className="text-gray-500">lorem ipsum dolor sit ammer</p>
                    </div>

                    {/* Recent Pods */}
                    <section>
                        <div className="flex items-center justify-between mb-4">
                            <h2 className="text-lg font-bold text-gray-900">Recent Pods</h2>
                            <button className="text-xs text-gray-400 hover:text-[#FF8811]">See all</button>
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
                            <h2 className="text-lg font-bold text-gray-900">Last Chatbot Logs</h2>
                            <button className="text-xs text-gray-400 hover:text-[#FF8811]">See all</button>
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
            </div>

            {/* Right Sidebar */}
            <RightSidebar />
        </div>
    );
}