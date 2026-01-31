'use client';
import { useAuth } from '@/lib/auth-context';
import { useTranslations } from 'next-intl';

import { Bell, BookText, LogOut, Settings, User, Users } from 'lucide-react'
import Link from 'next/link';
import { DropdownMenu, DropdownMenuContent, DropdownMenuGroup, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger } from '../ui/dropdown-menu';
import { useRouter } from 'next/navigation';
import DashboardSearchBar from './DashboardSearchBar';

interface TopbarProps {
    onRightMenuClick?: () => void; // New prop for right sidebar

}

const Topbar = ({ onRightMenuClick, }: TopbarProps) => {
    const { user, logout } = useAuth();
    const router = useRouter();
    const t = useTranslations('dashboard');
    const tCommon = useTranslations('common');
    const tSidebar = useTranslations('sidebar');


    const handleLogout = async () => {
        try {

            await logout();
            router.push('/');
        } catch (error) {
            console.error('Logout failed:', error);
            router.push('/');
        }
    }

    return (
        <header className='h-12 lg:h-16 px-4 lg:px-8 flex  items-center justify-between bg-[#fffbf7] sticky top-0 z-10 border-b border-black' suppressHydrationWarning>
            {/* Left Section */}
            <div className='flex items-center gap-4'>

                {/* Searchbar - Now with integrated search API */}
                <DashboardSearchBar className='w-48 sm:w-64 lg:w-96' />
            </div>

            {/* Right Actions */}
            <div className='flex items-center gap-3 lg:gap-6'>
                {/* Mobile Right Sidebar Toggle */}
                <button
                    onClick={onRightMenuClick}
                    className='p-2 text-gray-600 hover:text-[#FF8811] xl:hidden'
                >
                    <Users className="w-6 h-6 lg:w-8 md:h-8 font-bold text-black" />
                </button>




                {/* TODO Hackaton 50% */}
                <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                        <button className="relative p-2 text-gray-600 hover:text-[#FF8811] transition-colors hidden sm:block">
                            <Bell className="w-5 h-5 lg:w-6 lg:h-6" />
                            <span className="absolute top-1.5 right-2 w-2 h-2 bg-red-500 rounded-full border-2 border-[#FFFBF7]"></span>
                        </button>
                    </DropdownMenuTrigger>

                    <DropdownMenuContent className="w-80 p-0 border-2 border-[#2B2D42] shadow-[4px_4px_0px_0px_#2B2D42] bg-white rounded-xl overflow-hidden" align="end">
                        <div className="p-4 border-b border-gray-100 bg-[#FFFBF7]">
                            <div className="flex items-center justify-between">
                                <h3 className="font-bold text-[#2B2D42]">Notifications</h3>
                                <span className="text-xs text-[#FF8811] font-medium cursor-pointer hover:underline">Mark all as read</span>
                            </div>
                        </div>
                        
                        <div className="max-h-[300px] overflow-y-auto">
                    
                            <DropdownMenuItem className="p-4 cursor-pointer hover:bg-gray-50 focus:bg-gray-50 border-b border-gray-100 last:border-0 outline-none">
                                <div className="flex gap-3">
                                    <div className="mt-1 shrink-0 w-2 h-2 rounded-full bg-[#FF8811]"></div>
                                    <div className="flex-1 space-y-1">
                                        <p className="text-sm font-medium text-[#2B2D42] leading-none">New Team Member</p>
                                        <p className="text-xs text-gray-500">Alex joined the design team.</p>
                                        <p className="text-[10px] text-gray-400">2 min ago</p>
                                    </div>
                                </div>
                            </DropdownMenuItem>

                            <DropdownMenuItem className="p-4 cursor-pointer hover:bg-gray-50 focus:bg-gray-50 border-b border-gray-100 last:border-0 outline-none">
                                <div className="flex gap-3">
                                    <div className="mt-1 shrink-0 w-2 h-2 rounded-full bg-[#FF8811]"></div>
                                    <div className="flex-1 space-y-1">
                                        <p className="text-sm font-medium text-[#2B2D42] leading-none">Project Update</p>
                                        <p className="text-xs text-gray-500">Dashboard redesign is complete.</p>
                                        <p className="text-[10px] text-gray-400">1 hour ago</p>
                                    </div>
                                </div>
                            </DropdownMenuItem>

                            <DropdownMenuItem className="p-4 cursor-pointer hover:bg-gray-50 focus:bg-gray-50 outline-none">
                                <div className="flex gap-3">
                                    <div className="mt-1 shrink-0 w-2 h-2 rounded-full bg-gray-300"></div>
                                    <div className="flex-1 space-y-1">
                                        <p className="text-sm font-medium text-[#2B2D42] leading-none">System Alert</p>
                                        <p className="text-xs text-gray-500">Maintenance scheduled for tonight.</p>
                                        <p className="text-[10px] text-gray-400">5 hours ago</p>
                                    </div>
                                </div>
                            </DropdownMenuItem>
                        </div>

                        <div className="p-2 border-t border-gray-100 bg-gray-50 text-center">
                            <Link href="/dashboard/notifications" className="text-xs font-medium text-[#2B2D42] hover:text-[#FF8811] transition-colors">
                                View all notifications
                            </Link>
                        </div>
                    </DropdownMenuContent>
                </DropdownMenu>


                <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                        <button
                            type="button"
                            suppressHydrationWarning
                            className="flex items-center justify-center w-8 h-8 lg:w-9 lg:h-9 rounded-full bg-[#FF8811] text-white font-bold  shadow-sm cursor-pointer select-none border-2 border-[#2B2D42] hover:opacity-90 transition-opacity text-sm lg:text-base"
                        >
                            {user?.name?.charAt(0).toUpperCase() || 'U'}
                        </button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent className="p-4 w-64 text-3xl border  border-[#2B2D42] shadow-[4px_4px_0px_0px_#2B2D42]  " align="end" >
                        <DropdownMenuLabel>
                            <div className='flex items-center gap-2 mb-4'>
                                <div className="shrink-0 w-8 h-8 sm:w-8 sm:h-8 lg:w-9 lg:h-9 rounded-full bg-[#FF8811] flex items-center justify-center text-white font-bold shadow-sm cursor-pointer hover:opacity-90 transition-opacity text-sm lg:text-base border-2 border-[#2B2D42]">
                                    {user?.name?.charAt(0).toUpperCase() || ''}
                                </div>
                                <div className='flex flex-col '>
                                    <span className='text-gray-600 text-sm font-family-name:var(--font-inter)'>Hello, {user?.name || 'User'}</span>
                                    <span className='text-gray-400 text-xs font-family-name:var(--font-inter)'>{user?.role || 'Student'}</span>
                                </div>

                            </div>
                        </DropdownMenuLabel>
                        <DropdownMenuSeparator />
                        <DropdownMenuGroup>
                            <DropdownMenuItem>
                                <button>
                                    <Link href={'/dashboard/my-pods'} className='flex' >
                                        <BookText className='size-5 mr-2' />
                                        <span className='text-md '>{tSidebar('myPods')}</span>
                                    </Link>
                                </button>
                            </DropdownMenuItem>
                            <DropdownMenuItem>
                                <button disabled={true} className='flex cursor-not-allowed'>
                                        <User className='size-5 mr-2 text-gray-300' />
                                        <span className='text-md text-gray-300'>{tCommon('profile')}</span>
                                </button>
                            </DropdownMenuItem>
                            <DropdownMenuItem>
                                <button disabled={true} className='flex cursor-not-allowed'>
                                        <Settings className='size-5 mr-2 text-gray-300' />
                                        <span className='text-md text-gray-300'>{tSidebar('settings')}</span>
                                </button>
                            </DropdownMenuItem>
                        </DropdownMenuGroup>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem>
                            <button onClick={handleLogout} className='flex cursor-pointer'>
                                <LogOut className='size-5 mr-2' />
                                <span className='text-md font-semibold text-red-500'>{tCommon('logout')}</span>
                            </button>
                        </DropdownMenuItem>
                    </DropdownMenuContent>
                </DropdownMenu>
            </div>
        </header>
    )
}

export default Topbar