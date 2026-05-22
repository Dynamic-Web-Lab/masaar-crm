package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/maidulcu/masaar-crm/internal/domain"
	"github.com/maidulcu/masaar-crm/internal/repo"
	"github.com/maidulcu/masaar-crm/internal/ws"
)

// MockLeadRepo implements LeadRepository for testing.
type MockLeadRepo struct {
	leads map[uuid.UUID]*domain.Lead
}

func NewMockLeadRepo() *MockLeadRepo {
	return &MockLeadRepo{
		leads: make(map[uuid.UUID]*domain.Lead),
	}
}

func (m *MockLeadRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Lead, error) {
	if lead, ok := m.leads[id]; ok {
		return lead, nil
	}
	return nil, fiber.NewError(fiber.StatusNotFound, "lead not found")
}

func (m *MockLeadRepo) Create(ctx context.Context, lead *domain.Lead) error {
	if lead.ID == uuid.Nil {
		lead.ID = uuid.New()
	}
	m.leads[lead.ID] = lead
	return nil
}

func (m *MockLeadRepo) UpdateStage(ctx context.Context, id uuid.UUID, stage domain.LeadStage, reason string) error {
	if lead, ok := m.leads[id]; ok {
		lead.Stage = stage
		lead.ClosedReason = reason
		return nil
	}
	return fiber.NewError(fiber.StatusNotFound, "lead not found")
}

func (m *MockLeadRepo) UpdateNotes(ctx context.Context, id uuid.UUID, notes string) error {
	return nil
}

func (m *MockLeadRepo) Assign(ctx context.Context, id uuid.UUID, userID *uuid.UUID) error {
	return nil
}

func (m *MockLeadRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockLeadRepo) List(ctx context.Context, filter repo.LeadFilter) ([]domain.Lead, error) {
	return nil, nil
}

func (m *MockLeadRepo) KanbanBoard(ctx context.Context) (map[domain.LeadStage][]domain.Lead, error) {
	return make(map[domain.LeadStage][]domain.Lead), nil
}

// MockContactRepo implements ContactRepository for testing.
type MockContactRepo struct {
	contacts map[uuid.UUID]*domain.Contact
}

func NewMockContactRepo() *MockContactRepo {
	return &MockContactRepo{
		contacts: make(map[uuid.UUID]*domain.Contact),
	}
}

func (m *MockContactRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Contact, error) {
	if contact, ok := m.contacts[id]; ok {
		return contact, nil
	}
	return nil, fiber.NewError(fiber.StatusNotFound, "contact not found")
}

// stubCommHistRepo implements CommunicationHistoryRepository for testing.
type stubCommHistRepo struct{}

func (s *stubCommHistRepo) GetByLead(ctx context.Context, leadID uuid.UUID, limit int) ([]domain.CommunicationHistory, error) {
	return nil, nil
}

// stubScoringService implements ScoringService for testing.
type stubScoringService struct{}

func (s *stubScoringService) UpdateScoreOnStageChange(ctx context.Context, leadID uuid.UUID, newStage domain.LeadStage) error {
	return nil
}

// stubAuditLog implements AuditLogRepository for testing.
type stubAuditLog struct{}

func (s *stubAuditLog) Log(ctx context.Context, actorID uuid.UUID, action, entityType string, entityID uuid.UUID, diff any) {
}

func newTestLeadHandler(leadRepo LeadRepository, contactRepo ContactRepository, hub *ws.Hub) *LeadHandler {
	return NewLeadHandler(
		leadRepo,
		contactRepo,
		&stubCommHistRepo{},
		&stubScoringService{},
		hub,
		&stubAuditLog{},
		nil, // dispatcher — nil-safe in UpdateStage
	)
}

