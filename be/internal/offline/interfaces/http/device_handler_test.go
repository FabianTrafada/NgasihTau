package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/common/middleware"
	"ngasihtau/internal/offline/application"
	"ngasihtau/internal/offline/domain"
)

// MockDeviceService is a mock implementation of DeviceService.
type MockDeviceService struct {
	mock.Mock
}

func (m *MockDeviceService) RegisterDevice(ctx context.Context, input application.RegisterDeviceInput) (*domain.Device, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Device), args.Error(1)
}

func (m *MockDeviceService) ListDevices(ctx context.Context, userID uuid.UUID) ([]*domain.Device, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Device), args.Error(1)
}

func (m *MockDeviceService) DeregisterDevice(ctx context.Context, userID, deviceID uuid.UUID) error {
	args := m.Called(ctx, userID, deviceID)
	return args.Error(0)
}

func (m *MockDeviceService) ValidateDevice(ctx context.Context, userID uuid.UUID, fingerprint string) (*domain.Device, error) {
	args := m.Called(ctx, userID, fingerprint)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Device), args.Error(1)
}

func (m *MockDeviceService) UpdateLastUsed(ctx context.Context, deviceID uuid.UUID) error {
	args := m.Called(ctx, deviceID)
	return args.Error(0)
}

// setupTestApp creates a test Fiber app with the device handler.
func setupTestApp(deviceService application.DeviceService) *fiber.App {
	app := fiber.New()
	handler := NewDeviceHandler(deviceService)

	// Add a middleware to set user ID for testing
	app.Use(func(c *fiber.Ctx) error {
		userIDStr := c.Get("X-Test-User-ID")
		if userIDStr != "" {
			userID, err := uuid.Parse(userIDStr)
			if err == nil {
				c.Locals(middleware.UserIDKey, userID)
			}
		}
		return c.Next()
	})

	// Register routes
	app.Post("/devices", handler.RegisterDevice)
	app.Get("/devices", handler.ListDevices)
	app.Delete("/devices/:device_id", handler.DeregisterDevice)

	return app
}

// createTestDevice creates a test device.
func createTestDevice(userID uuid.UUID) *domain.Device {
	return &domain.Device{
		ID:          uuid.New(),
		UserID:      userID,
		Fingerprint: "test-fingerprint-12345678901234567890",
		Name:        "Test Device",
		Platform:    domain.PlatformIOS,
		LastUsedAt:  time.Now(),
		CreatedAt:   time.Now(),
	}
}

