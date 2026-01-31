// Package application contains unit tests for the Notification Service.
package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/notification/domain"
	"ngasihtau/pkg/email"
)

// Mock implementations for repositories

type mockNotificationRepo struct {
	notifications map[uuid.UUID]*domain.Notification
	createErr     error
	findErr       error
	markReadErr   error
}

func newMockNotificationRepo() *mockNotificationRepo {
	return &mockNotificationRepo{
		notifications: make(map[uuid.UUID]*domain.Notification),
	}
}

func (m *mockNotificationRepo) Create(ctx context.Context, notification *domain.Notification) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.notifications[notification.ID] = notification
	return nil
}

func (m *mockNotificationRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	notification, ok := m.notifications[id]
	if !ok {
		return nil, errors.NotFound("notification", id.String())
	}
	return notification, nil
}

func (m *mockNotificationRepo) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Notification, int, error) {
	var result []*domain.Notification
	for _, n := range m.notifications {
		if n.UserID == userID {
			result = append(result, n)
		}
	}
	// Apply pagination
	total := len(result)
	if offset >= len(result) {
		return []*domain.Notification{}, total, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], total, nil
}

func (m *mockNotificationRepo) FindUnreadByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Notification, int, error) {
	var result []*domain.Notification
	for _, n := range m.notifications {
		if n.UserID == userID && !n.Read {
			result = append(result, n)
		}
	}
	return result, len(result), nil
}

func (m *mockNotificationRepo) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	if m.markReadErr != nil {
		return m.markReadErr
	}
	notification, ok := m.notifications[id]
	if ok {
		notification.Read = true
	}
	return nil
}

func (m *mockNotificationRepo) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	for _, n := range m.notifications {
		if n.UserID == userID {
			n.Read = true
		}
	}
	return nil
}

func (m *mockNotificationRepo) CountUnread(ctx context.Context, userID uuid.UUID) (int, error) {
	count := 0
	for _, n := range m.notifications {
		if n.UserID == userID && !n.Read {
			count++
		}
	}
	return count, nil
}

func (m *mockNotificationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.notifications, id)
	return nil
}

func (m *mockNotificationRepo) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	var count int64
	for id, n := range m.notifications {
		if n.CreatedAt.Before(before) {
			delete(m.notifications, id)
			count++
		}
	}
	return count, nil
}

// Mock Preference Repository
type mockPreferenceRepo struct {
	preferences map[uuid.UUID]*domain.NotificationPreference
	createErr   error
	findErr     error
	upsertErr   error
}

func newMockPreferenceRepo() *mockPreferenceRepo {
	return &mockPreferenceRepo{
		preferences: make(map[uuid.UUID]*domain.NotificationPreference),
	}
}

func (m *mockPreferenceRepo) Create(ctx context.Context, pref *domain.NotificationPreference) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.preferences[pref.UserID] = pref
	return nil
}

func (m *mockPreferenceRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*domain.NotificationPreference, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	pref, ok := m.preferences[userID]
	if !ok {
		return nil, errors.NotFound("preference", userID.String())
	}
	return pref, nil
}

func (m *mockPreferenceRepo) Update(ctx context.Context, pref *domain.NotificationPreference) error {
	m.preferences[pref.UserID] = pref
	return nil
}

func (m *mockPreferenceRepo) Upsert(ctx context.Context, pref *domain.NotificationPreference) error {
	if m.upsertErr != nil {
		return m.upsertErr
	}
	m.preferences[pref.UserID] = pref
	return nil
}

// Mock Email Provider
type mockEmailProvider struct {
	sentEmails []email.Email
	sendErr    error
}

func newMockEmailProvider() *mockEmailProvider {
	return &mockEmailProvider{
		sentEmails: make([]email.Email, 0),
	}
}

func (m *mockEmailProvider) Send(ctx context.Context, e *email.Email) error {
	if m.sendErr != nil {
		return m.sendErr
	}
	m.sentEmails = append(m.sentEmails, *e)
	return nil
}

// Helper to create a test service
func newTestService() (*NotificationService, *mockNotificationRepo, *mockPreferenceRepo) {
	notificationRepo := newMockNotificationRepo()
	preferenceRepo := newMockPreferenceRepo()

	svc := NewNotificationService(notificationRepo, preferenceRepo)

	return svc, notificationRepo, preferenceRepo
}

// Test: Get Notifications
func TestGetNotifications_Success(t *testing.T) {
	svc, notificationRepo, _ := newTestService()
	ctx := context.Background()

	userID := uuid.New()

	// Create some notifications
	for i := 0; i < 5; i++ {
		n := domain.NewNotification(userID, domain.NotificationTypePodInvite, "Test Title", "Test Message", nil)
		notificationRepo.notifications[n.ID] = n
	}

	input := GetNotificationsInput{
		UserID: userID,
		Limit:  10,
		Offset: 0,
	}

	result, err := svc.GetNotifications(ctx, input)
	if err != nil {
		t.Fatalf("GetNotifications failed: %v", err)
	}

	if result.Total != 5 {
		t.Errorf("Expected total 5, got %d", result.Total)
	}
	if len(result.Notifications) != 5 {
		t.Errorf("Expected 5 notifications, got %d", len(result.Notifications))
	}
}

