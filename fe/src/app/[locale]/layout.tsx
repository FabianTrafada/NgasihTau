import type { Metadata } from "next";
import { Plus_Jakarta_Sans, Inter } from "next/font/google";
import { NextIntlClientProvider } from 'next-intl';
import { getMessages, setRequestLocale } from 'next-intl/server';
import { notFound } from 'next/navigation';
import { routing } from '@/i18n/routing';
import { Locale } from '@/i18n/config';
import "../globals.css";
import SmoothScroll from "@/components/SmoothScroll";
import { AuthProvider } from "@/lib/auth-context";

const plusJakartaSans = Plus_Jakarta_Sans({
    variable: "--font-plus-jakarta-sans",
    subsets: ["latin"],
    weight: ["400", "500", "600", "700", "800"],
});

const inter = Inter({
    variable: "--font-inter",
    subsets: ["latin"],
    weight: ["400", "500", "600"],
});

export const metadata: Metadata = {
    title: "NgasihTau",
    description: "NgasihTau - Your learning platform",
};

/**
 * Generate static params for all supported locales
 * Required for static generation
 */

interface LocaleLayoutProps {
    children: React.ReactNode;
    params: Promise<{ locale: string }>;
}

/**
 * Locale Layout - Server Component
 * Handles i18n setup and provides translations to children
 */
export default async function LocaleLayout({
    children,
    params,
}: LocaleLayoutProps) {

    // Enable static rendering for this locale

    // Fetch messages for current locale
    const messages = await getMessages();

    return (
        <html lang={(await params).locale} suppressHydrationWarning>
            <body
                className={`${plusJakartaSans.variable} ${inter.variable} antialiased`}
                suppressHydrationWarning
            >
                <NextIntlClientProvider messages={messages}>
                    <AuthProvider>
                        <SmoothScroll />
                        {children}
                    </AuthProvider>
                </NextIntlClientProvider>
            </body>
        </html>
    );
}
