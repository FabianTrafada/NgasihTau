// Package docs provides Swagger/OpenAPI documentation for NgasihTau API.
// This file contains the main API information and common types used across all services.
//
// @title NgasihTau API
// @version 1.0
// @description NgasihTau is a learning platform API that enables teachers to create Knowledge Pods for sharing learning materials with students.
// @description The platform uses a microservices architecture with Go as the primary backend language.
//
// @contact.name NgasihTau API Support
// @contact.email support@ngasihtau.com
//
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
//
// @host localhost:8001
// @BasePath /api/v1
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Bearer token. Format: "Bearer {token}"
//
// @tag.name Auth
// @tag.description Authentication and authorization endpoints
//
// @tag.name Users
// @tag.description User profile and social features
//
// @tag.name Pods
// @tag.description Knowledge Pod management and collaboration
//
// @tag.name Materials
// @tag.description Learning material upload, versioning, and management
//
// @tag.name Comments
// @tag.description Material comments and discussions
//
// @tag.name Ratings
// @tag.description Material ratings and reviews
//
// @tag.name Bookmarks
// @tag.description Material bookmarking
//
// @tag.name Search
// @tag.description Full-text and semantic search
//
// @tag.name AI
// @tag.description AI-powered chat and Q&A
//
// @tag.name Notifications
// @tag.description In-app and email notifications
package docs

import "time"

// ============================================================================
// Common Response Types
// ============================================================================

// Meta contains metadata for API responses.
type Meta struct {
	Timestamp string `json:"timestamp" example:"2025-12-08T10:00:00Z"`
	RequestID string `json:"request_id" example:"req_abc123"`
}

// Pagination contains pagination information for list responses.
type Pagination struct {
	Page       int `json:"page" example:"1"`
	PerPage    int `json:"per_page" example:"20"`
	Total      int `json:"total" example:"100"`
	TotalPages int `json:"total_pages" example:"5"`
}

// ErrorBody contains error details in the response.
type ErrorBody struct {
	Code    string        `json:"code" example:"VALIDATION_ERROR"`
	Message string        `json:"message" example:"Invalid input data"`
	Details []ErrorDetail `json:"details,omitempty"`
}

// ErrorDetail contains field-level error information.
type ErrorDetail struct {
	Field   string `json:"field" example:"email"`
	Message string `json:"message" example:"Invalid email format"`
}

// ErrorResponse represents the standard error response envelope.
// @Description Standard error response format
type ErrorResponse struct {
	Success bool      `json:"success" example:"false"`
	Error   ErrorBody `json:"error"`
	Meta    Meta      `json:"meta"`
}

// SuccessResponse represents a generic success response.
// @Description Standard success response format
type SuccessResponse struct {
	Success bool        `json:"success" example:"true"`
	Data    interface{} `json:"data"`
	Meta    Meta        `json:"meta"`
}

// EmptyResponse represents a success response with no data.
// @Description Success response with no data
type EmptyResponse struct {
	Success bool        `json:"success" example:"true"`
	Data    interface{} `json:"data" swaggertype:"object"`
	Meta    Meta        `json:"meta"`
}

// MessageResponse represents a success response with a message.
// @Description Success response with a message
type MessageResponse struct {
	Success bool              `json:"success" example:"true"`
	Data    map[string]string `json:"data"`
	Meta    Meta              `json:"meta"`
}

// ============================================================================
// Auth Types
// ============================================================================

// RegisterRequest represents the request body for user registration.
// @Description User registration request
type RegisterRequest struct {
	Email    string `json:"email" example:"user@example.com" validate:"required,email"`
	Password string `json:"password" example:"SecurePass123!" validate:"required,min=8"`
	Name     string `json:"name" example:"John Doe" validate:"required,min=2,max=100"`
}

// LoginRequest represents the request body for user login.
// @Description User login request
type LoginRequest struct {
	Email    string `json:"email" example:"user@example.com" validate:"required,email"`
	Password string `json:"password" example:"SecurePass123!" validate:"required"`
}

// GoogleLoginRequest represents the request body for Google OAuth login.
// @Description Google OAuth login request
type GoogleLoginRequest struct {
	Code        string `json:"code" example:"4/0AX4XfWh..." validate:"required"`
	RedirectURI string `json:"redirect_uri" example:"https://app.ngasihtau.com/auth/callback" validate:"required,url"`
}

