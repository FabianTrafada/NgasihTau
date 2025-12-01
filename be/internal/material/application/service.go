// Package application contains the business logic and use cases for the Material Service.
// This layer orchestrates domain entities and repository operations.
package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"ngasihtau/internal/material/domain"
	"ngasihtau/pkg/nats"
)

// Service provides material-related business operations.
type Service struct {
	materialRepo   domain.MaterialRepository
	versionRepo    domain.MaterialVersionRepository
	commentRepo    domain.CommentRepository
	ratingRepo     domain.RatingRepository
	bookmarkRepo   domain.BookmarkRepository
	minioClient    MinIOClient
	eventPublisher nats.EventPublisher
}

// MinIOClient defines the interface for MinIO operations.
type MinIOClient interface {
	// GeneratePresignedPutURL generates a presigned URL for uploading a file.
	GeneratePresignedPutURL(ctx context.Context, objectKey string, contentType string, expiry time.Duration) (string, error)

	// GeneratePresignedGetURL generates a presigned URL for downloading a file.
	GeneratePresignedGetURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error)

	// FileExists checks if a file exists in the bucket.
	FileExists(ctx context.Context, objectKey string) (bool, error)

	// GetFileInfo returns file information (size, content type).
	GetFileInfo(ctx context.Context, objectKey string) (*FileInfo, error)

	// DeleteFile deletes a file from the bucket.
	DeleteFile(ctx context.Context, objectKey string) error
}

// FileInfo represents file information from MinIO.
type FileInfo struct {
	Size        int64
	ContentType string
	ETag        string
}

// NewService creates a new Material Service.
func NewService(
	materialRepo domain.MaterialRepository,
	versionRepo domain.MaterialVersionRepository,
	commentRepo domain.CommentRepository,
	ratingRepo domain.RatingRepository,
	bookmarkRepo domain.BookmarkRepository,
	minioClient MinIOClient,
	eventPublisher nats.EventPublisher,
) *Service {
	return &Service{
		materialRepo:   materialRepo,
		versionRepo:    versionRepo,
		commentRepo:    commentRepo,
		ratingRepo:     ratingRepo,
		bookmarkRepo:   bookmarkRepo,
		minioClient:    minioClient,
		eventPublisher: eventPublisher,
	}
}

// UploadURLInput represents input for generating an upload URL.
type UploadURLInput struct {
	Filename    string
	ContentType string
	Size        int64
}

// UploadURLOutput represents the output of generating an upload URL.
type UploadURLOutput struct {
	UploadURL string    `json:"upload_url"`
	ObjectKey string    `json:"object_key"`
	ExpiresAt time.Time `json:"expires_at"`
}

// GetUploadURL generates a presigned URL for uploading a file.
// Implements requirement 5: Material Upload.
func (s *Service) GetUploadURL(ctx context.Context, input UploadURLInput) (*UploadURLOutput, error) {
	// Generate unique object key
	objectKey := fmt.Sprintf("materials/%s/%s_%s",
		time.Now().Format("2006/01/02"),
		uuid.New().String(),
		sanitizeFilename(input.Filename))

	// Generate presigned URL (5 min expiry)
	expiry := 5 * time.Minute
	uploadURL, err := s.minioClient.GeneratePresignedPutURL(ctx, objectKey, input.ContentType, expiry)
	if err != nil {
		log.Error().Err(err).Str("object_key", objectKey).Msg("failed to generate presigned URL")
		return nil, fmt.Errorf("failed to generate upload URL: %w", err)
	}

	return &UploadURLOutput{
		UploadURL: uploadURL,
		ObjectKey: objectKey,
		ExpiresAt: time.Now().Add(expiry),
	}, nil
}

// ConfirmUploadInput represents input for confirming an upload.
type ConfirmUploadInput struct {
	ObjectKey   string
	PodID       uuid.UUID
	UploaderID  uuid.UUID
	Title       string
	Description *string
}

