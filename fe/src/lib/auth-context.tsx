// =============================================================================
// AUTH CONTEXT
// =============================================================================
// This creates a React Context to manage authentication state globally.
// Context allows you to share state across all components without prop drilling.
//
// HOW IT WORKS:
// 1. AuthProvider wraps your entire app (in layout.tsx)
// 2. Any component can access auth state using the useAuth() hook
// 3. When auth state changes, all components using useAuth() re-render

"use client";

import React, { createContext, useContext, useState, useEffect, useCallback } from "react";
import {
    User,
    AuthResponse,
    TwoFactorRequiredResponse,
    login as loginApi,
    register as registerApi,
    logout as logoutApi,
    googleLogin as googleLoginApi,
    getCurrentUser,
    getStoredUser,
    verify2FALogin as verify2FALoginApi,
    isTwoFactorRequired,
    getGoogleAuthUrl,
    getGoogleRedirectUri,
} from "./api/auth";
import { TokenStorage } from "./api-client";

// =============================================================================
// TYPES
// =============================================================================

/**
 * Error that can be displayed to users
 */
export interface AuthError {
    message: string;
    code?: string;
}

/**
 * Result of a login attempt - could require 2FA
 */
export interface LoginResult {
    success: boolean;
    requires2FA: boolean;
    tempToken?: string; // Only present if 2FA required
}

/**
 * Everything the AuthContext provides to components
 */
interface AuthContextType {
    // State
    user: User | null;
    loading: boolean;
    error: AuthError | null;

    // Auth actions
    login: (email: string, password: string) => Promise<LoginResult>;
    register: (email: string, password: string, name: string) => Promise<void>;
    logout: () => Promise<void>;
    verify2FA: (tempToken: string, code: string) => Promise<void>;

    // Google OAuth
    initiateGoogleLogin: () => void;
    handleGoogleCallback: (code: string) => Promise<LoginResult>;

    // Utility
    refreshUser: () => Promise<void>;
    clearError: () => void;
    isAuthenticated: boolean;
}

// =============================================================================
// CONTEXT
// =============================================================================

// Create context with undefined default (will be provided by AuthProvider)
const AuthContext = createContext<AuthContextType | undefined>(undefined);

// =============================================================================
// ERROR HELPER
// =============================================================================

/**
 * Extract user-friendly error message from various error types
 */
function getErrorMessage(error: unknown): string {
    if (error instanceof Error) {
        // Check for Axios error with response
        const axiosError = error as { response?: { data?: { error?: { message?: string } } } };
        if (axiosError.response?.data?.error?.message) {
            return axiosError.response.data.error.message;
        }
        return error.message;
    }
    return "An unexpected error occurred";
}

// =============================================================================
// AUTH PROVIDER COMPONENT
// =============================================================================

/**
 * AuthProvider component wraps your app and provides authentication state
 *
 * Usage in layout.tsx:
 * ```tsx
 * <AuthProvider>
 *   <App />
 * </AuthProvider>
 * ```
 */