// RefreshTokenRequest represents the request body for token refresh.
// @Description Token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIs..." validate:"required"`
}

// LogoutRequest represents the request body for logout.
// @Description Logout request
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIs..." validate:"required"`
}

// Verify2FARequest represents the request body for 2FA verification.
// @Description 2FA verification request
type Verify2FARequest struct {
	Code string `json:"code" example:"123456" validate:"required,min=6,max=10"`
}

// Verify2FALoginRequest represents the request body for 2FA login verification.
// @Description 2FA login verification request
type Verify2FALoginRequest struct {
	TempToken string `json:"temp_token" example:"temp_abc123" validate:"required"`
	Code      string `json:"code" example:"123456" validate:"required,min=6,max=10"`
}

// PasswordResetRequest represents the request body for password reset request.
// @Description Password reset request
type PasswordResetRequest struct {
	Email string `json:"email" example:"user@example.com" validate:"required,email"`
}

// PasswordResetConfirmRequest represents the request body for password reset confirmation.
// @Description Password reset confirmation request
type PasswordResetConfirmRequest struct {
	Token       string `json:"token" example:"reset_token_abc123" validate:"required"`
	NewPassword string `json:"new_password" example:"NewSecurePass123!" validate:"required,min=8"`
}

// VerifyEmailRequest represents the request body for email verification.
// @Description Email verification request
type VerifyEmailRequest struct {
	Token string `json:"token" example:"verify_token_abc123" validate:"required"`
}

// UserResponse represents a user in API responses.
// @Description User information
type UserResponse struct {
	ID               string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email            string    `json:"email,omitempty" example:"user@example.com"`
	Name             string    `json:"name" example:"John Doe"`
	AvatarURL        *string   `json:"avatar_url,omitempty" example:"https://example.com/avatar.jpg"`
	Bio              *string   `json:"bio,omitempty" example:"Software developer and educator"`
	Role             string    `json:"role" example:"teacher"`
	EmailVerified    bool      `json:"email_verified,omitempty" example:"true"`
	TwoFactorEnabled bool      `json:"two_factor_enabled,omitempty" example:"false"`
	Language         string    `json:"language,omitempty" example:"en"`
	CreatedAt        time.Time `json:"created_at" example:"2025-01-15T10:00:00Z"`
}

