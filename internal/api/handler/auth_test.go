package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/maidulcu/masaar-crm/internal/config"
	"github.com/maidulcu/masaar-crm/internal/domain"
	"github.com/maidulcu/masaar-crm/internal/repo"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepo implements a mock repository for testing
type MockUserRepo struct {
	users map[uuid.UUID]*domain.User
}

func NewMockUserRepo() *MockUserRepo {
	return &MockUserRepo{
		users: make(map[uuid.UUID]*domain.User),
	}
}

func (m *MockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, fiber.NewError(fiber.StatusNotFound, "user not found")
}

func (m *MockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, fiber.NewError(fiber.StatusNotFound, "user not found")
}

func (m *MockUserRepo) UpdatePassword(ctx context.Context, userID uuid.UUID, hash string) error {
	if u, ok := m.users[userID]; ok {
		u.PasswordHash = hash
		return nil
	}
	return fiber.NewError(fiber.StatusNotFound, "user not found")
}

func (m *MockUserRepo) UpdateLangPref(ctx context.Context, userID uuid.UUID, lang string) error {
	if u, ok := m.users[userID]; ok {
		u.LangPref = lang
		return nil
	}
	return fiber.NewError(fiber.StatusNotFound, "user not found")
}

func TestAuthLogin_Success(t *testing.T) {
	// Setup
	app := fiber.New()
	mockRepo := NewMockUserRepo()
	mockRedis, _ := redismock.NewClientMock()
	cfg := &config.Config{
		JWTSecret:             "test-secret-key-32-characters!",
		JWTAccessExpiryMin:    15,
		JWTRefreshExpiryDays:  7,
	}

	// Create test user
	password := "testpassword123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	userID := uuid.New()
	user := &domain.User{
		ID:           userID,
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: string(hash),
		Role:         domain.RoleAgent,
		LangPref:     "en",
	}
	mockRepo.users[userID] = user

	// Mock Redis Set
	mockRedis.ExpectSet("refresh:*", userID.String(), time.Duration(7*24)*time.Hour).SetVal("OK")

	handler := NewAuthHandler(mockRepo, mockRedis, cfg)
	app.Post("/login", handler.Login)

	// Test
	body := bytes.NewReader([]byte(`{"email":"test@example.com","password":"testpassword123"}`))
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if result["access_token"] == nil {
		t.Error("expected access_token in response")
	}
	if result["refresh_token"] == nil {
		t.Error("expected refresh_token in response")
	}
}

func TestAuthLogin_InvalidCredentials(t *testing.T) {
	app := fiber.New()
	mockRepo := NewMockUserRepo()
	mockRedis, _ := redismock.NewClientMock()
	cfg := &config.Config{
		JWTSecret:             "test-secret-key-32-characters!",
		JWTAccessExpiryMin:    15,
		JWTRefreshExpiryDays:  7,
	}

	handler := NewAuthHandler(mockRepo, mockRedis, cfg)
	app.Post("/login", handler.Login)

	body := bytes.NewReader([]byte(`{"email":"nonexistent@example.com","password":"password"}`))
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestAuthLogin_WrongPassword(t *testing.T) {
	app := fiber.New()
	mockRepo := NewMockUserRepo()
	mockRedis, _ := redismock.NewClientMock()
	cfg := &config.Config{
		JWTSecret:             "test-secret-key-32-characters!",
		JWTAccessExpiryMin:    15,
		JWTRefreshExpiryDays:  7,
	}

	// Create test user
	hash, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	userID := uuid.New()
	user := &domain.User{
		ID:           userID,
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: string(hash),
		Role:         domain.RoleAgent,
	}
	mockRepo.users[userID] = user

	handler := NewAuthHandler(mockRepo, mockRedis, cfg)
	app.Post("/login", handler.Login)

	body := bytes.NewReader([]byte(`{"email":"test@example.com","password":"wrongpassword"}`))
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestAuthLogin_InvalidRequest(t *testing.T) {
	app := fiber.New()
	mockRepo := NewMockUserRepo()
	mockRedis, _ := redismock.NewClientMock()
	cfg := &config.Config{
		JWTSecret:             "test-secret-key-32-characters!",
		JWTAccessExpiryMin:    15,
		JWTRefreshExpiryDays:  7,
	}

	handler := NewAuthHandler(mockRepo, mockRedis, cfg)
	app.Post("/login", handler.Login)

	// Invalid JSON
	body := bytes.NewReader([]byte(`{invalid json}`))
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}
