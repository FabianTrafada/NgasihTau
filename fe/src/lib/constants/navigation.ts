import { LayoutDashboard, BookOpen, File, Users, Settings, LucideIcon } from 'lucide-react';

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
                label: "Knowledge Spot",
                href: "/dashboard/pod",
                icon: BookOpen,
            },
        ],
    },
    {
        title: "Your Pods",
        items: [
            {
                label: "My Knowledge Pods",
                href: "/dashboard/pods",
                icon: File,
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
