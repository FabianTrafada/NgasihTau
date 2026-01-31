// Package application contains unit tests for the Pod Service.
package application

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/pod/domain"
)

// Mock implementations for repositories

type mockPodRepo struct {
	pods      map[uuid.UUID]*domain.Pod
	slugIndex map[string]*domain.Pod
	createErr error
	findErr   error
	updateErr error
	deleteErr error
}

func newMockPodRepo() *mockPodRepo {
	return &mockPodRepo{
		pods:      make(map[uuid.UUID]*domain.Pod),
		slugIndex: make(map[string]*domain.Pod),
	}
}

func (m *mockPodRepo) Create(ctx context.Context, pod *domain.Pod) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.pods[pod.ID] = pod
	m.slugIndex[pod.Slug] = pod
	return nil
}

func (m *mockPodRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Pod, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	pod, ok := m.pods[id]
	if !ok {
		return nil, errors.NotFound("pod", id.String())
	}
	return pod, nil
}

func (m *mockPodRepo) FindBySlug(ctx context.Context, slug string) (*domain.Pod, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	pod, ok := m.slugIndex[slug]
	if !ok {
		return nil, errors.NotFound("pod", slug)
	}
	return pod, nil
}

func (m *mockPodRepo) FindByOwnerID(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*domain.Pod, int, error) {
	var result []*domain.Pod
	for _, pod := range m.pods {
		if pod.OwnerID == ownerID {
			result = append(result, pod)
		}
	}
	return result, len(result), nil
}

func (m *mockPodRepo) Update(ctx context.Context, pod *domain.Pod) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.pods[pod.ID] = pod
	m.slugIndex[pod.Slug] = pod
	return nil
}

func (m *mockPodRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	pod, ok := m.pods[id]
	if ok {
		delete(m.slugIndex, pod.Slug)
		delete(m.pods, id)
	}
	return nil
}

func (m *mockPodRepo) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	_, ok := m.slugIndex[slug]
	return ok, nil
}

func (m *mockPodRepo) IncrementStarCount(ctx context.Context, id uuid.UUID) error {
	pod, ok := m.pods[id]
	if ok {
		pod.StarCount++
	}
	return nil
}

func (m *mockPodRepo) DecrementStarCount(ctx context.Context, id uuid.UUID) error {
	pod, ok := m.pods[id]
	if ok && pod.StarCount > 0 {
		pod.StarCount--
	}
	return nil
}

func (m *mockPodRepo) IncrementForkCount(ctx context.Context, id uuid.UUID) error {
	pod, ok := m.pods[id]
	if ok {
		pod.ForkCount++
	}
	return nil
}

func (m *mockPodRepo) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	pod, ok := m.pods[id]
	if ok {
		pod.ViewCount++
	}
	return nil
}

func (m *mockPodRepo) IncrementUpvoteCount(ctx context.Context, id uuid.UUID) error {
	pod, ok := m.pods[id]
	if ok {
		pod.UpvoteCount++
	}
	return nil
}

func (m *mockPodRepo) DecrementUpvoteCount(ctx context.Context, id uuid.UUID) error {
	pod, ok := m.pods[id]
	if ok && pod.UpvoteCount > 0 {
		pod.UpvoteCount--
	}
	return nil
}

func (m *mockPodRepo) Search(ctx context.Context, query string, filters domain.PodFilters, limit, offset int) ([]*domain.Pod, int, error) {
	var result []*domain.Pod
	for _, pod := range m.pods {
		result = append(result, pod)
	}
	return result, len(result), nil
}

func (m *mockPodRepo) GetPublicPods(ctx context.Context, limit, offset int) ([]*domain.Pod, int, error) {
	var result []*domain.Pod
	for _, pod := range m.pods {
		if pod.IsPublic() {
			result = append(result, pod)
		}
	}
	return result, len(result), nil
}

type mockCollaboratorRepo struct {
	collaborators map[uuid.UUID]*domain.Collaborator
	podUserIndex  map[string]*domain.Collaborator
	createErr     error
	findErr       error
}

func newMockCollaboratorRepo() *mockCollaboratorRepo {
	return &mockCollaboratorRepo{
		collaborators: make(map[uuid.UUID]*domain.Collaborator),
		podUserIndex:  make(map[string]*domain.Collaborator),
	}
}

func (m *mockCollaboratorRepo) Create(ctx context.Context, collaborator *domain.Collaborator) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.collaborators[collaborator.ID] = collaborator
	key := collaborator.PodID.String() + ":" + collaborator.UserID.String()
	m.podUserIndex[key] = collaborator
	return nil
}

func (m *mockCollaboratorRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Collaborator, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	collab, ok := m.collaborators[id]
	if !ok {
		return nil, errors.NotFound("collaborator", id.String())
	}
	return collab, nil
}

func (m *mockCollaboratorRepo) FindByPodAndUser(ctx context.Context, podID, userID uuid.UUID) (*domain.Collaborator, error) {
	key := podID.String() + ":" + userID.String()
	collab, ok := m.podUserIndex[key]
	if !ok {
		return nil, errors.NotFound("collaborator", key)
	}
	return collab, nil
}

