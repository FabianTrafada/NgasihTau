import { LayoutDashboard, BookOpen, File, Users, Settings, LucideIcon, BookOpenText, BotMessageSquare } from 'lucide-react';

export interface NavItem {
    label: string;
    href: string;
    icon: LucideIcon;
}

export interface NavGroup {
    title: string;
    items: NavItem[];
}

export const USER_SIDEBAR_GROUPS: NavGroup[] = [
    {
        title: "Navigation",
        items: [
            {
                label: "Home",
                href: "/dashboard",
                icon: LayoutDashboard,
            },
        ],
    },
    {
        title: "Knowledge",
        items: [
            {
                label: "Browse Knowledge",
                href: "/dashboard/pods",
                icon: BookOpenText 
,
            },
        ],
    },
    {
        title: "AI Tools",
        items: [
            {
                label: "Chatbot",
                href: "/dashboard/chatbot",
                icon: BotMessageSquare  
            },
        ],
    },
];

export const TEACHER_SIDEBAR_GROUPS: NavGroup[] = [
    {
        title: "Menu",
        items: [
            {
                label: "Dashboard",
                href: "/teacher/dashboard",
                icon: LayoutDashboard,
            },
            {
                label: "My Pods",
                href: "/teacher/pods",
                icon: Users,
            },
          
        ],
    },
    {
        title: "Settings",
        items: [
            {
                label: "Settings",
                href: "/teacher/settings",
                icon: Settings,
            },
        ],
    },
];
