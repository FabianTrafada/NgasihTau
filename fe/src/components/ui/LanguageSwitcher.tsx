'use client';

import { useLocale } from 'next-intl';
import { useRouter, usePathname } from 'next/navigation';
import { useTransition } from 'react';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Button } from '@/components/ui/button';
import { Globe } from 'lucide-react';
import { locales, localeNames, localeFlags, Locale } from '@/i18n/config';

/**
 * Language Switcher Component
 * Provides a dropdown menu to switch between available locales
 * Automatically updates URL and reloads translations
 */
export function LanguageSwitcher() {
    const locale = useLocale() as Locale;
    const router = useRouter();
    const pathname = usePathname();
    const [isPending, startTransition] = useTransition();

    /**
     * Handle locale change
     * Replace current locale in pathname with new locale
     */
    const handleLocaleChange = (newLocale: Locale) => {
        if (newLocale === locale) return;

        const segments = pathname.split('/');

        // Replace locale in URL path
        if (locales.includes(segments[1] as Locale)) {
            segments[1] = newLocale;
        } else {
            segments.splice(1, 0, newLocale);
        }

        const newPath = segments.join('/') || `/${newLocale}`;

        startTransition(() => {
            router.replace(newPath);
            router.refresh();
        });
    };

    return (
        <DropdownMenu>
            <DropdownMenuTrigger asChild>
                <Button
                    variant="ghost"
                    size="sm"
                    className="flex items-center gap-2 hover:bg-gray-100 transition-colors"
                    disabled={isPending}
                    aria-label="Change language"
                >
                    <Globe className="h-4 w-4" />
                    <span className="hidden sm:inline font-semibold">
                        {localeFlags[locale]} {localeNames[locale]}
                    </span>
                    <span className="sm:hidden text-lg">{localeFlags[locale]}</span>
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="min-w-[150px]">
                {locales.map((loc) => (
                    <DropdownMenuItem
                        key={loc}
                        onClick={() => handleLocaleChange(loc)}
                        className={`cursor-pointer ${locale === loc ? 'bg-accent font-semibold' : ''
                            }`}
                    >
                        <span className="mr-2 text-lg">{localeFlags[loc]}</span>
                        <span>{localeNames[loc]}</span>
                    </DropdownMenuItem>
                ))}
            </DropdownMenuContent>
        </DropdownMenu>
    );
}
