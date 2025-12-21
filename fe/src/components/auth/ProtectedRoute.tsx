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
import { useRouter } from "next/navigation";
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
 * 4. If authenticated, render children
 */
export function ProtectedRoute({
    children,
    loadingComponent,
    redirectTo = "/sign-in",
}: ProtectedRouteProps) {
    const router = useRouter();
    const { isAuthenticated, loading } = useAuth();

    useEffect(() => {
        // Wait for loading to complete before making redirect decision
        if (!loading && !isAuthenticated) {
            // Save the current URL to redirect back after login
            const currentPath = window.location.pathname;
            const redirectUrl = `${redirectTo}?redirect=${encodeURIComponent(currentPath)}`;
            router.push(redirectUrl);
        }
    }, [loading, isAuthenticated, router, redirectTo]);

    // Show loading state while checking auth
    if (loading) {
        return <>{loadingComponent || <DefaultLoading />}</>;
    }

    // If not authenticated, don't render children (redirect is in progress)
    if (!isAuthenticated) {
        return <>{loadingComponent || <DefaultLoading />}</>;
    }

    // User is authenticated, render the protected content
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
 */
export function PublicOnlyRoute({
    children,
    redirectTo = "/dashboard",
}: PublicOnlyRouteProps) {
    const router = useRouter();
    const { isAuthenticated, loading } = useAuth();

    useEffect(() => {
        if (!loading && isAuthenticated) {
            router.push(redirectTo);
        }
    }, [loading, isAuthenticated, router, redirectTo]);

    // Show loading while checking auth
    if (loading) {
        return <DefaultLoading />;
    }

    // If authenticated, don't render (redirect in progress)
    if (isAuthenticated) {
        return <DefaultLoading />;
    }

    // Not authenticated, show the public page
    return <>{children}</>;
}