func TestGetNotifications_DefaultLimit(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	input := GetNotificationsInput{
		UserID: uuid.New(),
		Limit:  0, // Should default to 20
		Offset: 0,
	}

	_, err := svc.GetNotifications(ctx, input)
	if err != nil {
		t.Fatalf("GetNotifications failed: %v", err)
	}
}

func TestGetNotifications_MaxLimit(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	input := GetNotificationsInput{
		UserID: uuid.New(),
		Limit:  200, // Should be capped to 100
		Offset: 0,
	}

	_, err := svc.GetNotifications(ctx, input)
	if err != nil {
		t.Fatalf("GetNotifications failed: %v", err)
	}
}

func TestGetNotifications_WithUnreadCount(t *testing.T) {
	svc, notificationRepo, _ := newTestService()
	ctx := context.Background()

	userID := uuid.New()

	// Create 3 unread and 2 read notifications
	for i := 0; i < 3; i++ {
		n := domain.NewNotification(userID, domain.NotificationTypePodInvite, "Unread", "Message", nil)
		notificationRepo.notifications[n.ID] = n
	}
	for i := 0; i < 2; i++ {
		n := domain.NewNotification(userID, domain.NotificationTypePodInvite, "Read", "Message", nil)
		n.Read = true
		notificationRepo.notifications[n.ID] = n
	}

	input := GetNotificationsInput{
		UserID: userID,
		Limit:  10,
		Offset: 0,
	}

	result, err := svc.GetNotifications(ctx, input)
	if err != nil {
		t.Fatalf("GetNotifications failed: %v", err)
	}

	if result.UnreadCount != 3 {
		t.Errorf("Expected unread count 3, got %d", result.UnreadCount)
	}
}

// Test: Mark As Read
func TestMarkAsRead_Success(t *testing.T) {
	svc, notificationRepo, _ := newTestService()
	ctx := context.Background()

	userID := uuid.New()
	notification := domain.NewNotification(userID, domain.NotificationTypePodInvite, "Test", "Message", nil)
	notificationRepo.notifications[notification.ID] = notification

	err := svc.MarkAsRead(ctx, userID, notification.ID)
	if err != nil {
		t.Fatalf("MarkAsRead failed: %v", err)
	}

	if !notification.Read {
		t.Error("Expected notification to be marked as read")
	}
}

func TestMarkAsRead_NotOwner(t *testing.T) {
	svc, notificationRepo, _ := newTestService()
	ctx := context.Background()

	ownerID := uuid.New()
	otherUserID := uuid.New()
	notification := domain.NewNotification(ownerID, domain.NotificationTypePodInvite, "Test", "Message", nil)
	notificationRepo.notifications[notification.ID] = notification

	err := svc.MarkAsRead(ctx, otherUserID, notification.ID)
	if err == nil {
		t.Fatal("Expected error when non-owner marks as read")
	}
	if err != ErrNotificationNotFound {
		t.Errorf("Expected ErrNotificationNotFound, got %v", err)
	}
}

func TestMarkAsRead_NotFound(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	err := svc.MarkAsRead(ctx, uuid.New(), uuid.New())
	if err == nil {
		t.Fatal("Expected error for non-existent notification")
	}
}

// Test: Mark All As Read
func TestMarkAllAsRead_Success(t *testing.T) {
	svc, notificationRepo, _ := newTestService()
	ctx := context.Background()

	userID := uuid.New()

	// Create multiple unread notifications
	for i := 0; i < 5; i++ {
		n := domain.NewNotification(userID, domain.NotificationTypePodInvite, "Test", "Message", nil)
		notificationRepo.notifications[n.ID] = n
	}

	err := svc.MarkAllAsRead(ctx, userID)
	if err != nil {
		t.Fatalf("MarkAllAsRead failed: %v", err)
	}

	// Verify all are marked as read
	for _, n := range notificationRepo.notifications {
		if n.UserID == userID && !n.Read {
			t.Error("Expected all notifications to be marked as read")
		}
	}
}

// Test: Get Preferences
func TestGetPreferences_Existing(t *testing.T) {
	svc, _, preferenceRepo := newTestService()
	ctx := context.Background()

	userID := uuid.New()
	pref := domain.NewDefaultNotificationPreference(userID)
	pref.EmailPodInvite = false
	preferenceRepo.preferences[userID] = pref

	result, err := svc.GetPreferences(ctx, userID)
	if err != nil {
		t.Fatalf("GetPreferences failed: %v", err)
	}

	if result.EmailPodInvite != false {
		t.Error("Expected EmailPodInvite to be false")
	}
}

