// Package application contains unit tests for the Material Service.
package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	apperrors "ngasihtau/internal/common/errors"
	"ngasihtau/internal/material/domain"
	"ngasihtau/pkg/nats"
)

// Mock implementations for repositories and clients

type mockMaterialRepo struct {
	materials map[uuid.UUID]*domain.Material
	createErr error
	findErr   error
	updateErr error
	deleteErr error
}

func newMockMaterialRepo() *mockMaterialRepo {
	return &mockMaterialRepo{
		materials: make(map[uuid.UUID]*domain.Material),
	}
}

func (m *mockMaterialRepo) Create(ctx context.Context, material *domain.Material) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.materials[material.ID] = material
	return nil
}

func (m *mockMaterialRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Material, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	material, ok := m.materials[id]
	if !ok {
		return nil, apperrors.NotFound("material", id.String())
	}
	return material, nil
}

func (m *mockMaterialRepo) FindByPodID(ctx context.Context, podID uuid.UUID, limit, offset int) ([]*domain.Material, int, error) {
	var result []*domain.Material
	for _, mat := range m.materials {
		if mat.PodID == podID {
			result = append(result, mat)
		}
	}
	return result, len(result), nil
}

func (m *mockMaterialRepo) FindByUploaderID(ctx context.Context, uploaderID uuid.UUID, limit, offset int) ([]*domain.Material, int, error) {
	var result []*domain.Material
	for _, mat := range m.materials {
		if mat.UploaderID == uploaderID {
			result = append(result, mat)
		}
	}
	return result, len(result), nil
}

func (m *mockMaterialRepo) Update(ctx context.Context, material *domain.Material) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.materials[material.ID] = material
	return nil
}

func (m *mockMaterialRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.materials, id)
	return nil
}

func (m *mockMaterialRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.MaterialStatus) error {
	mat, ok := m.materials[id]
	if ok {
		mat.Status = status
	}
	return nil
}

func (m *mockMaterialRepo) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	mat, ok := m.materials[id]
	if ok {
		mat.ViewCount++
	}
	return nil
}

func (m *mockMaterialRepo) IncrementDownloadCount(ctx context.Context, id uuid.UUID) error {
	mat, ok := m.materials[id]
	if ok {
		mat.DownloadCount++
	}
	return nil
}

func (m *mockMaterialRepo) UpdateRatingStats(ctx context.Context, id uuid.UUID, avgRating float64, ratingCount int) error {
	mat, ok := m.materials[id]
	if ok {
		mat.AverageRating = avgRating
		mat.RatingCount = ratingCount
	}
	return nil
}

func (m *mockMaterialRepo) IncrementVersion(ctx context.Context, id uuid.UUID) error {
	mat, ok := m.materials[id]
	if ok {
		mat.CurrentVersion++
	}
	return nil
}

// Mock Version Repository
type mockVersionRepo struct {
	versions  map[uuid.UUID]*domain.MaterialVersion
	createErr error
}

func newMockVersionRepo() *mockVersionRepo {
	return &mockVersionRepo{
		versions: make(map[uuid.UUID]*domain.MaterialVersion),
	}
}

func (m *mockVersionRepo) Create(ctx context.Context, version *domain.MaterialVersion) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.versions[version.ID] = version
	return nil
}

func (m *mockVersionRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.MaterialVersion, error) {
	version, ok := m.versions[id]
	if !ok {
		return nil, apperrors.NotFound("version", id.String())
	}
	return version, nil
}

func (m *mockVersionRepo) FindByMaterialID(ctx context.Context, materialID uuid.UUID) ([]*domain.MaterialVersion, error) {
	var result []*domain.MaterialVersion
	for _, v := range m.versions {
		if v.MaterialID == materialID {
			result = append(result, v)
		}
	}
	return result, nil
}