func TestLeadUpdateStage_ValidStage(t *testing.T) {
	app := fiber.New()
	mockLeadRepo := NewMockLeadRepo()
	mockContactRepo := NewMockContactRepo()
	hub := ws.NewHub()

	leadID := uuid.New()
	lead := &domain.Lead{
		ID:      leadID,
		Stage:   domain.StageNew,
		Contact: &domain.Contact{ID: uuid.New()},
	}
	mockLeadRepo.leads[leadID] = lead

	handler := newTestLeadHandler(mockLeadRepo, mockContactRepo, hub)
	app.Patch("/leads/:id/stage", handler.UpdateStage)

	body := bytes.NewReader([]byte(`{"stage":"contacted"}`))
	req := httptest.NewRequest(http.MethodPatch, "/leads/"+leadID.String()+"/stage", body)
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	if mockLeadRepo.leads[leadID].Stage != domain.StageContacted {
		t.Errorf("expected stage to be 'contacted', got '%s'", mockLeadRepo.leads[leadID].Stage)
	}
}

func TestLeadUpdateStage_InvalidStage(t *testing.T) {
	app := fiber.New()
	mockLeadRepo := NewMockLeadRepo()
	mockContactRepo := NewMockContactRepo()
	hub := ws.NewHub()

	leadID := uuid.New()
	lead := &domain.Lead{
		ID:    leadID,
		Stage: domain.StageNew,
	}
	mockLeadRepo.leads[leadID] = lead

	handler := newTestLeadHandler(mockLeadRepo, mockContactRepo, hub)
	app.Patch("/leads/:id/stage", handler.UpdateStage)

	body := bytes.NewReader([]byte(`{"stage":"invalid_stage"}`))
	req := httptest.NewRequest(http.MethodPatch, "/leads/"+leadID.String()+"/stage", body)
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid stage, got %d", resp.StatusCode)
	}

	if mockLeadRepo.leads[leadID].Stage != domain.StageNew {
		t.Errorf("stage should not have been updated to invalid value")
	}
}

func TestLeadUpdateStage_AllValidStages(t *testing.T) {
	validStages := []domain.LeadStage{
		domain.StageNew,
		domain.StageContacted,
		domain.StageQualified,
		domain.StageProposal,
		domain.StageWon,
		domain.StageLost,
	}

	for _, stage := range validStages {
		app := fiber.New()
		mockLeadRepo := NewMockLeadRepo()
		mockContactRepo := NewMockContactRepo()
		hub := ws.NewHub()

		leadID := uuid.New()
		lead := &domain.Lead{ID: leadID, Stage: domain.StageNew}
		mockLeadRepo.leads[leadID] = lead

		handler := newTestLeadHandler(mockLeadRepo, mockContactRepo, hub)
		app.Patch("/leads/:id/stage", handler.UpdateStage)

		body := bytes.NewReader([]byte(`{"stage":"` + string(stage) + `"}`))
		req := httptest.NewRequest(http.MethodPatch, "/leads/"+leadID.String()+"/stage", body)
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("stage '%s' should be valid, got status %d", stage, resp.StatusCode)
		}
	}
}

func TestLeadUpdateStage_MissingStage(t *testing.T) {
	app := fiber.New()
	mockLeadRepo := NewMockLeadRepo()
	mockContactRepo := NewMockContactRepo()
	hub := ws.NewHub()

	leadID := uuid.New()
	lead := &domain.Lead{ID: leadID, Stage: domain.StageNew}
	mockLeadRepo.leads[leadID] = lead

	handler := newTestLeadHandler(mockLeadRepo, mockContactRepo, hub)
	app.Patch("/leads/:id/stage", handler.UpdateStage)

	body := bytes.NewReader([]byte(`{}`))
	req := httptest.NewRequest(http.MethodPatch, "/leads/"+leadID.String()+"/stage", body)
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 when stage is missing, got %d", resp.StatusCode)
	}
}

func TestLeadUpdateStage_InvalidID(t *testing.T) {
	app := fiber.New()
	mockLeadRepo := NewMockLeadRepo()
	mockContactRepo := NewMockContactRepo()
	hub := ws.NewHub()

	handler := newTestLeadHandler(mockLeadRepo, mockContactRepo, hub)
	app.Patch("/leads/:id/stage", handler.UpdateStage)

	body := bytes.NewReader([]byte(`{"stage":"contacted"}`))
	req := httptest.NewRequest(http.MethodPatch, "/leads/invalid-id/stage", body)
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid ID, got %d", resp.StatusCode)
	}
}
