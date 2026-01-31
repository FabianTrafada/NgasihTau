import axios, { AxiosError, InternalAxiosRequestConfig } from "axios";

// =============================================================================
// CONFIGURATION
// =============================================================================

/**
 * Base URL for API requests
 * - Uses environment variable if available (for production)
 * - Falls back to localhost:8000 (Traefik gateway) for development
 */
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8000";

// =============================================================================
// TOKEN STORAGE HELPERS
// =============================================================================

/**
 * Helper functions to safely access localStorage
 * These handle SSR (Server-Side Rendering) where localStorage doesn't exist
 */
const TokenStorage = {
    getAccessToken: (): string | null => {
        if (typeof window === "undefined") return null;
        return localStorage.getItem("access_token");
    },

    getRefreshToken: (): string | null => {
        if (typeof window === "undefined") return null;
        return localStorage.getItem("refresh_token");
    },

    setTokens: (accessToken: string, refreshToken: string): void => {
        if (typeof window === "undefined") return;

        // Store in localStorage
        localStorage.setItem("access_token", accessToken);
        localStorage.setItem("refresh_token", refreshToken);

        // ALSO store in cookies for middleware
        document.cookie = `access_token=${accessToken};path=/;max-age=${7 * 24 * 60 * 60};SameSite=Lax`;
        document.cookie = `refresh_token=${refreshToken};path=/;max-age=${30 * 24 * 60 * 60};SameSite=Lax`;
    },

    clearAll: (): void => {
        if (typeof window === "undefined") return;

        // Clear localStorage
        localStorage.removeItem("access_token");
        localStorage.removeItem("refresh_token");
        localStorage.removeItem("user");

        // Clear cookies
        document.cookie = "access_token=;path=/;max-age=0";
        document.cookie = "refresh_token=;path=/;max-age=0";
    },
};

// =============================================================================
// AXIOS INSTANCE
// =============================================================================

/**
 * Configured Axios instance for API calls
 * 
 * Features:
 * - Automatically includes Content-Type header
 * - Uses base URL from environment/config
 * - 10 second timeout (adjust as needed)
 */
export const apiClient = axios.create({
    baseURL: API_BASE_URL,
    headers: {
        "Content-Type": "application/json",
    },
    timeout: 10000, // 10 seconds
});

// =============================================================================
// REQUEST INTERCEPTOR
// =============================================================================

/**
 * Request Interceptor - runs BEFORE every request is sent
 * 
 * Purpose: Automatically attach the access token to requests
 * 
 * Flow:
 * 1. Check if access token exists in localStorage
 * 2. If yes, add "Authorization: Bearer <token>" header
 * 3. Request continues to the server
 */
apiClient.interceptors.request.use(
    (config: InternalAxiosRequestConfig) => {
        const accessToken = TokenStorage.getAccessToken();

        if (accessToken) {
            config.headers.Authorization = `Bearer ${accessToken}`;
        }

        return config;
    },
    (error: AxiosError) => {
        return Promise.reject(error);
    }
);

// =============================================================================
// RESPONSE INTERCEPTOR
// =============================================================================

/**
 * Track if we're currently refreshing to prevent multiple refresh calls
 */
let isRefreshing = false;
let failedQueue: Array<{
    resolve: (token: string) => void;
    reject: (error: Error) => void;
}> = [];

/**
 * Process queued requests after token refresh
 */
const processQueue = (error: Error | null, token: string | null = null) => {
    failedQueue.forEach((promise) => {
        if (error) {
            promise.reject(error);
        } else if (token) {
            promise.resolve(token);
        }
    });
    failedQueue = [];
};

/**
 * Response Interceptor - runs AFTER every response is received
 * 
 * Purpose: Handle token expiration gracefully
 * 
 * Flow:
 * 1. If response is successful (2xx), return it
 * 2. If 401 Unauthorized:
 *    a. Check if this is already a retry (prevent infinite loop)
 *    b. Use refresh token to get new access token
 *    c. Retry the original request with new token
 *    d. If refresh fails, clear tokens and redirect to login
 */
apiClient.interceptors.response.use(
    // Success handler - just return the response
    (response) => response,

    // Error handler - handle 401 and token refresh
    async (error: AxiosError) => {
        const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

        // Only handle 401 errors (Unauthorized)
        if (error.response?.status !== 401 || originalRequest._retry) {
            return Promise.reject(error);
        }

        // If already refreshing, queue this request
        if (isRefreshing) {
            return new Promise((resolve, reject) => {
                failedQueue.push({ resolve, reject });
            }).then((token) => {
                originalRequest.headers.Authorization = `Bearer ${token}`;
                return apiClient(originalRequest);
            });
        }

        originalRequest._retry = true;
        isRefreshing = true;

        try {
            const refreshToken = TokenStorage.getRefreshToken();

            if (!refreshToken) {
                throw new Error("No refresh token available");
            }

            // Call refresh endpoint (note: correct endpoint is /auth/refresh)
            const response = await axios.post(`${API_BASE_URL}/api/v1/auth/refresh`, {
                refresh_token: refreshToken,
            });

            const { access_token, refresh_token: newRefreshToken } = response.data.data;

            // Store new tokens
            TokenStorage.setTokens(access_token, newRefreshToken);

            // Process queued requests with new token
            processQueue(null, access_token);

            // Retry original request
            originalRequest.headers.Authorization = `Bearer ${access_token}`;
            return apiClient(originalRequest);

        } catch (refreshError) {
            // Refresh failed - clear everything and redirect to login
            processQueue(refreshError as Error, null);
            TokenStorage.clearAll();

            if (typeof window !== "undefined") {
                window.location.href = "/";
            }

            return Promise.reject(refreshError);
        } finally {
            isRefreshing = false;
        }
    }
);

// =============================================================================
// EXPORTS
// =============================================================================

export { TokenStorage };
export default apiClient;