// AuthResponse represents the response for successful authentication.
// @Description Authentication response with tokens
type AuthResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	RefreshToken string        `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	ExpiresIn    int64         `json:"expires_in" example:"900"`
}

// TwoFactorSetupResponse represents the response for 2FA setup.
// @Description 2FA setup response with secret and backup codes
type TwoFactorSetupResponse struct {
	Secret      string   `json:"secret" example:"JBSWY3DPEHPK3PXP"`
	QRCodeURL   string   `json:"qr_code_url" example:"otpauth://totp/NgasihTau:user@example.com?secret=JBSWY3DPEHPK3PXP&issuer=NgasihTau"`
	BackupCodes []string `json:"backup_codes" example:"ABCD-1234,EFGH-5678"`
}

// TwoFactorRequiredResponse represents the response when 2FA is required.
// @Description Response indicating 2FA verification is required
type TwoFactorRequiredResponse struct {
	TempToken       string `json:"temp_token" example:"temp_abc123"`
	TwoFactorMethod string `json:"two_factor_method" example:"totp"`
}

// ============================================================================
// User Types
// ============================================================================

// UserProfileResponse represents a user's public profile.
// @Description User public profile
type UserProfileResponse struct {
	ID             string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name           string    `json:"name" example:"John Doe"`
	AvatarURL      *string   `json:"avatar_url,omitempty" example:"https://example.com/avatar.jpg"`
	Bio            *string   `json:"bio,omitempty" example:"Software developer and educator"`
	Role           string    `json:"role" example:"teacher"`
	FollowerCount  int       `json:"follower_count" example:"150"`
	FollowingCount int       `json:"following_count" example:"75"`
	PodCount       int       `json:"pod_count" example:"10"`
	MaterialCount  int       `json:"material_count" example:"45"`
	CreatedAt      time.Time `json:"created_at" example:"2025-01-15T10:00:00Z"`
	IsFollowing    bool      `json:"is_following,omitempty" example:"false"`
}

// UpdateProfileRequest represents the request body for profile update.
// @Description Profile update request
type UpdateProfileRequest struct {
	Name      *string `json:"name,omitempty" example:"John Doe" validate:"omitempty,min=2,max=100"`
	Bio       *string `json:"bio,omitempty" example:"Software developer" validate:"omitempty,max=500"`
	AvatarURL *string `json:"avatar_url,omitempty" example:"https://example.com/avatar.jpg" validate:"omitempty,url"`
	Language  *string `json:"language,omitempty" example:"en" validate:"omitempty,oneof=id en"`
}

// ============================================================================
// Pod Types
// ============================================================================

// PodResponse represents a Knowledge Pod in API responses.
// @Description Knowledge Pod information
type PodResponse struct {
	ID           string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	OwnerID      string    `json:"owner_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	Name         string    `json:"name" example:"Introduction to Go Programming"`
	Slug         string    `json:"slug" example:"introduction-to-go-programming"`
	Description  *string   `json:"description,omitempty" example:"A comprehensive guide to Go programming language"`
	Visibility   string    `json:"visibility" example:"public"`
	Categories   []string  `json:"categories" example:"programming,golang"`
	Tags         []string  `json:"tags" example:"beginner,tutorial"`
	StarCount    int       `json:"star_count" example:"42"`
	ForkCount    int       `json:"fork_count" example:"5"`
	ViewCount    int       `json:"view_count" example:"1500"`
	ForkedFromID *string   `json:"forked_from_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440002"`
	CreatedAt    time.Time `json:"created_at" example:"2025-01-15T10:00:00Z"`
	UpdatedAt    time.Time `json:"updated_at" example:"2025-02-20T15:30:00Z"`
}

// CreatePodRequest represents the request body for creating a pod.
// @Description Create pod request
type CreatePodRequest struct {
	Name        string   `json:"name" example:"Introduction to Go Programming" validate:"required,min=3,max=255"`
	Description *string  `json:"description,omitempty" example:"A comprehensive guide to Go programming language" validate:"omitempty,max=2000"`
	Visibility  string   `json:"visibility" example:"public" validate:"required,oneof=public private"`
	Categories  []string `json:"categories,omitempty" example:"programming,golang"`
	Tags        []string `json:"tags,omitempty" example:"beginner,tutorial"`
}

// UpdatePodRequest represents the request body for updating a pod.
// @Description Update pod request
type UpdatePodRequest struct {
	Name        *string  `json:"name,omitempty" example:"Advanced Go Programming" validate:"omitempty,min=3,max=255"`
	Description *string  `json:"description,omitempty" example:"Advanced topics in Go" validate:"omitempty,max=2000"`
	Visibility  *string  `json:"visibility,omitempty" example:"public" validate:"omitempty,oneof=public private"`
	Categories  []string `json:"categories,omitempty" example:"programming,golang"`
	Tags        []string `json:"tags,omitempty" example:"advanced,tutorial"`
}

// CollaboratorResponse represents a collaborator in API responses.
// @Description Pod collaborator information
type CollaboratorResponse struct {
	ID        string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	PodID     string    `json:"pod_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	UserID    string    `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	Role      string    `json:"role" example:"contributor"`
	Status    string    `json:"status" example:"verified"`
	InvitedBy string    `json:"invited_by" example:"550e8400-e29b-41d4-a716-446655440003"`
	CreatedAt time.Time `json:"created_at" example:"2025-01-15T10:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2025-01-16T12:00:00Z"`
}

// InviteCollaboratorRequest represents the request body for inviting a collaborator.
// @Description Invite collaborator request
type InviteCollaboratorRequest struct {
	Email string `json:"email" example:"collaborator@example.com" validate:"required,email"`
	Role  string `json:"role" example:"contributor" validate:"required,oneof=viewer contributor admin"`
}

// UpdateCollaboratorRequest represents the request body for updating a collaborator.
// @Description Update collaborator request
type UpdateCollaboratorRequest struct {
	Action string `json:"action" example:"verify" validate:"required,oneof=verify update_role"`
	Role   string `json:"role,omitempty" example:"admin"`
}