func (m *mockVersionRepo) FindByMaterialIDAndVersion(ctx context.Context, materialID uuid.UUID, version int) (*domain.MaterialVersion, error) {
	for _, v := range m.versions {
		if v.MaterialID == materialID && v.Version == version {
			return v, nil
		}
	}
	return nil, apperrors.NotFound("version", materialID.String())
}

func (m *mockVersionRepo) GetLatestVersion(ctx context.Context, materialID uuid.UUID) (int, error) {
	maxVersion := 0
	for _, v := range m.versions {
		if v.MaterialID == materialID && v.Version > maxVersion {
			maxVersion = v.Version
		}
	}
	return maxVersion, nil
}

func (m *mockVersionRepo) DeleteByMaterialID(ctx context.Context, materialID uuid.UUID) error {
	for id, v := range m.versions {
		if v.MaterialID == materialID {
			delete(m.versions, id)
		}
	}
	return nil
}

// Mock Comment Repository
type mockCommentRepo struct {
	comments  map[uuid.UUID]*domain.Comment
	createErr error
	findErr   error
	updateErr error
}

func newMockCommentRepo() *mockCommentRepo {
	return &mockCommentRepo{
		comments: make(map[uuid.UUID]*domain.Comment),
	}
}

func (m *mockCommentRepo) Create(ctx context.Context, comment *domain.Comment) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.comments[comment.ID] = comment
	return nil
}

func (m *mockCommentRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Comment, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	comment, ok := m.comments[id]
	if !ok {
		return nil, apperrors.NotFound("comment", id.String())
	}
	return comment, nil
}

func (m *mockCommentRepo) FindByMaterialID(ctx context.Context, materialID uuid.UUID, limit, offset int) ([]*domain.Comment, int, error) {
	var result []*domain.Comment
	for _, c := range m.comments {
		if c.MaterialID == materialID {
			result = append(result, c)
		}
	}
	return result, len(result), nil
}

func (m *mockCommentRepo) FindByMaterialIDWithUsers(ctx context.Context, materialID uuid.UUID, limit, offset int) ([]*domain.CommentWithUser, int, error) {
	var result []*domain.CommentWithUser
	for _, c := range m.comments {
		if c.MaterialID == materialID {
			result = append(result, &domain.CommentWithUser{
				Comment:  *c,
				UserName: "Test User",
			})
		}
	}
	return result, len(result), nil
}

func (m *mockCommentRepo) FindReplies(ctx context.Context, parentID uuid.UUID, limit, offset int) ([]*domain.CommentWithUser, int, error) {
	return nil, 0, nil
}

func (m *mockCommentRepo) Update(ctx context.Context, comment *domain.Comment) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.comments[comment.ID] = comment
	return nil
}

func (m *mockCommentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	comment, ok := m.comments[id]
	if ok {
		now := time.Now()
		comment.DeletedAt = &now
	}
	return nil
}

func (m *mockCommentRepo) CountByMaterialID(ctx context.Context, materialID uuid.UUID) (int, error) {
	count := 0
	for _, c := range m.comments {
		if c.MaterialID == materialID {
			count++
		}
	}
	return count, nil
}

// Mock Rating Repository
type mockRatingRepo struct {
	ratings   map[uuid.UUID]*domain.Rating
	createErr error
	updateErr error
}

func newMockRatingRepo() *mockRatingRepo {
	return &mockRatingRepo{
		ratings: make(map[uuid.UUID]*domain.Rating),
	}
}

func (m *mockRatingRepo) Create(ctx context.Context, rating *domain.Rating) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.ratings[rating.ID] = rating
	return nil
}

func (m *mockRatingRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Rating, error) {
	rating, ok := m.ratings[id]
	if !ok {
		return nil, apperrors.NotFound("rating", id.String())
	}
	return rating, nil
}

func (m *mockRatingRepo) FindByMaterialAndUser(ctx context.Context, materialID, userID uuid.UUID) (*domain.Rating, error) {
	for _, r := range m.ratings {
		if r.MaterialID == materialID && r.UserID == userID {
			return r, nil
		}
	}
	return nil, apperrors.NotFound("rating", materialID.String())
}

