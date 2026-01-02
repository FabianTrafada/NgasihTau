// Package application contains unit tests for the User Service.
package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/user/domain"
	"ngasihtau/pkg/hash"
	"ngasihtau/pkg/jwt"
)

// Mock implementations for repositories

type mockUserRepo struct {
	users         map[uuid.UUID]*domain.User
	emailIndex    map[string]*domain.User
	createErr     error
	findErr       error
	updateErr     error
	enable2FAErr  error
	disable2FAErr error
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:      make(map[uuid.UUID]*domain.User),
		emailIndex: make(map[string]*domain.User),
	}
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.users[user.ID] = user
	m.emailIndex[user.Email] = user
	return nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	user, ok := m.users[id]
	if !ok {
		return nil, errors.NotFound("user", id.String())
	}
	return user, nil
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	user, ok := m.emailIndex[email]
	if !ok {
		return nil, errors.NotFound("user", email)
	}
	return user, nil
}

func (m *mockUserRepo) Update(ctx context.Context, user *domain.User) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.users[user.ID] = user
	m.emailIndex[user.Email] = user
	return nil
}

func (m *mockUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.users, id)
	return nil
}

func (m *mockUserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	_, ok := m.emailIndex[email]
	return ok, nil
}

func (m *mockUserRepo) GetProfile(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, errors.NotFound("user", id.String())
	}
	return &domain.UserProfile{
		ID:   user.ID,
		Name: user.Name,
		Role: user.Role,
	}, nil
}

func (m *mockUserRepo) UpdateProfile(ctx context.Context, id uuid.UUID, name string, bio *string, avatarURL *string) error {
	return nil
}

func (m *mockUserRepo) Enable2FA(ctx context.Context, id uuid.UUID, secret string) error {
	if m.enable2FAErr != nil {
		return m.enable2FAErr
	}
	user, ok := m.users[id]
	if !ok {
		return errors.NotFound("user", id.String())
	}
	user.TwoFactorEnabled = true
	user.TwoFactorSecret = &secret
	return nil
}

func (m *mockUserRepo) Disable2FA(ctx context.Context, id uuid.UUID) error {
	if m.disable2FAErr != nil {
		return m.disable2FAErr
	}
	user, ok := m.users[id]
	if !ok {
		return errors.NotFound("user", id.String())
	}
	user.TwoFactorEnabled = false
	user.TwoFactorSecret = nil
	return nil
}

func (m *mockUserRepo) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	user, ok := m.users[id]
	if !ok {
		return errors.NotFound("user", id.String())
	}
	user.PasswordHash = passwordHash
	return nil
}

func (m *mockUserRepo) VerifyEmail(ctx context.Context, id uuid.UUID) error {
	user, ok := m.users[id]
	if !ok {
		return errors.NotFound("user", id.String())
	}
	user.EmailVerified = true
	return nil
}

func (m *mockUserRepo) SetOnboardingCompleted(ctx context.Context, id uuid.UUID, completed bool) error {
	user, ok := m.users[id]
	if !ok {
		return errors.NotFound("user", id.String())
	}
	user.OnboardingCompleted = completed
	return nil
}

type mockOAuthRepo struct{}

func (m *mockOAuthRepo) Create(ctx context.Context, account *domain.OAuthAccount) error { return nil }
func (m *mockOAuthRepo) FindByProvider(ctx context.Context, provider domain.OAuthProvider, providerUserID string) (*domain.OAuthAccount, error) {
	return nil, errors.NotFound("oauth_account", providerUserID)
}
func (m *mockOAuthRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.OAuthAccount, error) {
	return nil, nil
}
func (m *mockOAuthRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }

type mockRefreshTokenRepo struct {
	tokens    map[uuid.UUID]*domain.RefreshToken
	hashIndex map[string]*domain.RefreshToken
}

func newMockRefreshTokenRepo() *mockRefreshTokenRepo {
	return &mockRefreshTokenRepo{
		tokens:    make(map[uuid.UUID]*domain.RefreshToken),
		hashIndex: make(map[string]*domain.RefreshToken),
	}
}

func (m *mockRefreshTokenRepo) Create(ctx context.Context, token *domain.RefreshToken) error {
	m.tokens[token.ID] = token
	m.hashIndex[token.TokenHash] = token
	return nil
}

func (m *mockRefreshTokenRepo) FindByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	token, ok := m.hashIndex[tokenHash]
	if !ok {
		return nil, errors.NotFound("refresh_token", tokenHash)
	}
	return token, nil
}

func (m *mockRefreshTokenRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.RefreshToken, error) {
	var result []*domain.RefreshToken
	for _, t := range m.tokens {
		if t.UserID == userID {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *mockRefreshTokenRepo) GetActiveSessions(ctx context.Context, userID uuid.UUID, currentTokenHash string) ([]*domain.Session, error) {
	return nil, nil
}

func (m *mockRefreshTokenRepo) Delete(ctx context.Context, id uuid.UUID) error {
	token, ok := m.tokens[id]
	if ok {
		delete(m.hashIndex, token.TokenHash)
		delete(m.tokens, id)
	}
	return nil
}

func (m *mockRefreshTokenRepo) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	token, ok := m.hashIndex[tokenHash]
	if ok {
		delete(m.tokens, token.ID)
		delete(m.hashIndex, tokenHash)
	}
	return nil
}

func (m *mockRefreshTokenRepo) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	for id, t := range m.tokens {
		if t.UserID == userID {
			delete(m.hashIndex, t.TokenHash)
			delete(m.tokens, id)
		}
	}
	return nil
}

