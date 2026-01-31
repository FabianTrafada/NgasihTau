// =============================================================================
// ROLE PROTECTED ROUTE COMPONENT
// =============================================================================
// This component extends ProtectedRoute to check for specific user roles.
// If user is authenticated but doesn't have the required role, they are
// redirected to an access denied page.
//
// USAGE:
// ```tsx
// // In a teacher-only page (e.g., teacher/dashboard/page.tsx)
// export default function TeacherDashboardPage() {
//   return (
//     <RoleProtectedRoute requiredRole="teacher">
//       <TeacherDashboardContent />
//     </RoleProtectedRoute>
//   );
// }
// ```

"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth-context";

interface RoleProtectedRouteProps {
    children: React.ReactNode;
    /**
     * Required role(s) to access this route
     * Can be a single role or array of roles
     */
    requiredRole: string | string[];
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
    /**
     * Optional: Where to redirect if role check fails
     * Defaults to "/dashboard"
     */
    accessDeniedRedirect?: string;
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
                    Verifying permissions...
                </p>
            </div>
        </div>
    );
}

/**
 * Check if user has any of the required roles
 */
function hasRequiredRole(userRole: string, requiredRole: string | string[]): boolean {
    if (Array.isArray(requiredRole)) {
        return requiredRole.includes(userRole);
    }
    return userRole === requiredRole;
}

/**
 * RoleProtectedRoute Component
 *
 * HOW IT WORKS:
 * 1. First checks if user is authenticated
 * 2. If not authenticated, redirect to sign-in
 * 3. If authenticated, check if user has required role
 * 4. If role check fails, redirect to access denied page
 * 5. If all checks pass, render children
 */
export function RoleProtectedRoute({
    children,
    requiredRole,
    loadingComponent,
    redirectTo = "/sign-in",
    accessDeniedRedirect = "/access-denied",
}: RoleProtectedRouteProps) {
    const router = useRouter();
    const { isAuthenticated, loading, user } = useAuth();
    const [isRedirecting, setIsRedirecting] = useState(false);

    useEffect(() => {
        console.log('[RoleProtectedRoute] Auth state:', {
            loading,
            isAuthenticated,
            hasUser: !!user,
            userRole: user?.role,
            requiredRole,
        });

        // Wait for loading to complete before making redirect decision
        if (!loading) {
            // Check authentication first
            if (!isAuthenticated || !user) {
                console.warn('[RoleProtectedRoute] ⚠️ Not authenticated - redirecting to sign-in');
                const currentPath = window.location.pathname;
                const redirectUrl = `${redirectTo}?redirect=${encodeURIComponent(currentPath)}`;
                router.push(redirectUrl);
                return;
            }

            // Check role permission
            if (!hasRequiredRole(user.role, requiredRole)) {
                console.warn(
                    `[RoleProtectedRoute] ⚠️ Access denied - User role "${user.role}" doesn't match required role(s):`,
                    requiredRole
                );

                // Show loading for 4 seconds then redirect
                const timer = setTimeout(() => {
                    setIsRedirecting(true);
                    router.push(accessDeniedRedirect);
                }, 4000);
                return () => clearTimeout(timer);
                return;
            }

            console.log('[RoleProtectedRoute] ✅ Authenticated and authorized - rendering protected content');
        }
    }, [loading, isAuthenticated, user, requiredRole, router, redirectTo, accessDeniedRedirect]);

    // Show loading state while checking auth and role
    if (loading || isRedirecting) {
        return <>{loadingComponent || <DefaultLoading />}</>;
    }

    // If not authenticated or role doesn't match, don't render children (redirect is in progress)
    if (!isAuthenticated || !user || !hasRequiredRole(user.role, requiredRole)) {
        return <>{loadingComponent || <DefaultLoading />}</>;
    }

    // User is authenticated and has required role - render children
    return <>{children}</>;
}
