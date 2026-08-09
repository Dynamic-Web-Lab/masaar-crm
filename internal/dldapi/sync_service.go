package dldapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/dynamicweblab/masaar-crm/internal/repo"
	"github.com/dynamicweblab/masaar-crm/internal/ws"
)

// SyncService imports DLD marketplace data (listings + inquiries) into Masaar.
type SyncService struct {
	dldRepo   *repo.DLDIntegrationRepo
	contactRepo *repo.ContactRepo
	hub         *ws.Hub
}

func NewSyncService(
	dldRepo *repo.DLDIntegrationRepo,
	contactRepo *repo.ContactRepo,
	hub *ws.Hub,
) *SyncService {
	return &SyncService{
		dldRepo:   dldRepo,
		contactRepo: contactRepo,
		hub:         hub,
	}
}

// SyncResult holds counts from a sync operation.
type SyncResult struct {
	ListingsImported int
	ListingsUpdated  int
	InquiriesCreated int
}

// SyncAll runs a full delta sync for a company: listings updated since lastSyncAt
// and inquiries created since lastSyncAt.
// Pass a zero time.Time to do a full initial load.
func (s *SyncService) SyncAll(ctx context.Context, companyID uuid.UUID, apiKey string, lastSyncAt *time.Time) (*SyncResult, error) {
	client := NewIntegrationClient(apiKey)
	result := &SyncResult{}

	var sinceStr string
	if lastSyncAt != nil && !lastSyncAt.IsZero() {
		sinceStr = lastSyncAt.UTC().Format(time.RFC3339)
	}

	// Sync listings
	li, lu, err := s.syncListings(ctx, client, companyID, sinceStr)
	if err != nil {
		log.Printf("[DLD] company %s: listings sync error: %v", companyID, err)
		// don't abort — still attempt inquiries
	}
	result.ListingsImported = li
	result.ListingsUpdated = lu

	// Sync inquiries
	ic, err := s.syncInquiries(ctx, client, companyID, sinceStr)
	if err != nil {
		log.Printf("[DLD] company %s: inquiries sync error: %v", companyID, err)
	}
	result.InquiriesCreated = ic

	return result, nil
}

// syncListings pages through the DLD listing feed and upserts each into Masaar.
func (s *SyncService) syncListings(ctx context.Context, client *IntegrationClient, companyID uuid.UUID, updatedSince string) (imported, updated int, err error) {
	page := 1
	for {
		resp, e := client.FetchListings(ctx, "published", updatedSince, page, 100)
		if e != nil {
			return imported, updated, fmt.Errorf("page %d: %w", page, e)
		}
		for _, l := range resp.Data {
			_, upsertErr := s.dldRepo.UpsertListing(ctx, companyID,
				l.UUID,
				l.Title,
				l.Description,
				mapPropertyType(l.Category.Slug),
				"sale", // DLD listings are for sale by default
				l.City.Slug,
				mapListingStatus(l.Status),
				imageMediumURL(l.PrimaryImage),
				l.Currency,
				l.Price,
			)
			if upsertErr != nil {
				log.Printf("[DLD] upsert listing %s: %v", l.UUID, upsertErr)
				continue
			}
			imported++
		}

		if resp.Meta.CurrentPage >= resp.Meta.LastPage {
			break
		}
		page++
	}
	return imported, updated, nil
}

// syncInquiries pages through the DLD inquiry feed and creates contact + lead for each.
func (s *SyncService) syncInquiries(ctx context.Context, client *IntegrationClient, companyID uuid.UUID, createdSince string) (created int, err error) {
	page := 1
	for {
		resp, e := client.FetchInquiries(ctx, createdSince, page, 100)
		if e != nil {
			return created, fmt.Errorf("page %d: %w", page, e)
		}
		for _, inq := range resp.Data {
			ok, e2 := s.processInquiry(ctx, companyID, &inq)
			if e2 != nil {
				log.Printf("[DLD] process inquiry %d: %v", inq.ID, e2)
				continue
			}
			if ok {
				created++
			}
		}

		if resp.Meta.CurrentPage >= resp.Meta.LastPage {
			break
		}
		page++
	}
	return created, nil
}

