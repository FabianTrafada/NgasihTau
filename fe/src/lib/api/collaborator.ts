// =============================================================================
// COLLABORATOR API SERVICE
// =============================================================================
// API service for managing pod collaborators
// Based on Swagger API documentation endpoints:
// - GET    /pods/{id}/collaborators
// - POST   /pods/{id}/collaborators
// - PUT    /pods/{id}/collaborators/{userId}
// - DELETE /pods/{id}/collaborators/{userId}

import apiClient from "@/lib/api-client";
import {
    Collaborator,
    CollaboratorsListResponse,
    CollaboratorResponse,
    InviteCollaboratorInput,
    UpdateCollaboratorInput,
    CollaboratorError,
    CollaboratorErrorCode,
} from "@/types/collaborator";
import { AxiosError } from "axios";

// =============================================================================
// ERROR HANDLING
// =============================================================================

/**
 * Parse API error response into structured CollaboratorError
 */
function parseApiError(error: unknown): CollaboratorError {
    if (error instanceof AxiosError) {
        const status = error.response?.status;
        const errorData = error.response?.data?.error;
        const message = errorData?.message || error.message;

        switch (status) {
            case 400:
                if (message.toLowerCase().includes("already")) {
                    return {
                        code: CollaboratorErrorCode.ALREADY_COLLABORATOR,
                        message,
                    };
                }
                if (message.toLowerCase().includes("self")) {
                    return {
                        code: CollaboratorErrorCode.CANNOT_INVITE_SELF,
                        message,
                    };
                }
                if (message.toLowerCase().includes("role")) {
                    return {
                        code: CollaboratorErrorCode.INVALID_ROLE,
                        message,
                    };
                }
                return {
                    code: CollaboratorErrorCode.UNKNOWN_ERROR,
                    message,
                };
            case 401:
                return {
                    code: CollaboratorErrorCode.PERMISSION_DENIED,
                    message: "Authentication required",
                };
            case 403:
                return {
                    code: CollaboratorErrorCode.PERMISSION_DENIED,
                    message: "You don't have permission to perform this action",
                };
            case 404:
                if (message.toLowerCase().includes("collaborator")) {
                    return {
                        code: CollaboratorErrorCode.COLLABORATOR_NOT_FOUND,
                        message,
                    };
                }
                return {
                    code: CollaboratorErrorCode.USER_NOT_FOUND,
                    message: "User or pod not found",
                };
            case 409:
                return {
                    code: CollaboratorErrorCode.ALREADY_COLLABORATOR,
                    message: "User is already a collaborator",
                };
            default:
                if (!navigator.onLine || error.code === "ERR_NETWORK") {
                    return {
                        code: CollaboratorErrorCode.NETWORK_ERROR,
                        message: "Network error. Please check your connection.",
                    };
                }
                return {
                    code: CollaboratorErrorCode.UNKNOWN_ERROR,
                    message: message || "An unexpected error occurred",
                };
        }
    }

    return {
        code: CollaboratorErrorCode.UNKNOWN_ERROR,
        message: error instanceof Error ? error.message : "An unexpected error occurred",
    };
}

// =============================================================================
// API FUNCTIONS
// =============================================================================

/**
 * Get all collaborators for a pod
 * Endpoint: GET /api/v1/pods/{id}/collaborators
 * 
 * @param podId - The pod ID
 * @returns List of collaborators
 * @throws CollaboratorError
 */
export async function getPodCollaborators(podId: string): Promise<Collaborator[]> {
    try {
        const response = await apiClient.get<CollaboratorsListResponse>(
            `/api/v1/pods/${podId}/collaborators`
        );
        return response.data.data || [];
    } catch (error) {
        console.error("[Collaborator API] Error fetching collaborators:", error);
        throw parseApiError(error);
    }
}

/**
 * Invite a user to collaborate on a pod
 * Endpoint: POST /api/v1/pods/{id}/collaborators
 * 
 * @param podId - The pod ID
 * @param input - Invitation details (user_id and role)
 * @returns The created collaborator
 * @throws CollaboratorError
 */
