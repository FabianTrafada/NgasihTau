// =============================================================================
// COLLABORATOR PERMISSION UTILS
// =============================================================================
// Utility functions for checking collaborator permissions
// Can be used across different components to maintain consistency

import { Collaborator, CollaboratorRole, CollaboratorPermission } from "@/types/collaborator";

/**
 * Check if a user is the owner of a pod
 */
export function isPodOwner(ownerId: string, userId: string): boolean {
    return ownerId === userId;
}

/**
 * Find a user's collaborator entry in a pod
 */
export function findCollaborator(
    collaborators: Collaborator[],
    userId: string
): Collaborator | undefined {
    return collaborators.find(
        (c) => c.user_id === userId && c.status === "verified"
    );
}

/**
 * Check if a user is a verified collaborator
 */
export function isVerifiedCollaborator(
    collaborators: Collaborator[],
    userId: string
): boolean {
    return !!findCollaborator(collaborators, userId);
}

/**
 * Get a user's role in a pod
 */
export function getUserRole(
    ownerId: string,
    collaborators: Collaborator[],
    userId: string
): "owner" | CollaboratorRole | null {
    if (isPodOwner(ownerId, userId)) {
        return "owner";
    }

    const collaborator = findCollaborator(collaborators, userId);
    return collaborator?.role || null;
}

/**
 * Get full permission object for a user
 */
export function getCollaboratorPermissions(
    ownerId: string,
    collaborators: Collaborator[],
    userId: string
): CollaboratorPermission {
    const isOwner = isPodOwner(ownerId, userId);
    const collaborator = findCollaborator(collaborators, userId);
    const role = collaborator?.role;

    return {
        isOwner,
        isCollaborator: !!collaborator,
        isAdmin: role === "admin",
        isContributor: role === "contributor",
        isViewer: role === "viewer",
        canUpload: isOwner || role === "admin" || role === "contributor",
        canManageCollaborators: isOwner || role === "admin",
        canEdit: isOwner || role === "admin" || role === "contributor",
        canDelete: isOwner || role === "admin",
    };
}

/**
 * Role hierarchy for comparison
 */
const ROLE_HIERARCHY: Record<string, number> = {
    owner: 4,
    admin: 3,
    contributor: 2,
    viewer: 1,
};

/**
 * Check if a user has at least a certain role level
 */
export function hasMinimumRole(
    ownerId: string,
    collaborators: Collaborator[],
    userId: string,
    minimumRole: "owner" | CollaboratorRole
): boolean {
    const userRole = getUserRole(ownerId, collaborators, userId);
    if (!userRole) return false;

    return ROLE_HIERARCHY[userRole] >= ROLE_HIERARCHY[minimumRole];
}

/**
 * Validation: Check if a user can invite collaborators
 */
export function canInviteCollaborator(
    ownerId: string,
    collaborators: Collaborator[],
    inviterId: string,
    targetUserId: string
): { allowed: boolean; reason?: string } {
    // Cannot invite self
    if (inviterId === targetUserId) {
        return { allowed: false, reason: "cannotInviteSelf" };
    }

    // Check if inviter has permission
    const permissions = getCollaboratorPermissions(ownerId, collaborators, inviterId);
    if (!permissions.canManageCollaborators) {
        return { allowed: false, reason: "permissionDenied" };
    }

    // Check if target is already a collaborator
    if (isVerifiedCollaborator(collaborators, targetUserId)) {
        return { allowed: false, reason: "alreadyCollaborator" };
    }

    // Check if target is the owner
    if (isPodOwner(ownerId, targetUserId)) {
        return { allowed: false, reason: "cannotInviteOwner" };
    }

    return { allowed: true };
}

/**
 * Validation: Check if a user can remove a collaborator
 */
export function canRemoveCollaborator(
    ownerId: string,
    collaborators: Collaborator[],
    removerId: string,
    targetUserId: string
): { allowed: boolean; reason?: string } {
    // Cannot remove the owner
    if (isPodOwner(ownerId, targetUserId)) {
        return { allowed: false, reason: "cannotRemoveOwner" };
    }

    // Check if remover has permission
    const permissions = getCollaboratorPermissions(ownerId, collaborators, removerId);
    if (!permissions.canManageCollaborators) {
        return { allowed: false, reason: "permissionDenied" };
    }

    // Admins cannot remove other admins (only owner can)
    if (!permissions.isOwner) {
        const targetRole = getUserRole(ownerId, collaborators, targetUserId);
        if (targetRole === "admin") {
            return { allowed: false, reason: "cannotRemoveAdmin" };
        }
    }

    return { allowed: true };
}