func TestGetPreferences_Default(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	userID := uuid.New()

	result, err := svc.GetPreferences(ctx, userID)
	if err != nil {
		t.Fatalf("GetPreferences failed: %v", err)
	}

	// Should return default preferences
	if !result.EmailPodInvite {
		t.Error("Expected default EmailPodInvite to be true")
	}
	if !result.InAppPodInvite {
		t.Error("Expected default InAppPodInvite to be true")
	}
}

// Test: Update Preferences
func TestUpdatePreferences_Success(t *testing.T) {
	svc, _, preferenceRepo := newTestService()
	ctx := context.Background()

	userID := uuid.New()
	falseVal := false
	trueVal := true

	input := UpdatePreferencesInput{
		UserID:           userID,
		EmailPodInvite:   &falseVal,
		InAppNewMaterial: &trueVal,
	}

	result, err := svc.UpdatePreferences(ctx, input)
	if err != nil {
		t.Fatalf("UpdatePreferences failed: %v", err)
	}

	if result.EmailPodInvite != false {
		t.Error("Expected EmailPodInvite to be false")
	}
	if result.InAppNewMaterial != true {
		t.Error("Expected InAppNewMaterial to be true")
	}

	// Verify stored
	if len(preferenceRepo.preferences) != 1 {
		t.Errorf("Expected 1 preference in repo, got %d", len(preferenceRepo.preferences))
	}
}

func TestUpdatePreferences_PartialUpdate(t *testing.T) {
	svc, _, preferenceRepo := newTestService()
	ctx := context.Background()

	userID := uuid.New()

	// Create existing preference
	existingPref := domain.NewDefaultNotificationPreference(userID)
	existingPref.EmailPodInvite = true
	existingPref.EmailNewMaterial = true
	preferenceRepo.preferences[userID] = existingPref

	// Update only one field
	falseVal := false
	input := UpdatePreferencesInput{
		UserID:         userID,
		EmailPodInvite: &falseVal,
	}

	result, err := svc.UpdatePreferences(ctx, input)
	if err != nil {
		t.Fatalf("UpdatePreferences failed: %v", err)
	}

	// Updated field
	if result.EmailPodInvite != false {
		t.Error("Expected EmailPodInvite to be false")
	}
	// Unchanged field
	if result.EmailNewMaterial != true {
		t.Error("Expected EmailNewMaterial to remain true")
	}
}

// Test: Create Notification with Preference Filtering
func TestCreateNotification_Success(t *testing.T) {
	svc, notificationRepo, _ := newTestService()
	ctx := context.Background()

	userID := uuid.New()
	notification := domain.NewNotification(userID, domain.NotificationTypePodInvite, "Test", "Message", nil)

	err := svc.CreateNotification(ctx, notification)
	if err != nil {
		t.Fatalf("CreateNotification failed: %v", err)
	}

	if len(notificationRepo.notifications) != 1 {
		t.Errorf("Expected 1 notification in repo, got %d", len(notificationRepo.notifications))
	}
}

func TestCreateNotification_FilteredByPreference_PodInvite(t *testing.T) {
	svc, notificationRepo, preferenceRepo := newTestService()
	ctx := context.Background()

	userID := uuid.New()

	// Set preference to disable pod invite notifications
	pref := domain.NewDefaultNotificationPreference(userID)
	pref.InAppPodInvite = false
	preferenceRepo.preferences[userID] = pref

	notification := domain.NewNotification(userID, domain.NotificationTypePodInvite, "Test", "Message", nil)

	err := svc.CreateNotification(ctx, notification)
	if err != nil {
		t.Fatalf("CreateNotification failed: %v", err)
	}

	// Notification should be skipped
	if len(notificationRepo.notifications) != 0 {
		t.Errorf("Expected 0 notifications (filtered), got %d", len(notificationRepo.notifications))
	}
}

func TestCreateNotification_FilteredByPreference_NewMaterial(t *testing.T) {
	svc, notificationRepo, preferenceRepo := newTestService()
	ctx := context.Background()

	userID := uuid.New()

	// Set preference to disable new material notifications
	pref := domain.NewDefaultNotificationPreference(userID)
	pref.InAppNewMaterial = false
	preferenceRepo.preferences[userID] = pref

	notification := domain.NewNotification(userID, domain.NotificationTypeNewMaterial, "New Material", "Message", nil)

	err := svc.CreateNotification(ctx, notification)
	if err != nil {
		t.Fatalf("CreateNotification failed: %v", err)
	}

	// Notification should be skipped
	if len(notificationRepo.notifications) != 0 {
		t.Errorf("Expected 0 notifications (filtered), got %d", len(notificationRepo.notifications))
	}
}