export async function inviteCollaborator(
    podId: string,
    input: InviteCollaboratorInput
): Promise<Collaborator> {
    try {
        const response = await apiClient.post<CollaboratorResponse>(
            `/api/v1/pods/${podId}/collaborators`,
            input
        );
        return response.data.data;
    } catch (error) {
        console.error("[Collaborator API] Error inviting collaborator:", error);
        throw parseApiError(error);
    }
}

/**
 * Update a collaborator (verify or change role)
 * Endpoint: PUT /api/v1/pods/{id}/collaborators/{userId}
 * 
 * @param podId - The pod ID
 * @param userId - The collaborator's user ID
 * @param input - Update details (action and optional role)
 * @returns Success status
 * @throws CollaboratorError
 */
export async function updateCollaborator(
    podId: string,
    userId: string,
    input: UpdateCollaboratorInput
): Promise<boolean> {
    try {
        const response = await apiClient.put<{ success: boolean }>(
            `/api/v1/pods/${podId}/collaborators/${userId}`,
            input
        );
        return response.data.success;
    } catch (error) {
        console.error("[Collaborator API] Error updating collaborator:", error);
        throw parseApiError(error);
    }
}

/**
 * Remove a collaborator from a pod
 * Endpoint: DELETE /api/v1/pods/{id}/collaborators/{userId}
 * 
 * @param podId - The pod ID
 * @param userId - The collaborator's user ID to remove
 * @throws CollaboratorError
 */
export async function removeCollaborator(
    podId: string,
    userId: string
): Promise<void> {
    try {
        await apiClient.delete(`/api/v1/pods/${podId}/collaborators/${userId}`);
    } catch (error) {
        console.error("[Collaborator API] Error removing collaborator:", error);
        throw parseApiError(error);
    }
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

/**
 * Check if a user is a collaborator of a pod
 * 
 * @param collaborators - List of collaborators
 * @param userId - User ID to check
 * @returns true if user is a collaborator
 */
export function isUserCollaborator(
    collaborators: Collaborator[],
    userId: string
): boolean {
    return collaborators.some(
        (c) => c.user_id === userId && c.status === "verified"
    );
}

/**
 * Check if a user is a collaborator by email
 * 
 * @param collaborators - List of collaborators with extended user info
 * @param email - Email to check
 * @returns true if user is a collaborator
 */
export function isEmailCollaborator(
    collaborators: Collaborator[],
    email: string
): boolean {
    return collaborators.some(
        (c) => c.user_email?.toLowerCase() === email.toLowerCase() && c.status === "verified"
    );
}

/**
 * Get collaborator role for a user
 * 
 * @param collaborators - List of collaborators
 * @param userId - User ID to check
 * @returns CollaboratorRole or null if not a collaborator
 */
export function getCollaboratorRole(
    collaborators: Collaborator[],
    userId: string
): string | null {
    const collaborator = collaborators.find(
        (c) => c.user_id === userId && c.status === "verified"
    );
    return collaborator?.role || null;
}

/**
 * Check if user can manage collaborators (owner or admin)
 * 
 * @param ownerId - Pod owner ID
 * @param collaborators - List of collaborators
 * @param userId - Current user ID
 * @returns true if user can manage collaborators
 */
export function canManageCollaborators(
    ownerId: string,
    collaborators: Collaborator[],
    userId: string
): boolean {
    if (ownerId === userId) return true;
    const role = getCollaboratorRole(collaborators, userId);
    return role === "admin";
}

/**
 * Check if user can upload materials
 * 
 * @param ownerId - Pod owner ID
 * @param collaborators - List of collaborators
 * @param userId - Current user ID
 * @returns true if user can upload
 */
export function canUploadMaterials(
    ownerId: string,
    collaborators: Collaborator[],
    userId: string
): boolean {
    if (ownerId === userId) return true;
    const role = getCollaboratorRole(collaborators, userId);
    return role === "admin" || role === "contributor";
}