func (m *mockRatingRepo) FindByMaterialID(ctx context.Context, materialID uuid.UUID, limit, offset int) ([]*domain.Rating, int, error) {
	var result []*domain.Rating
	for _, r := range m.ratings {
		if r.MaterialID == materialID {
			result = append(result, r)
		}
	}
	return result, len(result), nil
}

func (m *mockRatingRepo) FindByMaterialIDWithUsers(ctx context.Context, materialID uuid.UUID, limit, offset int) ([]*domain.RatingWithUser, int, error) {
	var result []*domain.RatingWithUser
	for _, r := range m.ratings {
		if r.MaterialID == materialID {
			result = append(result, &domain.RatingWithUser{
				Rating:   *r,
				UserName: "Test User",
			})
		}
	}
	return result, len(result), nil
}

func (m *mockRatingRepo) Update(ctx context.Context, rating *domain.Rating) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.ratings[rating.ID] = rating
	return nil
}

func (m *mockRatingRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.ratings, id)
	return nil
}

func (m *mockRatingRepo) GetSummary(ctx context.Context, materialID uuid.UUID) (*domain.RatingSummary, error) {
	avg, count, _ := m.CalculateAverage(ctx, materialID)
	return &domain.RatingSummary{
		AverageRating: avg,
		RatingCount:   count,
	}, nil
}

func (m *mockRatingRepo) CalculateAverage(ctx context.Context, materialID uuid.UUID) (float64, int, error) {
	var total, count int
	for _, r := range m.ratings {
		if r.MaterialID == materialID {
			total += r.Score
			count++
		}
	}
	if count == 0 {
		return 0, 0, nil
	}
	return float64(total) / float64(count), count, nil
}

// Mock Bookmark Repository
type mockBookmarkRepo struct {
	bookmarks map[uuid.UUID]*domain.Bookmark
	createErr error
}

func newMockBookmarkRepo() *mockBookmarkRepo {
	return &mockBookmarkRepo{
		bookmarks: make(map[uuid.UUID]*domain.Bookmark),
	}
}

func (m *mockBookmarkRepo) Create(ctx context.Context, bookmark *domain.Bookmark) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.bookmarks[bookmark.ID] = bookmark
	return nil
}

func (m *mockBookmarkRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Bookmark, error) {
	bookmark, ok := m.bookmarks[id]
	if !ok {
		return nil, apperrors.NotFound("bookmark", id.String())
	}
	return bookmark, nil
}

func (m *mockBookmarkRepo) FindByUserAndMaterial(ctx context.Context, userID, materialID uuid.UUID) (*domain.Bookmark, error) {
	for _, b := range m.bookmarks {
		if b.UserID == userID && b.MaterialID == materialID {
			return b, nil
		}
	}
	return nil, apperrors.NotFound("bookmark", userID.String())
}

func (m *mockBookmarkRepo) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Bookmark, int, error) {
	var result []*domain.Bookmark
	for _, b := range m.bookmarks {
		if b.UserID == userID {
			result = append(result, b)
		}
	}
	return result, len(result), nil
}

func (m *mockBookmarkRepo) FindByUserIDWithMaterials(ctx context.Context, userID uuid.UUID, folder *string, limit, offset int) ([]*domain.MaterialWithUploader, int, error) {
	return nil, 0, nil
}

func (m *mockBookmarkRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.bookmarks, id)
	return nil
}

func (m *mockBookmarkRepo) DeleteByUserAndMaterial(ctx context.Context, userID, materialID uuid.UUID) error {
	for id, b := range m.bookmarks {
		if b.UserID == userID && b.MaterialID == materialID {
			delete(m.bookmarks, id)
			return nil
		}
	}
	return nil
}

