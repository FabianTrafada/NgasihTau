// Package postgres provides PostgreSQL implementation of interaction tracking.
package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/pod/domain"
)

// InteractionRepository implements domain.InteractionRepository.
type InteractionRepository struct {
	db *pgxpool.Pool
}

// NewInteractionRepository creates a new InteractionRepository.
func NewInteractionRepository(db *pgxpool.Pool) *InteractionRepository {
	return &InteractionRepository{db: db}
}

// Ensure InteractionRepository implements domain.InteractionRepository.
var _ domain.InteractionRepository = (*InteractionRepository)(nil)

// Create records a new interaction.
func (r *InteractionRepository) Create(ctx context.Context, interaction *domain.PodInteraction) error {
	var metadataJSON []byte
	var err error
	if interaction.Metadata != nil {
		metadataJSON, err = json.Marshal(interaction.Metadata)
		if err != nil {
			return errors.Internal("failed to marshal metadata", err)
		}
	}

	query := `
		INSERT INTO pod_interactions (id, user_id, pod_id, interaction_type, weight, metadata, session_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = r.db.Exec(ctx, query,
		interaction.ID,
		interaction.UserID,
		interaction.PodID,
		interaction.InteractionType,
		interaction.Weight,
		metadataJSON,
		interaction.SessionID,
		interaction.CreatedAt,
	)
	if err != nil {
		return errors.Internal("failed to create interaction", err)
	}

	return nil
}

// CreateBatch records multiple interactions at once.
func (r *InteractionRepository) CreateBatch(ctx context.Context, interactions []*domain.PodInteraction) error {
	if len(interactions) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `
		INSERT INTO pod_interactions (id, user_id, pod_id, interaction_type, weight, metadata, session_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	for _, interaction := range interactions {
		var metadataJSON []byte
		if interaction.Metadata != nil {
			metadataJSON, _ = json.Marshal(interaction.Metadata)
		}

		batch.Queue(query,
			interaction.ID,
			interaction.UserID,
			interaction.PodID,
			interaction.InteractionType,
			interaction.Weight,
			metadataJSON,
			interaction.SessionID,
			interaction.CreatedAt,
		)
	}

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	for range interactions {
		if _, err := br.Exec(); err != nil {
			return errors.Internal("failed to create interaction batch", err)
		}
	}

	return nil
}

// FindByUserID finds interactions for a user.
func (r *InteractionRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.PodInteraction, error) {
	query := `
		SELECT id, user_id, pod_id, interaction_type, weight, metadata, session_id, created_at
		FROM pod_interactions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, errors.Internal("failed to find interactions", err)
	}
	defer rows.Close()

	return r.scanInteractions(rows)
}

// FindByUserAndPod finds interactions for a user on a specific pod.
func (r *InteractionRepository) FindByUserAndPod(ctx context.Context, userID, podID uuid.UUID, limit int) ([]*domain.PodInteraction, error) {
	query := `
		SELECT id, user_id, pod_id, interaction_type, weight, metadata, session_id, created_at
		FROM pod_interactions
		WHERE user_id = $1 AND pod_id = $2
		ORDER BY created_at DESC
		LIMIT $3
	`

	rows, err := r.db.Query(ctx, query, userID, podID, limit)
	if err != nil {
		return nil, errors.Internal("failed to find interactions", err)
	}
	defer rows.Close()

	return r.scanInteractions(rows)
}

// CountByUserID returns the total interaction count for a user.
func (r *InteractionRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM pod_interactions WHERE user_id = $1`

	var count int
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, errors.Internal("failed to count interactions", err)
	}

	return count, nil
}

// GetRecentInteractionTypes returns distinct recent interaction types for a user-pod pair.
func (r *InteractionRepository) GetRecentInteractionTypes(ctx context.Context, userID, podID uuid.UUID, since time.Time) ([]domain.InteractionType, error) {
	query := `
		SELECT DISTINCT interaction_type
		FROM pod_interactions
		WHERE user_id = $1 AND pod_id = $2 AND created_at >= $3
		ORDER BY interaction_type
	`

	rows, err := r.db.Query(ctx, query, userID, podID, since)
	if err != nil {
		return nil, errors.Internal("failed to get interaction types", err)
	}
	defer rows.Close()

	var types []domain.InteractionType
	for rows.Next() {
		var t domain.InteractionType
		if err := rows.Scan(&t); err != nil {
			return nil, errors.Internal("failed to scan interaction type", err)
		}
		types = append(types, t)
	}

	return types, nil
}

// GetUserInteractedPodIDs returns pod IDs the user has interacted with.
func (r *InteractionRepository) GetUserInteractedPodIDs(ctx context.Context, userID uuid.UUID, limit int) ([]uuid.UUID, error) {
	query := `
		SELECT DISTINCT pod_id
		FROM pod_interactions
		WHERE user_id = $1
		ORDER BY MAX(created_at) DESC
		LIMIT $2
	`

	// Use a subquery to get distinct pods ordered by most recent interaction
	query = `
		SELECT pod_id FROM (
			SELECT pod_id, MAX(created_at) as last_interaction
			FROM pod_interactions
			WHERE user_id = $1
			GROUP BY pod_id
			ORDER BY last_interaction DESC
			LIMIT $2
		) sub
	`

	rows, err := r.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, errors.Internal("failed to get interacted pod IDs", err)
	}
	defer rows.Close()

	var podIDs []uuid.UUID
	for rows.Next() {
		var podID uuid.UUID
		if err := rows.Scan(&podID); err != nil {
			return nil, errors.Internal("failed to scan pod ID", err)
		}
		podIDs = append(podIDs, podID)
	}

	return podIDs, nil
}

// scanInteractions scans rows into PodInteraction slices.
func (r *InteractionRepository) scanInteractions(rows pgx.Rows) ([]*domain.PodInteraction, error) {
	var interactions []*domain.PodInteraction

	for rows.Next() {
		var (
			interaction  domain.PodInteraction
			metadataJSON []byte
		)

		err := rows.Scan(
			&interaction.ID,
			&interaction.UserID,
			&interaction.PodID,
			&interaction.InteractionType,
			&interaction.Weight,
			&metadataJSON,
			&interaction.SessionID,
			&interaction.CreatedAt,
		)
		if err != nil {
			return nil, errors.Internal("failed to scan interaction", err)
		}

		if len(metadataJSON) > 0 {
			var metadata domain.InteractionMetadata
			if err := json.Unmarshal(metadataJSON, &metadata); err == nil {
				interaction.Metadata = &metadata
			}
		}

		interactions = append(interactions, &interaction)
	}

	return interactions, nil
}
