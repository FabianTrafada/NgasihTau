// =============================================================================
// AUTH SERVICE
// =============================================================================
// This module handles all authentication-related API calls and token management.
// It's the "service layer" that communicates with the backend.

import apiClient, { TokenStorage } from "../api-client";

// =============================================================================
// TYPES & INTERFACES
// =============================================================================

/**
 * Data required for user registration
 */
export interface RegisterData {
    email: string;
    password: string;
    name: string;
}

/**
 * Data required for user login
 */
export interface LoginData {
    email: string;
    password: string;
}

/**
 * Data required for Google OAuth login
 */
export interface GoogleLoginData {
    code: string;
    redirect_uri: string;
}

/**
 * Response from successful authentication (login/register)
 * Contains user data and JWT tokens
 */
export interface AuthResponse {
    user: User;
    access_token: string;
    refresh_token: string;
    expires_in: number; // Token expiration in seconds
}

/**
 * Response when 2FA is required
 * User must complete 2FA verification with the temp_token
 */
export interface TwoFactorRequiredResponse {
    temp_token: string;
    two_factor_method: "totp" | "backup_code";
}

/**
 * User object from the backend
 */
export interface User {
    id: string;
    email: string;
    name: string;
    username?: string;
    bio?: string;
    avatar_url?: string;
    role: string;
    email_verified: boolean;
    two_factor_enabled: boolean;
    created_at: string;
    updated_at: string;
}

/**
 * Standard API response wrapper from the backend
 * All API responses follow this format
 */
interface ApiResponse<T> {
    success: boolean;
    data: T;
    request_id: string;
    timestamp: string;
}

/**
 * Union type for login response (either full auth or 2FA required)
 */
type LoginResponse = AuthResponse | TwoFactorRequiredResponse;

// =============================================================================
// GOOGLE OAUTH CONFIG
// =============================================================================

/**
 * Google OAuth configuration
 * These values must match your Google Cloud Console settings
 */
const GOOGLE_CLIENT_ID = process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID || "";
const GOOGLE_REDIRECT_URI =
    typeof window !== "undefined"
        ? `${window.location.origin}/auth/google/callback`
        : "";

/**
 * Google OAuth authorization URL
 * Redirect users here to start the OAuth flow
 */
export function getGoogleAuthUrl(): string {
    const params = new URLSearchParams({
        client_id: GOOGLE_CLIENT_ID,
        redirect_uri: GOOGLE_REDIRECT_URI,
        response_type: "code",
        scope: "openid email profile",
        access_type: "offline",
        prompt: "consent",
    });

    return `https://accounts.google.com/o/oauth2/v2/auth?${params.toString()}`;
}

/**
 * Get the redirect URI for Google OAuth
 * Used when exchanging the code
 */
export function getGoogleRedirectUri(): string {
    return GOOGLE_REDIRECT_URI;
}

// =============================================================================
// TYPE GUARDS
// =============================================================================

/**
 * Check if login response requires 2FA
 * Type guard to help TypeScript understand the response type
 */
export function isTwoFactorRequired(
    response: LoginResponse
): response is TwoFactorRequiredResponse {
    return "temp_token" in response;
}

// =============================================================================
// AUTHENTICATION FUNCTIONS
// =============================================================================

/**
 * Register a new user account
 *
 * Flow:
 * 1. Send registration data to backend
 * 2. Backend creates user and returns tokens
 * 3. Store tokens in localStorage
 * 4. Return auth data for context update
 *
 * @param data - User registration data (email, password, name)
 * @returns AuthResponse with user data and tokens
 * @throws Error if registration fails (e.g., email already exists)
 */
export async function register(data: RegisterData): Promise<AuthResponse> {
    const response = await apiClient.post<ApiResponse<AuthResponse>>(
        "/api/v1/auth/register",
        data
    );

    const authData = response.data.data;

    // Store tokens and user in localStorage
    TokenStorage.setTokens(authData.access_token, authData.refresh_token);
    localStorage.setItem("user", JSON.stringify(authData.user));

    return authData;
}

/**
 * Login with email and password
 *
 * Flow:
 * 1. Send credentials to backend
 * 2. Backend validates and returns either:
 *    a. AuthResponse (tokens + user) - if no 2FA
 *    b. TwoFactorRequiredResponse - if 2FA enabled
 * 3. If AuthResponse, store tokens in localStorage
 * 4. Return response for context/component handling
 *
 * @param data - Login credentials (email, password)
 * @returns Either AuthResponse or TwoFactorRequiredResponse
 * @throws Error if credentials are invalid
 */
export async function login(data: LoginData): Promise<LoginResponse> {
    const response = await apiClient.post<ApiResponse<LoginResponse>>(
        "/api/v1/auth/login",
        data
    );

    const loginData = response.data.data;

    // Only store tokens if it's a full auth response (no 2FA required)
    if (!isTwoFactorRequired(loginData)) {
        TokenStorage.setTokens(loginData.access_token, loginData.refresh_token);
        localStorage.setItem("user", JSON.stringify(loginData.user));
    }

    return loginData;
}