// ActivityResponse represents an activity in API responses.
// @Description Pod activity information
type ActivityResponse struct {
	ID        string                 `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	PodID     string                 `json:"pod_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	UserID    string                 `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	Action    string                 `json:"action" example:"material_uploaded"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at" example:"2025-01-15T10:00:00Z"`
}

// ============================================================================
// Material Types
// ============================================================================

// MaterialResponse represents a material in API responses.
// @Description Learning material information
type MaterialResponse struct {
	ID             string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	PodID          string    `json:"pod_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	UploaderID     string    `json:"uploader_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	Title          string    `json:"title" example:"Chapter 1: Introduction"`
	Description    *string   `json:"description,omitempty" example:"Introduction to the course"`
	FileType       string    `json:"file_type" example:"pdf"`
	FileURL        string    `json:"file_url" example:"https://storage.example.com/materials/file.pdf"`
	FileSize       int64     `json:"file_size" example:"1048576"`
	CurrentVersion int       `json:"current_version" example:"1"`
	Status         string    `json:"status" example:"ready"`
	ViewCount      int       `json:"view_count" example:"250"`
	DownloadCount  int       `json:"download_count" example:"50"`
	AverageRating  float64   `json:"average_rating" example:"4.5"`
	RatingCount    int       `json:"rating_count" example:"20"`
	CreatedAt      time.Time `json:"created_at" example:"2025-01-15T10:00:00Z"`
	UpdatedAt      time.Time `json:"updated_at" example:"2025-02-20T15:30:00Z"`
}

// UploadURLRequest represents the request body for getting an upload URL.
// @Description Upload URL request
type UploadURLRequest struct {
	Filename    string `json:"filename" example:"chapter1.pdf" validate:"required"`
	ContentType string `json:"content_type" example:"application/pdf" validate:"required"`
	Size        int64  `json:"size" example:"1048576" validate:"required,max=104857600"`
}

// UploadURLResponse represents the response for upload URL request.
// @Description Upload URL response
type UploadURLResponse struct {
	UploadURL string    `json:"upload_url" example:"https://storage.example.com/upload?signature=..."`
	ObjectKey string    `json:"object_key" example:"materials/abc123/chapter1.pdf"`
	ExpiresAt time.Time `json:"expires_at" example:"2025-01-15T10:05:00Z"`
}

// ConfirmUploadRequest represents the request body for confirming an upload.
// @Description Confirm upload request
type ConfirmUploadRequest struct {
	ObjectKey   string  `json:"object_key" example:"materials/abc123/chapter1.pdf" validate:"required"`
	PodID       string  `json:"pod_id" example:"550e8400-e29b-41d4-a716-446655440000" validate:"required,uuid"`
	Title       string  `json:"title" example:"Chapter 1: Introduction" validate:"required,max=255"`
	Description *string `json:"description,omitempty" example:"Introduction to the course" validate:"omitempty,max=2000"`
}

// UpdateMaterialRequest represents the request body for updating a material.
// @Description Update material request
type UpdateMaterialRequest struct {
	Title       *string `json:"title,omitempty" example:"Chapter 1: Getting Started" validate:"omitempty,max=255"`
	Description *string `json:"description,omitempty" example:"Updated introduction" validate:"omitempty,max=2000"`
}

// CreateVersionRequest represents the request body for creating a new version.
// @Description Create version request
type CreateVersionRequest struct {
	ObjectKey string  `json:"object_key" example:"materials/abc123/chapter1_v2.pdf" validate:"required"`
	Changelog *string `json:"changelog,omitempty" example:"Fixed typos and added examples" validate:"omitempty,max=1000"`
}

// VersionResponse represents a material version in API responses.
// @Description Material version information
type VersionResponse struct {
	ID         string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	MaterialID string    `json:"material_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	Version    int       `json:"version" example:"2"`
	FileURL    string    `json:"file_url" example:"https://storage.example.com/materials/file_v2.pdf"`
	FileSize   int64     `json:"file_size" example:"1100000"`
	UploaderID string    `json:"uploader_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	Changelog  *string   `json:"changelog,omitempty" example:"Fixed typos"`
	CreatedAt  time.Time `json:"created_at" example:"2025-02-20T15:30:00Z"`
}