func TestDeviceHandler_RegisterDevice(t *testing.T) {
	mockService := new(MockDeviceService)
	app := setupTestApp(mockService)

	userID := uuid.New()
	device := createTestDevice(userID)

	mockService.On("RegisterDevice", mock.Anything, mock.MatchedBy(func(input application.RegisterDeviceInput) bool {
		return input.UserID == userID &&
			input.Fingerprint == "test-fingerprint-12345678901234567890" &&
			input.Name == "My iPhone" &&
			input.Platform == domain.PlatformIOS
	})).Return(device, nil)

	reqBody := RegisterDeviceRequest{
		Fingerprint: "test-fingerprint-12345678901234567890",
		Name:        "My iPhone",
		Platform:    "ios",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/devices", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", userID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestDeviceHandler_RegisterDevice_ValidationError(t *testing.T) {
	mockService := new(MockDeviceService)
	app := setupTestApp(mockService)

	userID := uuid.New()

	// Missing required fields
	reqBody := RegisterDeviceRequest{
		Fingerprint: "short", // Too short
		Name:        "",      // Required
		Platform:    "invalid",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/devices", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", userID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestDeviceHandler_RegisterDevice_DeviceLimitExceeded(t *testing.T) {
	mockService := new(MockDeviceService)
	app := setupTestApp(mockService)

	userID := uuid.New()

	mockService.On("RegisterDevice", mock.Anything, mock.Anything).
		Return(nil, errors.New(errors.CodeForbidden, "DEVICE_LIMIT_EXCEEDED"))

	reqBody := RegisterDeviceRequest{
		Fingerprint: "test-fingerprint-12345678901234567890",
		Name:        "My iPhone",
		Platform:    "ios",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/devices", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", userID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestDeviceHandler_RegisterDevice_Unauthorized(t *testing.T) {
	mockService := new(MockDeviceService)
	app := setupTestApp(mockService)

	reqBody := RegisterDeviceRequest{
		Fingerprint: "test-fingerprint-12345678901234567890",
		Name:        "My iPhone",
		Platform:    "ios",
	}
	body, _ := json.Marshal(reqBody)

	// No user ID header
	req := httptest.NewRequest("POST", "/devices", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestDeviceHandler_ListDevices(t *testing.T) {
	mockService := new(MockDeviceService)
	app := setupTestApp(mockService)

	userID := uuid.New()
	devices := []*domain.Device{
		createTestDevice(userID),
		createTestDevice(userID),
	}

	mockService.On("ListDevices", mock.Anything, userID).Return(devices, nil)

	req := httptest.NewRequest("GET", "/devices", nil)
	req.Header.Set("X-Test-User-ID", userID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestDeviceHandler_ListDevices_Empty(t *testing.T) {
	mockService := new(MockDeviceService)
	app := setupTestApp(mockService)

	userID := uuid.New()
	devices := []*domain.Device{}

	mockService.On("ListDevices", mock.Anything, userID).Return(devices, nil)

	req := httptest.NewRequest("GET", "/devices", nil)
	req.Header.Set("X-Test-User-ID", userID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestDeviceHandler_ListDevices_Unauthorized(t *testing.T) {
	mockService := new(MockDeviceService)
	app := setupTestApp(mockService)

	// No user ID header
	req := httptest.NewRequest("GET", "/devices", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestDeviceHandler_DeregisterDevice(t *testing.T) {
	mockService := new(MockDeviceService)
	app := setupTestApp(mockService)

	userID := uuid.New()
	deviceID := uuid.New()

	mockService.On("DeregisterDevice", mock.Anything, userID, deviceID).Return(nil)

	req := httptest.NewRequest("DELETE", "/devices/"+deviceID.String(), nil)
	req.Header.Set("X-Test-User-ID", userID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestDeviceHandler_DeregisterDevice_NotFound(t *testing.T) {
	mockService := new(MockDeviceService)
	app := setupTestApp(mockService)

	userID := uuid.New()
	deviceID := uuid.New()

	mockService.On("DeregisterDevice", mock.Anything, userID, deviceID).
		Return(errors.NotFound("device", deviceID.String()))

	req := httptest.NewRequest("DELETE", "/devices/"+deviceID.String(), nil)
	req.Header.Set("X-Test-User-ID", userID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestDeviceHandler_DeregisterDevice_InvalidID(t *testing.T) {
	mockService := new(MockDeviceService)
	app := setupTestApp(mockService)

	userID := uuid.New()

	req := httptest.NewRequest("DELETE", "/devices/invalid-uuid", nil)
	req.Header.Set("X-Test-User-ID", userID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestDeviceHandler_DeregisterDevice_Unauthorized(t *testing.T) {
	mockService := new(MockDeviceService)
	app := setupTestApp(mockService)

	deviceID := uuid.New()

	// No user ID header
	req := httptest.NewRequest("DELETE", "/devices/"+deviceID.String(), nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestToDeviceResponse(t *testing.T) {
	userID := uuid.New()
	device := createTestDevice(userID)

	resp := ToDeviceResponse(device)

	assert.NotNil(t, resp)
	assert.Equal(t, device.ID, resp.ID)
	assert.Equal(t, device.UserID, resp.UserID)
	assert.Equal(t, device.Fingerprint, resp.Fingerprint)
	assert.Equal(t, device.Name, resp.Name)
	assert.Equal(t, string(device.Platform), resp.Platform)
	assert.Equal(t, string(device.Status()), resp.Status) // Status() is a method
}

func TestToDeviceResponse_Nil(t *testing.T) {
	resp := ToDeviceResponse(nil)
	assert.Nil(t, resp)
}

func TestToDeviceResponseList(t *testing.T) {
	userID := uuid.New()
	devices := []*domain.Device{
		createTestDevice(userID),
		createTestDevice(userID),
	}

	resp := ToDeviceResponseList(devices)

	assert.Len(t, resp, 2)
	for i, r := range resp {
		assert.Equal(t, devices[i].ID, r.ID)
	}
}
