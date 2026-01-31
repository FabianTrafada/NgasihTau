// =============================================================================
// USE COLLABORATORS HOOK
// =============================================================================
// Custom hook for managing pod collaborators
// Provides state management, API calls, and permission logic

"use client";

import { useState, useCallback, useEffect, useRef } from "react";
import {
    Collaborator,
    CollaboratorRole,
    CollaboratorPermission,
    CollaboratorError,
    CollaboratorErrorCode,
} from "@/types/collaborator";
import {
    getPodCollaborators,
    inviteCollaborator,
    updateCollaborator,
    removeCollaborator,
    isUserCollaborator,
    getCollaboratorRole,
} from "@/lib/api/collaborator";
import { getUserDetail } from "@/lib/api/user";

// =============================================================================
// TYPES
// =============================================================================

interface UseCollaboratorsState {
    collaborators: Collaborator[];
    loading: boolean;
    error: CollaboratorError | null;
    inviting: boolean;
    removing: string | null; // userId being removed
    updating: string | null; // userId being updated
}

interface UseCollaboratorsReturn extends UseCollaboratorsState {
    // Actions
    fetchCollaborators: () => Promise<void>;
    invite: (email: string, role: CollaboratorRole) => Promise<boolean>;
    remove: (userId: string) => Promise<boolean>;
    update: (userId: string, action: "verify" | "update_role", role?: CollaboratorRole) => Promise<boolean>;
    clearError: () => void;

    // Permission helpers
    getPermissions: (currentUserId: string, ownerId: string, currentUserEmail?: string) => CollaboratorPermission;
    isCollaborator: (userId: string) => boolean;

    // Computed values
    verifiedCollaborators: Collaborator[];
    pendingCollaborators: Collaborator[];
}

// =============================================================================
// HOOK IMPLEMENTATION
// =============================================================================