// ConfirmUpload confirms a file upload and creates a material record.
// Implements requirement 5: Material Upload.
func (s *Service) ConfirmUpload(ctx context.Context, input ConfirmUploadInput) (*domain.Material, error) {
	// Verify file exists in MinIO
	fileInfo, err := s.minioClient.GetFileInfo(ctx, input.ObjectKey)
	if err != nil {
		log.Error().Err(err).Str("object_key", input.ObjectKey).Msg("file not found in storage")
		return nil, fmt.Errorf("file not found or upload incomplete")
	}

	// Determine file type from object key
	fileType := getFileTypeFromKey(input.ObjectKey)
	if fileType == "" {
		return nil, fmt.Errorf("unsupported file type")
	}

	// Create material record
	material := domain.NewMaterial(
		input.PodID,
		input.UploaderID,
		input.Title,
		fileType,
		input.ObjectKey,
		fileInfo.Size,
	)
	material.Description = input.Description

	if err := s.materialRepo.Create(ctx, material); err != nil {
		log.Error().Err(err).Msg("failed to create material record")
		return nil, fmt.Errorf("failed to create material: %w", err)
	}

	// Create initial version
	version := domain.NewMaterialVersion(
		material.ID,
		1,
		input.ObjectKey,
		fileInfo.Size,
		input.UploaderID,
		nil,
	)
	if err := s.versionRepo.Create(ctx, version); err != nil {
		log.Error().Err(err).Msg("failed to create material version")
		// Don't fail the whole operation, version is secondary
	}

	log.Info().
		Str("material_id", material.ID.String()).
		Str("pod_id", input.PodID.String()).
		Str("uploader_id", input.UploaderID.String()).
		Msg("material created successfully")

	// Publish material.uploaded event for background processing (AI Service, Search Service)
	// Implements requirement 5, 6: Material Upload and Material Processing
	if s.eventPublisher != nil {
		event := nats.MaterialUploadedEvent{
			MaterialID: material.ID,
			PodID:      material.PodID,
			FileURL:    material.FileURL,
			FileType:   string(material.FileType),
			UploaderID: material.UploaderID,
		}
		if err := s.eventPublisher.PublishMaterialUploaded(ctx, event); err != nil {
			// Log error but don't fail the operation - event publishing is non-critical
			log.Error().Err(err).
				Str("material_id", material.ID.String()).
				Msg("failed to publish material.uploaded event")
		} else {
			log.Debug().
				Str("material_id", material.ID.String()).
				Msg("published material.uploaded event")
		}
	}

	return material, nil
}

// GetMaterial retrieves a material by ID.
func (s *Service) GetMaterial(ctx context.Context, id uuid.UUID) (*domain.Material, error) {
	material, err := s.materialRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Increment view count asynchronously
	go func() {
		if err := s.materialRepo.IncrementViewCount(context.Background(), id); err != nil {
			log.Error().Err(err).Str("material_id", id.String()).Msg("failed to increment view count")
		}
	}()

	return material, nil
}

// UpdateMaterialInput represents input for updating a material.
type UpdateMaterialInput struct {
	ID          uuid.UUID
	Title       *string
	Description *string
}

// UpdateMaterial updates a material's metadata.
func (s *Service) UpdateMaterial(ctx context.Context, input UpdateMaterialInput) (*domain.Material, error) {
	material, err := s.materialRepo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if input.Title != nil {
		material.Title = *input.Title
	}
	if input.Description != nil {
		material.Description = input.Description
	}
	material.UpdatedAt = time.Now()

	if err := s.materialRepo.Update(ctx, material); err != nil {
		return nil, fmt.Errorf("failed to update material: %w", err)
	}

	return material, nil
}

// DeleteMaterial soft-deletes a material.
func (s *Service) DeleteMaterial(ctx context.Context, id uuid.UUID) error {
	// Get material first to have PodID for the event
	material, err := s.materialRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.materialRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete material: %w", err)
	}

	log.Info().Str("material_id", id.String()).Msg("material deleted")

	// Publish material.deleted event for cleanup (AI Service removes vectors, Search Service removes index)
	// Implements requirement 5, 6: Material Upload and Material Processing
	if s.eventPublisher != nil {
		event := nats.MaterialDeletedEvent{
			MaterialID: material.ID,
			PodID:      material.PodID,
		}
		if err := s.eventPublisher.PublishMaterialDeleted(ctx, event); err != nil {
			// Log error but don't fail the operation - event publishing is non-critical
			log.Error().Err(err).
				Str("material_id", id.String()).
				Msg("failed to publish material.deleted event")
		} else {
			log.Debug().
				Str("material_id", id.String()).
				Msg("published material.deleted event")
		}
	}

	return nil
}

// GetMaterialsByPod retrieves all materials in a pod.
func (s *Service) GetMaterialsByPod(ctx context.Context, podID uuid.UUID, limit, offset int) ([]*domain.Material, int, error) {
	return s.materialRepo.FindByPodID(ctx, podID, limit, offset)
}

// GetPreviewURL generates a presigned URL for previewing a material.
func (s *Service) GetPreviewURL(ctx context.Context, id uuid.UUID) (string, error) {
	material, err := s.materialRepo.FindByID(ctx, id)
	if err != nil {
		return "", err
	}

	// Generate presigned URL (1 hour expiry for preview)
	url, err := s.minioClient.GeneratePresignedGetURL(ctx, material.FileURL, time.Hour)
	if err != nil {
		return "", fmt.Errorf("failed to generate preview URL: %w", err)
	}

	return url, nil
}