export function AuthProvider({ children }: { children: React.ReactNode }) {
    // ===========================================================================
    // STATE
    // ===========================================================================

    const [user, setUser] = useState<User | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<AuthError | null>(null);

    // ===========================================================================
    // INITIALIZATION
    // ===========================================================================

    /**
     * Effect runs once when app loads
     * Tries to restore session from localStorage
     */
    useEffect(() => {
        const initializeAuth = async () => {
            // Try to get user from localStorage first (fast)
            const storedUser = getStoredUser();

            if (storedUser) {
                // Show stored user immediately for better UX
                setUser(storedUser);

                try {
                    // Verify token is still valid by fetching fresh user data
                    const freshUser = await getCurrentUser();
                    setUser(freshUser);
                } catch {
                    // Token expired or invalid - clear everything
                    setUser(null);
                    TokenStorage.clearAll();
                }
            }

            setLoading(false);
        };

        initializeAuth();
    }, []);

    // ===========================================================================
    // AUTH ACTIONS
    // ===========================================================================

    /**
     * Login with email and password
     * Returns LoginResult indicating if 2FA is required
     */
    const login = useCallback(async (email: string, password: string): Promise<LoginResult> => {
        console.log('[AuthContext] ========== LOGIN ==========');
        console.log('[AuthContext] Email:', email);

        setError(null);
        setLoading(true);

        try {
            console.log('[AuthContext] Calling loginApi...');
            const result = await loginApi({ email, password });
            console.log('[AuthContext] Login API result:', result);

            // Check if 2FA is required
            if (isTwoFactorRequired(result)) {
                console.log('[AuthContext] 2FA is required');

                

                setLoading(false);
                return {
                    success: true,
                    requires2FA: true,
                    tempToken: result.temp_token,
                };
            }

            // Full login successful
            console.log('[AuthContext] ✅ Login successful! User:', result.user);
            console.log('[AuthContext] Setting user state...');
            setUser(result.user);

            // Verify tokens are saved
            const savedTokens = {
                access: localStorage.getItem('access_token'),
                refresh: localStorage.getItem('refresh_token'),
                user: localStorage.getItem('user')
            };
            console.log('[AuthContext] Tokens saved in localStorage:', {
                hasAccessToken: !!savedTokens.access,
                hasRefreshToken: !!savedTokens.refresh,
                hasUser: !!savedTokens.user
            });

            setLoading(false);
            console.log('[AuthContext] ========== LOGIN SUCCESS ==========');
            return {
                success: true,
                requires2FA: false,
            };
        } catch (err) {
            console.error('[AuthContext] ❌ Login failed:', err);
            setError({ message: getErrorMessage(err) });
            setLoading(false);
            throw err;
        }
    }, []);

    /**
     * Complete 2FA verification
     */
    const verify2FA = useCallback(async (tempToken: string, code: string): Promise<void> => {
        setError(null);
        setLoading(true);

        try {
            const result = await verify2FALoginApi(tempToken, code);
            setUser(result.user);
        } catch (err) {
            setError({ message: getErrorMessage(err) });
            throw err;
        } finally {
            setLoading(false);
        }
    }, []);

    /**
     * Register new user
     */
    const register = useCallback(async (email: string, password: string, name: string): Promise<void> => {
        setError(null);
        setLoading(true);

        try {
            const result = await registerApi({ email, password, name });
            setUser(result.user);
        } catch (err) {
            setError({ message: getErrorMessage(err) });
            throw err;
        } finally {
            setLoading(false);
        }
    }, []);

    // ===========================================================================
    // GOOGLE OAUTH
    // ===========================================================================

    /**
     * Initiate Google OAuth flow
     * Redirects user to Google's authorization page
     */
    const initiateGoogleLogin = useCallback(() => {
        const authUrl = getGoogleAuthUrl();
        window.location.href = authUrl;
    }, []);

    /**
     * Handle Google OAuth callback
     * Called when user returns from Google with authorization code
     */
    const handleGoogleCallback = useCallback(async (code: string): Promise<LoginResult> => {
        setError(null);
        setLoading(true);

        try {
            const redirectUri = getGoogleRedirectUri();
            const result = await googleLoginApi(code, redirectUri);

            // Check if 2FA is required
            if (isTwoFactorRequired(result)) {
                setLoading(false);
                return {
                    success: true,
                    requires2FA: true,
                    tempToken: result.temp_token,
                };
            }

            // Full login successful
            setUser(result.user);
            setLoading(false);
            return {
                success: true,
                requires2FA: false,
            };
        } catch (err) {
            setError({ message: getErrorMessage(err) });
            setLoading(false);
            throw err;
        }
    }, []);

    /**
     * Logout current user
     */
    const logout = useCallback(async (): Promise<void> => {
        setError(null);

        try {
            await logoutApi();
        } finally {
            // Always clear user state
            setUser(null);
        }
    }, []);

    /**
     * Refresh user data from server
     */
    const refreshUser = useCallback(async (): Promise<void> => {
        try {
            const freshUser = await getCurrentUser();
            setUser(freshUser);
        } catch (err) {
            setError({ message: getErrorMessage(err) });
            throw err;
        }
    }, []);

    /**
     * Clear any displayed error
     */
    const clearError = useCallback(() => {
        setError(null);
    }, []);

    // ===========================================================================
    // CONTEXT VALUE
    // ===========================================================================

    const value: AuthContextType = {
        // State
        user,
        loading,
        error,

        // Actions
        login,
        register,
        logout,
        verify2FA,

        // Google OAuth
        initiateGoogleLogin,
        handleGoogleCallback,

        // Utility
        refreshUser,
        clearError,
        isAuthenticated: !!user,
    };

    return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

// =============================================================================
// HOOK
// =============================================================================

/**
 * Custom hook to access authentication context
 *
 * Usage in any component:
 * ```tsx
 * function MyComponent() {
 *   const { user, login, logout, isAuthenticated } = useAuth();
 *
 *   if (!isAuthenticated) {
 *     return <LoginButton onClick={() => login(email, password)} />;
 *   }
 *
 *   return <p>Welcome, {user.name}!</p>;
 * }
 * ```
 */
export function useAuth(): AuthContextType {
    const context = useContext(AuthContext);

    if (context === undefined) {
        throw new Error("useAuth must be used within an AuthProvider");
    }

    return context;
}