// ProcessEvent handles a real-time DLD webhook event.
func (s *SyncService) ProcessEvent(ctx context.Context, companyID uuid.UUID, event *DLDWebhookEvent) error {
	switch event.Event {
	case "inquiry.created":
		var inq DLDInquiry
		if err := json.Unmarshal(event.Data, &inq); err != nil {
			return fmt.Errorf("parse inquiry payload: %w", err)
		}
		_, err := s.processInquiry(ctx, companyID, &inq)
		return err

	case "listing.created", "listing.updated", "listing.published":
		var listing DLDListing
		if err := json.Unmarshal(event.Data, &listing); err != nil {
			return fmt.Errorf("parse listing payload: %w", err)
		}
		_, err := s.dldRepo.UpsertListing(ctx, companyID,
			listing.UUID,
			listing.Title,
			listing.Description,
			mapPropertyType(listing.Category.Slug),
			"sale",
			listing.City.Slug,
			mapListingStatus(listing.Status),
			imageMediumURL(listing.PrimaryImage),
			listing.Currency,
			listing.Price,
		)
		return err

	case "listing.deactivated", "listing.deleted":
		if event.Resource.UUID != "" {
			return s.dldRepo.DeactivateListing(ctx, companyID, event.Resource.UUID)
		}

	default:
		log.Printf("[DLD] unknown event type: %s", event.Event)
	}
	return nil
}

// processInquiry creates a contact (upsert by phone) and a lead for a DLD inquiry.
// Returns true if a new lead was created, false if it was a duplicate.
func (s *SyncService) processInquiry(ctx context.Context, companyID uuid.UUID, inq *DLDInquiry) (bool, error) {
	phone := sanitizePhone(inq.BuyerPhone)
	if phone == "" {
		return false, nil // no phone — skip
	}

	// Upsert contact by phone number (reuse existing ContactRepo.Upsert)
	contact, err := s.contactRepo.Upsert(ctx, phone, inq.BuyerName)
	if err != nil {
		return false, fmt.Errorf("upsert contact: %w", err)
	}

	// Update email if provided and contact doesn't already have one
	if inq.BuyerEmail != "" && contact.Email == "" {
		_ = s.dldRepo.UpdateContactEmailIfEmpty(ctx, contact.ID, inq.BuyerEmail)
	}

	// Build notes from inquiry message + listing reference
	notes := inq.Message
	if inq.Listing.Title != "" {
		notes = fmt.Sprintf("DLD inquiry on: %s\n\n%s", inq.Listing.Title, inq.Message)
	}

	// Create lead (idempotent — ON CONFLICT dld_inquiry_id DO NOTHING)
	created, leadID, err := s.dldRepo.CreateLeadFromInquiry(ctx, contact.ID, inq.ID, notes)
	if err != nil {
		return false, fmt.Errorf("create lead: %w", err)
	}

	// Broadcast real-time notification for new leads
	if created && s.hub != nil {
		s.hub.Broadcast(ws.Event{
			Type: "lead.created",
			Payload: map[string]interface{}{
				"lead_id":    leadID,
				"contact":    contact.FullName,
				"phone":      contact.PhoneWA,
				"source":     "dld",
				"company_id": companyID,
			},
		})
	}

	return created, nil
}

// ── Mapping helpers ───────────────────────────────────────────────────────────

// mapPropertyType converts DLD category slug to Masaar property_type.
func mapPropertyType(categorySlug string) string {
	slug := strings.ToLower(categorySlug)
	switch {
	case strings.Contains(slug, "apartment") || strings.Contains(slug, "flat"):
		return "apartment"
	case strings.Contains(slug, "villa"):
		return "villa"
	case strings.Contains(slug, "townhouse"):
		return "townhouse"
	case strings.Contains(slug, "commercial") || strings.Contains(slug, "office"):
		return "commercial"
	default:
		return "apartment" // safe default
	}
}

// mapListingStatus converts DLD listing status to Masaar listing status.
func mapListingStatus(dldStatus string) string {
	switch strings.ToLower(dldStatus) {
	case "published", "active":
		return "active"
	case "pending":
		return "draft"
	case "inactive", "expired", "deleted":
		return "inactive"
	default:
		return "draft"
	}
}

// imageMediumURL returns the medium URL from a DLD primary image, or empty string.
func imageMediumURL(img *DLDImage) string {
	if img == nil {
		return ""
	}
	if img.URLMedium != "" {
		return img.URLMedium
	}
	return img.URL
}

// sanitizePhone ensures the phone number is non-empty and reasonably formatted.
func sanitizePhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return ""
	}
	// Ensure it starts with + for international format
	if !strings.HasPrefix(phone, "+") {
		phone = "+" + phone
	}
	return phone
}
