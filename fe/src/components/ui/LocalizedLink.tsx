'use client';

import Link from 'next/link';
import { useLocale } from 'next-intl';
import { Locale } from '@/i18n/config';
import { ComponentProps } from 'react';

type LocalizedLinkProps = Omit<ComponentProps<typeof Link>, 'href'> & {
    href: string;
};

/**
 * Localized Link Component
 * Automatically prefixes URLs with current locale
 * Use this instead of Next.js Link for internal navigation
 * 
 * @example
 * <LocalizedLink href="/dashboard">Dashboard</LocalizedLink>
 * // Renders: /en/dashboard or /id/dashboard based on current locale
 */
export function LocalizedLink({ href, ...props }: LocalizedLinkProps) {
    const locale = useLocale() as Locale;

    // Don't localize external links or already localized URLs
    const shouldLocalize =
        !href.startsWith('http') &&
        !href.startsWith('mailto:') &&
        !href.startsWith('tel:') &&
        !href.startsWith(`/${locale}`);

    const localizedHref = shouldLocalize
        ? `/${locale}${href.startsWith('/') ? href : `/${href}`}`
        : href;

    return <Link href={localizedHref} {...props} />;
}