func (m *mockRefreshTokenRepo) DeleteExpired(ctx context.Context) (int64, error) {
	return 0, nil
}

type mockBackupCodeRepo struct {
	codes map[uuid.UUID][]*domain.BackupCode
}

func newMockBackupCodeRepo() *mockBackupCodeRepo {
	return &mockBackupCodeRepo{
		codes: make(map[uuid.UUID][]*domain.BackupCode),
	}
}

func (m *mockBackupCodeRepo) CreateBatch(ctx context.Context, codes []*domain.BackupCode) error {
	if len(codes) > 0 {
		m.codes[codes[0].UserID] = codes
	}
	return nil
}

func (m *mockBackupCodeRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.BackupCode, error) {
	return m.codes[userID], nil
}

func (m *mockBackupCodeRepo) FindUnusedByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.BackupCode, error) {
	var unused []*domain.BackupCode
	for _, c := range m.codes[userID] {
		if !c.Used {
			unused = append(unused, c)
		}
	}
	return unused, nil
}

func (m *mockBackupCodeRepo) MarkAsUsed(ctx context.Context, id uuid.UUID) error {
	for _, codes := range m.codes {
		for _, c := range codes {
			if c.ID == id {
				c.Used = true
				return nil
			}
		}
	}
	return nil
}

func (m *mockBackupCodeRepo) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	delete(m.codes, userID)
	return nil
}

type mockFollowRepo struct {
	follows map[string]bool
}

func newMockFollowRepo() *mockFollowRepo {
	return &mockFollowRepo{
		follows: make(map[string]bool),
	}
}

func (m *mockFollowRepo) Create(ctx context.Context, follow *domain.Follow) error {
	key := follow.FollowerID.String() + ":" + follow.FollowingID.String()
	m.follows[key] = true
	return nil
}

func (m *mockFollowRepo) Delete(ctx context.Context, followerID, followingID uuid.UUID) error {
	key := followerID.String() + ":" + followingID.String()
	delete(m.follows, key)
	return nil
}

func (m *mockFollowRepo) Exists(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	key := followerID.String() + ":" + followingID.String()
	return m.follows[key], nil
}

func (m *mockFollowRepo) GetFollowers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.User, int, error) {
	return nil, 0, nil
}

func (m *mockFollowRepo) GetFollowing(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.User, int, error) {
	return nil, 0, nil
}

func (m *mockFollowRepo) CountFollowers(ctx context.Context, userID uuid.UUID) (int, error) {
	return 0, nil
}

func (m *mockFollowRepo) CountFollowing(ctx context.Context, userID uuid.UUID) (int, error) {
	return 0, nil
}

type mockVerificationTokenRepo struct {
	tokens    map[uuid.UUID]*domain.VerificationToken
	hashIndex map[string]*domain.VerificationToken
}

func newMockVerificationTokenRepo() *mockVerificationTokenRepo {
	return &mockVerificationTokenRepo{
		tokens:    make(map[uuid.UUID]*domain.VerificationToken),
		hashIndex: make(map[string]*domain.VerificationToken),
	}
}

func (m *mockVerificationTokenRepo) Create(ctx context.Context, token *domain.VerificationToken) error {
	m.tokens[token.ID] = token
	m.hashIndex[token.TokenHash] = token
	return nil
}

func (m *mockVerificationTokenRepo) FindByTokenHash(ctx context.Context, tokenHash string) (*domain.VerificationToken, error) {
	token, ok := m.hashIndex[tokenHash]
	if !ok {
		return nil, errors.NotFound("verification_token", tokenHash)
	}
	return token, nil
}

func (m *mockVerificationTokenRepo) FindActiveByUserIDAndType(ctx context.Context, userID uuid.UUID, tokenType domain.TokenType) (*domain.VerificationToken, error) {
	return nil, errors.NotFound("verification_token", userID.String())
}

func (m *mockVerificationTokenRepo) MarkAsUsed(ctx context.Context, id uuid.UUID) error {
	token, ok := m.tokens[id]
	if ok {
		now := time.Now()
		token.UsedAt = &now
	}
	return nil
}

func (m *mockVerificationTokenRepo) DeleteByUserIDAndType(ctx context.Context, userID uuid.UUID, tokenType domain.TokenType) error {
	return nil
}

func (m *mockVerificationTokenRepo) DeleteExpired(ctx context.Context) (int64, error) {
	return 0, nil
}

// mockTeacherVerificationRepo is a mock implementation of TeacherVerificationRepository.
type mockTeacherVerificationRepo struct {
	verifications map[uuid.UUID]*domain.TeacherVerification
	userIndex     map[uuid.UUID]*domain.TeacherVerification
}

func newMockTeacherVerificationRepo() *mockTeacherVerificationRepo {
	return &mockTeacherVerificationRepo{
		verifications: make(map[uuid.UUID]*domain.TeacherVerification),
		userIndex:     make(map[uuid.UUID]*domain.TeacherVerification),
	}
}

func (m *mockTeacherVerificationRepo) Create(ctx context.Context, verification *domain.TeacherVerification) error {
	m.verifications[verification.ID] = verification
	m.userIndex[verification.UserID] = verification
	return nil
}

