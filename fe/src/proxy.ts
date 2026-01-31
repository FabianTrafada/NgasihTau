import createMiddleware from 'next-intl/middleware';
import { routing } from '@/i18n/routing';
import { NextRequest } from 'next/server';

const intlMiddleware = createMiddleware(routing);

export default function middleware(request: NextRequest) {
    const { pathname } = request.nextUrl;

    // Run intl middleware first
    const intlResponse = intlMiddleware(request);

    // Get locale from pathname (e.g., /en/dashboard -> en)
    const pathnameLocale = routing.locales.find(
        (locale) => pathname.startsWith(`/${locale}/`) || pathname === `/${locale}`
    );
    const locale = pathnameLocale || routing.defaultLocale;

    // Auth guard - check paths without locale prefix
    const pathnameWithoutLocale = pathnameLocale
        ? pathname.replace(`/${pathnameLocale}`, '') || '/'
        : pathname;

    // Let individual pages handle authentication checks via ProtectedRoute component
    // This provides more flexibility for:
    // - Email verification redirects
    // - Onboarding flow
    // - Role-based access
    // Protected pages MUST wrap content with <ProtectedRoute> component

    return intlResponse;
}

export const config = {
    // Match all pathnames except:
    // - API routes
    // - Static files (images, fonts, etc.)
    // - Next.js internals
    matcher: [
        // Match all pathnames except static files
        '/((?!api|_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp|ico|css|js|woff|woff2|ttf|eot)$).*)',
        // Match root
        '/'
    ]
};