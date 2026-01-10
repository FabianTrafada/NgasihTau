import { NextRequest, NextResponse } from 'next/server';
import createMiddleware from 'next-intl/middleware';
import { routing } from '@/i18n/routing';

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

    const accessToken = request.cookies.get('access_token')?.value;

    const authRoutes = ['/sign-in', '/sign-up'];
    const protectedRoutes = ['/dashboard', '/profile', '/onboarding', '/teacher'];

    // Redirect authenticated users away from auth pages
    if (accessToken && authRoutes.some((r) => pathnameWithoutLocale.startsWith(r))) {
        const url = request.nextUrl.clone();
        url.pathname = `/${locale}/dashboard`;
        return NextResponse.redirect(url);
    }

    // Redirect unauthenticated users away from protected pages
    if (!accessToken && protectedRoutes.some((r) => pathnameWithoutLocale.startsWith(r))) {
        const url = request.nextUrl.clone();
        url.pathname = `/${locale}/sign-in`;
        return NextResponse.redirect(url);
    }

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

