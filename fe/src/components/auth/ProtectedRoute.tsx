// =============================================================================
// PROTECTED ROUTE COMPONENT
// =============================================================================
// This component wraps pages that require authentication.
// If user is not authenticated, they are redirected to sign-in page.
//
// USAGE:
// ```tsx
// // In any protected page (e.g., dashboard/page.tsx)
// export default function DashboardPage() {
//   return (
//     <ProtectedRoute>
//       <DashboardContent />
//     </ProtectedRoute>
//   );
// }
// ```

"use client";

import { useEffect } from "react";
import { useRouter, usePathname } from "next/navigation";
import { useAuth } from "@/lib/auth-context";

interface ProtectedRouteProps {
    children: React.ReactNode;
    /**
     * Optional: Fallback component to show while checking auth
     * Defaults to a simple loading spinner
     */
    loadingComponent?: React.ReactNode;
    /**
     * Optional: Where to redirect if not authenticated
     * Defaults to "/sign-in"
     */
    redirectTo?: string;
}

/**
 * Default loading component
 */
function DefaultLoading() {
    return (
        <div className="min-h-screen flex items-center justify-center bg-[#FFFBF7]">
            <div className="flex flex-col items-center gap-4">
                {/* Spinner */}
                <div className="w-12 h-12 border-4 border-[#FF8811] border-t-transparent rounded-full animate-spin" />
                {/* Loading text */}
                <p className="text-gray-500 text-sm font-[family-name:var(--font-inter)]">
                    Loading...
                </p>
            </div>
        </div>
    );
}

/**
 * ProtectedRoute Component
 *
 * HOW IT WORKS:
 * 1. On mount, check if auth context is still loading
 * 2. If loading, show loading component
 * 3. If not authenticated, redirect to sign-in
 * 4. If authenticated but email not verified, redirect to verify-waiting
 * 5. If authenticated, render children (children can do their own checks like onboarding)
 */
export function ProtectedRoute({
    children,
    loadingComponent,
    redirectTo = "/sign-in",
}: ProtectedRouteProps) {
    const router = useRouter();
    const pathname = usePathname();
    const { isAuthenticated, loading, user } = useAuth();

    // Extract locale from pathname (e.g., /en/dashboard -> en)
    const pathParts = pathname.split('/').filter(Boolean);
    const locale = pathParts[0] && ['en', 'id'].includes(pathParts[0]) ? pathParts[0] : 'id';

    useEffect(() => {
        console.log('[ProtectedRoute] Auth state:', {
            loading,
            isAuthenticated,
            hasUser: !!user,
            userEmail: user?.email,
            emailVerified: user?.email_verified,
            pathname,
            locale
        });

        // Wait for loading to complete before making redirect decision
        if (!loading) {
            if (!isAuthenticated) {
                console.warn('[ProtectedRoute] ⚠️ Not authenticated - redirecting to sign-in');
                // Use relative path without adding locale again (router handles it)
                const currentPath = window.location.pathname;
                const redirectUrl = `${redirectTo}?redirect=${encodeURIComponent(currentPath)}`;
                console.log('[ProtectedRoute] Redirect URL:', redirectUrl);
                router.push(redirectUrl);
            } else if (user && !user.email_verified) {
                console.warn('[ProtectedRoute] ⚠️ Email not verified - redirecting to verify-waiting');
                router.push(`/verify-waiting?email=${encodeURIComponent(user.email)}`);
            } else if (isAuthenticated) {
                console.log('[ProtectedRoute] ✅ Authenticated and email verified - rendering protected content');
            }
        }
    }, [loading, isAuthenticated, user, router, redirectTo, pathname, locale]);

    // Show loading state while checking auth
    if (loading) {
        return <>{loadingComponent || <DefaultLoading />}</>;
    }

    // If not authenticated or email not verified, don't render children (redirect is in progress)
    if (!isAuthenticated || (user && !user.email_verified)) {
        return <>{loadingComponent || <DefaultLoading />}</>;
    }

    // User is authenticated and email verified, render the protected content
    return <>{children}</>;
}

/**
 * Higher-order component version (alternative API)
 *
 * USAGE:
 * ```tsx
 * function DashboardPage() {
 *   return <div>Dashboard</div>;
 * }
 *
 * export default withAuth(DashboardPage);
 * ```
 */
export function withAuth<P extends object>(
    Component: React.ComponentType<P>,
    options?: { redirectTo?: string }
) {
    return function AuthenticatedComponent(props: P) {
        return (
            <ProtectedRoute redirectTo={options?.redirectTo}>
                <Component {...props} />
            </ProtectedRoute>
        );
    };
}

// =============================================================================
// PUBLIC ONLY ROUTE (opposite of ProtectedRoute)
// =============================================================================
// Use this for pages that should only be accessible when NOT logged in
// (e.g., sign-in, sign-up pages)

interface PublicOnlyRouteProps {
    children: React.ReactNode;
    /**
     * Where to redirect if user IS authenticated
     * Defaults to "/dashboard"
     */
    redirectTo?: string;
}

/**
 * PublicOnlyRoute Component
 *
 * Redirects authenticated users away from public pages (like login/register)
 * to prevent them from seeing those pages when already logged in.
 * 
 * However, if email is not verified, allow them to stay on public pages
 * so they can access verify-waiting or sign-in again.
 */
export function PublicOnlyRoute({
    children,
    redirectTo = "/dashboard",
}: PublicOnlyRouteProps) {
    const router = useRouter();
    const pathname = usePathname();
    const { isAuthenticated, loading, user } = useAuth();

    // Extract locale from pathname (e.g., /en/sign-in -> en)
    const pathParts = pathname.split('/').filter(Boolean);
    const locale = pathParts[0] && ['en', 'id'].includes(pathParts[0]) ? pathParts[0] : 'id';

    useEffect(() => {
        console.log('[PublicOnlyRoute] Auth state:', {
            loading,
            isAuthenticated,
            hasUser: !!user,
            emailVerified: user?.email_verified,
            redirectTo,
            pathname,
            locale
        });

        if (!loading && isAuthenticated && user) {
            // Only redirect if email is verified
            // If email not verified, let them stay on public pages
            if (user.email_verified) {
                console.log('[PublicOnlyRoute] ✅ User authenticated and verified - redirecting to:', redirectTo);
                // Use relative path without adding locale again (router handles it)
                router.push(redirectTo);
            } else {
                console.log('[PublicOnlyRoute] User authenticated but email not verified - staying on public page');
            }
        } else if (!loading && !isAuthenticated) {
            console.log('[PublicOnlyRoute] User not authenticated - showing public page');
        }
    }, [loading, isAuthenticated, router, redirectTo, user, pathname, locale]);

    // Show loading while checking auth
    if (loading) {
        return <DefaultLoading />;
    }

    // If authenticated and verified, don't render (redirect in progress)
    if (isAuthenticated && user?.email_verified) {
        return <DefaultLoading />;
    }

    // Not authenticated OR not verified, show the public page
    return <>{children}</>;
}
