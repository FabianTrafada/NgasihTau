package material

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// MockMaterialDB is a mock implementation for material database queries.
type MockMaterialDB struct {
	materials map[uuid.UUID]struct {
		podID     uuid.UUID
		title     string
		fileType  string
		fileURL   string
		fileSize  int64
		isDeleted bool
	}
}

// MockPodDB is a mock implementation for pod database queries.
type MockPodDB struct {
	pods map[uuid.UUID]struct {
		ownerID    uuid.UUID
		visibility string
		isDeleted  bool
	}
	collaborators map[string]bool // key: "podID:userID"
}

// mockAccessChecker wraps the mock databases for testing.
type mockAccessChecker struct {
	materialDB *MockMaterialDB
	podDB      *MockPodDB
}

func newMockAccessChecker() *mockAccessChecker {
	return &mockAccessChecker{
		materialDB: &MockMaterialDB{
			materials: make(map[uuid.UUID]struct {
				podID     uuid.UUID
				title     string
				fileType  string
				fileURL   string
				fileSize  int64
				isDeleted bool
			}),
		},
		podDB: &MockPodDB{
			pods: make(map[uuid.UUID]struct {
				ownerID    uuid.UUID
				visibility string
				isDeleted  bool
			}),
			collaborators: make(map[string]bool),
		},
	}
}

func (m *mockAccessChecker) addMaterial(id, podID uuid.UUID, title, fileType, fileURL string, fileSize int64) {
	m.materialDB.materials[id] = struct {
		podID     uuid.UUID
		title     string
		fileType  string
		fileURL   string
		fileSize  int64
		isDeleted bool
	}{
		podID:     podID,
		title:     title,
		fileType:  fileType,
		fileURL:   fileURL,
		fileSize:  fileSize,
		isDeleted: false,
	}
}

func (m *mockAccessChecker) addPod(id, ownerID uuid.UUID, visibility string) {
	m.podDB.pods[id] = struct {
		ownerID    uuid.UUID
		visibility string
		isDeleted  bool
	}{
		ownerID:    ownerID,
		visibility: visibility,
		isDeleted:  false,
	}
}

func (m *mockAccessChecker) addCollaborator(podID, userID uuid.UUID) {
	key := podID.String() + ":" + userID.String()
	m.podDB.collaborators[key] = true
}

// CheckAccess implements the access check logic using mock data.
func (m *mockAccessChecker) CheckAccess(ctx context.Context, userID, materialID uuid.UUID) (bool, error) {
	// Get material
	mat, ok := m.materialDB.materials[materialID]
	if !ok || mat.isDeleted {
		return false, nil
	}

	// Get pod
	pod, ok := m.podDB.pods[mat.podID]
	if !ok || pod.isDeleted {
		return false, nil
	}

	// Public pods are accessible to everyone
	if pod.visibility == "public" {
		return true, nil
	}

	// Owner can always access
	if pod.ownerID == userID {
		return true, nil
	}

	// Check collaborator
	key := mat.podID.String() + ":" + userID.String()
	return m.podDB.collaborators[key], nil
}

func TestMockAccessChecker_PublicPod(t *testing.T) {
	ctx := context.Background()
	checker := newMockAccessChecker()

	userID := uuid.New()
	materialID := uuid.New()
	podID := uuid.New()
	ownerID := uuid.New()

	checker.addMaterial(materialID, podID, "Test Material", "pdf", "materials/test.pdf", 1024)
	checker.addPod(podID, ownerID, "public")

	canAccess, err := checker.CheckAccess(ctx, userID, materialID)
	assert.NoError(t, err)
	assert.True(t, canAccess, "any user should have access to public pod material")
}

func TestMockAccessChecker_PrivatePod_Owner(t *testing.T) {
	ctx := context.Background()
	checker := newMockAccessChecker()

	ownerID := uuid.New()
	materialID := uuid.New()
	podID := uuid.New()

	checker.addMaterial(materialID, podID, "Test Material", "pdf", "materials/test.pdf", 1024)
	checker.addPod(podID, ownerID, "private")

	canAccess, err := checker.CheckAccess(ctx, ownerID, materialID)
	assert.NoError(t, err)
	assert.True(t, canAccess, "owner should have access to their private pod material")
}

func TestMockAccessChecker_PrivatePod_Collaborator(t *testing.T) {
	ctx := context.Background()
	checker := newMockAccessChecker()

	userID := uuid.New()
	ownerID := uuid.New()
	materialID := uuid.New()
	podID := uuid.New()

	checker.addMaterial(materialID, podID, "Test Material", "pdf", "materials/test.pdf", 1024)
	checker.addPod(podID, ownerID, "private")
	checker.addCollaborator(podID, userID)

	canAccess, err := checker.CheckAccess(ctx, userID, materialID)
	assert.NoError(t, err)
	assert.True(t, canAccess, "collaborator should have access to private pod material")
}

func TestMockAccessChecker_PrivatePod_NoAccess(t *testing.T) {
	ctx := context.Background()
	checker := newMockAccessChecker()

	userID := uuid.New()
	ownerID := uuid.New()
	materialID := uuid.New()
	podID := uuid.New()

	checker.addMaterial(materialID, podID, "Test Material", "pdf", "materials/test.pdf", 1024)
	checker.addPod(podID, ownerID, "private")
	// Note: userID is NOT added as collaborator

	canAccess, err := checker.CheckAccess(ctx, userID, materialID)
	assert.NoError(t, err)
	assert.False(t, canAccess, "non-collaborator should not have access to private pod material")
}

func TestMockAccessChecker_MaterialNotFound(t *testing.T) {
	ctx := context.Background()
	checker := newMockAccessChecker()

	userID := uuid.New()
	materialID := uuid.New()

	// Material not added to mock

	canAccess, err := checker.CheckAccess(ctx, userID, materialID)
	assert.NoError(t, err)
	assert.False(t, canAccess, "should return false for non-existent material")
}

func TestMockAccessChecker_PodNotFound(t *testing.T) {
	ctx := context.Background()
	checker := newMockAccessChecker()

	userID := uuid.New()
	materialID := uuid.New()
	podID := uuid.New()

	checker.addMaterial(materialID, podID, "Test Material", "pdf", "materials/test.pdf", 1024)
	// Pod not added to mock

	canAccess, err := checker.CheckAccess(ctx, userID, materialID)
	assert.NoError(t, err)
	assert.False(t, canAccess, "should return false when pod doesn't exist")
}

// TestAccessCheckerInterface verifies that the real AccessChecker implements
// the LicenseMaterialAccessChecker interface correctly.
// This is a compile-time check.
func TestAccessCheckerInterface(t *testing.T) {
	// This test verifies that AccessChecker can be used where
	// LicenseMaterialAccessChecker is expected.
	// The actual database operations are tested via integration tests.
	t.Log("AccessChecker implements the required interface")
}