func TestCreateNotification_FilteredByPreference_CommentReply(t *testing.T) {
	svc, notificationRepo, preferenceRepo := newTestService()
	ctx := context.Background()

	userID := uuid.New()

	// Set preference to disable comment reply notifications
	pref := domain.NewDefaultNotificationPreference(userID)
	pref.InAppCommentReply = false
	preferenceRepo.preferences[userID] = pref

	notification := domain.NewNotification(userID, domain.NotificationTypeCommentReply, "Reply", "Message", nil)

	err := svc.CreateNotification(ctx, notification)
	if err != nil {
		t.Fatalf("CreateNotification failed: %v", err)
	}

	// Notification should be skipped
	if len(notificationRepo.notifications) != 0 {
		t.Errorf("Expected 0 notifications (filtered), got %d", len(notificationRepo.notifications))
	}
}

func TestCreateNotification_NotFilteredWhenEnabled(t *testing.T) {
	svc, notificationRepo, preferenceRepo := newTestService()
	ctx := context.Background()

	userID := uuid.New()

	// Set preference with all enabled
	pref := domain.NewDefaultNotificationPreference(userID)
	preferenceRepo.preferences[userID] = pref

	notification := domain.NewNotification(userID, domain.NotificationTypePodInvite, "Test", "Message", nil)

	err := svc.CreateNotification(ctx, notification)
	if err != nil {
		t.Fatalf("CreateNotification failed: %v", err)
	}

	// Notification should be created
	if len(notificationRepo.notifications) != 1 {
		t.Errorf("Expected 1 notification, got %d", len(notificationRepo.notifications))
	}
}

// Test: Email Worker with Mock Provider
func TestEmailWorker_SendWithRetry_Success(t *testing.T) {
	emailProvider := newMockEmailProvider()
	config := EmailWorkerConfig{
		AppName:        "TestApp",
		AppUrl:         "http://localhost:3000",
		SupportEmail:   "support@test.com",
		MaxRetries:     3,
		RetryBaseDelay: time.Millisecond, // Fast for tests
	}

	worker := &EmailWorker{
		emailProvider:    emailProvider,
		templateRenderer: email.NewTemplateRenderer(),
		config:           config,
	}

	ctx := context.Background()
	emailMsg := &email.Email{
		To:       "test@example.com",
		Subject:  "Test Subject",
		HTMLBody: "<p>Test</p>",
		TextBody: "Test",
	}

	err := worker.sendWithRetry(ctx, emailMsg)
	if err != nil {
		t.Fatalf("sendWithRetry failed: %v", err)
	}

	if len(emailProvider.sentEmails) != 1 {
		t.Errorf("Expected 1 email sent, got %d", len(emailProvider.sentEmails))
	}
}

func TestEmailWorker_SendWithRetry_FailsAfterMaxRetries(t *testing.T) {
	emailProvider := newMockEmailProvider()
	emailProvider.sendErr = errors.Internal("email send failed", nil)

	config := EmailWorkerConfig{
		AppName:        "TestApp",
		AppUrl:         "http://localhost:3000",
		SupportEmail:   "support@test.com",
		MaxRetries:     2,
		RetryBaseDelay: time.Millisecond,
	}

	worker := &EmailWorker{
		emailProvider:    emailProvider,
		templateRenderer: email.NewTemplateRenderer(),
		config:           config,
	}

	ctx := context.Background()
	emailMsg := &email.Email{
		To:       "test@example.com",
		Subject:  "Test Subject",
		HTMLBody: "<p>Test</p>",
		TextBody: "Test",
	}

	err := worker.sendWithRetry(ctx, emailMsg)
	if err == nil {
		t.Fatal("Expected error after max retries")
	}
}

func TestEmailWorker_SendWithRetry_ContextCancelled(t *testing.T) {
	emailProvider := newMockEmailProvider()
	emailProvider.sendErr = errors.Internal("email send failed", nil)

	config := EmailWorkerConfig{
		AppName:        "TestApp",
		AppUrl:         "http://localhost:3000",
		SupportEmail:   "support@test.com",
		MaxRetries:     5,
		RetryBaseDelay: time.Second, // Longer delay
	}

	worker := &EmailWorker{
		emailProvider:    emailProvider,
		templateRenderer: email.NewTemplateRenderer(),
		config:           config,
	}

	ctx, cancel := context.WithCancel(context.Background())
	emailMsg := &email.Email{
		To:       "test@example.com",
		Subject:  "Test Subject",
		HTMLBody: "<p>Test</p>",
		TextBody: "Test",
	}

	// Cancel context immediately
	cancel()

	err := worker.sendWithRetry(ctx, emailMsg)
	if err == nil {
		t.Fatal("Expected error when context is cancelled")
	}
}