// ============================================================================
// Comment Types
// ============================================================================

// CommentResponse represents a comment in API responses.
// @Description Material comment information
type CommentResponse struct {
	ID         string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	MaterialID string    `json:"material_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	UserID     string    `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	ParentID   *string   `json:"parent_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440003"`
	Content    string    `json:"content" example:"Great explanation! Very helpful."`
	Edited     bool      `json:"edited" example:"false"`
	CreatedAt  time.Time `json:"created_at" example:"2025-01-15T10:00:00Z"`
	UpdatedAt  time.Time `json:"updated_at" example:"2025-01-15T10:00:00Z"`
}

// AddCommentRequest represents the request body for adding a comment.
// @Description Add comment request
type AddCommentRequest struct {
	Content  string  `json:"content" example:"Great explanation! Very helpful." validate:"required,max=2000"`
	ParentID *string `json:"parent_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000" validate:"omitempty,uuid"`
}

// UpdateCommentRequest represents the request body for updating a comment.
// @Description Update comment request
type UpdateCommentRequest struct {
	Content string `json:"content" example:"Updated comment content" validate:"required,max=2000"`
}

// ============================================================================
// Rating Types
// ============================================================================

// RatingResponse represents a rating in API responses.
// @Description Material rating information
type RatingResponse struct {
	ID         string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	MaterialID string    `json:"material_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	UserID     string    `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	Score      int       `json:"score" example:"5"`
	Review     *string   `json:"review,omitempty" example:"Excellent material, very comprehensive!"`
	CreatedAt  time.Time `json:"created_at" example:"2025-01-15T10:00:00Z"`
	UpdatedAt  time.Time `json:"updated_at" example:"2025-01-15T10:00:00Z"`
}

// RateMaterialRequest represents the request body for rating a material.
// @Description Rate material request
type RateMaterialRequest struct {
	Score  int     `json:"score" example:"5" validate:"required,min=1,max=5"`
	Review *string `json:"review,omitempty" example:"Excellent material!" validate:"omitempty,max=2000"`
}

// RatingSummaryResponse represents a rating summary.
// @Description Rating summary information
type RatingSummaryResponse struct {
	AverageRating float64        `json:"average_rating" example:"4.5"`
	TotalRatings  int            `json:"total_ratings" example:"100"`
	Distribution  map[string]int `json:"distribution" example:"1:5,2:10,3:15,4:30,5:40"`
}

// ============================================================================
// Bookmark Types
// ============================================================================

// BookmarkResponse represents a bookmark in API responses.
// @Description Material bookmark information
type BookmarkResponse struct {
	ID         string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID     string    `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	MaterialID string    `json:"material_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	Folder     *string   `json:"folder,omitempty" example:"Study Materials"`
	CreatedAt  time.Time `json:"created_at" example:"2025-01-15T10:00:00Z"`
}

// BookmarkMaterialRequest represents the request body for bookmarking a material.
// @Description Bookmark material request
type BookmarkMaterialRequest struct {
	Folder *string `json:"folder,omitempty" example:"Study Materials" validate:"omitempty,max=100"`
}

// ============================================================================
// Search Types
// ============================================================================

// SearchResult represents a search result item.
// @Description Search result item
type SearchResult struct {
	ID          string   `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Type        string   `json:"type" example:"material"`
	Title       string   `json:"title" example:"Introduction to Go"`
	Description *string  `json:"description,omitempty" example:"A beginner's guide"`
	Score       float64  `json:"score" example:"0.95"`
	Highlights  []string `json:"highlights,omitempty" example:"...Go programming..."`
}

// SemanticSearchResult represents a semantic search result.
// @Description Semantic search result
type SemanticSearchResult struct {
	MaterialID string  `json:"material_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ChunkText  string  `json:"chunk_text" example:"Go is a statically typed language..."`
	Score      float64 `json:"score" example:"0.92"`
	Page       *int    `json:"page,omitempty" example:"5"`
}

// SuggestionsResponse represents autocomplete suggestions.
// @Description Autocomplete suggestions
type SuggestionsResponse struct {
	Suggestions []string `json:"suggestions" example:"golang,go programming,go tutorial"`
}