// GetDownloadURL generates a presigned URL for downloading a material.
func (s *Service) GetDownloadURL(ctx context.Context, id uuid.UUID) (string, error) {
	material, err := s.materialRepo.FindByID(ctx, id)
	if err != nil {
		return "", err
	}

	// Increment download count
	go func() {
		if err := s.materialRepo.IncrementDownloadCount(context.Background(), id); err != nil {
			log.Error().Err(err).Str("material_id", id.String()).Msg("failed to increment download count")
		}
	}()

	// Generate presigned URL (1 hour expiry for download)
	url, err := s.minioClient.GeneratePresignedGetURL(ctx, material.FileURL, time.Hour)
	if err != nil {
		return "", fmt.Errorf("failed to generate download URL: %w", err)
	}

	return url, nil
}

// CreateVersionInput represents input for creating a new version.
type CreateVersionInput struct {
	MaterialID uuid.UUID
	ObjectKey  string
	UploaderID uuid.UUID
	Changelog  *string
}

// CreateVersion creates a new version of a material.
// Implements requirement 5.1: Material Versioning.
func (s *Service) CreateVersion(ctx context.Context, input CreateVersionInput) (*domain.MaterialVersion, error) {
	// Get file info
	fileInfo, err := s.minioClient.GetFileInfo(ctx, input.ObjectKey)
	if err != nil {
		return nil, fmt.Errorf("file not found or upload incomplete")
	}

	// Get latest version number
	latestVersion, err := s.versionRepo.GetLatestVersion(ctx, input.MaterialID)
	if err != nil {
		latestVersion = 0
	}
	newVersion := latestVersion + 1

	// Create version record
	version := domain.NewMaterialVersion(
		input.MaterialID,
		newVersion,
		input.ObjectKey,
		fileInfo.Size,
		input.UploaderID,
		input.Changelog,
	)

	if err := s.versionRepo.Create(ctx, version); err != nil {
		return nil, fmt.Errorf("failed to create version: %w", err)
	}

	// Update material's current version and file URL
	if err := s.materialRepo.IncrementVersion(ctx, input.MaterialID); err != nil {
		log.Error().Err(err).Msg("failed to increment material version")
	}

	log.Info().
		Str("material_id", input.MaterialID.String()).
		Int("version", newVersion).
		Msg("new version created")

	return version, nil
}

// GetVersionHistory retrieves all versions of a material.
func (s *Service) GetVersionHistory(ctx context.Context, materialID uuid.UUID) ([]*domain.MaterialVersion, error) {
	return s.versionRepo.FindByMaterialID(ctx, materialID)
}

// AddCommentInput represents input for adding a comment.
type AddCommentInput struct {
	MaterialID uuid.UUID
	UserID     uuid.UUID
	Content    string
	ParentID   *uuid.UUID
}

// AddComment adds a comment to a material.
// Implements requirement 5.2: Material Comments.
func (s *Service) AddComment(ctx context.Context, input AddCommentInput) (*domain.Comment, error) {
	comment := domain.NewComment(input.MaterialID, input.UserID, input.Content, input.ParentID)

	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	return comment, nil
}

// GetComments retrieves comments for a material.
func (s *Service) GetComments(ctx context.Context, materialID uuid.UUID, limit, offset int) ([]*domain.CommentWithUser, int, error) {
	return s.commentRepo.FindByMaterialIDWithUsers(ctx, materialID, limit, offset)
}

// UpdateCommentInput represents input for updating a comment.
type UpdateCommentInput struct {
	ID      uuid.UUID
	UserID  uuid.UUID
	Content string
}

// UpdateComment updates a comment.
func (s *Service) UpdateComment(ctx context.Context, input UpdateCommentInput) (*domain.Comment, error) {
	comment, err := s.commentRepo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if comment.UserID != input.UserID {
		return nil, fmt.Errorf("not authorized to update this comment")
	}

	comment.Content = input.Content
	comment.Edited = true
	comment.UpdatedAt = time.Now()

	if err := s.commentRepo.Update(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to update comment: %w", err)
	}

	return comment, nil
}

// DeleteComment soft-deletes a comment.
func (s *Service) DeleteComment(ctx context.Context, id, userID uuid.UUID) error {
	comment, err := s.commentRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Verify ownership
	if comment.UserID != userID {
		return fmt.Errorf("not authorized to delete this comment")
	}

	return s.commentRepo.Delete(ctx, id)
}

// RateMaterialInput represents input for rating a material.
type RateMaterialInput struct {
	MaterialID uuid.UUID
	UserID     uuid.UUID
	Score      int
	Review     *string
}

