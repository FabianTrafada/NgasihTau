package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/pod/domain"
)

// UploadRequestRepository implements domain.UploadRequestRepository using PostgreSQL.
// Enables teacher-to-teacher collaboration for quality educational content.
// Implements requirements 4.1, 4.3, 4.6.
type UploadRequestRepository struct {
	db *pgxpool.Pool
}

// NewUploadRequestRepository creates a new UploadRequestRepository.
func NewUploadRequestRepository(db *pgxpool.Pool) *UploadRequestRepository {
	return &UploadRequestRepository{db: db}
}

// Create creates a new upload request.
// Implements requirement 4.1: WHEN a teacher submits upload request to another teacher's pod.
func (r *UploadRequestRepository) Create(ctx context.Context, request *domain.UploadRequest) error {
	query := `
		INSERT INTO upload_requests (id, requester_id, pod_id, pod_owner_id, status, message, rejection_reason, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Exec(ctx, query,
		request.ID, request.RequesterID, request.PodID, request.PodOwnerID,
		request.Status, request.Message, request.RejectionReason, request.ExpiresAt,
		request.CreatedAt, request.UpdatedAt,
	)
	if err != nil {
		return errors.Internal("failed to create upload request", err)
	}
	return nil
}

// FindByID finds an upload request by ID.
func (r *UploadRequestRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.UploadRequest, error) {
	query := `
		SELECT id, requester_id, pod_id, pod_owner_id, status, message, rejection_reason, expires_at, created_at, updated_at
		FROM upload_requests WHERE id = $1
	`
	return r.scanUploadRequest(r.db.QueryRow(ctx, query, id))
}

// FindByRequesterAndPod finds an upload request by requester ID and pod ID.
func (r *UploadRequestRepository) FindByRequesterAndPod(ctx context.Context, requesterID, podID uuid.UUID) (*domain.UploadRequest, error) {
	query := `
		SELECT id, requester_id, pod_id, pod_owner_id, status, message, rejection_reason, expires_at, created_at, updated_at
		FROM upload_requests WHERE requester_id = $1 AND pod_id = $2
	`
	return r.scanUploadRequest(r.db.QueryRow(ctx, query, requesterID, podID))
}

// FindByPodOwner finds upload requests for a pod owner with optional status filter.
// Implements requirement 4.3: Pod owner can view and manage upload requests.
func (r *UploadRequestRepository) FindByPodOwner(ctx context.Context, ownerID uuid.UUID, status *domain.UploadRequestStatus, limit, offset int) ([]*domain.UploadRequest, int, error) {
	var total int
	var rows pgx.Rows
	var err error

	if status != nil {
		// With status filter
		countQuery := `SELECT COUNT(*) FROM upload_requests WHERE pod_owner_id = $1 AND status = $2`
		if err := r.db.QueryRow(ctx, countQuery, ownerID, *status).Scan(&total); err != nil {
			return nil, 0, errors.Internal("failed to count upload requests", err)
		}

		query := `
			SELECT id, requester_id, pod_id, pod_owner_id, status, message, rejection_reason, expires_at, created_at, updated_at
			FROM upload_requests WHERE pod_owner_id = $1 AND status = $2
			ORDER BY created_at DESC LIMIT $3 OFFSET $4
		`
		rows, err = r.db.Query(ctx, query, ownerID, *status, limit, offset)
	} else {
		// Without status filter
		countQuery := `SELECT COUNT(*) FROM upload_requests WHERE pod_owner_id = $1`
		if err := r.db.QueryRow(ctx, countQuery, ownerID).Scan(&total); err != nil {
			return nil, 0, errors.Internal("failed to count upload requests", err)
		}

		query := `
			SELECT id, requester_id, pod_id, pod_owner_id, status, message, rejection_reason, expires_at, created_at, updated_at
			FROM upload_requests WHERE pod_owner_id = $1
			ORDER BY created_at DESC LIMIT $2 OFFSET $3
		`
		rows, err = r.db.Query(ctx, query, ownerID, limit, offset)
	}

	if err != nil {
		return nil, 0, errors.Internal("failed to query upload requests", err)
	}
	defer rows.Close()

	requests, err := r.scanUploadRequests(rows)
	if err != nil {
		return nil, 0, err
	}

	return requests, total, nil
}

// FindByRequester finds upload requests made by a requester.
func (r *UploadRequestRepository) FindByRequester(ctx context.Context, requesterID uuid.UUID, limit, offset int) ([]*domain.UploadRequest, int, error) {
	countQuery := `SELECT COUNT(*) FROM upload_requests WHERE requester_id = $1`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, requesterID).Scan(&total); err != nil {
		return nil, 0, errors.Internal("failed to count upload requests", err)
	}

	query := `
		SELECT id, requester_id, pod_id, pod_owner_id, status, message, rejection_reason, expires_at, created_at, updated_at
		FROM upload_requests WHERE requester_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, requesterID, limit, offset)
	if err != nil {
		return nil, 0, errors.Internal("failed to query upload requests", err)
	}
	defer rows.Close()

	requests, err := r.scanUploadRequests(rows)
	if err != nil {
		return nil, 0, err
	}

	return requests, total, nil
}

