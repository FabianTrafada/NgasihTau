package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/pod/domain"
)

// SharedPodRepository implements domain.SharedPodRepository using PostgreSQL.
// Enables teachers to share pods with students for guided learning.
// Implements requirement 7.2.
type SharedPodRepository struct {
	db *pgxpool.Pool
}

// NewSharedPodRepository creates a new SharedPodRepository.
func NewSharedPodRepository(db *pgxpool.Pool) *SharedPodRepository {
	return &SharedPodRepository{db: db}
}

// Create creates a new shared pod record.
// Implements requirement 7.2: THE Pod Service SHALL support a "shared with me" section.
func (r *SharedPodRepository) Create(ctx context.Context, share *domain.SharedPod) error {
	query := `INSERT INTO shared_pods (id, pod_id, teacher_id, student_id, message, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, query, share.ID, share.PodID, share.TeacherID, share.StudentID, share.Message, share.CreatedAt)
	if err != nil {
		return errors.Internal("failed to create shared pod", err)
	}
	return nil
}

// Delete removes a shared pod record.
func (r *SharedPodRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM shared_pods WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return errors.Internal("failed to delete shared pod", err)
	}
	if result.RowsAffected() == 0 {
		return errors.NotFound("shared pod", id.String())
	}
	return nil
}

// FindByStudent finds shared pods for a student with pagination.
// Implements requirement 7.2: showing pods that teachers have explicitly shared with the student.
func (r *SharedPodRepository) FindByStudent(ctx context.Context, studentID uuid.UUID, limit, offset int) ([]*domain.SharedPod, int, error) {
	countQuery := `SELECT COUNT(*) FROM shared_pods WHERE student_id = $1`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, studentID).Scan(&total); err != nil {
		return nil, 0, errors.Internal("failed to count shared pods", err)
	}

	query := `
		SELECT id, pod_id, teacher_id, student_id, message, created_at
		FROM shared_pods
		WHERE student_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, studentID, limit, offset)
	if err != nil {
		return nil, 0, errors.Internal("failed to query shared pods", err)
	}
	defer rows.Close()

	var sharedPods []*domain.SharedPod
	for rows.Next() {
		var sp domain.SharedPod
		err := rows.Scan(&sp.ID, &sp.PodID, &sp.TeacherID, &sp.StudentID, &sp.Message, &sp.CreatedAt)
		if err != nil {
			return nil, 0, errors.Internal("failed to scan shared pod", err)
		}
		sharedPods = append(sharedPods, &sp)
	}
	return sharedPods, total, nil
}

// FindByStudentWithDetails finds shared pods for a student with pod and teacher details.
// Implements requirement 7.2: showing pods with teacher info in "shared with me" section.
func (r *SharedPodRepository) FindByStudentWithDetails(ctx context.Context, studentID uuid.UUID, limit, offset int) ([]*domain.SharedPodWithDetails, int, error) {
	countQuery := `SELECT COUNT(*) FROM shared_pods WHERE student_id = $1`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, studentID).Scan(&total); err != nil {
		return nil, 0, errors.Internal("failed to count shared pods", err)
	}

	query := `
		SELECT sp.id, sp.pod_id, sp.teacher_id, sp.student_id, sp.message, sp.created_at,
			p.name AS pod_name, p.slug AS pod_slug,
			u.name AS teacher_name, u.avatar_url AS teacher_avatar
		FROM shared_pods sp
		JOIN pods p ON sp.pod_id = p.id
		JOIN users u ON sp.teacher_id = u.id
		WHERE sp.student_id = $1 AND p.deleted_at IS NULL
		ORDER BY sp.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, studentID, limit, offset)
	if err != nil {
		return nil, 0, errors.Internal("failed to query shared pods with details", err)
	}
	defer rows.Close()

	var sharedPods []*domain.SharedPodWithDetails
	for rows.Next() {
		var sp domain.SharedPodWithDetails
		err := rows.Scan(
			&sp.ID, &sp.PodID, &sp.TeacherID, &sp.StudentID, &sp.Message, &sp.CreatedAt,
			&sp.PodName, &sp.PodSlug, &sp.TeacherName, &sp.TeacherAvatar,
		)
		if err != nil {
			return nil, 0, errors.Internal("failed to scan shared pod with details", err)
		}
		sharedPods = append(sharedPods, &sp)
	}
	return sharedPods, total, nil
}

// FindByTeacherAndStudent finds shared pods between a specific teacher and student.
func (r *SharedPodRepository) FindByTeacherAndStudent(ctx context.Context, teacherID, studentID uuid.UUID) ([]*domain.SharedPod, error) {
	query := `
		SELECT id, pod_id, teacher_id, student_id, message, created_at
		FROM shared_pods
		WHERE teacher_id = $1 AND student_id = $2
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, teacherID, studentID)
	if err != nil {
		return nil, errors.Internal("failed to query shared pods by teacher and student", err)
	}
	defer rows.Close()

	var sharedPods []*domain.SharedPod
	for rows.Next() {
		var sp domain.SharedPod
		err := rows.Scan(&sp.ID, &sp.PodID, &sp.TeacherID, &sp.StudentID, &sp.Message, &sp.CreatedAt)
		if err != nil {
			return nil, errors.Internal("failed to scan shared pod", err)
		}
		sharedPods = append(sharedPods, &sp)
	}
	return sharedPods, nil
}

// Exists checks if a pod is already shared with a student.
// Used to prevent duplicate shares.
func (r *SharedPodRepository) Exists(ctx context.Context, podID, studentID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM shared_pods WHERE pod_id = $1 AND student_id = $2)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, podID, studentID).Scan(&exists); err != nil {
		return false, errors.Internal("failed to check shared pod existence", err)
	}
	return exists, nil
}
