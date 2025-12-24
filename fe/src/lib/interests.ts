import apiClient from "./api-client";

export interface Interest {
    id: string;
    name: string;
    slug?: string;
    description?: string;
    icon?: string;
    category: string;
    display_order?: number;
    is_active?: boolean;
    created_at?: string;
    updated_at?: string;
}

export interface InterestCategory {
    name: string;
    interests: Interest[];
}

export interface UserInterest {
    interest_id: string;
    interest_name: string;
    category?: string;
}

// Response wrapper from backend
interface PredefinedInterestsResponse {
    interests: Interest[];
    total: number;
}

// =============================================================
// API FUNCTIONS 
// =============================================================

/**
 * Get all predefined interests
 * Endpoint: GET /api/v1/interests/predefined
 */
export const getAllInterests = async (): Promise<Record<string, Interest[]>> => {
    const response = await apiClient.get<{ data: { categories: { category: string; interests: Interest[] }[] } }>('/api/v1/interests/predefined/categories');

    // Transform the array response to the Record format expected by the component
    const interestsRecord: Record<string, Interest[]> = {};
    if (response.data?.data?.categories) {
        response.data.data.categories.forEach((cat) => {
            interestsRecord[cat.category] = cat.interests;
        });
    }
    return interestsRecord;
};

/**
 * Get all interest categories
 * Endpoint: GET /api/v1/interests/predefined/categories
 */
export const getInterestCategories = async (): Promise<InterestCategory[]> => {
    const response = await apiClient.get('/api/v1/interests/predefined/categories');
    return response.data.data;
};

/**
 * Get current user's interests
 * Endpoint: GET /api/v1/interests/me
 */
export const getUserInterests = async (): Promise<UserInterest[]> => {
    const response = await apiClient.get('/api/v1/interests/me');
    return response.data.data;
};

/**
 * Set user interests in batch (replaces all existing interests)
 * Endpoint: PUT /api/v1/interests/me
 * @param predefinedInterestIds - Array of predefined interest UUIDs
 * @param customInterests - Array of custom interest names (strings)
 */
export const setUserInterests = async (
    predefinedInterestIds: string[],
    customInterests: string[] = []
): Promise<void> => {
    await apiClient.put('/api/v1/interests/me', {
        predefined_interest_ids: predefinedInterestIds,
        custom_interests: customInterests
    });
};

/**
 * Add a single interest to user
 * Endpoint: POST /api/v1/interests/me
 */
export const addUserInterest = async (interestId: string): Promise<void> => {
    await apiClient.post('/api/v1/interests/me', {
        predefined_interest_id: interestId
    });
};

/**
 * Remove a single interest from user
 * Endpoint: DELETE /api/v1/interests/me/:id
 */
export const removeUserInterest = async (interestId: string): Promise<void> => {
    await apiClient.delete(`/api/v1/interests/me/${interestId}`);
};

/**
 * Complete onboarding process
 * Endpoint: POST /api/v1/interests/onboarding/complete
 */
export const completeOnboarding = async (): Promise<void> => {
    await apiClient.post("/api/v1/interests/onboarding/complete");
};

/**
 * Check onboarding status
 * Endpoint: GET /api/v1/interests/onboarding/status
 */
export const getOnboardingStatus = async (): Promise<{ completed: boolean }> => {
    const response = await apiClient.get("/api/v1/interests/onboarding/status");
    return { completed: response.data.data.onboarding_completed };
};

