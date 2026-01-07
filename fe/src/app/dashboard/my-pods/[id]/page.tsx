'use client';
import { useState } from 'react';
import { Eye, BadgeCheck, FileText, Plus, Search } from 'lucide-react';
import FileListItem from '@/components/knowledge-pod/FileListItem'; // Sesuaikan path-nya
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Switch } from "@/components/ui/switch"
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

const KnowledgePodDetail = () => {
    const [isPublic, setIsPublic] = useState(true);
    const [isSaving, setIsSaving] = useState(false);
    const collaborators = [
        { name: "Rahmat hadiwibowo", avatar: "https://github.com/shadcn.png" },
        { name: "Edi Hadiwibowo", avatar: "https://github.com/shadcn.png" },
        { name: "Slamet Oli samping", avatar: "https://github.com/shadcn.png" },
    ];

    const files = [
        {
            title: "Limit Turunan",
            description: "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam.",
            likes: "2,3 K",
            date: "3 Jan 2025"
        },
        {
            title: "Limit Turunan",
            description: "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam.",
            likes: "2,3 K",
            date: "3 Jan 2025"
        },
        {
            title: "Limit Turunan",
            description: "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam.",
            likes: "2,3 K",
            date: "3 Jan 2025"
        }
    ];

    return (
        <div className="max-w-6xl mx-auto space-y-8 p-4 md:p-8">
            {/* saving button */}

            {isSaving && (
                <div className='items-center'>
                    <button className="bg-orange-500 hover:bg-orange-600 text-white font-bold py-2 px-4 rounded-lg shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] border-2 border-black">
                        Save Changes
                    </button>
                </div>
            )}

            {/* Header Section */}
            <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                <h1 className="text-4xl font-black text-black tracking-tight">
                    Kalkulus
                </h1>

                <div className="flex flex-wrap items-center gap-4 md:gap-6">
                    {/* Avatar Group */}
                    <div className="flex items-center -space-x-3">
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <button className="relative flex items-center justify-center w-10 h-10 rounded-full border-2 border-black bg-orange-500 text-white hover:bg-orange-600 transition-colors z-20 focus:outline-none shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]">
                                    <Plus size={20} strokeWidth={3} />
                                </button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent className="w-72 p-4 border-2 border-black shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] rounded-xl bg-[#FFFBF7]">
                                <DropdownMenuLabel className="text-lg font-bold text-center mb-2">Select Collaborators</DropdownMenuLabel>

                                {/* Search Bar */}
                                <div className="relative mb-4">
                                    <input
                                        type="text"
                                        placeholder="search"
                                        className="w-full bg-white border-2 border-[#2B2D42] rounded-md py-2 px-3 pl-3 text-sm shadow-[2px_2px_0px_0px_#2B2D42] focus:outline-none focus:translate-x-[1px] focus:translate-y-[1px] focus:shadow-none transition-all font-[family-name:var(--font-inter)]"
                                    />
                                </div>

                                <div className="space-y-2">
                                    {collaborators.map((collab, idx) => (
                                        <DropdownMenuItem key={idx} className="flex items-center gap-3 p-2 cursor-pointer hover:bg-orange-100 rounded-lg focus:bg-orange-100 focus:text-black">
                                            <Avatar className="w-8 h-8 border-2 border-black shadow-none">
                                                <AvatarImage src={collab.avatar} />
                                                <AvatarFallback>{collab.name.substring(0, 2).toUpperCase()}</AvatarFallback>
                                            </Avatar>
                                            <span className="font-bold text-sm">{collab.name}</span>
                                        </DropdownMenuItem>
                                    ))}
                                </div>
                            </DropdownMenuContent>
                        </DropdownMenu>

                        <Avatar className="border-2 border-black w-10 h-10 bg-white">
                            <AvatarImage src="https://github.com/shadcn.png" />
                            <AvatarFallback>AB</AvatarFallback>
                        </Avatar>
                        <Avatar className="border-2 border-black w-10 h-10 bg-white">
                            <AvatarImage src="https://github.com/shadcn.png" />
                            <AvatarFallback>CD</AvatarFallback>
                        </Avatar>
                        <div className="flex items-center justify-center w-10 h-10 rounded-full border-2 border-black bg-[#FFFBF7] text-xs font-bold text-black z-10 shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]">
                            +3
                        </div>
                    </div>

                    {/* Public/Private Switch */}
                    <div className="flex items-center gap-2 bg-white px-4 py-2 border-2 border-black rounded-full shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] sm:justify-between">
                        <span className={`text-sm font-bold transition-colors ${!isPublic ? 'text-black' : 'text-zinc-400'}`}>
                            Private
                        </span>
                        <Switch
                            checked={isPublic}
                            onCheckedChange={setIsPublic}
                            className="data-[state=checked]:bg-orange-500 border-2 border-transparent"
                        />
                        <span className={`text-sm font-bold transition-colors ${isPublic ? 'text-black' : 'text-zinc-400'}`}>
                            Public
                        </span>
                    </div>
                </div>
            </div>

            {/* Description Card */}
            <div className="bg-white border-2 border-black rounded-2xl p-6 pb-4 shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]">
                <p className='text-xs font-semibold text-zinc-400 mb-1'>Deskripsi</p>
                <p className="text-sm font-medium text-black mb-6">
                    Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
                </p>

                <div className="flex justify-between items-center pt-4 border-t border-zinc-200">
                    <div className='flex gap-6'>
                        <div className="flex items-center gap-2 text-zinc-500">
                            <Eye size={16} />
                            <span className="text-xs font-mono font-bold">999K</span>
                        </div>

                        <div className="flex items-center gap-2 text-zinc-500">
                            <FileText size={16} />
                            <span className="text-xs font-mono font-bold">12 files</span>
                        </div>
                    </div>

                    <div className="flex items-center gap-2 px-3 py-1 bg-zinc-100 border-2 border-black rounded-full">
                        <BadgeCheck size={16} className="fill-black text-white" />
                        <span className="font-bold text-xs">Guru Besar ITB</span>
                    </div>
                </div>
            </div>

            {/* <SearchSection/> */}

            {/* Files List Container */}
            <div className="bg-white border-2 border-black rounded-2xl overflow-hidden shadow-[4px_4px_0px_0px_#FF8811]">
                <div className="flex flex-col">
                    {files.map((file, index) => (
                        <FileListItem
                            key={index}
                            materialId={`dummy-${index}`}
                            podId=''
                            userId=''
                            title={file.title}
                            description={file.description}
                            likes={file.likes}
                            date={file.date}
                            isLast={index === files.length - 1}
                        />
                    ))}
                </div>
            </div>

        </div>
    );
};

export default KnowledgePodDetail;