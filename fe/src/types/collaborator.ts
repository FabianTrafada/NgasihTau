// =============================================================================
// COLLABORATOR TYPES
// =============================================================================
// Type definitions for the Collaborative Pod feature
// Based on Swagger API documentation

/**
 * Collaborator roles in a pod
 * - viewer: Can view materials only
 * - contributor: Can view and upload materials
 * - admin: Full access including managing collaborators
 */
export type CollaboratorRole = "viewer" | "contributor" | "admin";

/**
 * Collaborator status
 * - pending: Invitation sent, waiting for acceptance
 * - pending_verification: User needs verification
 * - verified: Fully active collaborator
 */
export type CollaboratorStatus = "pending" | "pending_verification" | "verified";

/**
 * Collaborator entity from the API
 */
export interface Collaborator {
    id: string;
    pod_id: string;
    user_id: string;
    role: CollaboratorRole;
    status: CollaboratorStatus;
    invited_by: string;
    created_at: string;
    updated_at: string;
    // Extended fields (populated by API or frontend)
    user_name?: string;
    user_email?: string;
    user_avatar_url?: string;
}

/**
 * Input for inviting a collaborator
 * Based on application.InviteCollaboratorInput from Swagger
 * Uses email instead of user_id - backend resolves email to user internally
 */
export interface InviteCollaboratorInput {
    email: string;
    role: CollaboratorRole;
}

/**
 * Input for updating a collaborator
 */
export interface UpdateCollaboratorInput {
    action: "verify" | "update_role";
    role?: CollaboratorRole;
}

/**
 * API Response wrapper
 */
export interface CollaboratorResponse {
    success: boolean;
    data: Collaborator;
    meta?: {
        request_id: string;
        timestamp: string;
    };
}

/**
 * API Response wrapper for list of collaborators
 */
export interface CollaboratorsListResponse {
    success: boolean;
    data: Collaborator[];
    meta?: {
        request_id: string;
        timestamp: string;
    };
}

/**
 * User search result for collaborator invitation
 */
export interface UserSearchResult {
    id: string;
    name: string;
    username: string;
    email: string;
    avatar_url?: string;
    role: string;
    email_verified: boolean;
}

/**
 * Collaborator permission check result
 */
export interface CollaboratorPermission {
    isOwner: boolean;
    isCollaborator: boolean;
    isAdmin: boolean;
    isContributor: boolean;
    isViewer: boolean;
    canUpload: boolean;
    canManageCollaborators: boolean;
    canEdit: boolean;
    canDelete: boolean;
}

/**
 * Error codes for collaborator operations
 */
export enum CollaboratorErrorCode {
    USER_NOT_FOUND = "USER_NOT_FOUND",
    USER_NOT_VERIFIED = "USER_NOT_VERIFIED",
    ALREADY_COLLABORATOR = "ALREADY_COLLABORATOR",
    CANNOT_INVITE_SELF = "CANNOT_INVITE_SELF",
    PERMISSION_DENIED = "PERMISSION_DENIED",
    COLLABORATOR_NOT_FOUND = "COLLABORATOR_NOT_FOUND",
    INVALID_ROLE = "INVALID_ROLE",
    NETWORK_ERROR = "NETWORK_ERROR",
    UNKNOWN_ERROR = "UNKNOWN_ERROR",
}

/**
 * Structured error for collaborator operations
 */
export interface CollaboratorError {
    code: CollaboratorErrorCode;
    message: string;
    field?: string;
}