func (m *mockBookmarkRepo) Exists(ctx context.Context, userID, materialID uuid.UUID) (bool, error) {
	for _, b := range m.bookmarks {
		if b.UserID == userID && b.MaterialID == materialID {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockBookmarkRepo) GetFolders(ctx context.Context, userID uuid.UUID) ([]string, error) {
	folders := make(map[string]bool)
	for _, b := range m.bookmarks {
		if b.UserID == userID && b.Folder != nil {
			folders[*b.Folder] = true
		}
	}
	var result []string
	for f := range folders {
		result = append(result, f)
	}
	return result, nil
}

// Mock MinIO Client
type mockMinIOClient struct {
	files         map[string]*FileInfo
	presignedURLs map[string]string
	putURLErr     error
	getURLErr     error
	fileInfoErr   error
}

func newMockMinIOClient() *mockMinIOClient {
	return &mockMinIOClient{
		files:         make(map[string]*FileInfo),
		presignedURLs: make(map[string]string),
	}
}

func (m *mockMinIOClient) GeneratePresignedPutURL(ctx context.Context, objectKey string, contentType string, expiry time.Duration) (string, error) {
	if m.putURLErr != nil {
		return "", m.putURLErr
	}
	url := "https://minio.example.com/materials/" + objectKey + "?presigned=put"
	m.presignedURLs[objectKey] = url
	return url, nil
}

func (m *mockMinIOClient) GeneratePresignedGetURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	if m.getURLErr != nil {
		return "", m.getURLErr
	}
	return "https://minio.example.com/materials/" + objectKey + "?presigned=get", nil
}

func (m *mockMinIOClient) FileExists(ctx context.Context, objectKey string) (bool, error) {
	_, ok := m.files[objectKey]
	return ok, nil
}

func (m *mockMinIOClient) GetFileInfo(ctx context.Context, objectKey string) (*FileInfo, error) {
	if m.fileInfoErr != nil {
		return nil, m.fileInfoErr
	}
	info, ok := m.files[objectKey]
	if !ok {
		return nil, errors.New("file not found")
	}
	return info, nil
}

func (m *mockMinIOClient) DeleteFile(ctx context.Context, objectKey string) error {
	delete(m.files, objectKey)
	return nil
}

// Mock Event Publisher
type mockEventPublisher struct {
	materialUploadedEvents []nats.MaterialUploadedEvent
	materialDeletedEvents  []nats.MaterialDeletedEvent
}

func newMockEventPublisher() *mockEventPublisher {
	return &mockEventPublisher{
		materialUploadedEvents: make([]nats.MaterialUploadedEvent, 0),
		materialDeletedEvents:  make([]nats.MaterialDeletedEvent, 0),
	}
}

func (m *mockEventPublisher) PublishEmailVerification(ctx context.Context, event nats.EmailVerificationEvent) error {
	return nil
}
func (m *mockEventPublisher) PublishEmailPasswordReset(ctx context.Context, event nats.EmailPasswordResetEvent) error {
	return nil
}
func (m *mockEventPublisher) PublishUserCreated(ctx context.Context, event nats.UserCreatedEvent) error {
	return nil
}
func (m *mockEventPublisher) PublishUserUpdated(ctx context.Context, userID uuid.UUID) error {
	return nil
}
func (m *mockEventPublisher) PublishUserFollowed(ctx context.Context, followerID, followingID uuid.UUID) error {
	return nil
}
func (m *mockEventPublisher) PublishMaterialUploaded(ctx context.Context, event nats.MaterialUploadedEvent) error {
	m.materialUploadedEvents = append(m.materialUploadedEvents, event)
	return nil
}
func (m *mockEventPublisher) PublishMaterialDeleted(ctx context.Context, event nats.MaterialDeletedEvent) error {
	m.materialDeletedEvents = append(m.materialDeletedEvents, event)
	return nil
}
func (m *mockEventPublisher) PublishPodCreated(ctx context.Context, event nats.PodCreatedEvent) error {
	return nil
}
func (m *mockEventPublisher) PublishPodUpdated(ctx context.Context, event nats.PodUpdatedEvent) error {
	return nil
}
func (m *mockEventPublisher) PublishCollaboratorInvited(ctx context.Context, event nats.CollaboratorInvitedEvent) error {
	return nil
}
func (m *mockEventPublisher) PublishCommentCreated(ctx context.Context, event nats.CommentCreatedEvent) error {
	return nil
}

// Helper to create a test service
func newTestService() (*Service, *mockMaterialRepo, *mockVersionRepo, *mockCommentRepo, *mockRatingRepo, *mockBookmarkRepo, *mockMinIOClient, *mockEventPublisher) {
	materialRepo := newMockMaterialRepo()
	versionRepo := newMockVersionRepo()
	commentRepo := newMockCommentRepo()
	ratingRepo := newMockRatingRepo()
	bookmarkRepo := newMockBookmarkRepo()
	minioClient := newMockMinIOClient()
	eventPublisher := newMockEventPublisher()

	svc := NewService(
		materialRepo,
		versionRepo,
		commentRepo,
		ratingRepo,
		bookmarkRepo,
		minioClient,
		eventPublisher,
	)

	return svc, materialRepo, versionRepo, commentRepo, ratingRepo, bookmarkRepo, minioClient, eventPublisher
}

// Test: Presigned URL Generation
func TestGetUploadURL_Success(t *testing.T) {
	svc, _, _, _, _, _, _, _ := newTestService()
	ctx := context.Background()

	input := UploadURLInput{
		Filename:    "test-document.pdf",
		ContentType: "application/pdf",
		Size:        1024 * 1024, // 1MB
	}

	result, err := svc.GetUploadURL(ctx, input)
	if err != nil {
		t.Fatalf("GetUploadURL failed: %v", err)
	}

	if result.UploadURL == "" {
		t.Error("Expected upload URL")
	}
	if result.ObjectKey == "" {
		t.Error("Expected object key")
	}
	if result.ExpiresAt.Before(time.Now()) {
		t.Error("Expected expiry time in the future")
	}
}

func TestGetUploadURL_MinIOError(t *testing.T) {
	svc, _, _, _, _, _, minioClient, _ := newTestService()
	ctx := context.Background()

	minioClient.putURLErr = errors.New("minio connection error")

	input := UploadURLInput{
		Filename:    "test-document.pdf",
		ContentType: "application/pdf",
		Size:        1024 * 1024,
	}

	_, err := svc.GetUploadURL(ctx, input)
	if err == nil {
		t.Fatal("Expected error when MinIO fails")
	}
}

// Test: Upload Confirmation Flow
func TestConfirmUpload_Success(t *testing.T) {
	svc, materialRepo, versionRepo, _, _, _, minioClient, eventPublisher := newTestService()
	ctx := context.Background()

	// Simulate file uploaded to MinIO
	objectKey := "2025/01/01/test-uuid_document.pdf"
	minioClient.files[objectKey] = &FileInfo{
		Size:        1024 * 1024,
		ContentType: "application/pdf",
		ETag:        "abc123",
	}

	podID := uuid.New()
	uploaderID := uuid.New()
	input := ConfirmUploadInput{
		ObjectKey:   objectKey,
		PodID:       podID,
		UploaderID:  uploaderID,
		Title:       "Test Document",
		Description: strPtr("A test document"),
	}

	material, err := svc.ConfirmUpload(ctx, input)
	if err != nil {
		t.Fatalf("ConfirmUpload failed: %v", err)
	}

	// Verify material was created
	if material == nil {
		t.Fatal("Expected material in result")
	}
	if material.Title != input.Title {
		t.Errorf("Expected title %s, got %s", input.Title, material.Title)
	}
	if material.PodID != podID {
		t.Errorf("Expected pod ID %s, got %s", podID, material.PodID)
	}
	if material.UploaderID != uploaderID {
		t.Errorf("Expected uploader ID %s, got %s", uploaderID, material.UploaderID)
	}
	if material.FileType != domain.FileTypePDF {
		t.Errorf("Expected file type pdf, got %s", material.FileType)
	}
	if material.Status != domain.MaterialStatusProcessing {
		t.Errorf("Expected status processing, got %s", material.Status)
	}

	// Verify material was stored
	if len(materialRepo.materials) != 1 {
		t.Errorf("Expected 1 material in repo, got %d", len(materialRepo.materials))
	}

	// Verify version was created
	if len(versionRepo.versions) != 1 {
		t.Errorf("Expected 1 version in repo, got %d", len(versionRepo.versions))
	}

	// Verify event was published
	if len(eventPublisher.materialUploadedEvents) != 1 {
		t.Errorf("Expected 1 material.uploaded event, got %d", len(eventPublisher.materialUploadedEvents))
	}
}

func TestConfirmUpload_FileNotFound(t *testing.T) {
	svc, _, _, _, _, _, _, _ := newTestService()
	ctx := context.Background()

	input := ConfirmUploadInput{
		ObjectKey:  "nonexistent.pdf",
		PodID:      uuid.New(),
		UploaderID: uuid.New(),
		Title:      "Test Document",
	}

	_, err := svc.ConfirmUpload(ctx, input)
	if err == nil {
		t.Fatal("Expected error when file not found")
	}
}

func TestConfirmUpload_UnsupportedFileType(t *testing.T) {
	svc, _, _, _, _, _, minioClient, _ := newTestService()
	ctx := context.Background()

	// Simulate file with unsupported extension
	objectKey := "2025/01/01/test-uuid_document.txt"
	minioClient.files[objectKey] = &FileInfo{
		Size:        1024,
		ContentType: "text/plain",
		ETag:        "abc123",
	}

	input := ConfirmUploadInput{
		ObjectKey:  objectKey,
		PodID:      uuid.New(),
		UploaderID: uuid.New(),
		Title:      "Test Document",
	}

	_, err := svc.ConfirmUpload(ctx, input)
	if err == nil {
		t.Fatal("Expected error for unsupported file type")
	}
}

// Helper function
func strPtr(s string) *string {
	return &s
}

// Test: Comment CRUD Operations
func TestAddComment_Success(t *testing.T) {
	svc, _, _, commentRepo, _, _, _, _ := newTestService()
	ctx := context.Background()

	materialID := uuid.New()
	userID := uuid.New()
	input := AddCommentInput{
		MaterialID: materialID,
		UserID:     userID,
		Content:    "This is a great document!",
		ParentID:   nil,
	}

	comment, err := svc.AddComment(ctx, input)
	if err != nil {
		t.Fatalf("AddComment failed: %v", err)
	}

	if comment == nil {
		t.Fatal("Expected comment in result")
	}
	if comment.Content != input.Content {
		t.Errorf("Expected content %s, got %s", input.Content, comment.Content)
	}
	if comment.MaterialID != materialID {
		t.Errorf("Expected material ID %s, got %s", materialID, comment.MaterialID)
	}
	if comment.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, comment.UserID)
	}
	if comment.Edited {
		t.Error("Expected edited to be false for new comment")
	}

	// Verify comment was stored
	if len(commentRepo.comments) != 1 {
		t.Errorf("Expected 1 comment in repo, got %d", len(commentRepo.comments))
	}
}