// FindApprovedByRequesterAndPod finds an approved upload request for a requester and pod.
// Implements requirement 4.5: WHILE an upload request is approved, THE Material Service SHALL allow the requesting teacher to upload.
func (r *UploadRequestRepository) FindApprovedByRequesterAndPod(ctx context.Context, requesterID, podID uuid.UUID) (*domain.UploadRequest, error) {
	query := `
		SELECT id, requester_id, pod_id, pod_owner_id, status, message, rejection_reason, expires_at, created_at, updated_at
		FROM upload_requests 
		WHERE requester_id = $1 AND pod_id = $2 AND status = $3
	`
	return r.scanUploadRequest(r.db.QueryRow(ctx, query, requesterID, podID, domain.UploadRequestStatusApproved))
}

// Update updates an upload request.
func (r *UploadRequestRepository) Update(ctx context.Context, request *domain.UploadRequest) error {
	query := `
		UPDATE upload_requests 
		SET status = $2, message = $3, rejection_reason = $4, expires_at = $5, updated_at = $6
		WHERE id = $1
	`
	result, err := r.db.Exec(ctx, query,
		request.ID, request.Status, request.Message, request.RejectionReason, request.ExpiresAt, time.Now(),
	)
	if err != nil {
		return errors.Internal("failed to update upload request", err)
	}
	if result.RowsAffected() == 0 {
		return errors.NotFound("upload request", request.ID.String())
	}
	return nil
}

// UpdateStatus updates the status of an upload request with optional rejection reason.
// Implements requirements 4.3 (approve), 4.4 (reject), 4.6 (revoke).
func (r *UploadRequestRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.UploadRequestStatus, reason *string) error {
	query := `
		UPDATE upload_requests 
		SET status = $2, rejection_reason = $3, updated_at = $4
		WHERE id = $1
	`
	result, err := r.db.Exec(ctx, query, id, status, reason, time.Now())
	if err != nil {
		return errors.Internal("failed to update upload request status", err)
	}
	if result.RowsAffected() == 0 {
		return errors.NotFound("upload request", id.String())
	}
	return nil
}

// scanUploadRequest scans a single upload request from a row.
func (r *UploadRequestRepository) scanUploadRequest(row pgx.Row) (*domain.UploadRequest, error) {
	var req domain.UploadRequest
	err := row.Scan(
		&req.ID, &req.RequesterID, &req.PodID, &req.PodOwnerID,
		&req.Status, &req.Message, &req.RejectionReason, &req.ExpiresAt,
		&req.CreatedAt, &req.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NotFound("upload request", "")
		}
		return nil, errors.Internal("failed to scan upload request", err)
	}
	return &req, nil
}

// scanUploadRequests scans multiple upload requests from rows.
func (r *UploadRequestRepository) scanUploadRequests(rows pgx.Rows) ([]*domain.UploadRequest, error) {
	var requests []*domain.UploadRequest
	for rows.Next() {
		var req domain.UploadRequest
		err := rows.Scan(
			&req.ID, &req.RequesterID, &req.PodID, &req.PodOwnerID,
			&req.Status, &req.Message, &req.RejectionReason, &req.ExpiresAt,
			&req.CreatedAt, &req.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Internal("failed to scan upload request", err)
		}
		requests = append(requests, &req)
	}
	return requests, nil
}
