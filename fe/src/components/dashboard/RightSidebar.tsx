import React, { useState } from "react";
import { Search, MoreHorizontal, Users, X } from "lucide-react";
import { cn } from "@/lib/utils";


interface RightSidebarProps {
    onClose?: () => void;
    isOpen?: boolean;
}

const RightSidebar = ({ onClose, isOpen }: RightSidebarProps) => {


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
        <>

            {/* mobile overlay */}
            <div className={cn("fixed inset-0  z-40 xl:hidden transition-opacity duration-300", isOpen ? "opacity-100" : "opacity-0 pointer-events-none")} onClick={onClose} />


            {/* sidebar */}
            <aside className={cn("w-80 bg-[#FFFBF7] border  border-gray-200 p-6 flex flex-col h-screen overflow-hidden fixed ", "right-0 top-0 z-20  transition-transform duration-300 xl:translate-x-0 xl:sticky xl:top-0 xl:h-screen xl:z-0", isOpen ? "translate-x-0" : "translate-x-full")}>

                <div className="flex items-center justify-between mb-6">
                    <h2 className="text-lg font-bold text-[#2B2D42]">Knowledge Community</h2>

                    <button
                        onClick={onClose}
                        className="p-1 rounded-lg hover:bg-gray-100 xl:hidden">
                        <X className="w-5 h-5 text-[#2B2D42]" />
                    </button>

                </div>

                {/* Search */}
                <div className="relative mb-8">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[#2B2D42]" />
                    <input
                        type="text"
                        placeholder="Search community..."
                        className="w-full pl-9 pr-4 py-2 bg-white border-2 border-[#2B2D42] rounded-lg text-sm focus:outline-none focus:shadow-[2px_2px_0px_0px_#FF8811] transition-all placeholder:text-gray-400 text-[#2B2D42]"
                    />
                </div>

                {/* List */}
                <div className="space-y-6 overflow-y-auto flex-1 pr-2 custom-scrollbar">
                    {communities.map((comm, idx) => (
                        <div key={idx} className="group cursor-pointer">
                            <div className="flex items-start justify-between mb-1">
                                <div className="flex gap-3">
                                    <div className="w-10 h-10 rounded-lg bg-white flex items-center justify-center text-xl border-2 border-[#2B2D42] shadow-[2px_2px_0px_0px_#2B2D42] group-hover:translate-x-[1px] group-hover:translate-y-[1px] group-hover:shadow-none transition-all">
                                        {comm.icon}
                                    </div>
                                    <div>
                                        <h3 className="text-sm font-bold text-[#2B2D42] leading-tight mb-0.5 group-hover:text-[#FF8811] transition-colors">
                                            {comm.title}
                                        </h3>
                                        <div className="flex items-center gap-1 text-[10px] text-gray-500">
                                            <span>{comm.author}</span>
                                            <span>👑</span>
                                        </div>
                                        <div className="flex items-center gap-1 text-[10px] text-gray-500 mt-1">
                                            <span className="font-medium text-[#2B2D42]">{comm.members}</span>
                                            <Users className="w-3 h-3 text-[#2B2D42]" />
                                        </div>
                                    </div>
                                </div>
                                <button className="text-gray-400 hover:text-[#FF8811]">
                                    <MoreHorizontal className="w-5 h-5" />
                                </button>
                            </div>
                            <div className="h-px bg-[#2B2D42]/10 mt-4 group-last:hidden" />
                        </div>
                    ))}
                </div>
            </aside>
        </>


    );
};

export default RightSidebar;