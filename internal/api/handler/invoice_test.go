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
)

// MockInvoiceRepo implements InvoiceRepository for testing.
type MockInvoiceRepo struct {
	invoices map[uuid.UUID]*domain.VATInvoice
	counter  int
}

func NewMockInvoiceRepo() *MockInvoiceRepo {
	return &MockInvoiceRepo{
		invoices: make(map[uuid.UUID]*domain.VATInvoice),
		counter:  1000,
	}
}

func (m *MockInvoiceRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.VATInvoice, error) {
	if inv, ok := m.invoices[id]; ok {
		return inv, nil
	}
	return nil, fiber.NewError(fiber.StatusNotFound, "invoice not found")
}

func (m *MockInvoiceRepo) Create(ctx context.Context, inv *domain.VATInvoice) error {
	if inv.ID == uuid.Nil {
		inv.ID = uuid.New()
	}
	m.invoices[inv.ID] = inv
	return nil
}

func (m *MockInvoiceRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.InvoiceStatus) error {
	if inv, ok := m.invoices[id]; ok {
		inv.Status = status
		return nil
	}
	return fiber.NewError(fiber.StatusNotFound, "invoice not found")
}

func (m *MockInvoiceRepo) NextInvoiceNo(ctx context.Context) (string, error) {
	m.counter++
	return "INV-2026-" + string(rune(m.counter)), nil
}

// MockDealRepo implements DealRepository for testing.
type MockDealRepo struct {
	deals map[uuid.UUID]*domain.Deal
}

func NewMockDealRepo() *MockDealRepo {
	return &MockDealRepo{
		deals: make(map[uuid.UUID]*domain.Deal),
	}
}

func (m *MockDealRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Deal, error) {
	if deal, ok := m.deals[id]; ok {
		return deal, nil
	}
	return nil, fiber.NewError(fiber.StatusNotFound, "deal not found")
}

// stubCompanySettings implements CompanySettingsRepository for testing.
type stubCompanySettings struct{}

func (s *stubCompanySettings) Get(ctx context.Context) (*domain.CompanySettings, error) {
	return &domain.CompanySettings{
		Name:            "Test Company",
		VATNumber:       "AE123456",
		BusinessAddress: "Dubai, UAE",
	}, nil
}

func TestInvoiceUpdateStatus_ValidStatus(t *testing.T) {
	app := fiber.New()
	mockInvoiceRepo := NewMockInvoiceRepo()
	mockDealRepo := NewMockDealRepo()

	invoiceID := uuid.New()
	invoice := &domain.VATInvoice{
		ID:     invoiceID,
		Status: domain.InvoiceDraft,
	}
	mockInvoiceRepo.invoices[invoiceID] = invoice

	handler := NewInvoiceHandler(mockInvoiceRepo, mockDealRepo, &stubCompanySettings{})
	app.Patch("/invoices/:id/status", handler.UpdateStatus)

	body := bytes.NewReader([]byte(`{"status":"sent"}`))
	req := httptest.NewRequest(http.MethodPatch, "/invoices/"+invoiceID.String()+"/status", body)
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	if mockInvoiceRepo.invoices[invoiceID].Status != domain.InvoiceSent {
		t.Errorf("expected status to be 'sent', got '%s'", mockInvoiceRepo.invoices[invoiceID].Status)
	}
}

func TestInvoiceUpdateStatus_InvalidStatus(t *testing.T) {
	app := fiber.New()
	mockInvoiceRepo := NewMockInvoiceRepo()
	mockDealRepo := NewMockDealRepo()

	invoiceID := uuid.New()
	invoice := &domain.VATInvoice{
		ID:     invoiceID,
		Status: domain.InvoiceDraft,
	}
	mockInvoiceRepo.invoices[invoiceID] = invoice

	handler := NewInvoiceHandler(mockInvoiceRepo, mockDealRepo, &stubCompanySettings{})
	app.Patch("/invoices/:id/status", handler.UpdateStatus)

	body := bytes.NewReader([]byte(`{"status":"invalid_status"}`))
	req := httptest.NewRequest(http.MethodPatch, "/invoices/"+invoiceID.String()+"/status", body)
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid status, got %d", resp.StatusCode)
	}

	if mockInvoiceRepo.invoices[invoiceID].Status != domain.InvoiceDraft {
		t.Errorf("status should not have been updated to invalid value")
	}
}

func TestInvoiceUpdateStatus_AllValidStatuses(t *testing.T) {
	validStatuses := []domain.InvoiceStatus{
		domain.InvoiceDraft,
		domain.InvoiceSent,
		domain.InvoicePaid,
	}

	for _, status := range validStatuses {
		app := fiber.New()
		mockInvoiceRepo := NewMockInvoiceRepo()
		mockDealRepo := NewMockDealRepo()

		invoiceID := uuid.New()
		invoice := &domain.VATInvoice{ID: invoiceID, Status: domain.InvoiceDraft}
		mockInvoiceRepo.invoices[invoiceID] = invoice

		handler := NewInvoiceHandler(mockInvoiceRepo, mockDealRepo, &stubCompanySettings{})
		app.Patch("/invoices/:id/status", handler.UpdateStatus)

		body := bytes.NewReader([]byte(`{"status":"` + string(status) + `"}`))
		req := httptest.NewRequest(http.MethodPatch, "/invoices/"+invoiceID.String()+"/status", body)
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("status '%s' should be valid, got status %d", status, resp.StatusCode)
		}
	}
}

func TestInvoiceUpdateStatus_MissingStatus(t *testing.T) {
	app := fiber.New()
	mockInvoiceRepo := NewMockInvoiceRepo()
	mockDealRepo := NewMockDealRepo()

	invoiceID := uuid.New()
	invoice := &domain.VATInvoice{ID: invoiceID, Status: domain.InvoiceDraft}
	mockInvoiceRepo.invoices[invoiceID] = invoice

	handler := NewInvoiceHandler(mockInvoiceRepo, mockDealRepo, &stubCompanySettings{})
	app.Patch("/invoices/:id/status", handler.UpdateStatus)

	body := bytes.NewReader([]byte(`{}`))
	req := httptest.NewRequest(http.MethodPatch, "/invoices/"+invoiceID.String()+"/status", body)
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 when status is missing, got %d", resp.StatusCode)
	}
}

func TestInvoiceUpdateStatus_InvalidID(t *testing.T) {
	app := fiber.New()
	mockInvoiceRepo := NewMockInvoiceRepo()
	mockDealRepo := NewMockDealRepo()

	handler := NewInvoiceHandler(mockInvoiceRepo, mockDealRepo, &stubCompanySettings{})
	app.Patch("/invoices/:id/status", handler.UpdateStatus)

	body := bytes.NewReader([]byte(`{"status":"sent"}`))
	req := httptest.NewRequest(http.MethodPatch, "/invoices/invalid-id/status", body)
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid ID, got %d", resp.StatusCode)
	}
}