func (m *mockTeacherVerificationRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.TeacherVerification, error) {
	v, ok := m.verifications[id]
	if !ok {
		return nil, errors.NotFound("teacher_verification", id.String())
	}
	return v, nil
}

func (m *mockTeacherVerificationRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*domain.TeacherVerification, error) {
	v, ok := m.userIndex[userID]
	if !ok {
		return nil, errors.NotFound("teacher_verification", userID.String())
	}
	return v, nil
}

func (m *mockTeacherVerificationRepo) FindPending(ctx context.Context, limit, offset int) ([]*domain.TeacherVerification, int, error) {
	var pending []*domain.TeacherVerification
	for _, v := range m.verifications {
		if v.Status == domain.VerificationStatusPending {
			pending = append(pending, v)
		}
	}
	total := len(pending)
	if offset >= total {
		return []*domain.TeacherVerification{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return pending[offset:end], total, nil
}

func (m *mockTeacherVerificationRepo) Update(ctx context.Context, verification *domain.TeacherVerification) error {
	m.verifications[verification.ID] = verification
	m.userIndex[verification.UserID] = verification
	return nil
}

func (m *mockTeacherVerificationRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.VerificationStatus, reviewedBy uuid.UUID, reason *string) error {
	v, ok := m.verifications[id]
	if !ok {
		return errors.NotFound("teacher_verification", id.String())
	}
	v.Status = status
	v.ReviewedBy = &reviewedBy
	now := time.Now()
	v.ReviewedAt = &now
	v.RejectionReason = reason
	return nil
}

func (m *mockTeacherVerificationRepo) ExistsByUserID(ctx context.Context, userID uuid.UUID) (bool, error) {
	_, ok := m.userIndex[userID]
	return ok, nil
}

func (m *mockTeacherVerificationRepo) ExistsPendingByUserID(ctx context.Context, userID uuid.UUID) (bool, error) {
	v, ok := m.userIndex[userID]
	if !ok {
		return false, nil
	}
	return v.Status == domain.VerificationStatusPending, nil
}

// Helper to create a test service
func newTestService() (UserService, *mockUserRepo, *mockRefreshTokenRepo) {
	userRepo := newMockUserRepo()
	oauthRepo := &mockOAuthRepo{}
	refreshTokenRepo := newMockRefreshTokenRepo()
	backupCodeRepo := newMockBackupCodeRepo()
	followRepo := newMockFollowRepo()
	verificationTokenRepo := newMockVerificationTokenRepo()
	teacherVerificationRepo := newMockTeacherVerificationRepo()

	jwtManager := jwt.NewManager(jwt.Config{
		Secret:             "test-secret-key-for-testing-only",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "ngasihtau-test",
	})

	svc := NewUserService(
		userRepo,
		oauthRepo,
		refreshTokenRepo,
		backupCodeRepo,
		followRepo,
		verificationTokenRepo,
		teacherVerificationRepo,
		jwtManager,
		nil, // No Google client for tests
		nil, // No event publisher for tests (will use nil checks)
	)

	return svc, userRepo, refreshTokenRepo
}

// Test: User Registration Flow
func TestRegister_Success(t *testing.T) {
	svc, userRepo, _ := newTestService()
	ctx := context.Background()

	input := RegisterInput{
		Email:    "test@example.com",
		Password: "SecurePass123!",
		Name:     "Test User",
	}

	result, err := svc.Register(ctx, input)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Verify result
	if result.User == nil {
		t.Fatal("Expected user in result")
	}
	if result.User.Email != input.Email {
		t.Errorf("Expected email %s, got %s", input.Email, result.User.Email)
	}
	if result.User.Name != input.Name {
		t.Errorf("Expected name %s, got %s", input.Name, result.User.Name)
	}
	if result.User.Role != domain.RoleStudent {
		t.Errorf("Expected role %s, got %s", domain.RoleStudent, result.User.Role)
	}
	if result.AccessToken == "" {
		t.Error("Expected access token")
	}
	if result.RefreshToken == "" {
		t.Error("Expected refresh token")
	}

	// Verify user was stored
	if len(userRepo.users) != 1 {
		t.Errorf("Expected 1 user in repo, got %d", len(userRepo.users))
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc, userRepo, _ := newTestService()
	ctx := context.Background()

	// Create existing user
	existingUser := domain.NewUser("test@example.com", "hash", "Existing")
	userRepo.users[existingUser.ID] = existingUser
	userRepo.emailIndex[existingUser.Email] = existingUser

	input := RegisterInput{
		Email:    "test@example.com",
		Password: "SecurePass123!",
		Name:     "Test User",
	}

	_, err := svc.Register(ctx, input)
	if err == nil {
		t.Fatal("Expected error for duplicate email")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeConflict {
		t.Errorf("Expected conflict error code, got %s", appErr.Code)
	}
}

// Test: Login Flow
func TestLogin_Success(t *testing.T) {
	svc, userRepo, _ := newTestService()
	ctx := context.Background()

	// Create user with hashed password
	passwordHash, _ := hash.Password("SecurePass123!")
	user := domain.NewUser("test@example.com", passwordHash, "Test User")
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user

	input := LoginInput{
		Email:    "test@example.com",
		Password: "SecurePass123!",
	}

	result, err := svc.Login(ctx, input)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if result.User == nil {
		t.Fatal("Expected user in result")
	}
	if result.AccessToken == "" {
		t.Error("Expected access token")
	}
	if result.RefreshToken == "" {
		t.Error("Expected refresh token")
	}
	if result.Requires2FA {
		t.Error("Expected Requires2FA to be false")
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	svc, userRepo, _ := newTestService()
	ctx := context.Background()

	// Create user with hashed password
	passwordHash, _ := hash.Password("SecurePass123!")
	user := domain.NewUser("test@example.com", passwordHash, "Test User")
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user

	input := LoginInput{
		Email:    "test@example.com",
		Password: "WrongPassword!",
	}

	_, err := svc.Login(ctx, input)
	if err == nil {
		t.Fatal("Expected error for invalid password")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeUnauthorized {
		t.Errorf("Expected unauthorized error code, got %s", appErr.Code)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	input := LoginInput{
		Email:    "nonexistent@example.com",
		Password: "SecurePass123!",
	}

	_, err := svc.Login(ctx, input)
	if err == nil {
		t.Fatal("Expected error for non-existent user")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeUnauthorized {
		t.Errorf("Expected unauthorized error code, got %s", appErr.Code)
	}
}

func TestLogin_With2FAEnabled(t *testing.T) {
	svc, userRepo, _ := newTestService()
	ctx := context.Background()

	// Create user with 2FA enabled
	passwordHash, _ := hash.Password("SecurePass123!")
	user := domain.NewUser("test@example.com", passwordHash, "Test User")
	user.TwoFactorEnabled = true
	secret := "JBSWY3DPEHPK3PXP"
	user.TwoFactorSecret = &secret
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user

	input := LoginInput{
		Email:    "test@example.com",
		Password: "SecurePass123!",
	}

	result, err := svc.Login(ctx, input)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if !result.Requires2FA {
		t.Error("Expected Requires2FA to be true")
	}
	if result.TempToken == "" {
		t.Error("Expected temp token for 2FA flow")
	}
	if result.AccessToken != "" {
		t.Error("Expected no access token when 2FA is required")
	}
}

// Test: Token Refresh Flow
func TestRefreshToken_Success(t *testing.T) {
	svc, userRepo, refreshTokenRepo := newTestService()
	ctx := context.Background()

	// Create user
	user := domain.NewUser("test@example.com", "hash", "Test User")
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user

	// Create refresh token
	rawToken := "test-refresh-token"
	tokenHash, _ := hash.SHA256(rawToken)
	refreshToken := domain.NewRefreshToken(
		user.ID,
		tokenHash,
		nil,
		time.Now().Add(7*24*time.Hour),
	)
	refreshTokenRepo.tokens[refreshToken.ID] = refreshToken
	refreshTokenRepo.hashIndex[tokenHash] = refreshToken

	result, err := svc.RefreshToken(ctx, rawToken)
	if err != nil {
		t.Fatalf("RefreshToken failed: %v", err)
	}

	if result.User == nil {
		t.Fatal("Expected user in result")
	}
	if result.AccessToken == "" {
		t.Error("Expected new access token")
	}
	if result.RefreshToken == "" {
		t.Error("Expected new refresh token")
	}

	// Verify old token was deleted (token rotation)
	if _, ok := refreshTokenRepo.hashIndex[tokenHash]; ok {
		t.Error("Expected old refresh token to be deleted")
	}
}

func TestRefreshToken_ExpiredToken(t *testing.T) {
	svc, userRepo, refreshTokenRepo := newTestService()
	ctx := context.Background()

	// Create user
	user := domain.NewUser("test@example.com", "hash", "Test User")
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user

	// Create expired refresh token
	rawToken := "test-refresh-token"
	tokenHash, _ := hash.SHA256(rawToken)
	refreshToken := domain.NewRefreshToken(
		user.ID,
		tokenHash,
		nil,
		time.Now().Add(-1*time.Hour), // Expired
	)
	refreshTokenRepo.tokens[refreshToken.ID] = refreshToken
	refreshTokenRepo.hashIndex[tokenHash] = refreshToken

	_, err := svc.RefreshToken(ctx, rawToken)
	if err == nil {
		t.Fatal("Expected error for expired token")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeUnauthorized {
		t.Errorf("Expected unauthorized error code, got %s", appErr.Code)
	}
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	_, err := svc.RefreshToken(ctx, "invalid-token")
	if err == nil {
		t.Fatal("Expected error for invalid token")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeUnauthorized {
		t.Errorf("Expected unauthorized error code, got %s", appErr.Code)
	}
}

// Test: Logout
func TestLogout_Success(t *testing.T) {
	svc, _, refreshTokenRepo := newTestService()
	ctx := context.Background()

	// Create refresh token
	rawToken := "test-refresh-token"
	tokenHash, _ := hash.SHA256(rawToken)
	userID := uuid.New()
	refreshToken := domain.NewRefreshToken(
		userID,
		tokenHash,
		nil,
		time.Now().Add(7*24*time.Hour),
	)
	refreshTokenRepo.tokens[refreshToken.ID] = refreshToken
	refreshTokenRepo.hashIndex[tokenHash] = refreshToken

	err := svc.Logout(ctx, rawToken)
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	// Verify token was deleted
	if _, ok := refreshTokenRepo.hashIndex[tokenHash]; ok {
		t.Error("Expected refresh token to be deleted")
	}
}

// Test: 2FA Enable Flow
func TestEnable2FA_Success(t *testing.T) {
	svc, userRepo, _ := newTestService()
	ctx := context.Background()

	// Create user without 2FA
	user := domain.NewUser("test@example.com", "hash", "Test User")
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user

	result, err := svc.Enable2FA(ctx, user.ID)
	if err != nil {
		t.Fatalf("Enable2FA failed: %v", err)
	}

	if result.Secret == "" {
		t.Error("Expected TOTP secret")
	}
	if result.QRCodeURL == "" {
		t.Error("Expected QR code URL")
	}
	if len(result.BackupCodes) == 0 {
		t.Error("Expected backup codes")
	}

	// Verify secret was stored (but 2FA not yet enabled)
	storedUser := userRepo.users[user.ID]
	if storedUser.TwoFactorSecret == nil {
		t.Error("Expected TOTP secret to be stored")
	}
	if storedUser.TwoFactorEnabled {
		t.Error("Expected 2FA to not be enabled yet (needs verification)")
	}
}

func TestEnable2FA_AlreadyEnabled(t *testing.T) {
	svc, userRepo, _ := newTestService()
	ctx := context.Background()

	// Create user with 2FA already enabled
	user := domain.NewUser("test@example.com", "hash", "Test User")
	user.TwoFactorEnabled = true
	secret := "existing-secret"
	user.TwoFactorSecret = &secret
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user

	_, err := svc.Enable2FA(ctx, user.ID)
	if err == nil {
		t.Fatal("Expected error when 2FA is already enabled")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeBadRequest {
		t.Errorf("Expected bad request error code, got %s", appErr.Code)
	}
}

// =============================================================================
// Teacher Verification Service Tests
// =============================================================================

// Test: SubmitTeacherVerification - Success
func TestSubmitTeacherVerification_Success(t *testing.T) {
	svc, userRepo, teacherVerificationRepo := newTestServiceWithTeacherVerification()
	ctx := context.Background()

	// Create a student user
	user := domain.NewUser("student@example.com", "hash", "Student User")
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user

	input := TeacherVerificationInput{
		FullName:       "John Doe Teacher",
		IDNumber:       "1234567890123456",
		CredentialType: domain.CredentialTypeEducatorCard,
		DocumentRef:    "ref://documents/educator-card-123",
	}

	result, err := svc.SubmitTeacherVerification(ctx, user.ID, input)
	if err != nil {
		t.Fatalf("SubmitTeacherVerification failed: %v", err)
	}

	// Verify result
	if result == nil {
		t.Fatal("Expected verification result")
	}
	if result.UserID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, result.UserID)
	}
	if result.FullName != input.FullName {
		t.Errorf("Expected full name %s, got %s", input.FullName, result.FullName)
	}
	if result.IDNumber != input.IDNumber {
		t.Errorf("Expected ID number %s, got %s", input.IDNumber, result.IDNumber)
	}
	if result.CredentialType != input.CredentialType {
		t.Errorf("Expected credential type %s, got %s", input.CredentialType, result.CredentialType)
	}
	if result.DocumentRef != input.DocumentRef {
		t.Errorf("Expected document ref %s, got %s", input.DocumentRef, result.DocumentRef)
	}
	if result.Status != domain.VerificationStatusPending {
		t.Errorf("Expected status %s, got %s", domain.VerificationStatusPending, result.Status)
	}

	// Verify verification was stored
	if len(teacherVerificationRepo.verifications) != 1 {
		t.Errorf("Expected 1 verification in repo, got %d", len(teacherVerificationRepo.verifications))
	}
}

// Test: SubmitTeacherVerification - User Already Teacher
func TestSubmitTeacherVerification_UserAlreadyTeacher(t *testing.T) {
	svc, userRepo, _ := newTestServiceWithTeacherVerification()
	ctx := context.Background()

	// Create a teacher user
	user := domain.NewUser("teacher@example.com", "hash", "Teacher User")
	user.Role = domain.RoleTeacher
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user

	input := TeacherVerificationInput{
		FullName:       "John Doe Teacher",
		IDNumber:       "1234567890123456",
		CredentialType: domain.CredentialTypeEducatorCard,
		DocumentRef:    "ref://documents/educator-card-123",
	}

	_, err := svc.SubmitTeacherVerification(ctx, user.ID, input)
	if err == nil {
		t.Fatal("Expected error for user already teacher")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeBadRequest {
		t.Errorf("Expected bad request error code, got %s", appErr.Code)
	}
}

// Test: SubmitTeacherVerification - Pending Verification Exists
func TestSubmitTeacherVerification_PendingExists(t *testing.T) {
	svc, userRepo, teacherVerificationRepo := newTestServiceWithTeacherVerification()
	ctx := context.Background()

	// Create a student user
	user := domain.NewUser("student@example.com", "hash", "Student User")
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user

	// Create existing pending verification
	existingVerification := domain.NewTeacherVerification(
		user.ID,
		"Existing Name",
		"1234567890123456",
		domain.CredentialTypeGovernmentID,
		"ref://existing",
	)
	teacherVerificationRepo.verifications[existingVerification.ID] = existingVerification
	teacherVerificationRepo.userIndex[user.ID] = existingVerification

	input := TeacherVerificationInput{
		FullName:       "John Doe Teacher",
		IDNumber:       "1234567890123456",
		CredentialType: domain.CredentialTypeEducatorCard,
		DocumentRef:    "ref://documents/educator-card-123",
	}

	_, err := svc.SubmitTeacherVerification(ctx, user.ID, input)
	if err == nil {
		t.Fatal("Expected error for pending verification exists")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeConflict {
		t.Errorf("Expected conflict error code, got %s", appErr.Code)
	}
}

// Test: SubmitTeacherVerification - Validation Errors
func TestSubmitTeacherVerification_ValidationErrors(t *testing.T) {
	svc, userRepo, _ := newTestServiceWithTeacherVerification()
	ctx := context.Background()

	// Create a student user
	user := domain.NewUser("student@example.com", "hash", "Student User")
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user

	tests := []struct {
		name  string
		input TeacherVerificationInput
	}{
		{
			name: "empty full name",
			input: TeacherVerificationInput{
				FullName:       "",
				IDNumber:       "1234567890123456",
				CredentialType: domain.CredentialTypeEducatorCard,
				DocumentRef:    "ref://documents/123",
			},
		},
		{
			name: "short full name",
			input: TeacherVerificationInput{
				FullName:       "AB",
				IDNumber:       "1234567890123456",
				CredentialType: domain.CredentialTypeEducatorCard,
				DocumentRef:    "ref://documents/123",
			},
		},
		{
			name: "empty ID number",
			input: TeacherVerificationInput{
				FullName:       "John Doe Teacher",
				IDNumber:       "",
				CredentialType: domain.CredentialTypeEducatorCard,
				DocumentRef:    "ref://documents/123",
			},
		},
		{
			name: "short ID number",
			input: TeacherVerificationInput{
				FullName:       "John Doe Teacher",
				IDNumber:       "123456789",
				CredentialType: domain.CredentialTypeEducatorCard,
				DocumentRef:    "ref://documents/123",
			},
		},
		{
			name: "empty credential type",
			input: TeacherVerificationInput{
				FullName:       "John Doe Teacher",
				IDNumber:       "1234567890123456",
				CredentialType: "",
				DocumentRef:    "ref://documents/123",
			},
		},
		{
			name: "invalid credential type",
			input: TeacherVerificationInput{
				FullName:       "John Doe Teacher",
				IDNumber:       "1234567890123456",
				CredentialType: "invalid_type",
				DocumentRef:    "ref://documents/123",
			},
		},
		{
			name: "empty document ref",
			input: TeacherVerificationInput{
				FullName:       "John Doe Teacher",
				IDNumber:       "1234567890123456",
				CredentialType: domain.CredentialTypeEducatorCard,
				DocumentRef:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.SubmitTeacherVerification(ctx, user.ID, tt.input)
			if err == nil {
				t.Fatal("Expected validation error")
			}

			appErr, ok := err.(*errors.AppError)
			if !ok {
				t.Fatalf("Expected AppError, got %T", err)
			}
			if appErr.Code != errors.CodeValidationError {
				t.Errorf("Expected validation error code, got %s", appErr.Code)
			}
		})
	}
}

// Test: ApproveVerification - Success
func TestApproveVerification_Success(t *testing.T) {
	svc, userRepo, teacherVerificationRepo := newTestServiceWithTeacherVerification()
	ctx := context.Background()

	// Create a student user
	user := domain.NewUser("student@example.com", "hash", "Student User")
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user

	// Create pending verification
	verification := domain.NewTeacherVerification(
		user.ID,
		"John Doe Teacher",
		"1234567890123456",
		domain.CredentialTypeEducatorCard,
		"ref://documents/123",
	)
	teacherVerificationRepo.verifications[verification.ID] = verification
	teacherVerificationRepo.userIndex[user.ID] = verification

	reviewerID := uuid.New()

	err := svc.ApproveVerification(ctx, verification.ID, reviewerID)
	if err != nil {
		t.Fatalf("ApproveVerification failed: %v", err)
	}

	// Verify verification status was updated
	updatedVerification := teacherVerificationRepo.verifications[verification.ID]
	if updatedVerification.Status != domain.VerificationStatusApproved {
		t.Errorf("Expected status %s, got %s", domain.VerificationStatusApproved, updatedVerification.Status)
	}
	if updatedVerification.ReviewedBy == nil || *updatedVerification.ReviewedBy != reviewerID {
		t.Error("Expected reviewer ID to be set")
	}
	if updatedVerification.ReviewedAt == nil {
		t.Error("Expected reviewed at to be set")
	}

	// Verify user role was updated to teacher
	updatedUser := userRepo.users[user.ID]
	if updatedUser.Role != domain.RoleTeacher {
		t.Errorf("Expected user role %s, got %s", domain.RoleTeacher, updatedUser.Role)
	}
}

// Test: ApproveVerification - Already Reviewed
func TestApproveVerification_AlreadyReviewed(t *testing.T) {
	svc, userRepo, teacherVerificationRepo := newTestServiceWithTeacherVerification()
	ctx := context.Background()

	// Create a student user
	user := domain.NewUser("student@example.com", "hash", "Student User")
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user

	// Create already approved verification
	verification := domain.NewTeacherVerification(
		user.ID,
		"John Doe Teacher",
		"1234567890123456",
		domain.CredentialTypeEducatorCard,
		"ref://documents/123",
	)
	verification.Approve(uuid.New())
	teacherVerificationRepo.verifications[verification.ID] = verification
	teacherVerificationRepo.userIndex[user.ID] = verification

	reviewerID := uuid.New()

	err := svc.ApproveVerification(ctx, verification.ID, reviewerID)
	if err == nil {
		t.Fatal("Expected error for already reviewed verification")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeBadRequest {
		t.Errorf("Expected bad request error code, got %s", appErr.Code)
	}
}

// Test: ApproveVerification - Not Found
func TestApproveVerification_NotFound(t *testing.T) {
	svc, _, _ := newTestServiceWithTeacherVerification()
	ctx := context.Background()

	nonExistentID := uuid.New()
	reviewerID := uuid.New()

	err := svc.ApproveVerification(ctx, nonExistentID, reviewerID)
	if err == nil {
		t.Fatal("Expected error for non-existent verification")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeNotFound {
		t.Errorf("Expected not found error code, got %s", appErr.Code)
	}
}

// Test: RejectVerification - Success
func TestRejectVerification_Success(t *testing.T) {
	svc, userRepo, teacherVerificationRepo := newTestServiceWithTeacherVerification()
	ctx := context.Background()

	// Create a student user
	user := domain.NewUser("student@example.com", "hash", "Student User")
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user

	// Create pending verification
	verification := domain.NewTeacherVerification(
		user.ID,
		"John Doe Teacher",
		"1234567890123456",
		domain.CredentialTypeEducatorCard,
		"ref://documents/123",
	)
	teacherVerificationRepo.verifications[verification.ID] = verification
	teacherVerificationRepo.userIndex[user.ID] = verification

	reviewerID := uuid.New()
	reason := "Invalid credentials provided"

	err := svc.RejectVerification(ctx, verification.ID, reviewerID, reason)
	if err != nil {
		t.Fatalf("RejectVerification failed: %v", err)
	}

	// Verify verification status was updated
	updatedVerification := teacherVerificationRepo.verifications[verification.ID]
	if updatedVerification.Status != domain.VerificationStatusRejected {
		t.Errorf("Expected status %s, got %s", domain.VerificationStatusRejected, updatedVerification.Status)
	}
	if updatedVerification.ReviewedBy == nil || *updatedVerification.ReviewedBy != reviewerID {
		t.Error("Expected reviewer ID to be set")
	}
	if updatedVerification.ReviewedAt == nil {
		t.Error("Expected reviewed at to be set")
	}
	if updatedVerification.RejectionReason == nil || *updatedVerification.RejectionReason != reason {
		t.Errorf("Expected rejection reason %s, got %v", reason, updatedVerification.RejectionReason)
	}

	// Verify user role was NOT changed (should remain student)
	updatedUser := userRepo.users[user.ID]
	if updatedUser.Role != domain.RoleStudent {
		t.Errorf("Expected user role to remain %s, got %s", domain.RoleStudent, updatedUser.Role)
	}
}

// Test: RejectVerification - Empty Reason
func TestRejectVerification_EmptyReason(t *testing.T) {
	svc, userRepo, teacherVerificationRepo := newTestServiceWithTeacherVerification()
	ctx := context.Background()

	// Create a student user
	user := domain.NewUser("student@example.com", "hash", "Student User")
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user

	// Create pending verification
	verification := domain.NewTeacherVerification(
		user.ID,
		"John Doe Teacher",
		"1234567890123456",
		domain.CredentialTypeEducatorCard,
		"ref://documents/123",
	)
	teacherVerificationRepo.verifications[verification.ID] = verification
	teacherVerificationRepo.userIndex[user.ID] = verification

	reviewerID := uuid.New()

	err := svc.RejectVerification(ctx, verification.ID, reviewerID, "")
	if err == nil {
		t.Fatal("Expected error for empty rejection reason")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeBadRequest {
		t.Errorf("Expected bad request error code, got %s", appErr.Code)
	}
}

// Test: RejectVerification - Already Reviewed
func TestRejectVerification_AlreadyReviewed(t *testing.T) {
	svc, userRepo, teacherVerificationRepo := newTestServiceWithTeacherVerification()
	ctx := context.Background()

	// Create a student user
	user := domain.NewUser("student@example.com", "hash", "Student User")
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user

	// Create already rejected verification
	verification := domain.NewTeacherVerification(
		user.ID,
		"John Doe Teacher",
		"1234567890123456",
		domain.CredentialTypeEducatorCard,
		"ref://documents/123",
	)
	verification.Reject(uuid.New(), "Previous rejection")
	teacherVerificationRepo.verifications[verification.ID] = verification
	teacherVerificationRepo.userIndex[user.ID] = verification

	reviewerID := uuid.New()

	err := svc.RejectVerification(ctx, verification.ID, reviewerID, "New rejection reason")
	if err == nil {
		t.Fatal("Expected error for already reviewed verification")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeBadRequest {
		t.Errorf("Expected bad request error code, got %s", appErr.Code)
	}
}

// Test: RejectVerification - Not Found
func TestRejectVerification_NotFound(t *testing.T) {
	svc, _, _ := newTestServiceWithTeacherVerification()
	ctx := context.Background()

	nonExistentID := uuid.New()
	reviewerID := uuid.New()

	err := svc.RejectVerification(ctx, nonExistentID, reviewerID, "Some reason")
	if err == nil {
		t.Fatal("Expected error for non-existent verification")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeNotFound {
		t.Errorf("Expected not found error code, got %s", appErr.Code)
	}
}

// Test: GetVerificationStatus - Success
func TestGetVerificationStatus_Success(t *testing.T) {
	svc, userRepo, teacherVerificationRepo := newTestServiceWithTeacherVerification()
	ctx := context.Background()

	// Create a student user
	user := domain.NewUser("student@example.com", "hash", "Student User")
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user

	// Create pending verification
	verification := domain.NewTeacherVerification(
		user.ID,
		"John Doe Teacher",
		"1234567890123456",
		domain.CredentialTypeEducatorCard,
		"ref://documents/123",
	)
	teacherVerificationRepo.verifications[verification.ID] = verification
	teacherVerificationRepo.userIndex[user.ID] = verification

	result, err := svc.GetVerificationStatus(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetVerificationStatus failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected verification result")
	}
	if result.ID != verification.ID {
		t.Errorf("Expected verification ID %s, got %s", verification.ID, result.ID)
	}
	if result.Status != domain.VerificationStatusPending {
		t.Errorf("Expected status %s, got %s", domain.VerificationStatusPending, result.Status)
	}
}

// Test: GetVerificationStatus - Not Found
func TestGetVerificationStatus_NotFound(t *testing.T) {
	svc, userRepo, _ := newTestServiceWithTeacherVerification()
	ctx := context.Background()

	// Create a student user without verification
	user := domain.NewUser("student@example.com", "hash", "Student User")
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user

	_, err := svc.GetVerificationStatus(ctx, user.ID)
	if err == nil {
		t.Fatal("Expected error for non-existent verification")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeNotFound {
		t.Errorf("Expected not found error code, got %s", appErr.Code)
	}
}

// Test: GetPendingVerifications - Success
func TestGetPendingVerifications_Success(t *testing.T) {
	svc, userRepo, teacherVerificationRepo := newTestServiceWithTeacherVerification()
	ctx := context.Background()

	// Create multiple users with pending verifications
	for i := 0; i < 3; i++ {
		user := domain.NewUser("student"+string(rune('0'+i))+"@example.com", "hash", "Student "+string(rune('0'+i)))
		userRepo.users[user.ID] = user
		userRepo.emailIndex[user.Email] = user

		verification := domain.NewTeacherVerification(
			user.ID,
			"Teacher "+string(rune('0'+i)),
			"123456789012345"+string(rune('0'+i)),
			domain.CredentialTypeEducatorCard,
			"ref://documents/"+string(rune('0'+i)),
		)
		teacherVerificationRepo.verifications[verification.ID] = verification
		teacherVerificationRepo.userIndex[user.ID] = verification
	}

	result, err := svc.GetPendingVerifications(ctx, 1, 10)
	if err != nil {
		t.Fatalf("GetPendingVerifications failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result")
	}
	if result.Total != 3 {
		t.Errorf("Expected total 3, got %d", result.Total)
	}
	if len(result.Verifications) != 3 {
		t.Errorf("Expected 3 verifications, got %d", len(result.Verifications))
	}
	if result.Page != 1 {
		t.Errorf("Expected page 1, got %d", result.Page)
	}
	if result.PerPage != 10 {
		t.Errorf("Expected per page 10, got %d", result.PerPage)
	}
}

// Test: GetPendingVerifications - Pagination
func TestGetPendingVerifications_Pagination(t *testing.T) {
	svc, userRepo, teacherVerificationRepo := newTestServiceWithTeacherVerification()
	ctx := context.Background()

	// Create 5 users with pending verifications
	for i := 0; i < 5; i++ {
		user := domain.NewUser("student"+string(rune('0'+i))+"@example.com", "hash", "Student "+string(rune('0'+i)))
		userRepo.users[user.ID] = user
		userRepo.emailIndex[user.Email] = user

		verification := domain.NewTeacherVerification(
			user.ID,
			"Teacher "+string(rune('0'+i)),
			"123456789012345"+string(rune('0'+i)),
			domain.CredentialTypeEducatorCard,
			"ref://documents/"+string(rune('0'+i)),
		)
		teacherVerificationRepo.verifications[verification.ID] = verification
		teacherVerificationRepo.userIndex[user.ID] = verification
	}

	// Get first page with 2 items
	result, err := svc.GetPendingVerifications(ctx, 1, 2)
	if err != nil {
		t.Fatalf("GetPendingVerifications failed: %v", err)
	}

	if result.Total != 5 {
		t.Errorf("Expected total 5, got %d", result.Total)
	}
	if len(result.Verifications) != 2 {
		t.Errorf("Expected 2 verifications on page 1, got %d", len(result.Verifications))
	}
	if result.TotalPages != 3 {
		t.Errorf("Expected 3 total pages, got %d", result.TotalPages)
	}
}

// Test: GetPendingVerifications - Empty
func TestGetPendingVerifications_Empty(t *testing.T) {
	svc, _, _ := newTestServiceWithTeacherVerification()
	ctx := context.Background()

	result, err := svc.GetPendingVerifications(ctx, 1, 10)
	if err != nil {
		t.Fatalf("GetPendingVerifications failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result")
	}
	if result.Total != 0 {
		t.Errorf("Expected total 0, got %d", result.Total)
	}
	if len(result.Verifications) != 0 {
		t.Errorf("Expected 0 verifications, got %d", len(result.Verifications))
	}
}
