package ai

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/maidulcu/masaar-crm/internal/domain"
	"github.com/maidulcu/masaar-crm/internal/repo"
)

type TaggingService struct {
	aiClient  *Client
	leadRepo  *repo.LeadRepo
	waRepo    *repo.WhatsAppRepo
	tagRepo   *repo.LeadTagRepo
	contactRepo *repo.ContactRepo
}

func NewTaggingService(
	aiClient *Client,
	leadRepo *repo.LeadRepo,
	waRepo *repo.WhatsAppRepo,
	tagRepo *repo.LeadTagRepo,
	contactRepo *repo.ContactRepo,
) *TaggingService {
	return &TaggingService{
		aiClient:    aiClient,
		leadRepo:    leadRepo,
		waRepo:      waRepo,
		tagRepo:     tagRepo,
		contactRepo: contactRepo,
	}
}

// AutoTagFromMessage analyzes a message and applies tags to related leads
func (s *TaggingService) AutoTagFromMessage(ctx context.Context, contactID uuid.UUID, messageBody string) error {
	if s.aiClient == nil || messageBody == "" {
		return nil
	}

	// Find leads linked to this contact
	// For now, we'll find the most recent lead for this contact
	// In the future, this could be more sophisticated (multiple leads per contact)
	boards, err := s.leadRepo.KanbanBoard(ctx)
	if err != nil {
		log.Printf("tagging: error fetching leads: %v", err)
		return nil
	}

	var targetLead *domain.Lead
	for _, stage := range boards {
		for _, lead := range stage {
			if lead.ContactID == contactID {
				// Use the most recent lead
				if targetLead == nil || lead.CreatedAt.After(targetLead.CreatedAt) {
					l := lead // Create a copy to avoid pointer issues
					targetLead = &l
				}
			}
		}
	}

	// No lead found for this contact, skip tagging
	if targetLead == nil {
		return nil
	}

	// Parse intent from message
	intentJSON, err := s.aiClient.ParseIntent(ctx, messageBody)
	if err != nil {
		log.Printf("tagging: intent parsing error: %v", err)
		return nil
	}

	var intent map[string]interface{}
	if err := json.Unmarshal([]byte(intentJSON), &intent); err != nil {
		return nil
	}

	// Apply tags based on intent
	tags := s.extractTagsFromIntent(intent)

	// Also apply tags based on urgency signal from message
	qualityTags := s.extractQualityTags(messageBody, intent)

	// Combine all tags
	for category, tagList := range tags {
		for _, tag := range tagList {
			_ = s.tagRepo.AddTag(ctx, targetLead.ID, tag, category, true, nil)
		}
	}

	for category, tagList := range qualityTags {
		for _, tag := range tagList {
			_ = s.tagRepo.AddTag(ctx, targetLead.ID, tag, category, true, nil)
		}
	}

	return nil
}

// extractTagsFromIntent extracts tags from parsed intent
func (s *TaggingService) extractTagsFromIntent(intent map[string]interface{}) map[string][]string {
	tags := make(map[string][]string)

	// Extract property type interest
	if propType, ok := intent["property_type"].(string); ok && propType != "" {
		if tags["interest"] == nil {
			tags["interest"] = []string{}
		}
		tags["interest"] = append(tags["interest"], propType)
	}

	// Extract area/location interest
	if area, ok := intent["area"].(string); ok && area != "" {
		if tags["interest"] == nil {
			tags["interest"] = []string{}
		}
		tags["interest"] = append(tags["interest"], area)
	}

	// Extract timeline urgency
	if urgency, ok := intent["urgency"].(string); ok && urgency != "" {
		if urgency == "urgent" {
			if tags["timeline"] == nil {
				tags["timeline"] = []string{}
			}
			tags["timeline"] = append(tags["timeline"], "urgent")
		} else if urgency == "flexible" {
			if tags["timeline"] == nil {
				tags["timeline"] = []string{}
			}
			tags["timeline"] = append(tags["timeline"], "flexible")
		}
	}

	// Extract transaction type (buy/sell/rent)
	if txnType, ok := intent["transaction_type"].(string); ok && txnType != "" {
		if tags["interest"] == nil {
			tags["interest"] = []string{}
		}
		tags["interest"] = append(tags["interest"], txnType)
	}

	// Extract budget segment
	if budget, ok := intent["budget_max"].(float64); ok && budget > 0 {
		segment := s.getBudgetSegment(budget)
		if segment != "" {
			if tags["budget"] == nil {
				tags["budget"] = []string{}
			}
			tags["budget"] = append(tags["budget"], segment)
		}
	}

	return tags
}

// extractQualityTags applies quality signals based on message content
func (s *TaggingService) extractQualityTags(message string, intent map[string]interface{}) map[string][]string {
	tags := make(map[string][]string)

	// Message indicates active interest = warm/hot lead
	if len(message) > 50 {
		// Longer messages suggest serious inquiry
		if tags["quality"] == nil {
			tags["quality"] = []string{}
		}
		tags["quality"] = append(tags["quality"], "warm")
	}

	// Multiple specific details = hot lead
	detailCount := 0
	if _, ok := intent["property_type"].(string); ok {
		detailCount++
	}
	if _, ok := intent["area"].(string); ok {
		detailCount++
	}
	if _, ok := intent["budget_max"].(float64); ok {
		detailCount++
	}

	if detailCount >= 2 {
		if tags["quality"] == nil {
			tags["quality"] = []string{}
		}
		// Replace warm with hot if we have multiple signals
		tags["quality"] = []string{"hot"}
	}

	return tags
}

// getBudgetSegment categorizes budget into segments
func (s *TaggingService) getBudgetSegment(budgetAED float64) string {
	if budgetAED < 1000000 {
		return "under-1M"
	} else if budgetAED < 2000000 {
		return "1M-2M"
	} else if budgetAED < 3000000 {
		return "2M-3M"
	} else {
		return "3M+"
	}
}
