import React from "react";
import { Search, MoreHorizontal, Users } from "lucide-react";

const RightSidebar = () => {
    const communities = [
        {
            title: "Kalkulus By guru besar ITB",
            author: "Prof. Dr. Cornelius",
            members: "127,k",
            icon: "🧮",
        },
        {
            title: "IPA Biology By Guru besar UGM",
            author: "Prof. Dr. Cornelius",
            members: "127,k",
            icon: "🧬",
        },
        {
            title: "Kalkulus By guru besar ITB",
            author: "Prof. Dr. Cornelius",
            members: "127,k",
            icon: "📐",
        },
        {
            title: "Kalkulus By guru besar ITB",
            author: "Prof. Dr. Cornelius",
            members: "127,k",
            icon: "📚",
        },
    ];

    return (
        <aside className="w-80 bg-[#FFFBF7] border-l border-gray-200 p-6 flex flex-col hidden xl:flex sticky top-0 h-screen overflow-y-auto">
            <h2 className="text-lg font-bold text-gray-900 mb-6">Knowledge Community</h2>

            {/* Search */}
            <div className="relative mb-8">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
                <input
                    type="text"
                    placeholder="Search community, or etc"
                    className="w-full pl-9 pr-4 py-2 bg-gray-100/50 border border-gray-200 rounded-lg text-sm focus:outline-none focus:border-[#FF8811] transition-all"
                />
            </div>

            {/* List */}
            <div className="space-y-6">
                {communities.map((comm, idx) => (
                    <div key={idx} className="group cursor-pointer">
                        <div className="flex items-start justify-between mb-1">
                            <div className="flex gap-3">
                                <div className="w-10 h-10 rounded-lg bg-gray-100 flex items-center justify-center text-xl border border-gray-200">
                                    {comm.icon}
                                </div>
                                <div>
                                    <h3 className="text-sm font-bold text-gray-900 leading-tight mb-0.5 group-hover:text-[#FF8811] transition-colors">
                                        {comm.title}
                                    </h3>
                                    <div className="flex items-center gap-1 text-[10px] text-gray-500">
                                        <span>{comm.author}</span>
                                        <span>👑</span>
                                    </div>
                                    <div className="flex items-center gap-1 text-[10px] text-gray-500 mt-1">
                                        <span className="font-medium text-gray-700">{comm.members}</span>
                                        <Users className="w-3 h-3" />
                                    </div>
                                </div>
                            </div>
                            <button className="text-gray-400 hover:text-gray-600">
                                <MoreHorizontal className="w-5 h-5" />
                            </button>
                        </div>
                        <div className="h-px bg-gray-100 mt-4 group-last:hidden" />
                    </div>
                ))}
            </div>
        </aside>
    );
};

export default RightSidebar;