export function useCollaborators(podId: string): UseCollaboratorsReturn {
    // State
    const [state, setState] = useState<UseCollaboratorsState>({
        collaborators: [],
        loading: false,
        error: null,
        inviting: false,
        removing: null,
        updating: null,
    });

    // Prevent duplicate fetches
    const fetchingRef = useRef(false);
    const invitingRef = useRef(false);

    // ==========================================================================
    // FETCH COLLABORATORS
    // ==========================================================================

    const fetchCollaborators = useCallback(async () => {
        if (!podId || fetchingRef.current) return;

        fetchingRef.current = true;
        setState((prev) => ({ ...prev, loading: true, error: null }));

        try {
            // Backend now returns collaborators with user details (user_name, user_email, user_avatar_url)
            const collaborators = await getPodCollaborators(podId);

            setState((prev) => ({
                ...prev,
                collaborators,
                loading: false,
            }));
        } catch (error) {
            setState((prev) => ({
                ...prev,
                loading: false,
                error: error as CollaboratorError,
            }));
        } finally {
            fetchingRef.current = false;
        }
    }, [podId]);

    // ==========================================================================
    // INVITE COLLABORATOR
    // ==========================================================================

    const invite = useCallback(
        async (email: string, role: CollaboratorRole): Promise<boolean> => {
            if (!podId || invitingRef.current) return false;

            invitingRef.current = true;
            setState((prev) => ({ ...prev, inviting: true, error: null }));

            try {
                // Backend resolves email to user_id internally and returns enriched collaborator
                const newCollaborator = await inviteCollaborator(podId, {
                    email,
                    role,
                });

                setState((prev) => ({
                    ...prev,
                    collaborators: [...prev.collaborators, newCollaborator],
                    inviting: false,
                }));

                return true;
            } catch (error) {
                setState((prev) => ({
                    ...prev,
                    inviting: false,
                    error: error as CollaboratorError,
                }));
                return false;
            } finally {
                invitingRef.current = false;
            }
        },
        [podId]
    );

    // ==========================================================================
    // REMOVE COLLABORATOR
    // ==========================================================================

    const remove = useCallback(
        async (userId: string): Promise<boolean> => {
            if (!podId) return false;

            setState((prev) => ({ ...prev, removing: userId, error: null }));

            try {
                await removeCollaborator(podId, userId);

                setState((prev) => ({
                    ...prev,
                    collaborators: prev.collaborators.filter((c) => c.user_id !== userId),
                    removing: null,
                }));

                return true;
            } catch (error) {
                setState((prev) => ({
                    ...prev,
                    removing: null,
                    error: error as CollaboratorError,
                }));
                return false;
            }
        },
        [podId]
    );

    // ==========================================================================
    // UPDATE COLLABORATOR
    // ==========================================================================

    const update = useCallback(
        async (
            userId: string,
            action: "verify" | "update_role",
            role?: CollaboratorRole
        ): Promise<boolean> => {
            if (!podId) return false;

            setState((prev) => ({ ...prev, updating: userId, error: null }));

            try {
                await updateCollaborator(podId, userId, { action, role });

                setState((prev) => ({
                    ...prev,
                    collaborators: prev.collaborators.map((c) =>
                        c.user_id === userId
                            ? {
                                ...c,
                                status: action === "verify" ? "verified" : c.status,
                                role: role || c.role,
                            }
                            : c
                    ),
                    updating: null,
                }));

                return true;
            } catch (error) {
                setState((prev) => ({
                    ...prev,
                    updating: null,
                    error: error as CollaboratorError,
                }));
                return false;
            }
        },
        [podId]
    );

    // ==========================================================================
    // CLEAR ERROR
    // ==========================================================================

    const clearError = useCallback(() => {
        setState((prev) => ({ ...prev, error: null }));
    }, []);

    // ==========================================================================
    // PERMISSION HELPERS
    // ==========================================================================

    const getPermissions = useCallback(
        (currentUserId: string, ownerId: string, currentUserEmail?: string): CollaboratorPermission => {
            const isOwner = currentUserId === ownerId;

            // Check collaborator status by user_id OR email (email is more reliable for cross-service validation)
            let collaborator = state.collaborators.find(
                (c) => c.user_id === currentUserId && c.status === "verified"
            );

            // Fallback: Check by email if provided (handles cross-database consistency)
            if (!collaborator && currentUserEmail) {
                collaborator = state.collaborators.find(
                    (c) => c.user_email?.toLowerCase() === currentUserEmail.toLowerCase() && c.status === "verified"
                );
            }

            const isCollaboratorUser = !!collaborator;
            const role = collaborator?.role;

            return {
                isOwner,
                isCollaborator: isCollaboratorUser,
                isAdmin: role === "admin",
                isContributor: role === "contributor",
                isViewer: role === "viewer",
                canUpload: isOwner || role === "admin" || role === "contributor",
                canManageCollaborators: isOwner || role === "admin",
                canEdit: isOwner || role === "admin" || role === "contributor",
                canDelete: isOwner || role === "admin",
            };
        },
        [state.collaborators]
    );

    const isCollaboratorCheck = useCallback(
        (userId: string): boolean => {
            return isUserCollaborator(state.collaborators, userId);
        },
        [state.collaborators]
    );

    // ==========================================================================
    // COMPUTED VALUES
    // ==========================================================================

    const verifiedCollaborators = state.collaborators.filter(
        (c) => c.status === "verified"
    );

    const pendingCollaborators = state.collaborators.filter(
        (c) => c.status === "pending" || c.status === "pending_verification"
    );

    // ==========================================================================
    // RETURN
    // ==========================================================================

    return {
        ...state,
        fetchCollaborators,
        invite,
        remove,
        update,
        clearError,
        getPermissions,
        isCollaborator: isCollaboratorCheck,
        verifiedCollaborators,
        pendingCollaborators,
    };
}

// =============================================================================
// ERROR MESSAGE HELPER
// =============================================================================

/**
 * Get user-friendly error message for collaborator errors
 * To be used with i18n translations
 */
export function getCollaboratorErrorKey(error: CollaboratorError): string {
    switch (error.code) {
        case CollaboratorErrorCode.USER_NOT_FOUND:
            return "collaborator.errors.userNotFound";
        case CollaboratorErrorCode.USER_NOT_VERIFIED:
            return "collaborator.errors.userNotVerified";
        case CollaboratorErrorCode.ALREADY_COLLABORATOR:
            return "collaborator.errors.alreadyCollaborator";
        case CollaboratorErrorCode.CANNOT_INVITE_SELF:
            return "collaborator.errors.cannotInviteSelf";
        case CollaboratorErrorCode.PERMISSION_DENIED:
            return "collaborator.errors.permissionDenied";
        case CollaboratorErrorCode.COLLABORATOR_NOT_FOUND:
            return "collaborator.errors.collaboratorNotFound";
        case CollaboratorErrorCode.INVALID_ROLE:
            return "collaborator.errors.invalidRole";
        case CollaboratorErrorCode.NETWORK_ERROR:
            return "collaborator.errors.networkError";
        default:
            return "collaborator.errors.unknown";
    }
}