/**
 * Login/Register with Google OAuth
 *
 * Flow:
 * 1. User clicks "Sign in with Google" → Redirected to Google
 * 2. User authorizes → Google redirects back with authorization code
 * 3. Frontend catches the code → Sends to this function
 * 4. Backend exchanges code for Google tokens → Returns our JWT tokens
 *
 * Note: This handles both login AND registration!
 * If the Google email doesn't exist, backend creates a new account.
 *
 * @param code - Authorization code from Google OAuth redirect
 * @param redirectUri - The redirect URI used in the OAuth flow
 * @returns Either AuthResponse or TwoFactorRequiredResponse
 */
export async function googleLogin(
    code: string,
    redirectUri: string
): Promise<LoginResponse> {
    const response = await apiClient.post<ApiResponse<LoginResponse>>(
        "/api/v1/auth/google",
        {
            code,
            redirect_uri: redirectUri,
        }
    );

    const loginData = response.data.data;

    // Only store tokens if it's a full auth response (no 2FA required)
    if (!isTwoFactorRequired(loginData)) {
        TokenStorage.setTokens(loginData.access_token, loginData.refresh_token);
        localStorage.setItem("user", JSON.stringify(loginData.user));
    }

    return loginData;
}

/**
 * Complete 2FA login verification
 *
 * Flow:
 * 1. User enters TOTP code or backup code
 * 2. Send temp_token + code to backend
 * 3. Backend validates and returns full AuthResponse
 * 4. Store tokens and return auth data
 *
 * @param tempToken - Temporary token from login response
 * @param code - 6-digit TOTP code or backup code (XXXX-XXXX format)
 * @returns AuthResponse with user data and tokens
 */
export async function verify2FALogin(
    tempToken: string,
    code: string
): Promise<AuthResponse> {
    const response = await apiClient.post<ApiResponse<AuthResponse>>(
        "/api/v1/auth/2fa/login",
        {
            temp_token: tempToken,
            code: code,
        }
    );

    const authData = response.data.data;

    // Now store the real tokens
    TokenStorage.setTokens(authData.access_token, authData.refresh_token);
    localStorage.setItem("user", JSON.stringify(authData.user));

    return authData;
}

/**
 * Logout the current user
 *
 * Flow:
 * 1. Get refresh token from localStorage
 * 2. Call backend to invalidate the refresh token (revoke it)
 * 3. Clear all local storage (tokens + user)
 * 4. Always clear local data even if backend call fails
 *
 * Note: We don't redirect here - let the calling component handle navigation
 */
export async function logout(): Promise<void> {
    try {
        const refreshToken = TokenStorage.getRefreshToken();

        if (refreshToken) {
            // Tell backend to invalidate this refresh token
            await apiClient.post("/api/v1/auth/logout", {
                refresh_token: refreshToken,
            });
        }
    } catch (error) {
        // Log but don't throw - we still want to clear local data
        console.error("Logout API error:", error);
    } finally {
        // Always clear local storage
        TokenStorage.clearAll();
    }
}

// =============================================================================
// USER DATA FUNCTIONS
// =============================================================================

/**
 * Get the current authenticated user from the backend
 *
 * This is used to:
 * 1. Verify the token is still valid
 * 2. Get fresh user data (might have changed on another device)
 *
 * @returns User object
 * @throws Error if not authenticated or token expired
 */
export async function getCurrentUser(): Promise<User> {
    const response = await apiClient.get<ApiResponse<User>>("/api/v1/users/me");

    const user = response.data.data;

    // Update stored user data with fresh data
    localStorage.setItem("user", JSON.stringify(user));

    return user;
}

/**
 * Check if user is authenticated
 * Simple check - just looks for access token in localStorage
 *
 * Note: This doesn't verify the token is valid!
 * Use getCurrentUser() for that.
 *
 * @returns true if access token exists
 */
export function isAuthenticated(): boolean {
    return !!TokenStorage.getAccessToken();
}

/**
 * Get stored user from localStorage
 * Used for initial page load before API validation
 *
 * @returns User object or null if not found/invalid
 */
export function getStoredUser(): User | null {
    if (typeof window === "undefined") return null;

    const userStr = localStorage.getItem("user");
    if (!userStr) return null;

    try {
        return JSON.parse(userStr) as User;
    } catch {
        // Invalid JSON, clear it
        localStorage.removeItem("user");
        return null;
    }
}

// =============================================================================
// PASSWORD RESET FUNCTIONS
// =============================================================================

/**
 * Request a password reset email
 *
 * @param email - User's email address
 */
export async function requestPasswordReset(email: string): Promise<void> {
    await apiClient.post("/api/v1/auth/password/forgot", { email });
}

/**
 * Reset password with token from email
 *
 * @param token - Reset token from email link
 * @param newPassword - New password to set
 */
export async function resetPassword(
    token: string,
    newPassword: string
): Promise<void> {
    await apiClient.post("/api/v1/auth/password/reset", {
        token,
        new_password: newPassword,
    });
}

// =============================================================================
// EMAIL VERIFICATION FUNCTIONS
// =============================================================================

/**
 * Verify email with token from verification email
 *
 * @param token - Verification token from email link
 */
export async function verifyEmail(token: string): Promise<void> {
    await apiClient.post("/api/v1/auth/verify-email", { token });
}

/**
 * Request a new verification email
 * Requires user to be authenticated
 */
export async function sendVerificationEmail(): Promise<void> {
    await apiClient.post("/api/v1/auth/send-verification");
}