// SearchHistoryItem represents a search history item.
// @Description Search history item
type SearchHistoryItem struct {
	Query     string    `json:"query" example:"golang tutorial"`
	Timestamp time.Time `json:"timestamp" example:"2025-01-15T10:00:00Z"`
}

// ============================================================================
// AI Chat Types
// ============================================================================

// ChatRequest represents the request body for AI chat.
// @Description AI chat request
type ChatRequest struct {
	Message string `json:"message" example:"What is the main topic of this material?" validate:"required,min=1,max=2000"`
}

// ChatMessageResponse represents a chat message in API responses.
// @Description AI chat message
type ChatMessageResponse struct {
	ID           string       `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	SessionID    string       `json:"session_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	Role         string       `json:"role" example:"assistant"`
	Content      string       `json:"content" example:"The main topic of this material is..."`
	Sources      []ChatSource `json:"sources,omitempty"`
	Feedback     *string      `json:"feedback,omitempty" example:"thumbs_up"`
	FeedbackText *string      `json:"feedback_text,omitempty"`
	CreatedAt    time.Time    `json:"created_at" example:"2025-01-15T10:00:00Z"`
}

// ChatSource represents a source reference in a chat response.
// @Description Chat source reference
type ChatSource struct {
	MaterialID string `json:"material_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ChunkIndex int    `json:"chunk_index" example:"5"`
	Page       *int   `json:"page,omitempty" example:"10"`
	Text       string `json:"text" example:"...relevant excerpt..."`
}

// ExportChatRequest represents the request body for exporting chat.
// @Description Export chat request
type ExportChatRequest struct {
	Format string `json:"format" example:"pdf" validate:"required,oneof=pdf markdown"`
}

// FeedbackRequest represents the request body for submitting feedback.
// @Description Chat feedback request
type FeedbackRequest struct {
	Feedback     string  `json:"feedback" example:"thumbs_up" validate:"required,oneof=thumbs_up thumbs_down"`
	FeedbackText *string `json:"feedback_text,omitempty" example:"Very helpful response!"`
}

// SuggestedQuestionsResponse represents suggested questions.
// @Description Suggested questions for AI chat
type SuggestedQuestionsResponse struct {
	Questions []string `json:"questions" example:"What are the key concepts?,Can you explain the main idea?,What are the practical applications?"`
}

// ============================================================================
// Notification Types
// ============================================================================

// NotificationResponse represents a notification in API responses.
// @Description Notification information
type NotificationResponse struct {
	ID        string                 `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID    string                 `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	Type      string                 `json:"type" example:"pod_invite"`
	Title     string                 `json:"title" example:"New Collaboration Invite"`
	Message   *string                `json:"message,omitempty" example:"You have been invited to collaborate on..."`
	Data      map[string]interface{} `json:"data,omitempty"`
	Read      bool                   `json:"read" example:"false"`
	CreatedAt time.Time              `json:"created_at" example:"2025-01-15T10:00:00Z"`
}

// NotificationPreferencesResponse represents notification preferences.
// @Description Notification preferences
type NotificationPreferencesResponse struct {
	UserID            string `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	EmailPodInvite    bool   `json:"email_pod_invite" example:"true"`
	EmailNewMaterial  bool   `json:"email_new_material" example:"true"`
	EmailCommentReply bool   `json:"email_comment_reply" example:"true"`
	InAppPodInvite    bool   `json:"inapp_pod_invite" example:"true"`
	InAppNewMaterial  bool   `json:"inapp_new_material" example:"true"`
	InAppCommentReply bool   `json:"inapp_comment_reply" example:"true"`
}

// UpdatePreferencesRequest represents the request body for updating preferences.
// @Description Update notification preferences request
type UpdatePreferencesRequest struct {
	EmailPodInvite    *bool `json:"email_pod_invite,omitempty" example:"true"`
	EmailNewMaterial  *bool `json:"email_new_material,omitempty" example:"true"`
	EmailCommentReply *bool `json:"email_comment_reply,omitempty" example:"false"`
	InAppPodInvite    *bool `json:"inapp_pod_invite,omitempty" example:"true"`
	InAppNewMaterial  *bool `json:"inapp_new_material,omitempty" example:"true"`
	InAppCommentReply *bool `json:"inapp_comment_reply,omitempty" example:"true"`
}
