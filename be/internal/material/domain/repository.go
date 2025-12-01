package domain

import (
	"context"

	"github.com/google/uuid"
)

// MaterialRepository defines the interface for material data access.
// Implements the Repository pattern for data access abstraction (requirement 10.3).
type MaterialRepository interface {
	// Create creates a new material.
	Create(ctx context.Context, material *Material) error

	// FindByID finds a material by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*Material, error)

	// FindByPodID finds all materials in a pod.
	FindByPodID(ctx context.Context, podID uuid.UUID, limit, offset int) ([]*Material, int, error)

	// FindByUploaderID finds all materials uploaded by a user.
	FindByUploaderID(ctx context.Context, uploaderID uuid.UUID, limit, offset int) ([]*Material, int, error)

	// Update updates an existing material.
	Update(ctx context.Context, material *Material) error

	// Delete soft-deletes a material.
	Delete(ctx context.Context, id uuid.UUID) error

	// UpdateStatus updates a material's processing status.
	UpdateStatus(ctx context.Context, id uuid.UUID, status MaterialStatus) error

	// IncrementViewCount increments the view count for a material.
	IncrementViewCount(ctx context.Context, id uuid.UUID) error

	// IncrementDownloadCount increments the download count for a material.
	IncrementDownloadCount(ctx context.Context, id uuid.UUID) error

	// UpdateRatingStats updates the average rating and rating count.
	UpdateRatingStats(ctx context.Context, id uuid.UUID, avgRating float64, ratingCount int) error

	// IncrementVersion increments the current version number.
	IncrementVersion(ctx context.Context, id uuid.UUID) error
}

// MaterialVersionRepository defines the interface for material version data access.
type MaterialVersionRepository interface {
	// Create creates a new material version.
	Create(ctx context.Context, version *MaterialVersion) error

	// FindByID finds a version by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*MaterialVersion, error)

	// FindByMaterialID finds all versions for a material.
	FindByMaterialID(ctx context.Context, materialID uuid.UUID) ([]*MaterialVersion, error)

	// FindByMaterialIDAndVersion finds a specific version of a material.
	FindByMaterialIDAndVersion(ctx context.Context, materialID uuid.UUID, version int) (*MaterialVersion, error)

	// GetLatestVersion gets the latest version number for a material.
	GetLatestVersion(ctx context.Context, materialID uuid.UUID) (int, error)

	// DeleteByMaterialID deletes all versions for a material.
	DeleteByMaterialID(ctx context.Context, materialID uuid.UUID) error
}

// CommentRepository defines the interface for comment data access.
type CommentRepository interface {
	// Create creates a new comment.
	Create(ctx context.Context, comment *Comment) error

	// FindByID finds a comment by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*Comment, error)

	// FindByMaterialID finds all comments for a material.
	FindByMaterialID(ctx context.Context, materialID uuid.UUID, limit, offset int) ([]*Comment, int, error)

	// FindByMaterialIDWithUsers finds all comments for a material with user details.
	FindByMaterialIDWithUsers(ctx context.Context, materialID uuid.UUID, limit, offset int) ([]*CommentWithUser, int, error)

	// FindReplies finds all replies to a comment.
	FindReplies(ctx context.Context, parentID uuid.UUID, limit, offset int) ([]*CommentWithUser, int, error)

	// Update updates a comment.
	Update(ctx context.Context, comment *Comment) error

	// Delete soft-deletes a comment.
	Delete(ctx context.Context, id uuid.UUID) error

	// CountByMaterialID returns the comment count for a material.
	CountByMaterialID(ctx context.Context, materialID uuid.UUID) (int, error)
}

// RatingRepository defines the interface for rating data access.
type RatingRepository interface {
	// Create creates a new rating.
	Create(ctx context.Context, rating *Rating) error

	// FindByID finds a rating by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*Rating, error)

	// FindByMaterialAndUser finds a rating by material and user.
	FindByMaterialAndUser(ctx context.Context, materialID, userID uuid.UUID) (*Rating, error)

	// FindByMaterialID finds all ratings for a material.
	FindByMaterialID(ctx context.Context, materialID uuid.UUID, limit, offset int) ([]*Rating, int, error)

	// FindByMaterialIDWithUsers finds all ratings for a material with user details.
	FindByMaterialIDWithUsers(ctx context.Context, materialID uuid.UUID, limit, offset int) ([]*RatingWithUser, int, error)

	// Update updates a rating.
	Update(ctx context.Context, rating *Rating) error

	// Delete removes a rating.
	Delete(ctx context.Context, id uuid.UUID) error

	// GetSummary returns the rating summary for a material.
	GetSummary(ctx context.Context, materialID uuid.UUID) (*RatingSummary, error)

	// CalculateAverage calculates the average rating for a material.
	CalculateAverage(ctx context.Context, materialID uuid.UUID) (float64, int, error)
}

// BookmarkRepository defines the interface for bookmark data access.
type BookmarkRepository interface {
	// Create creates a new bookmark.
	Create(ctx context.Context, bookmark *Bookmark) error

	// FindByID finds a bookmark by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*Bookmark, error)

	// FindByUserAndMaterial finds a bookmark by user and material.
	FindByUserAndMaterial(ctx context.Context, userID, materialID uuid.UUID) (*Bookmark, error)

	// FindByUserID finds all bookmarks for a user.
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Bookmark, int, error)

	// FindByUserIDWithMaterials finds all bookmarks for a user with material details.
	FindByUserIDWithMaterials(ctx context.Context, userID uuid.UUID, folder *string, limit, offset int) ([]*MaterialWithUploader, int, error)

	// Delete removes a bookmark.
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteByUserAndMaterial removes a bookmark by user and material.
	DeleteByUserAndMaterial(ctx context.Context, userID, materialID uuid.UUID) error

	// Exists checks if a bookmark exists.
	Exists(ctx context.Context, userID, materialID uuid.UUID) (bool, error)

	// GetFolders returns all unique folders for a user.
	GetFolders(ctx context.Context, userID uuid.UUID) ([]string, error)
}