func TestAddComment_WithParent(t *testing.T) {
	svc, _, _, commentRepo, _, _, _, _ := newTestService()
	ctx := context.Background()

	materialID := uuid.New()
	userID := uuid.New()

	// Create parent comment
	parentComment := domain.NewComment(materialID, userID, "Parent comment", nil)
	commentRepo.comments[parentComment.ID] = parentComment

	// Create reply
	input := AddCommentInput{
		MaterialID: materialID,
		UserID:     uuid.New(),
		Content:    "This is a reply",
		ParentID:   &parentComment.ID,
	}

	comment, err := svc.AddComment(ctx, input)
	if err != nil {
		t.Fatalf("AddComment failed: %v", err)
	}

	if comment.ParentID == nil || *comment.ParentID != parentComment.ID {
		t.Error("Expected parent ID to be set")
	}
}

func TestUpdateComment_Success(t *testing.T) {
	svc, _, _, commentRepo, _, _, _, _ := newTestService()
	ctx := context.Background()

	materialID := uuid.New()
	userID := uuid.New()

	// Create existing comment
	existingComment := domain.NewComment(materialID, userID, "Original content", nil)
	commentRepo.comments[existingComment.ID] = existingComment

	input := UpdateCommentInput{
		ID:      existingComment.ID,
		UserID:  userID,
		Content: "Updated content",
	}

	comment, err := svc.UpdateComment(ctx, input)
	if err != nil {
		t.Fatalf("UpdateComment failed: %v", err)
	}

	if comment.Content != input.Content {
		t.Errorf("Expected content %s, got %s", input.Content, comment.Content)
	}
	if !comment.Edited {
		t.Error("Expected edited to be true after update")
	}
}

