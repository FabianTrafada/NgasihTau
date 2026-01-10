import React, { useTransition } from 'react'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '../ui/dropdown-menu'
import { Button } from '../ui/button'
import { Globe } from 'lucide-react'
import { usePathname, useRouter } from 'next/navigation'
import { useLocale } from 'next-intl';

import { locales, localeFlags, Locale, localNames } from '@/i18n/config';

const LanguageSwitch = () => {
    const locale = useLocale() as Locale;
    const router = useRouter();
    const pathname = usePathname();
    const [isPending, startTransition] = useTransition();

    const handleChange = (newLocale: Locale) => {
        const segments = pathname.split('/');

        if(locale.includes(segments[1] as Locale)) {
            segments[1] = newLocale;
        } else {
            segments.splice(1, 0, newLocale);
        }

        const newPath = segments.join('/') || '/';

        startTransition(() => {
            router.replace(newPath);
        } );
    }

    return (
        <DropdownMenu>
            <DropdownMenuTrigger asChild>
                <Button
                    variant="ghost"
                    size="sm"
                    className="flex items-center gap-2"
                    disabled={isPending}
                >
                    <Globe className="h-4 w-4" />
                    <span className="hidden sm:inline">
                        {localeFlags[locale]} {localNames[locale]}
                    </span>
                    <span className="sm:hidden">{localeFlags[locale]}</span>
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
                {locales.map((loc) => (
                    <DropdownMenuItem
                        key={loc}
                        onClick={() => handleChange(loc)}
                        className={locale === loc ? 'bg-accent' : ''}
                    >
                        <span className="mr-2">{localeFlags[loc]}</span>
                        {localNames[loc]}
                    </DropdownMenuItem>
                ))}
            </DropdownMenuContent>
        </DropdownMenu>)
}

export default LanguageSwitch