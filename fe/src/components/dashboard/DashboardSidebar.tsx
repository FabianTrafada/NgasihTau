'use client'
import { Sidebar, SidebarContent, SidebarGroup, SidebarGroupContent, SidebarGroupLabel, SidebarHeader, SidebarMenu, SidebarMenuButton, SidebarMenuItem, useSidebar } from '../ui/sidebar'
import Link from 'next/dist/client/link'
import { USER_SIDEBAR_GROUPS } from '@/lib/constants/navigation'
import { usePathname } from 'next/navigation'
import { useState } from 'react'

const DashboardSidebar = ({ groups }: { groups: typeof USER_SIDEBAR_GROUPS }) => {

    const pathname = usePathname();
    const { toggleSidebar } = useSidebar();

    const [isEnglish, setIsEnglish] = useState(false);
    
    const handleLanguageChange = (checked: boolean) => {
        setIsEnglish(checked);
    }

    return (

        <Sidebar variant='sidebar' collapsible='icon' className="border-r-2 border-black bg-[#FFFBF7] transition-[width] duration-300 ease-in-out">
            <SidebarHeader className="mt-2 bg-[#FFFBF7]">
                <div className="flex  gap-2 font-bold  ">
                        <button onClick={() => toggleSidebar()} className="size-8 aspect-square  border-2 bg-[#FF8811] border-black  text-white flex items-center justify-center shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]  font-family-name:var(--font-plus-jakarta-sans) cursor-pointer">N</button>
                    <Link href="/" className='items-center'>
                        <div>
                            <span className="group-data-[collapsible=icon]:hidden text-xl tracking-tight text-[#FF8811] ">Ngasih</span>
                            <span className="group-data-[collapsible=icon]:hidden text-xl tracking-tight">Tau</span>
                        </div>
                    </Link>
                </div>
            </SidebarHeader>

            <SidebarContent className="bg-[#FFFBF7]  ">
                {groups.map((group) => (
                    <SidebarGroup key={group.title}>
                        <SidebarGroupLabel className="text-black font-bold uppercase tracking-wider text-xs mb-2">{group.title}</SidebarGroupLabel>
                        <SidebarGroupContent>
                            <SidebarMenu>
                                {group.items.map((item) => (
                                    <SidebarMenuItem key={item.label} className="mb-1">
                                        <SidebarMenuButton
                                            asChild
                                            tooltip={item.label}
                                            isActive={pathname === item.href} // Menandai menu yang aktif
                                            className={`
                                                border-2 border-transparent hover:border-black hover:bg-[#FF8811] hover:text-white hover:shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] transition-all duration-200 
                                                data-[active=true]:bg-[#FF8811] data-[active=true]:text-white data-[active=true]:border-black data-[active=true]:shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] group-data-[collapsible=icon]:justify-center
                                            `}
                                        >
                                            <Link href={item.href} className="flex font-medium">
                                                <item.icon className="size-4" />
                                                <span className="group-data-[collapsible=icon]:hidden">{item.label}</span>
                                            </Link>
                                        </SidebarMenuButton>
                                    </SidebarMenuItem>
                                ))}
                            </SidebarMenu>
                        </SidebarGroupContent>
                    </SidebarGroup>
                ))}
            </SidebarContent>
        </Sidebar>
    )
}

export default DashboardSidebar