func TestUpdateComment_NotOwner(t *testing.T) {
	svc, _, _, commentRepo, _, _, _, _ := newTestService()
	ctx := context.Background()

	materialID := uuid.New()
	ownerID := uuid.New()
	otherUserID := uuid.New()

	// Create existing comment
	existingComment := domain.NewComment(materialID, ownerID, "Original content", nil)
	commentRepo.comments[existingComment.ID] = existingComment

	input := UpdateCommentInput{
		ID:      existingComment.ID,
		UserID:  otherUserID, // Different user
		Content: "Trying to update",
	}

	_, err := svc.UpdateComment(ctx, input)
	if err == nil {
		t.Fatal("Expected error when non-owner updates comment")
	}
}

func TestDeleteComment_Success(t *testing.T) {
	svc, _, _, commentRepo, _, _, _, _ := newTestService()
	ctx := context.Background()

	materialID := uuid.New()
	userID := uuid.New()

	// Create existing comment
	existingComment := domain.NewComment(materialID, userID, "To be deleted", nil)
	commentRepo.comments[existingComment.ID] = existingComment

	err := svc.DeleteComment(ctx, existingComment.ID, userID)
	if err != nil {
		t.Fatalf("DeleteComment failed: %v", err)
	}

	// Verify comment was soft-deleted
	if existingComment.DeletedAt == nil {
		t.Error("Expected comment to be soft-deleted")
	}
}

