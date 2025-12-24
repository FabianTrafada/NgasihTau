'use client';
import { useAuth } from '@/lib/auth-context';

import { Bell, LogOut, PanelLeft, PanelLeftClose, Search, Settings, User, Users } from 'lucide-react'
import Link from 'next/link';
import { DropdownMenu, DropdownMenuContent, DropdownMenuGroup, DropdownMenuItem, DropdownMenuLabel, DropdownMenuPortal, DropdownMenuSeparator, DropdownMenuShortcut, DropdownMenuSub, DropdownMenuSubContent, DropdownMenuSubTrigger, DropdownMenuTrigger } from '../ui/dropdown-menu';
import { Separator } from '@radix-ui/react-dropdown-menu';
interface TopbarProps {
    onMenuClick?: () => void;
    onRightMenuClick?: () => void; // New prop for right sidebar
    sidebarOpen?: boolean;
}

const Topbar = ({ onMenuClick, onRightMenuClick, sidebarOpen }: TopbarProps) => {
    const { user } = useAuth();

    return (
        <header className='h-16 lg:h-20 px-4 lg:px-8 flex  items-center justify-between bg-[#fffbf7] sticky top-0 z-10 border-b border-gray-200'>
            {/* Left Section */}
            <div className='flex items-center gap-4'>
                <button
                    onClick={onMenuClick}
                    className='p-2 rounded-lg hover:bg-gray-100 lg:hidden'
                    title={sidebarOpen ? "Hide sidebar" : "Show sidebar"}
                >
                    {sidebarOpen ? <PanelLeftClose className='w-6 h-6 text-gray-600' /> : <PanelLeft className='w-6 h-6 text-gray-600' />}
                </button>

                {/* Searchbar */}
                <div className='relative w-48 sm:w-64 lg:w-96'>
                    <Search className='absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400' />
                    <input
                        type="text"
                        placeholder='Search'
                        className='w-full pl-10 pr-4 py-2 lg:py-2.5 bg-gray-100/50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-[#ff8811]/20 focus:border-[#ff8811] transition-all text-sm lg:text-base'
                    />
                </div>
            </div>

            {/* Right Actions */}
            <div className='flex items-center gap-3 lg:gap-6'>
                {/* Mobile Right Sidebar Toggle */}
                <button
                    onClick={onRightMenuClick}
                    className='p-2 text-gray-600 hover:text-[#FF8811] xl:hidden'
                >
                    <Users className="w-6 h-6" />
                </button>

                

                <button className="relative p-2 text-gray-600 hover:text-[#FF8811] transition-colors hidden sm:block">
                    <Bell className="w-5 h-5 lg:w-6 lg:h-6" />
                    <span className="absolute top-1.5 right-2 w-2 h-2 bg-red-500 rounded-full border-2 border-[#FFFBF7]"></span>
                </button>

                <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                        <button className="w-8 h-8 lg:w-10 lg:h-10 rounded-full bg-[#FF8811] flex items-center justify-center text-white font-bold shadow-sm cursor-pointer hover:opacity-90 transition-opacity text-sm lg:text-base outline-none">
                            {user?.name?.charAt(0).toUpperCase() || 'U'}
                        </button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent className="p-4 w-64 text-3xl border  border-[#2B2D42] shadow-[4px_4px_0px_0px_#2B2D42]  " align="end" >
                        <DropdownMenuLabel>
                            <div className='flex items-center gap-2 mb-4'>
                                <div className="w-4 h-4 lg:w-10 lg:h-10 rounded-full bg-[#FF8811] flex items-center justify-center text-white font-bold shadow-sm cursor-pointer hover:opacity-90 transition-opacity text-sm lg:text-base">
                                    {user?.name?.charAt(0).toUpperCase() || 'U'}
                                </div>
                                <span className='font-semibold text-xl'>Hello, {user?.name || 'User'}</span>
                            </div>
                        </DropdownMenuLabel>
                        <DropdownMenuSeparator />
                        <DropdownMenuGroup>
                            <DropdownMenuItem>
                                <Link href={'/dashboard/profile'} className='flex'>
                                    <User className='size-5 mr-2' />
                                    <span className='text-md'>Profile</span>
                                </Link>
                            </DropdownMenuItem>
                            <DropdownMenuItem>
                                <Link href={'/dashboard/settings'} className='flex'>
                                    <Settings className='size-5 mr-2' />
                                    <span className='text-md'>Settings</span>
                                </Link>
                            </DropdownMenuItem>
                        </DropdownMenuGroup>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem>
                            <Link href={'/logout'} className='flex'>
                                <LogOut className='size-5 mr-2' />
                                <span className='text-md font-semibold text-red-500'>Log out</span>
                            </Link>
                        </DropdownMenuItem>
                    </DropdownMenuContent>
                </DropdownMenu>
            </div>
        </header>
    )
}

export default Topbar