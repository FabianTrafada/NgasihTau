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

// Helper to create a test service
func newTestService() (UserService, *mockUserRepo, *mockRefreshTokenRepo) {
	userRepo := newMockUserRepo()
	oauthRepo := &mockOAuthRepo{}
	refreshTokenRepo := newMockRefreshTokenRepo()
	backupCodeRepo := newMockBackupCodeRepo()
	followRepo := newMockFollowRepo()
	verificationTokenRepo := newMockVerificationTokenRepo()

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
