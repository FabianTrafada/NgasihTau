// Package domain contains the core business entities and repository interfaces
// for the Material Service. This layer is independent of external frameworks and databases.
package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// MaterialStatus represents the processing status of a material.
type MaterialStatus string

const (
	// MaterialStatusProcessing means the material is being processed.
	MaterialStatusProcessing MaterialStatus = "processing"
	// MaterialStatusReady means the material is ready for use.
	MaterialStatusReady MaterialStatus = "ready"
	// MaterialStatusProcessingFailed means processing failed.
	MaterialStatusProcessingFailed MaterialStatus = "processing_failed"
)

// FileType represents the type of uploaded file.
type FileType string

const (
	// FileTypePDF represents a PDF document.
	FileTypePDF FileType = "pdf"
	// FileTypeDOCX represents a Word document.
	FileTypeDOCX FileType = "docx"
	// FileTypePPTX represents a PowerPoint presentation.
	FileTypePPTX FileType = "pptx"
)

// Material represents a learning material in a Knowledge Pod.
// Implements requirements 5, 5.1.
type Material struct {
	ID             uuid.UUID      `json:"id"`
	PodID          uuid.UUID      `json:"pod_id"`
	UploaderID     uuid.UUID      `json:"uploader_id"`
	Title          string         `json:"title"`
	Description    *string        `json:"description,omitempty"`
	FileType       FileType       `json:"file_type"`
	FileURL        string         `json:"file_url"`
	FileSize       int64          `json:"file_size"`
	CurrentVersion int            `json:"current_version"`
	Status         MaterialStatus `json:"status"`
	ViewCount      int            `json:"view_count"`
	DownloadCount  int            `json:"download_count"`
	AverageRating  float64        `json:"average_rating"`
	RatingCount    int            `json:"rating_count"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      *time.Time     `json:"-"`
}

// IsReady returns true if the material is ready for use.
func (m *Material) IsReady() bool {
	return m.Status == MaterialStatusReady
}

// IsProcessing returns true if the material is being processed.
func (m *Material) IsProcessing() bool {
	return m.Status == MaterialStatusProcessing
}

// MaterialVersion represents a version of a material.
// Implements requirement 5.1.
type MaterialVersion struct {
	ID         uuid.UUID `json:"id"`
	MaterialID uuid.UUID `json:"material_id"`
	Version    int       `json:"version"`
	FileURL    string    `json:"file_url"`
	FileSize   int64     `json:"file_size"`
	UploaderID uuid.UUID `json:"uploader_id"`
	Changelog  *string   `json:"changelog,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// Comment represents a comment on a material.
// Implements requirement 5.2.
type Comment struct {
	ID         uuid.UUID  `json:"id"`
	MaterialID uuid.UUID  `json:"material_id"`
	UserID     uuid.UUID  `json:"user_id"`
	ParentID   *uuid.UUID `json:"parent_id,omitempty"`
	Content    string     `json:"content"`
	Edited     bool       `json:"edited"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"-"`
}

// IsDeleted returns true if the comment has been soft-deleted.
func (c *Comment) IsDeleted() bool {
	return c.DeletedAt != nil
}

// Rating represents a rating on a material.
// Implements requirement 5.3.
type Rating struct {
	ID         uuid.UUID `json:"id"`
	MaterialID uuid.UUID `json:"material_id"`
	UserID     uuid.UUID `json:"user_id"`
	Score      int       `json:"score"` // 1-5
	Review     *string   `json:"review,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Bookmark represents a user's bookmark on a material.
// Implements requirement 5.4.
type Bookmark struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	MaterialID uuid.UUID `json:"material_id"`
	Folder     *string   `json:"folder,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// CommentWithUser represents a comment with user information.
type CommentWithUser struct {
	Comment
	UserName      string  `json:"user_name"`
	UserAvatarURL *string `json:"user_avatar_url,omitempty"`
}

// RatingWithUser represents a rating with user information.
type RatingWithUser struct {
	Rating
	UserName      string  `json:"user_name"`
	UserAvatarURL *string `json:"user_avatar_url,omitempty"`
}

// MaterialWithUploader represents a material with uploader information.
type MaterialWithUploader struct {
	Material
	UploaderName      string  `json:"uploader_name"`
	UploaderAvatarURL *string `json:"uploader_avatar_url,omitempty"`
}

// RatingDistribution represents the distribution of ratings.
type RatingDistribution struct {
	OneStar   int `json:"one_star"`
	TwoStar   int `json:"two_star"`
	ThreeStar int `json:"three_star"`
	FourStar  int `json:"four_star"`
	FiveStar  int `json:"five_star"`
}

// RatingSummary represents a summary of ratings for a material.
type RatingSummary struct {
	AverageRating float64            `json:"average_rating"`
	RatingCount   int                `json:"rating_count"`
	Distribution  RatingDistribution `json:"distribution"`
}

// MaterialMetadata represents additional metadata for a material.
type MaterialMetadata map[string]any

// Value implements driver.Valuer for database storage.
func (m MaterialMetadata) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan implements sql.Scanner for database retrieval.
func (m *MaterialMetadata) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("cannot scan type %T into MaterialMetadata", value)
	}

	return json.Unmarshal(data, m)
}

// NewMaterial creates a new Material with default values.
func NewMaterial(podID, uploaderID uuid.UUID, title string, fileType FileType, fileURL string, fileSize int64) *Material {
	now := time.Now()
	return &Material{
		ID:             uuid.New(),
		PodID:          podID,
		UploaderID:     uploaderID,
		Title:          title,
		FileType:       fileType,
		FileURL:        fileURL,
		FileSize:       fileSize,
		CurrentVersion: 1,
		Status:         MaterialStatusProcessing,
		ViewCount:      0,
		DownloadCount:  0,
		AverageRating:  0,
		RatingCount:    0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// NewMaterialVersion creates a new MaterialVersion.
func NewMaterialVersion(materialID uuid.UUID, version int, fileURL string, fileSize int64, uploaderID uuid.UUID, changelog *string) *MaterialVersion {
	return &MaterialVersion{
		ID:         uuid.New(),
		MaterialID: materialID,
		Version:    version,
		FileURL:    fileURL,
		FileSize:   fileSize,
		UploaderID: uploaderID,
		Changelog:  changelog,
		CreatedAt:  time.Now(),
	}
}

// NewComment creates a new Comment.
func NewComment(materialID, userID uuid.UUID, content string, parentID *uuid.UUID) *Comment {
	now := time.Now()
	return &Comment{
		ID:         uuid.New(),
		MaterialID: materialID,
		UserID:     userID,
		ParentID:   parentID,
		Content:    content,
		Edited:     false,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// NewRating creates a new Rating.
func NewRating(materialID, userID uuid.UUID, score int, review *string) *Rating {
	now := time.Now()
	return &Rating{
		ID:         uuid.New(),
		MaterialID: materialID,
		UserID:     userID,
		Score:      score,
		Review:     review,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// NewBookmark creates a new Bookmark.
func NewBookmark(userID, materialID uuid.UUID, folder *string) *Bookmark {
	return &Bookmark{
		ID:         uuid.New(),
		UserID:     userID,
		MaterialID: materialID,
		Folder:     folder,
		CreatedAt:  time.Now(),
	}
}