// RateMaterial rates a material.
// Implements requirement 5.3: Material Ratings & Reviews.
func (s *Service) RateMaterial(ctx context.Context, input RateMaterialInput) (*domain.Rating, error) {
	// Validate score
	if input.Score < 1 || input.Score > 5 {
		return nil, fmt.Errorf("score must be between 1 and 5")
	}

	// Check if user already rated
	existing, err := s.ratingRepo.FindByMaterialAndUser(ctx, input.MaterialID, input.UserID)
	if err == nil && existing != nil {
		// Update existing rating
		existing.Score = input.Score
		existing.Review = input.Review
		existing.UpdatedAt = time.Now()

		if err := s.ratingRepo.Update(ctx, existing); err != nil {
			return nil, fmt.Errorf("failed to update rating: %w", err)
		}

		// Update material rating stats
		s.updateMaterialRatingStats(ctx, input.MaterialID)

		return existing, nil
	}

	// Create new rating
	rating := domain.NewRating(input.MaterialID, input.UserID, input.Score, input.Review)

	if err := s.ratingRepo.Create(ctx, rating); err != nil {
		return nil, fmt.Errorf("failed to create rating: %w", err)
	}

	// Update material rating stats
	s.updateMaterialRatingStats(ctx, input.MaterialID)

	return rating, nil
}

// GetRatings retrieves ratings for a material.
func (s *Service) GetRatings(ctx context.Context, materialID uuid.UUID, limit, offset int) ([]*domain.RatingWithUser, int, error) {
	return s.ratingRepo.FindByMaterialIDWithUsers(ctx, materialID, limit, offset)
}

// GetRatingSummary retrieves the rating summary for a material.
func (s *Service) GetRatingSummary(ctx context.Context, materialID uuid.UUID) (*domain.RatingSummary, error) {
	return s.ratingRepo.GetSummary(ctx, materialID)
}

// updateMaterialRatingStats updates the material's rating statistics.
func (s *Service) updateMaterialRatingStats(ctx context.Context, materialID uuid.UUID) {
	avgRating, count, err := s.ratingRepo.CalculateAverage(ctx, materialID)
	if err != nil {
		log.Error().Err(err).Str("material_id", materialID.String()).Msg("failed to calculate rating average")
		return
	}

	if err := s.materialRepo.UpdateRatingStats(ctx, materialID, avgRating, count); err != nil {
		log.Error().Err(err).Str("material_id", materialID.String()).Msg("failed to update rating stats")
	}
}

// BookmarkMaterialInput represents input for bookmarking a material.
type BookmarkMaterialInput struct {
	UserID     uuid.UUID
	MaterialID uuid.UUID
	Folder     *string
}

// BookmarkMaterial bookmarks a material.
// Implements requirement 5.4: Material Bookmarks.
func (s *Service) BookmarkMaterial(ctx context.Context, input BookmarkMaterialInput) (*domain.Bookmark, error) {
	// Check if already bookmarked
	exists, err := s.bookmarkRepo.Exists(ctx, input.UserID, input.MaterialID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("material already bookmarked")
	}

	bookmark := domain.NewBookmark(input.UserID, input.MaterialID, input.Folder)

	if err := s.bookmarkRepo.Create(ctx, bookmark); err != nil {
		return nil, fmt.Errorf("failed to create bookmark: %w", err)
	}

	return bookmark, nil
}

// RemoveBookmark removes a bookmark.
func (s *Service) RemoveBookmark(ctx context.Context, userID, materialID uuid.UUID) error {
	return s.bookmarkRepo.DeleteByUserAndMaterial(ctx, userID, materialID)
}

// GetBookmarks retrieves bookmarks for a user.
func (s *Service) GetBookmarks(ctx context.Context, userID uuid.UUID, folder *string, limit, offset int) ([]*domain.MaterialWithUploader, int, error) {
	return s.bookmarkRepo.FindByUserIDWithMaterials(ctx, userID, folder, limit, offset)
}

// GetBookmarkFolders retrieves all bookmark folders for a user.
func (s *Service) GetBookmarkFolders(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return s.bookmarkRepo.GetFolders(ctx, userID)
}

// Helper functions

// sanitizeFilename removes or replaces unsafe characters from a filename.
func sanitizeFilename(filename string) string {
	// Simple sanitization - in production, use a more robust solution
	result := ""
	for _, r := range filename {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			result += string(r)
		}
	}
	if result == "" {
		result = "file"
	}
	return result
}

// getFileTypeFromKey extracts the file type from an object key.
func getFileTypeFromKey(objectKey string) domain.FileType {
	// Get extension from object key
	for i := len(objectKey) - 1; i >= 0; i-- {
		if objectKey[i] == '.' {
			ext := objectKey[i+1:]
			switch ext {
			case "pdf":
				return domain.FileTypePDF
			case "docx":
				return domain.FileTypeDOCX
			case "pptx":
				return domain.FileTypePPTX
			}
			break
		}
	}
	return ""
}
