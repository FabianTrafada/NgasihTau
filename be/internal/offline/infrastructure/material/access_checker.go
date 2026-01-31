// Package material provides integration with the material-service for the offline module.
// This package implements the LicenseMaterialAccessChecker interface to verify
// user access to materials before issuing offline licenses.
package material

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// AccessChecker implements the LicenseMaterialAccessChecker interface.
// It verifies user access to materials by checking:
// 1. Material exists and is not deleted
// 2. User has access to the pod containing the material
//
// Implements Requirement 3.1: Verify user has access to material via material-service.
// Implements Requirement 6.2: Integrate with material-service to verify material access permissions.
type AccessChecker struct {
	materialDB *pgxpool.Pool
	podDB      *pgxpool.Pool
}

// NewAccessChecker creates a new AccessChecker instance.
// materialDB: Connection pool to the material database (materials table)
// podDB: Connection pool to the pod database (pods, collaborators tables)
func NewAccessChecker(materialDB, podDB *pgxpool.Pool) *AccessChecker {
	return &AccessChecker{
		materialDB: materialDB,
		podDB:      podDB,
	}
}

// CheckAccess verifies if a user has access to a material.
// Returns true if the user can access the material, false otherwise.
//
// Access is granted if:
// - The material exists and is not deleted
// - The pod is public, OR
// - The user is the pod owner, OR
// - The user is a collaborator on the pod
//
// Implements Property 11: License Access Control.
func (c *AccessChecker) CheckAccess(ctx context.Context, userID, materialID uuid.UUID) (bool, error) {
	// Step 1: Get the material's pod_id
	podID, err := c.getMaterialPodID(ctx, materialID)
	if err != nil {
		log.Debug().
			Err(err).
			Str("material_id", materialID.String()).
			Msg("material not found or deleted")
		return false, nil // Material doesn't exist or is deleted
	}

	// Step 2: Check if user can access the pod
	canAccess, err := c.canUserAccessPod(ctx, podID, userID)
	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", userID.String()).
			Str("pod_id", podID.String()).
			Msg("failed to check pod access")
		return false, fmt.Errorf("failed to check pod access: %w", err)
	}

	log.Debug().
		Str("user_id", userID.String()).
		Str("material_id", materialID.String()).
		Str("pod_id", podID.String()).
		Bool("can_access", canAccess).
		Msg("material access check completed")

	return canAccess, nil
}


// getMaterialPodID retrieves the pod_id for a material.
// Returns an error if the material doesn't exist or is deleted.
func (c *AccessChecker) getMaterialPodID(ctx context.Context, materialID uuid.UUID) (uuid.UUID, error) {
	query := `
		SELECT pod_id 
		FROM materials 
		WHERE id = $1 AND deleted_at IS NULL
	`

	var podID uuid.UUID
	err := c.materialDB.QueryRow(ctx, query, materialID).Scan(&podID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("material not found: %w", err)
	}

	return podID, nil
}

// canUserAccessPod checks if a user can access a pod.
// Access is granted if:
// - The pod is public (visibility = 'public')
// - The user is the pod owner
// - The user is a collaborator on the pod
func (c *AccessChecker) canUserAccessPod(ctx context.Context, podID, userID uuid.UUID) (bool, error) {
	// Check if pod is public or user is owner
	podQuery := `
		SELECT visibility, owner_id 
		FROM pods 
		WHERE id = $1 AND deleted_at IS NULL
	`

	var visibility string
	var ownerID uuid.UUID
	err := c.podDB.QueryRow(ctx, podQuery, podID).Scan(&visibility, &ownerID)
	if err != nil {
		return false, fmt.Errorf("pod not found: %w", err)
	}

	// Public pods are accessible to everyone
	if visibility == "public" {
		return true, nil
	}

	// Owner can always access
	if ownerID == userID {
		return true, nil
	}

	// Check if user is a collaborator
	collabQuery := `
		SELECT EXISTS(
			SELECT 1 FROM collaborators 
			WHERE pod_id = $1 AND user_id = $2 AND status = 'accepted'
		)
	`

	var isCollaborator bool
	err = c.podDB.QueryRow(ctx, collabQuery, podID, userID).Scan(&isCollaborator)
	if err != nil {
		return false, fmt.Errorf("failed to check collaborator status: %w", err)
	}

	return isCollaborator, nil
}

// GetMaterialInfo retrieves basic material information for offline access.
// This is a convenience method that can be used by other offline services.
type MaterialInfo struct {
	ID       uuid.UUID
	PodID    uuid.UUID
	Title    string
	FileType string
	FileURL  string
	FileSize int64
}

// GetMaterialInfo retrieves material information by ID.
func (c *AccessChecker) GetMaterialInfo(ctx context.Context, materialID uuid.UUID) (*MaterialInfo, error) {
	query := `
		SELECT id, pod_id, title, file_type, file_url, file_size
		FROM materials 
		WHERE id = $1 AND deleted_at IS NULL
	`

	var info MaterialInfo
	err := c.materialDB.QueryRow(ctx, query, materialID).Scan(
		&info.ID,
		&info.PodID,
		&info.Title,
		&info.FileType,
		&info.FileURL,
		&info.FileSize,
	)
	if err != nil {
		return nil, fmt.Errorf("material not found: %w", err)
	}

	return &info, nil
}