func TestDeleteComment_NotOwner(t *testing.T) {
	svc, _, _, commentRepo, _, _, _, _ := newTestService()
	ctx := context.Background()

	materialID := uuid.New()
	ownerID := uuid.New()
	otherUserID := uuid.New()

	// Create existing comment
	existingComment := domain.NewComment(materialID, ownerID, "To be deleted", nil)
	commentRepo.comments[existingComment.ID] = existingComment

	err := svc.DeleteComment(ctx, existingComment.ID, otherUserID)
	if err == nil {
		t.Fatal("Expected error when non-owner deletes comment")
	}
}

// Test: Rating Calculation
func TestRateMaterial_NewRating(t *testing.T) {
	svc, materialRepo, _, _, ratingRepo, _, _, _ := newTestService()
	ctx := context.Background()

	// Create material
	material := domain.NewMaterial(uuid.New(), uuid.New(), "Test Material", domain.FileTypePDF, "test.pdf", 1024)
	materialRepo.materials[material.ID] = material

	userID := uuid.New()
	input := RateMaterialInput{
		MaterialID: material.ID,
		UserID:     userID,
		Score:      5,
		Review:     strPtr("Excellent material!"),
	}

	rating, err := svc.RateMaterial(ctx, input)
	if err != nil {
		t.Fatalf("RateMaterial failed: %v", err)
	}

	if rating == nil {
		t.Fatal("Expected rating in result")
	}
	if rating.Score != 5 {
		t.Errorf("Expected score 5, got %d", rating.Score)
	}
	if rating.Review == nil || *rating.Review != "Excellent material!" {
		t.Error("Expected review to be set")
	}

	// Verify rating was stored
	if len(ratingRepo.ratings) != 1 {
		t.Errorf("Expected 1 rating in repo, got %d", len(ratingRepo.ratings))
	}

	// Verify material rating stats were updated
	if material.AverageRating != 5.0 {
		t.Errorf("Expected average rating 5.0, got %f", material.AverageRating)
	}
	if material.RatingCount != 1 {
		t.Errorf("Expected rating count 1, got %d", material.RatingCount)
	}
}