func (m *mockCollaboratorRepo) FindByPodID(ctx context.Context, podID uuid.UUID) ([]*domain.Collaborator, error) {
	var result []*domain.Collaborator
	for _, c := range m.collaborators {
		if c.PodID == podID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (m *mockCollaboratorRepo) FindByPodIDWithUsers(ctx context.Context, podID uuid.UUID) ([]*domain.CollaboratorWithUser, error) {
	return nil, nil
}

func (m *mockCollaboratorRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Collaborator, error) {
	var result []*domain.Collaborator
	for _, c := range m.collaborators {
		if c.UserID == userID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (m *mockCollaboratorRepo) Update(ctx context.Context, collaborator *domain.Collaborator) error {
	m.collaborators[collaborator.ID] = collaborator
	key := collaborator.PodID.String() + ":" + collaborator.UserID.String()
	m.podUserIndex[key] = collaborator
	return nil
}

func (m *mockCollaboratorRepo) Delete(ctx context.Context, id uuid.UUID) error {
	collab, ok := m.collaborators[id]
	if ok {
		key := collab.PodID.String() + ":" + collab.UserID.String()
		delete(m.podUserIndex, key)
		delete(m.collaborators, id)
	}
	return nil
}

func (m *mockCollaboratorRepo) DeleteByPodAndUser(ctx context.Context, podID, userID uuid.UUID) error {
	key := podID.String() + ":" + userID.String()
	collab, ok := m.podUserIndex[key]
	if ok {
		delete(m.collaborators, collab.ID)
		delete(m.podUserIndex, key)
	}
	return nil
}

func (m *mockCollaboratorRepo) Exists(ctx context.Context, podID, userID uuid.UUID) (bool, error) {
	key := podID.String() + ":" + userID.String()
	_, ok := m.podUserIndex[key]
	return ok, nil
}

func (m *mockCollaboratorRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.CollaboratorStatus) error {
	collab, ok := m.collaborators[id]
	if ok {
		collab.Status = status
	}
	return nil
}

func (m *mockCollaboratorRepo) UpdateRole(ctx context.Context, id uuid.UUID, role domain.CollaboratorRole) error {
	collab, ok := m.collaborators[id]
	if ok {
		collab.Role = role
	}
	return nil
}

type mockStarRepo struct {
	stars map[string]bool
}

func newMockStarRepo() *mockStarRepo {
	return &mockStarRepo{
		stars: make(map[string]bool),
	}
}

func (m *mockStarRepo) Create(ctx context.Context, star *domain.PodStar) error {
	key := star.UserID.String() + ":" + star.PodID.String()
	m.stars[key] = true
	return nil
}

func (m *mockStarRepo) Delete(ctx context.Context, userID, podID uuid.UUID) error {
	key := userID.String() + ":" + podID.String()
	delete(m.stars, key)
	return nil
}

func (m *mockStarRepo) Exists(ctx context.Context, userID, podID uuid.UUID) (bool, error) {
	key := userID.String() + ":" + podID.String()
	return m.stars[key], nil
}

func (m *mockStarRepo) GetStarredPods(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Pod, int, error) {
	return nil, 0, nil
}

func (m *mockStarRepo) CountByPodID(ctx context.Context, podID uuid.UUID) (int, error) {
	return 0, nil
}

type mockFollowRepo struct {
	follows map[string]bool
}

func newMockFollowRepo() *mockFollowRepo {
	return &mockFollowRepo{
		follows: make(map[string]bool),
	}
}

func (m *mockFollowRepo) Create(ctx context.Context, follow *domain.PodFollow) error {
	key := follow.UserID.String() + ":" + follow.PodID.String()
	m.follows[key] = true
	return nil
}

func (m *mockFollowRepo) Delete(ctx context.Context, userID, podID uuid.UUID) error {
	key := userID.String() + ":" + podID.String()
	delete(m.follows, key)
	return nil
}

func (m *mockFollowRepo) Exists(ctx context.Context, userID, podID uuid.UUID) (bool, error) {
	key := userID.String() + ":" + podID.String()
	return m.follows[key], nil
}

func (m *mockFollowRepo) GetFollowedPods(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Pod, int, error) {
	return nil, 0, nil
}

func (m *mockFollowRepo) GetFollowerIDs(ctx context.Context, podID uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

func (m *mockFollowRepo) CountByPodID(ctx context.Context, podID uuid.UUID) (int, error) {
	return 0, nil
}

type mockUpvoteRepo struct {
	upvotes map[string]bool
}

func newMockUpvoteRepo() *mockUpvoteRepo {
	return &mockUpvoteRepo{
		upvotes: make(map[string]bool),
	}
}

func (m *mockUpvoteRepo) Create(ctx context.Context, upvote *domain.PodUpvote) error {
	key := upvote.UserID.String() + ":" + upvote.PodID.String()
	m.upvotes[key] = true
	return nil
}

func (m *mockUpvoteRepo) Delete(ctx context.Context, userID, podID uuid.UUID) error {
	key := userID.String() + ":" + podID.String()
	delete(m.upvotes, key)
	return nil
}

func (m *mockUpvoteRepo) Exists(ctx context.Context, userID, podID uuid.UUID) (bool, error) {
	key := userID.String() + ":" + podID.String()
	return m.upvotes[key], nil
}

func (m *mockUpvoteRepo) CountByPodID(ctx context.Context, podID uuid.UUID) (int, error) {
	count := 0
	for key := range m.upvotes {
		if key[37:] == podID.String() {
			count++
		}
	}
	return count, nil
}

func (m *mockUpvoteRepo) GetUpvotedPods(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Pod, int, error) {
	return nil, 0, nil
}

type mockDownvoteRepo struct {
	downvotes map[string]bool
}

func newMockDownvoteRepo() *mockDownvoteRepo {
	return &mockDownvoteRepo{
		downvotes: make(map[string]bool),
	}
}

func (m *mockDownvoteRepo) Create(ctx context.Context, downvote *domain.PodDownvote) error {
	key := downvote.UserID.String() + ":" + downvote.PodID.String()
	m.downvotes[key] = true
	return nil
}

func (m *mockDownvoteRepo) Delete(ctx context.Context, userID, podID uuid.UUID) error {
	key := userID.String() + ":" + podID.String()
	delete(m.downvotes, key)
	return nil
}

func (m *mockDownvoteRepo) Exists(ctx context.Context, userID, podID uuid.UUID) (bool, error) {
	key := userID.String() + ":" + podID.String()
	return m.downvotes[key], nil
}

func (m *mockDownvoteRepo) CountByPodID(ctx context.Context, podID uuid.UUID) (int, error) {
	count := 0
	for key := range m.downvotes {
		if key[37:] == podID.String() {
			count++
		}
	}
	return count, nil
}

func (m *mockDownvoteRepo) GetDownvotedPods(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Pod, int, error) {
	return nil, 0, nil
}

type mockUploadRequestRepo struct {
	requests    map[uuid.UUID]*domain.UploadRequest
	reqPodIndex map[string]*domain.UploadRequest // requesterID:podID -> request
}

func newMockUploadRequestRepo() *mockUploadRequestRepo {
	return &mockUploadRequestRepo{
		requests:    make(map[uuid.UUID]*domain.UploadRequest),
		reqPodIndex: make(map[string]*domain.UploadRequest),
	}
}

func (m *mockUploadRequestRepo) Create(ctx context.Context, request *domain.UploadRequest) error {
	m.requests[request.ID] = request
	key := request.RequesterID.String() + ":" + request.PodID.String()
	m.reqPodIndex[key] = request
	return nil
}

func (m *mockUploadRequestRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.UploadRequest, error) {
	req, ok := m.requests[id]
	if !ok {
		return nil, errors.NotFound("upload request", id.String())
	}
	return req, nil
}

func (m *mockUploadRequestRepo) FindByRequesterAndPod(ctx context.Context, requesterID, podID uuid.UUID) (*domain.UploadRequest, error) {
	key := requesterID.String() + ":" + podID.String()
	req, ok := m.reqPodIndex[key]
	if !ok {
		return nil, errors.NotFound("upload request", key)
	}
	return req, nil
}

func (m *mockUploadRequestRepo) FindByPodOwner(ctx context.Context, ownerID uuid.UUID, status *domain.UploadRequestStatus, limit, offset int) ([]*domain.UploadRequest, int, error) {
	var result []*domain.UploadRequest
	for _, req := range m.requests {
		if req.PodOwnerID == ownerID {
			if status == nil || req.Status == *status {
				result = append(result, req)
			}
		}
	}
	return result, len(result), nil
}

func (m *mockUploadRequestRepo) FindByRequester(ctx context.Context, requesterID uuid.UUID, limit, offset int) ([]*domain.UploadRequest, int, error) {
	var result []*domain.UploadRequest
	for _, req := range m.requests {
		if req.RequesterID == requesterID {
			result = append(result, req)
		}
	}
	return result, len(result), nil
}

func (m *mockUploadRequestRepo) FindApprovedByRequesterAndPod(ctx context.Context, requesterID, podID uuid.UUID) (*domain.UploadRequest, error) {
	key := requesterID.String() + ":" + podID.String()
	req, ok := m.reqPodIndex[key]
	if !ok || req.Status != domain.UploadRequestStatusApproved {
		return nil, errors.NotFound("upload request", key)
	}
	return req, nil
}

func (m *mockUploadRequestRepo) Update(ctx context.Context, request *domain.UploadRequest) error {
	m.requests[request.ID] = request
	key := request.RequesterID.String() + ":" + request.PodID.String()
	m.reqPodIndex[key] = request
	return nil
}

func (m *mockUploadRequestRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.UploadRequestStatus, reason *string) error {
	req, ok := m.requests[id]
	if !ok {
		return errors.NotFound("upload request", id.String())
	}
	req.Status = status
	req.RejectionReason = reason
	return nil
}

type mockActivityRepo struct {
	activities map[uuid.UUID]*domain.Activity
}

func newMockActivityRepo() *mockActivityRepo {
	return &mockActivityRepo{
		activities: make(map[uuid.UUID]*domain.Activity),
	}
}

func (m *mockActivityRepo) Create(ctx context.Context, activity *domain.Activity) error {
	m.activities[activity.ID] = activity
	return nil
}

func (m *mockActivityRepo) FindByPodID(ctx context.Context, podID uuid.UUID, limit, offset int) ([]*domain.Activity, int, error) {
	var result []*domain.Activity
	for _, a := range m.activities {
		if a.PodID == podID {
			result = append(result, a)
		}
	}
	return result, len(result), nil
}

func (m *mockActivityRepo) FindByPodIDWithDetails(ctx context.Context, podID uuid.UUID, limit, offset int) ([]*domain.ActivityWithDetails, int, error) {
	return nil, 0, nil
}

func (m *mockActivityRepo) GetUserFeed(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.ActivityWithDetails, int, error) {
	return nil, 0, nil
}

func (m *mockActivityRepo) DeleteByPodID(ctx context.Context, podID uuid.UUID) error {
	for id, a := range m.activities {
		if a.PodID == podID {
			delete(m.activities, id)
		}
	}
	return nil
}

// mockSharedPodRepo implements domain.SharedPodRepository for testing
type mockSharedPodRepo struct {
	sharedPods    map[uuid.UUID]*domain.SharedPod
	podStudentIdx map[string]*domain.SharedPod // podID:studentID -> sharedPod
	studentShares map[uuid.UUID][]*domain.SharedPod
}

func newMockSharedPodRepo() *mockSharedPodRepo {
	return &mockSharedPodRepo{
		sharedPods:    make(map[uuid.UUID]*domain.SharedPod),
		podStudentIdx: make(map[string]*domain.SharedPod),
		studentShares: make(map[uuid.UUID][]*domain.SharedPod),
	}
}

func (m *mockSharedPodRepo) Create(ctx context.Context, share *domain.SharedPod) error {
	m.sharedPods[share.ID] = share
	key := share.PodID.String() + ":" + share.StudentID.String()
	m.podStudentIdx[key] = share
	m.studentShares[share.StudentID] = append(m.studentShares[share.StudentID], share)
	return nil
}

func (m *mockSharedPodRepo) Delete(ctx context.Context, id uuid.UUID) error {
	share, ok := m.sharedPods[id]
	if !ok {
		return errors.NotFound("shared pod", id.String())
	}
	key := share.PodID.String() + ":" + share.StudentID.String()
	delete(m.podStudentIdx, key)
	delete(m.sharedPods, id)
	// Remove from studentShares
	shares := m.studentShares[share.StudentID]
	for i, s := range shares {
		if s.ID == id {
			m.studentShares[share.StudentID] = append(shares[:i], shares[i+1:]...)
			break
		}
	}
	return nil
}

func (m *mockSharedPodRepo) FindByStudent(ctx context.Context, studentID uuid.UUID, limit, offset int) ([]*domain.SharedPod, int, error) {
	shares := m.studentShares[studentID]
	total := len(shares)
	if offset >= total {
		return []*domain.SharedPod{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return shares[offset:end], total, nil
}

func (m *mockSharedPodRepo) FindByStudentWithDetails(ctx context.Context, studentID uuid.UUID, limit, offset int) ([]*domain.SharedPodWithDetails, int, error) {
	shares := m.studentShares[studentID]
	total := len(shares)
	if offset >= total {
		return []*domain.SharedPodWithDetails{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	var result []*domain.SharedPodWithDetails
	for _, share := range shares[offset:end] {
		result = append(result, &domain.SharedPodWithDetails{
			SharedPod:     *share,
			PodName:       "Test Pod",
			PodSlug:       "test-pod",
			TeacherName:   "Test Teacher",
			TeacherAvatar: nil,
		})
	}
	return result, total, nil
}

func (m *mockSharedPodRepo) FindByTeacherAndStudent(ctx context.Context, teacherID, studentID uuid.UUID) ([]*domain.SharedPod, error) {
	var result []*domain.SharedPod
	for _, share := range m.sharedPods {
		if share.TeacherID == teacherID && share.StudentID == studentID {
			result = append(result, share)
		}
	}
	return result, nil
}

func (m *mockSharedPodRepo) Exists(ctx context.Context, podID, studentID uuid.UUID) (bool, error) {
	key := podID.String() + ":" + studentID.String()
	_, ok := m.podStudentIdx[key]
	return ok, nil
}

// Helper to create a test service
func newTestService() (PodService, *mockPodRepo, *mockCollaboratorRepo, *mockActivityRepo) {
	podRepo := newMockPodRepo()
	collaboratorRepo := newMockCollaboratorRepo()
	starRepo := newMockStarRepo()
	upvoteRepo := newMockUpvoteRepo()
	downvoteRepo := newMockDownvoteRepo()
	uploadReqRepo := newMockUploadRequestRepo()
	sharedPodRepo := newMockSharedPodRepo()
	followRepo := newMockFollowRepo()
	activityRepo := newMockActivityRepo()
	eventPublisher := NewNoOpEventPublisher()

	svc := NewPodService(
		podRepo,
		collaboratorRepo,
		starRepo,
		upvoteRepo,
		downvoteRepo,
		uploadReqRepo,
		sharedPodRepo,
		followRepo,
		activityRepo,
		eventPublisher,
		nil, // UserRoleChecker - nil for basic tests
	)

	return svc, podRepo, collaboratorRepo, activityRepo
}

// Test: Pod Creation
func TestCreatePod_Success(t *testing.T) {
	svc, podRepo, _, _ := newTestService()
	ctx := context.Background()

	ownerID := uuid.New()
	input := CreatePodInput{
		OwnerID:     ownerID,
		Name:        "My Learning Pod",
		Description: strPtr("A pod for learning"),
		Visibility:  domain.VisibilityPublic,
		Categories:  []string{"math", "science"},
		Tags:        []string{"beginner", "tutorial"},
	}

	pod, err := svc.CreatePod(ctx, input)
	if err != nil {
		t.Fatalf("CreatePod failed: %v", err)
	}

	// Verify pod was created
	if pod == nil {
		t.Fatal("Expected pod in result")
	}
	if pod.Name != input.Name {
		t.Errorf("Expected name %s, got %s", input.Name, pod.Name)
	}
	if pod.OwnerID != ownerID {
		t.Errorf("Expected owner ID %s, got %s", ownerID, pod.OwnerID)
	}
	if pod.Visibility != domain.VisibilityPublic {
		t.Errorf("Expected visibility public, got %s", pod.Visibility)
	}
	if pod.Slug == "" {
		t.Error("Expected slug to be generated")
	}
	if pod.Slug != "my-learning-pod" {
		t.Errorf("Expected slug 'my-learning-pod', got %s", pod.Slug)
	}

	// Verify pod was stored
	if len(podRepo.pods) != 1 {
		t.Errorf("Expected 1 pod in repo, got %d", len(podRepo.pods))
	}
}

func TestCreatePod_UniqueSlugGeneration(t *testing.T) {
	svc, podRepo, _, _ := newTestService()
	ctx := context.Background()

	ownerID := uuid.New()

	// Create first pod
	input1 := CreatePodInput{
		OwnerID:    ownerID,
		Name:       "Test Pod",
		Visibility: domain.VisibilityPublic,
	}
	pod1, err := svc.CreatePod(ctx, input1)
	if err != nil {
		t.Fatalf("CreatePod 1 failed: %v", err)
	}

	// Create second pod with same name
	input2 := CreatePodInput{
		OwnerID:    ownerID,
		Name:       "Test Pod",
		Visibility: domain.VisibilityPublic,
	}
	pod2, err := svc.CreatePod(ctx, input2)
	if err != nil {
		t.Fatalf("CreatePod 2 failed: %v", err)
	}

	// Verify slugs are different
	if pod1.Slug == pod2.Slug {
		t.Errorf("Expected different slugs, both got %s", pod1.Slug)
	}

	// Verify both pods were stored
	if len(podRepo.pods) != 2 {
		t.Errorf("Expected 2 pods in repo, got %d", len(podRepo.pods))
	}
}

func TestCreatePod_PrivateVisibility(t *testing.T) {
	svc, _, _, _ := newTestService()
	ctx := context.Background()

	ownerID := uuid.New()
	input := CreatePodInput{
		OwnerID:    ownerID,
		Name:       "Private Pod",
		Visibility: domain.VisibilityPrivate,
	}

	pod, err := svc.CreatePod(ctx, input)
	if err != nil {
		t.Fatalf("CreatePod failed: %v", err)
	}

	if pod.Visibility != domain.VisibilityPrivate {
		t.Errorf("Expected visibility private, got %s", pod.Visibility)
	}
	if pod.IsPublic() {
		t.Error("Expected IsPublic() to return false")
	}
}

// Test: Collaborator Invitation Flow
func TestInviteCollaborator_Success(t *testing.T) {
	svc, podRepo, collaboratorRepo, activityRepo := newTestService()
	ctx := context.Background()

	// Create a pod
	ownerID := uuid.New()
	pod := domain.NewPod(ownerID, "Test Pod", "test-pod", domain.VisibilityPublic, false)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Invite a collaborator
	inviteeID := uuid.New()
	input := InviteCollaboratorInput{
		PodID:     pod.ID,
		InviterID: ownerID,
		UserID:    inviteeID,
		Role:      domain.CollaboratorRoleContributor,
	}

	collab, err := svc.InviteCollaborator(ctx, input)
	if err != nil {
		t.Fatalf("InviteCollaborator failed: %v", err)
	}

	// Verify collaborator was created
	if collab == nil {
		t.Fatal("Expected collaborator in result")
	}
	if collab.PodID != pod.ID {
		t.Errorf("Expected pod ID %s, got %s", pod.ID, collab.PodID)
	}
	if collab.UserID != inviteeID {
		t.Errorf("Expected user ID %s, got %s", inviteeID, collab.UserID)
	}
	if collab.Role != domain.CollaboratorRoleContributor {
		t.Errorf("Expected role contributor, got %s", collab.Role)
	}
	if collab.Status != domain.CollaboratorStatusPending {
		t.Errorf("Expected status pending, got %s", collab.Status)
	}

	// Verify collaborator was stored
	if len(collaboratorRepo.collaborators) != 1 {
		t.Errorf("Expected 1 collaborator in repo, got %d", len(collaboratorRepo.collaborators))
	}

	// Verify activity was logged
	if len(activityRepo.activities) != 1 {
		t.Errorf("Expected 1 activity in repo, got %d", len(activityRepo.activities))
	}
}

func TestInviteCollaborator_NotOwner(t *testing.T) {
	svc, podRepo, _, _ := newTestService()
	ctx := context.Background()

	// Create a pod
	ownerID := uuid.New()
	pod := domain.NewPod(ownerID, "Test Pod", "test-pod", domain.VisibilityPublic, false)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Try to invite as non-owner
	nonOwnerID := uuid.New()
	inviteeID := uuid.New()
	input := InviteCollaboratorInput{
		PodID:     pod.ID,
		InviterID: nonOwnerID,
		UserID:    inviteeID,
		Role:      domain.CollaboratorRoleContributor,
	}

	_, err := svc.InviteCollaborator(ctx, input)
	if err == nil {
		t.Fatal("Expected error when non-owner invites")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeForbidden {
		t.Errorf("Expected forbidden error code, got %s", appErr.Code)
	}
}

func TestInviteCollaborator_AlreadyCollaborator(t *testing.T) {
	svc, podRepo, collaboratorRepo, _ := newTestService()
	ctx := context.Background()

	// Create a pod
	ownerID := uuid.New()
	pod := domain.NewPod(ownerID, "Test Pod", "test-pod", domain.VisibilityPublic, false)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Add existing collaborator
	existingCollabID := uuid.New()
	existingCollab := domain.NewCollaborator(pod.ID, existingCollabID, ownerID, domain.CollaboratorRoleViewer)
	collaboratorRepo.collaborators[existingCollab.ID] = existingCollab
	key := pod.ID.String() + ":" + existingCollabID.String()
	collaboratorRepo.podUserIndex[key] = existingCollab

	// Try to invite same user again
	input := InviteCollaboratorInput{
		PodID:     pod.ID,
		InviterID: ownerID,
		UserID:    existingCollabID,
		Role:      domain.CollaboratorRoleContributor,
	}

	_, err := svc.InviteCollaborator(ctx, input)
	if err == nil {
		t.Fatal("Expected error when inviting existing collaborator")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeConflict {
		t.Errorf("Expected conflict error code, got %s", appErr.Code)
	}
}

func TestInviteCollaborator_CannotInviteOwner(t *testing.T) {
	svc, podRepo, _, _ := newTestService()
	ctx := context.Background()

	// Create a pod
	ownerID := uuid.New()
	pod := domain.NewPod(ownerID, "Test Pod", "test-pod", domain.VisibilityPublic, false)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Try to invite the owner
	input := InviteCollaboratorInput{
		PodID:     pod.ID,
		InviterID: ownerID,
		UserID:    ownerID, // Same as owner
		Role:      domain.CollaboratorRoleContributor,
	}

	_, err := svc.InviteCollaborator(ctx, input)
	if err == nil {
		t.Fatal("Expected error when inviting owner")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeBadRequest {
		t.Errorf("Expected bad request error code, got %s", appErr.Code)
	}
}

func TestVerifyCollaborator_Success(t *testing.T) {
	svc, podRepo, collaboratorRepo, _ := newTestService()
	ctx := context.Background()

	// Create a pod
	ownerID := uuid.New()
	pod := domain.NewPod(ownerID, "Test Pod", "test-pod", domain.VisibilityPublic, false)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Add pending collaborator
	collabUserID := uuid.New()
	collab := domain.NewCollaborator(pod.ID, collabUserID, ownerID, domain.CollaboratorRoleContributor)
	collab.Status = domain.CollaboratorStatusPendingVerification
	collaboratorRepo.collaborators[collab.ID] = collab
	key := pod.ID.String() + ":" + collabUserID.String()
	collaboratorRepo.podUserIndex[key] = collab

	// Verify collaborator
	err := svc.VerifyCollaborator(ctx, pod.ID, collab.ID, ownerID)
	if err != nil {
		t.Fatalf("VerifyCollaborator failed: %v", err)
	}

	// Check status was updated
	if collab.Status != domain.CollaboratorStatusVerified {
		t.Errorf("Expected status verified, got %s", collab.Status)
	}
}

func TestVerifyCollaborator_NotOwner(t *testing.T) {
	svc, podRepo, collaboratorRepo, _ := newTestService()
	ctx := context.Background()

	// Create a pod
	ownerID := uuid.New()
	pod := domain.NewPod(ownerID, "Test Pod", "test-pod", domain.VisibilityPublic, false)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Add pending collaborator
	collabUserID := uuid.New()
	collab := domain.NewCollaborator(pod.ID, collabUserID, ownerID, domain.CollaboratorRoleContributor)
	collab.Status = domain.CollaboratorStatusPendingVerification
	collaboratorRepo.collaborators[collab.ID] = collab
	key := pod.ID.String() + ":" + collabUserID.String()
	collaboratorRepo.podUserIndex[key] = collab

	// Try to verify as non-owner
	nonOwnerID := uuid.New()
	err := svc.VerifyCollaborator(ctx, pod.ID, collab.ID, nonOwnerID)
	if err == nil {
		t.Fatal("Expected error when non-owner verifies")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeForbidden {
		t.Errorf("Expected forbidden error code, got %s", appErr.Code)
	}
}

// Helper function
func strPtr(s string) *string {
	return &s
}

// Test: Fork Functionality
func TestForkPod_Success(t *testing.T) {
	svc, podRepo, _, activityRepo := newTestService()
	ctx := context.Background()

	// Create original pod
	originalOwnerID := uuid.New()
	originalPod := domain.NewPod(originalOwnerID, "Original Pod", "original-pod", domain.VisibilityPublic, false)
	desc := "Original description"
	originalPod.Description = &desc
	originalPod.Categories = []string{"math"}
	originalPod.Tags = []string{"tutorial"}
	podRepo.pods[originalPod.ID] = originalPod
	podRepo.slugIndex[originalPod.Slug] = originalPod

	// Fork the pod
	forkerID := uuid.New()
	forkedPod, err := svc.ForkPod(ctx, originalPod.ID, forkerID)
	if err != nil {
		t.Fatalf("ForkPod failed: %v", err)
	}

	// Verify forked pod
	if forkedPod == nil {
		t.Fatal("Expected forked pod in result")
	}
	if forkedPod.OwnerID != forkerID {
		t.Errorf("Expected owner ID %s, got %s", forkerID, forkedPod.OwnerID)
	}
	if forkedPod.Name != originalPod.Name {
		t.Errorf("Expected name %s, got %s", originalPod.Name, forkedPod.Name)
	}
	if forkedPod.ForkedFromID == nil || *forkedPod.ForkedFromID != originalPod.ID {
		t.Error("Expected forked_from_id to reference original pod")
	}
	if forkedPod.Description == nil || *forkedPod.Description != *originalPod.Description {
		t.Error("Expected description to be copied")
	}
	if len(forkedPod.Categories) != len(originalPod.Categories) {
		t.Error("Expected categories to be copied")
	}
	if len(forkedPod.Tags) != len(originalPod.Tags) {
		t.Error("Expected tags to be copied")
	}

	// Verify original pod fork count was incremented
	if originalPod.ForkCount != 1 {
		t.Errorf("Expected fork count 1, got %d", originalPod.ForkCount)
	}

	// Verify activity was logged
	if len(activityRepo.activities) != 1 {
		t.Errorf("Expected 1 activity in repo, got %d", len(activityRepo.activities))
	}

	// Verify both pods exist
	if len(podRepo.pods) != 2 {
		t.Errorf("Expected 2 pods in repo, got %d", len(podRepo.pods))
	}
}

func TestForkPod_PrivatePodNoAccess(t *testing.T) {
	svc, podRepo, _, _ := newTestService()
	ctx := context.Background()

	// Create private pod
	originalOwnerID := uuid.New()
	originalPod := domain.NewPod(originalOwnerID, "Private Pod", "private-pod", domain.VisibilityPrivate, false)
	podRepo.pods[originalPod.ID] = originalPod
	podRepo.slugIndex[originalPod.Slug] = originalPod

	// Try to fork as non-collaborator
	forkerID := uuid.New()
	_, err := svc.ForkPod(ctx, originalPod.ID, forkerID)
	if err == nil {
		t.Fatal("Expected error when forking private pod without access")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeForbidden {
		t.Errorf("Expected forbidden error code, got %s", appErr.Code)
	}
}

func TestForkPod_PrivatePodWithAccess(t *testing.T) {
	svc, podRepo, collaboratorRepo, _ := newTestService()
	ctx := context.Background()

	// Create private pod
	originalOwnerID := uuid.New()
	originalPod := domain.NewPod(originalOwnerID, "Private Pod", "private-pod", domain.VisibilityPrivate, false)
	podRepo.pods[originalPod.ID] = originalPod
	podRepo.slugIndex[originalPod.Slug] = originalPod

	// Add collaborator
	forkerID := uuid.New()
	collab := domain.NewCollaborator(originalPod.ID, forkerID, originalOwnerID, domain.CollaboratorRoleViewer)
	collaboratorRepo.collaborators[collab.ID] = collab
	key := originalPod.ID.String() + ":" + forkerID.String()
	collaboratorRepo.podUserIndex[key] = collab

	// Fork as collaborator
	forkedPod, err := svc.ForkPod(ctx, originalPod.ID, forkerID)
	if err != nil {
		t.Fatalf("ForkPod failed: %v", err)
	}

	if forkedPod == nil {
		t.Fatal("Expected forked pod in result")
	}
	if forkedPod.OwnerID != forkerID {
		t.Errorf("Expected owner ID %s, got %s", forkerID, forkedPod.OwnerID)
	}
}

func TestForkPod_PodNotFound(t *testing.T) {
	svc, _, _, _ := newTestService()
	ctx := context.Background()

	// Try to fork non-existent pod
	forkerID := uuid.New()
	nonExistentPodID := uuid.New()
	_, err := svc.ForkPod(ctx, nonExistentPodID, forkerID)
	if err == nil {
		t.Fatal("Expected error when forking non-existent pod")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeNotFound {
		t.Errorf("Expected not found error code, got %s", appErr.Code)
	}
}

// Test: Permission Checks
func TestCanUserAccessPod_PublicPod(t *testing.T) {
	svc, podRepo, _, _ := newTestService()
	ctx := context.Background()

	// Create public pod
	ownerID := uuid.New()
	pod := domain.NewPod(ownerID, "Public Pod", "public-pod", domain.VisibilityPublic, false)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Anonymous user can access
	canAccess, err := svc.CanUserAccessPod(ctx, pod.ID, nil)
	if err != nil {
		t.Fatalf("CanUserAccessPod failed: %v", err)
	}
	if !canAccess {
		t.Error("Expected anonymous user to access public pod")
	}

	// Any authenticated user can access
	randomUserID := uuid.New()
	canAccess, err = svc.CanUserAccessPod(ctx, pod.ID, &randomUserID)
	if err != nil {
		t.Fatalf("CanUserAccessPod failed: %v", err)
	}
	if !canAccess {
		t.Error("Expected authenticated user to access public pod")
	}
}

func TestCanUserAccessPod_PrivatePod(t *testing.T) {
	svc, podRepo, collaboratorRepo, _ := newTestService()
	ctx := context.Background()

	// Create private pod
	ownerID := uuid.New()
	pod := domain.NewPod(ownerID, "Private Pod", "private-pod", domain.VisibilityPrivate, false)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Anonymous user cannot access
	canAccess, err := svc.CanUserAccessPod(ctx, pod.ID, nil)
	if err != nil {
		t.Fatalf("CanUserAccessPod failed: %v", err)
	}
	if canAccess {
		t.Error("Expected anonymous user to NOT access private pod")
	}

	// Random user cannot access
	randomUserID := uuid.New()
	canAccess, err = svc.CanUserAccessPod(ctx, pod.ID, &randomUserID)
	if err != nil {
		t.Fatalf("CanUserAccessPod failed: %v", err)
	}
	if canAccess {
		t.Error("Expected random user to NOT access private pod")
	}

	// Owner can access
	canAccess, err = svc.CanUserAccessPod(ctx, pod.ID, &ownerID)
	if err != nil {
		t.Fatalf("CanUserAccessPod failed: %v", err)
	}
	if !canAccess {
		t.Error("Expected owner to access private pod")
	}

	// Collaborator can access
	collabUserID := uuid.New()
	collab := domain.NewCollaborator(pod.ID, collabUserID, ownerID, domain.CollaboratorRoleViewer)
	collaboratorRepo.collaborators[collab.ID] = collab
	key := pod.ID.String() + ":" + collabUserID.String()
	collaboratorRepo.podUserIndex[key] = collab

	canAccess, err = svc.CanUserAccessPod(ctx, pod.ID, &collabUserID)
	if err != nil {
		t.Fatalf("CanUserAccessPod failed: %v", err)
	}
	if !canAccess {
		t.Error("Expected collaborator to access private pod")
	}
}

func TestCanUserUploadToPod_VerifiedContributor(t *testing.T) {
	svc, podRepo, collaboratorRepo, _ := newTestService()
	ctx := context.Background()

	// Create pod
	ownerID := uuid.New()
	pod := domain.NewPod(ownerID, "Test Pod", "test-pod", domain.VisibilityPublic, false)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Add verified contributor
	contributorID := uuid.New()
	collab := domain.NewCollaborator(pod.ID, contributorID, ownerID, domain.CollaboratorRoleContributor)
	collab.Status = domain.CollaboratorStatusVerified
	collaboratorRepo.collaborators[collab.ID] = collab
	key := pod.ID.String() + ":" + contributorID.String()
	collaboratorRepo.podUserIndex[key] = collab

	// Verified contributor can upload
	canUpload, err := svc.CanUserUploadToPod(ctx, pod.ID, contributorID)
	if err != nil {
		t.Fatalf("CanUserUploadToPod failed: %v", err)
	}
	if !canUpload {
		t.Error("Expected verified contributor to upload")
	}
}

func TestCanUserUploadToPod_UnverifiedContributor(t *testing.T) {
	svc, podRepo, collaboratorRepo, _ := newTestService()
	ctx := context.Background()

	// Create pod
	ownerID := uuid.New()
	pod := domain.NewPod(ownerID, "Test Pod", "test-pod", domain.VisibilityPublic, false)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Add unverified contributor
	contributorID := uuid.New()
	collab := domain.NewCollaborator(pod.ID, contributorID, ownerID, domain.CollaboratorRoleContributor)
	collab.Status = domain.CollaboratorStatusPendingVerification // Not verified
	collaboratorRepo.collaborators[collab.ID] = collab
	key := pod.ID.String() + ":" + contributorID.String()
	collaboratorRepo.podUserIndex[key] = collab

	// Unverified contributor cannot upload
	canUpload, err := svc.CanUserUploadToPod(ctx, pod.ID, contributorID)
	if err != nil {
		t.Fatalf("CanUserUploadToPod failed: %v", err)
	}
	if canUpload {
		t.Error("Expected unverified contributor to NOT upload")
	}
}

// Test: CanUserUploadToPod with approved upload request
// Implements requirement 4.5: WHILE an upload request is approved, THE Material Service SHALL allow the requesting teacher to upload.

func TestCanUserUploadToPod_ApprovedUploadRequest(t *testing.T) {
	roleChecker := newMockUserRoleChecker()
	svc, podRepo, _, uploadReqRepo := newTestServiceWithRoleChecker(roleChecker)
	ctx := context.Background()

	// Create a pod owned by a teacher
	ownerID := uuid.New()
	roleChecker.teacherIDs[ownerID] = true
	pod := domain.NewPod(ownerID, "Teacher Pod", "teacher-pod", domain.VisibilityPublic, true)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Create a teacher who will request upload permission
	requesterID := uuid.New()
	roleChecker.teacherIDs[requesterID] = true

	// Create an approved upload request
	uploadReq := domain.NewUploadRequest(requesterID, pod.ID, ownerID, nil)
	uploadReq.Status = domain.UploadRequestStatusApproved
	uploadReqRepo.requests[uploadReq.ID] = uploadReq
	key := requesterID.String() + ":" + pod.ID.String()
	uploadReqRepo.reqPodIndex[key] = uploadReq

	// Teacher with approved upload request can upload
	canUpload, err := svc.CanUserUploadToPod(ctx, pod.ID, requesterID)
	if err != nil {
		t.Fatalf("CanUserUploadToPod failed: %v", err)
	}
	if !canUpload {
		t.Error("Expected teacher with approved upload request to be able to upload")
	}
}

func TestCanUserUploadToPod_PendingUploadRequest(t *testing.T) {
	roleChecker := newMockUserRoleChecker()
	svc, podRepo, _, uploadReqRepo := newTestServiceWithRoleChecker(roleChecker)
	ctx := context.Background()

	// Create a pod owned by a teacher
	ownerID := uuid.New()
	roleChecker.teacherIDs[ownerID] = true
	pod := domain.NewPod(ownerID, "Teacher Pod", "teacher-pod", domain.VisibilityPublic, true)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Create a teacher who will request upload permission
	requesterID := uuid.New()
	roleChecker.teacherIDs[requesterID] = true

	// Create a pending upload request (not approved yet)
	uploadReq := domain.NewUploadRequest(requesterID, pod.ID, ownerID, nil)
	uploadReq.Status = domain.UploadRequestStatusPending
	uploadReqRepo.requests[uploadReq.ID] = uploadReq
	key := requesterID.String() + ":" + pod.ID.String()
	uploadReqRepo.reqPodIndex[key] = uploadReq

	// Teacher with pending upload request cannot upload
	canUpload, err := svc.CanUserUploadToPod(ctx, pod.ID, requesterID)
	if err != nil {
		t.Fatalf("CanUserUploadToPod failed: %v", err)
	}
	if canUpload {
		t.Error("Expected teacher with pending upload request to NOT be able to upload")
	}
}

func TestCanUserUploadToPod_RevokedUploadRequest(t *testing.T) {
	roleChecker := newMockUserRoleChecker()
	svc, podRepo, _, uploadReqRepo := newTestServiceWithRoleChecker(roleChecker)
	ctx := context.Background()

	// Create a pod owned by a teacher
	ownerID := uuid.New()
	roleChecker.teacherIDs[ownerID] = true
	pod := domain.NewPod(ownerID, "Teacher Pod", "teacher-pod", domain.VisibilityPublic, true)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Create a teacher who had upload permission
	requesterID := uuid.New()
	roleChecker.teacherIDs[requesterID] = true

	// Create a revoked upload request
	uploadReq := domain.NewUploadRequest(requesterID, pod.ID, ownerID, nil)
	uploadReq.Status = domain.UploadRequestStatusRevoked
	uploadReqRepo.requests[uploadReq.ID] = uploadReq
	key := requesterID.String() + ":" + pod.ID.String()
	uploadReqRepo.reqPodIndex[key] = uploadReq

	// Teacher with revoked upload request cannot upload
	canUpload, err := svc.CanUserUploadToPod(ctx, pod.ID, requesterID)
	if err != nil {
		t.Fatalf("CanUserUploadToPod failed: %v", err)
	}
	if canUpload {
		t.Error("Expected teacher with revoked upload request to NOT be able to upload")
	}
}

// Test: Upvote Functionality
// Implements requirements 5.1, 5.3

func TestUpvotePod_Success(t *testing.T) {
	svc, podRepo, _, _ := newTestService()
	ctx := context.Background()

	// Create a public pod
	ownerID := uuid.New()
	pod := domain.NewPod(ownerID, "Test Pod", "test-pod", domain.VisibilityPublic, false)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Upvote the pod
	userID := uuid.New()
	err := svc.UpvotePod(ctx, pod.ID, userID)
	if err != nil {
		t.Fatalf("UpvotePod failed: %v", err)
	}

	// Verify upvote count was incremented (requirement 5.1)
	if pod.UpvoteCount != 1 {
		t.Errorf("Expected upvote count 1, got %d", pod.UpvoteCount)
	}

	// Verify user has upvoted
	hasUpvoted, err := svc.HasUpvoted(ctx, pod.ID, userID)
	if err != nil {
		t.Fatalf("HasUpvoted failed: %v", err)
	}
	if !hasUpvoted {
		t.Error("Expected HasUpvoted to return true")
	}
}

func TestUpvotePod_DuplicateUpvote(t *testing.T) {
	svc, podRepo, _, _ := newTestService()
	ctx := context.Background()

	// Create a public pod
	ownerID := uuid.New()
	pod := domain.NewPod(ownerID, "Test Pod", "test-pod", domain.VisibilityPublic, false)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// First upvote
	userID := uuid.New()
	err := svc.UpvotePod(ctx, pod.ID, userID)
	if err != nil {
		t.Fatalf("First UpvotePod failed: %v", err)
	}

	// Try to upvote again (requirement 5.3: each user can upvote only once)
	err = svc.UpvotePod(ctx, pod.ID, userID)
	if err == nil {
		t.Fatal("Expected error when upvoting twice")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeConflict {
		t.Errorf("Expected conflict error code, got %s", appErr.Code)
	}

	// Verify upvote count is still 1
	if pod.UpvoteCount != 1 {
		t.Errorf("Expected upvote count 1, got %d", pod.UpvoteCount)
	}
}

func TestUpvotePod_PrivatePodNoAccess(t *testing.T) {
	svc, podRepo, _, _ := newTestService()
	ctx := context.Background()

	// Create a private pod
	ownerID := uuid.New()
	pod := domain.NewPod(ownerID, "Private Pod", "private-pod", domain.VisibilityPrivate, false)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Try to upvote as non-collaborator
	userID := uuid.New()
	err := svc.UpvotePod(ctx, pod.ID, userID)
	if err == nil {
		t.Fatal("Expected error when upvoting private pod without access")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeForbidden {
		t.Errorf("Expected forbidden error code, got %s", appErr.Code)
	}
}

func TestUpvotePod_PodNotFound(t *testing.T) {
	svc, _, _, _ := newTestService()
	ctx := context.Background()

	// Try to upvote non-existent pod
	userID := uuid.New()
	nonExistentPodID := uuid.New()
	err := svc.UpvotePod(ctx, nonExistentPodID, userID)
	if err == nil {
		t.Fatal("Expected error when upvoting non-existent pod")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeNotFound {
		t.Errorf("Expected not found error code, got %s", appErr.Code)
	}
}

func TestUpvotePod_OwnerCanUpvote(t *testing.T) {
	svc, podRepo, _, _ := newTestService()
	ctx := context.Background()

	// Create a public pod
	ownerID := uuid.New()
	pod := domain.NewPod(ownerID, "Test Pod", "test-pod", domain.VisibilityPublic, false)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Owner can upvote their own pod
	err := svc.UpvotePod(ctx, pod.ID, ownerID)
	if err != nil {
		t.Fatalf("UpvotePod by owner failed: %v", err)
	}

	// Verify upvote count was incremented
	if pod.UpvoteCount != 1 {
		t.Errorf("Expected upvote count 1, got %d", pod.UpvoteCount)
	}
}

// Test: RemoveUpvote Functionality
// Implements requirement 5.2: WHEN a user removes their upvote

func TestRemoveUpvote_Success(t *testing.T) {
	svc, podRepo, _, _ := newTestService()
	ctx := context.Background()

	// Create a public pod
	ownerID := uuid.New()
	pod := domain.NewPod(ownerID, "Test Pod", "test-pod", domain.VisibilityPublic, false)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// First upvote the pod
	userID := uuid.New()
	err := svc.UpvotePod(ctx, pod.ID, userID)
	if err != nil {
		t.Fatalf("UpvotePod failed: %v", err)
	}

	// Verify upvote count is 1
	if pod.UpvoteCount != 1 {
		t.Errorf("Expected upvote count 1, got %d", pod.UpvoteCount)
	}

	// Remove the upvote (requirement 5.2)
	err = svc.RemoveUpvote(ctx, pod.ID, userID)
	if err != nil {
		t.Fatalf("RemoveUpvote failed: %v", err)
	}

	// Verify upvote count was decremented (requirement 5.2)
	if pod.UpvoteCount != 0 {
		t.Errorf("Expected upvote count 0, got %d", pod.UpvoteCount)
	}

	// Verify user no longer has upvoted
	hasUpvoted, err := svc.HasUpvoted(ctx, pod.ID, userID)
	if err != nil {
		t.Fatalf("HasUpvoted failed: %v", err)
	}
	if hasUpvoted {
		t.Error("Expected HasUpvoted to return false after removing upvote")
	}
}

func TestRemoveUpvote_NotUpvoted(t *testing.T) {
	svc, podRepo, _, _ := newTestService()
	ctx := context.Background()

	// Create a public pod
	ownerID := uuid.New()
	pod := domain.NewPod(ownerID, "Test Pod", "test-pod", domain.VisibilityPublic, false)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Try to remove upvote without having upvoted
	userID := uuid.New()
	err := svc.RemoveUpvote(ctx, pod.ID, userID)
	if err == nil {
		t.Fatal("Expected error when removing upvote that doesn't exist")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeNotFound {
		t.Errorf("Expected not found error code, got %s", appErr.Code)
	}
}

func TestRemoveUpvote_PodNotFound(t *testing.T) {
	svc, _, _, _ := newTestService()
	ctx := context.Background()

	// Try to remove upvote from non-existent pod
	userID := uuid.New()
	nonExistentPodID := uuid.New()
	err := svc.RemoveUpvote(ctx, nonExistentPodID, userID)
	if err == nil {
		t.Fatal("Expected error when removing upvote from non-existent pod")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	// The error could be NotFound for the upvote check
	if appErr.Code != errors.CodeNotFound {
		t.Errorf("Expected not found error code, got %s", appErr.Code)
	}
}

// Mock UserRoleChecker for testing teacher role validation
type mockUserRoleChecker struct {
	teacherIDs map[uuid.UUID]bool
}

func newMockUserRoleChecker() *mockUserRoleChecker {
	return &mockUserRoleChecker{
		teacherIDs: make(map[uuid.UUID]bool),
	}
}

func (m *mockUserRoleChecker) IsTeacher(ctx context.Context, userID uuid.UUID) (bool, error) {
	return m.teacherIDs[userID], nil
}

// Helper to create a test service with UserRoleChecker
func newTestServiceWithRoleChecker(roleChecker UserRoleChecker) (PodService, *mockPodRepo, *mockCollaboratorRepo, *mockUploadRequestRepo) {
	podRepo := newMockPodRepo()
	collaboratorRepo := newMockCollaboratorRepo()
	starRepo := newMockStarRepo()
	upvoteRepo := newMockUpvoteRepo()
	downvoteRepo := newMockDownvoteRepo()
	uploadReqRepo := newMockUploadRequestRepo()
	sharedPodRepo := newMockSharedPodRepo()
	followRepo := newMockFollowRepo()
	activityRepo := newMockActivityRepo()
	eventPublisher := NewNoOpEventPublisher()

	svc := NewPodService(
		podRepo,
		collaboratorRepo,
		starRepo,
		upvoteRepo,
		downvoteRepo,
		uploadReqRepo,
		sharedPodRepo,
		followRepo,
		activityRepo,
		eventPublisher,
		roleChecker,
	)

	return svc, podRepo, collaboratorRepo, uploadReqRepo
}

// Test: CreateUploadRequest with teacher role validation
// Implements requirement 4.1: WHEN a teacher submits upload request to another teacher's pod

func TestCreateUploadRequest_Success(t *testing.T) {
	roleChecker := newMockUserRoleChecker()
	svc, podRepo, _, uploadReqRepo := newTestServiceWithRoleChecker(roleChecker)
	ctx := context.Background()

	// Create a pod owned by a teacher
	ownerID := uuid.New()
	roleChecker.teacherIDs[ownerID] = true
	pod := domain.NewPod(ownerID, "Teacher Pod", "teacher-pod", domain.VisibilityPublic, true)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Create upload request from another teacher
	requesterID := uuid.New()
	roleChecker.teacherIDs[requesterID] = true
	message := "I would like to contribute to your pod"

	request, err := svc.CreateUploadRequest(ctx, requesterID, pod.ID, &message)
	if err != nil {
		t.Fatalf("CreateUploadRequest failed: %v", err)
	}

	// Verify request was created
	if request == nil {
		t.Fatal("Expected upload request in result")
	}
	if request.RequesterID != requesterID {
		t.Errorf("Expected requester ID %s, got %s", requesterID, request.RequesterID)
	}
	if request.PodID != pod.ID {
		t.Errorf("Expected pod ID %s, got %s", pod.ID, request.PodID)
	}
	if request.PodOwnerID != ownerID {
		t.Errorf("Expected pod owner ID %s, got %s", ownerID, request.PodOwnerID)
	}
	if request.Status != domain.UploadRequestStatusPending {
		t.Errorf("Expected status pending, got %s", request.Status)
	}
	if request.Message == nil || *request.Message != message {
		t.Error("Expected message to be set")
	}

	// Verify request was stored
	if len(uploadReqRepo.requests) != 1 {
		t.Errorf("Expected 1 upload request in repo, got %d", len(uploadReqRepo.requests))
	}
}

func TestCreateUploadRequest_RequesterNotTeacher(t *testing.T) {
	roleChecker := newMockUserRoleChecker()
	svc, podRepo, _, _ := newTestServiceWithRoleChecker(roleChecker)
	ctx := context.Background()

	// Create a pod owned by a teacher
	ownerID := uuid.New()
	roleChecker.teacherIDs[ownerID] = true
	pod := domain.NewPod(ownerID, "Teacher Pod", "teacher-pod", domain.VisibilityPublic, true)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Try to create upload request from a student (not a teacher)
	requesterID := uuid.New()
	// requesterID is NOT in teacherIDs, so they are a student

	_, err := svc.CreateUploadRequest(ctx, requesterID, pod.ID, nil)
	if err == nil {
		t.Fatal("Expected error when student creates upload request")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeForbidden {
		t.Errorf("Expected forbidden error code, got %s", appErr.Code)
	}
}

func TestCreateUploadRequest_PodOwnerNotTeacher(t *testing.T) {
	roleChecker := newMockUserRoleChecker()
	svc, podRepo, _, _ := newTestServiceWithRoleChecker(roleChecker)
	ctx := context.Background()

	// Create a pod owned by a student (not a teacher)
	ownerID := uuid.New()
	// ownerID is NOT in teacherIDs, so they are a student
	pod := domain.NewPod(ownerID, "Student Pod", "student-pod", domain.VisibilityPublic, false)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Try to create upload request from a teacher to a student's pod
	requesterID := uuid.New()
	roleChecker.teacherIDs[requesterID] = true

	_, err := svc.CreateUploadRequest(ctx, requesterID, pod.ID, nil)
	if err == nil {
		t.Fatal("Expected error when requesting upload to student's pod")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeBadRequest {
		t.Errorf("Expected bad request error code, got %s", appErr.Code)
	}
}

func TestCreateUploadRequest_CannotRequestOwnPod(t *testing.T) {
	roleChecker := newMockUserRoleChecker()
	svc, podRepo, _, _ := newTestServiceWithRoleChecker(roleChecker)
	ctx := context.Background()

	// Create a pod owned by a teacher
	ownerID := uuid.New()
	roleChecker.teacherIDs[ownerID] = true
	pod := domain.NewPod(ownerID, "Teacher Pod", "teacher-pod", domain.VisibilityPublic, true)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Try to create upload request to own pod
	_, err := svc.CreateUploadRequest(ctx, ownerID, pod.ID, nil)
	if err == nil {
		t.Fatal("Expected error when requesting upload to own pod")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeBadRequest {
		t.Errorf("Expected bad request error code, got %s", appErr.Code)
	}
}

func TestCreateUploadRequest_DuplicatePendingRequest(t *testing.T) {
	roleChecker := newMockUserRoleChecker()
	svc, podRepo, _, uploadReqRepo := newTestServiceWithRoleChecker(roleChecker)
	ctx := context.Background()

	// Create a pod owned by a teacher
	ownerID := uuid.New()
	roleChecker.teacherIDs[ownerID] = true
	pod := domain.NewPod(ownerID, "Teacher Pod", "teacher-pod", domain.VisibilityPublic, true)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Create first upload request
	requesterID := uuid.New()
	roleChecker.teacherIDs[requesterID] = true

	_, err := svc.CreateUploadRequest(ctx, requesterID, pod.ID, nil)
	if err != nil {
		t.Fatalf("First CreateUploadRequest failed: %v", err)
	}

	// Try to create another request (should fail because pending request exists)
	_, err = svc.CreateUploadRequest(ctx, requesterID, pod.ID, nil)
	if err == nil {
		t.Fatal("Expected error when creating duplicate pending request")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeConflict {
		t.Errorf("Expected conflict error code, got %s", appErr.Code)
	}

	// Verify only one request exists
	if len(uploadReqRepo.requests) != 1 {
		t.Errorf("Expected 1 upload request in repo, got %d", len(uploadReqRepo.requests))
	}
}

func TestCreateUploadRequest_WithoutRoleChecker(t *testing.T) {
	// Test that CreateUploadRequest works without role checker (nil)
	svc, podRepo, _, uploadReqRepo := newTestServiceWithRoleChecker(nil)
	ctx := context.Background()

	// Create a pod
	ownerID := uuid.New()
	pod := domain.NewPod(ownerID, "Test Pod", "test-pod", domain.VisibilityPublic, false)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Create upload request (should work without role validation)
	requesterID := uuid.New()
	request, err := svc.CreateUploadRequest(ctx, requesterID, pod.ID, nil)
	if err != nil {
		t.Fatalf("CreateUploadRequest failed: %v", err)
	}

	// Verify request was created
	if request == nil {
		t.Fatal("Expected upload request in result")
	}
	if len(uploadReqRepo.requests) != 1 {
		t.Errorf("Expected 1 upload request in repo, got %d", len(uploadReqRepo.requests))
	}
}