func TestRateMaterial_UpdateExistingRating(t *testing.T) {
	svc, materialRepo, _, _, ratingRepo, _, _, _ := newTestService()
	ctx := context.Background()

	// Create material
	material := domain.NewMaterial(uuid.New(), uuid.New(), "Test Material", domain.FileTypePDF, "test.pdf", 1024)
	materialRepo.materials[material.ID] = material

	userID := uuid.New()

	// Create existing rating
	existingRating := domain.NewRating(material.ID, userID, 3, nil)
	ratingRepo.ratings[existingRating.ID] = existingRating

	// Update rating
	input := RateMaterialInput{
		MaterialID: material.ID,
		UserID:     userID,
		Score:      5,
		Review:     strPtr("Changed my mind, it's excellent!"),
	}

	rating, err := svc.RateMaterial(ctx, input)
	if err != nil {
		t.Fatalf("RateMaterial failed: %v", err)
	}

	if rating.Score != 5 {
		t.Errorf("Expected score 5, got %d", rating.Score)
	}

	// Verify only one rating exists (updated, not created new)
	if len(ratingRepo.ratings) != 1 {
		t.Errorf("Expected 1 rating in repo, got %d", len(ratingRepo.ratings))
	}
}

func TestRateMaterial_InvalidScore(t *testing.T) {
	svc, _, _, _, _, _, _, _ := newTestService()
	ctx := context.Background()

	// Test score too low
	input := RateMaterialInput{
		MaterialID: uuid.New(),
		UserID:     uuid.New(),
		Score:      0,
	}

	_, err := svc.RateMaterial(ctx, input)
	if err == nil {
		t.Fatal("Expected error for score 0")
	}

	// Test score too high
	input.Score = 6
	_, err = svc.RateMaterial(ctx, input)
	if err == nil {
		t.Fatal("Expected error for score 6")
	}
}

func TestRateMaterial_AverageCalculation(t *testing.T) {
	svc, materialRepo, _, _, ratingRepo, _, _, _ := newTestService()
	ctx := context.Background()

	// Create material
	material := domain.NewMaterial(uuid.New(), uuid.New(), "Test Material", domain.FileTypePDF, "test.pdf", 1024)
	materialRepo.materials[material.ID] = material

	// Add multiple ratings
	users := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	scores := []int{5, 4, 3}

	for i, userID := range users {
		input := RateMaterialInput{
			MaterialID: material.ID,
			UserID:     userID,
			Score:      scores[i],
		}
		_, err := svc.RateMaterial(ctx, input)
		if err != nil {
			t.Fatalf("RateMaterial failed: %v", err)
		}
	}

	// Verify ratings count
	if len(ratingRepo.ratings) != 3 {
		t.Errorf("Expected 3 ratings in repo, got %d", len(ratingRepo.ratings))
	}

	// Verify average calculation (5+4+3)/3 = 4.0
	expectedAvg := 4.0
	if material.AverageRating != expectedAvg {
		t.Errorf("Expected average rating %f, got %f", expectedAvg, material.AverageRating)
	}
	if material.RatingCount != 3 {
		t.Errorf("Expected rating count 3, got %d", material.RatingCount)
	}
}
