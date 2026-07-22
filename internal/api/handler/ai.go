package handler

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/dynamicweblab/masaar-crm/internal/ai"
	"github.com/dynamicweblab/masaar-crm/internal/domain"
	"github.com/dynamicweblab/masaar-crm/internal/repo"
)

// ollamaTimeout caps local inference calls (CPU can be slow, 90 s is generous).
const ollamaTimeout = 90 * time.Second

// cloudTimeout caps cloud API calls — they're fast, fail quickly if down.
const cloudTimeout = 15 * time.Second

type AIHandler struct {
	// sensitive handles anything containing customer PII — always local Ollama.
	sensitive *ai.Client
	// cloud handles non-PII tasks (property copy, market summaries).
	// May be Gemini or fall back to Ollama if Gemini is not configured.
	cloud    *ai.Client
	contacts *repo.ContactRepo
	leads    *repo.LeadRepo
	wa       *repo.WhatsAppRepo
}

func NewAIHandler(sensitive, cloud *ai.Client, contacts *repo.ContactRepo, leads *repo.LeadRepo, wa *repo.WhatsAppRepo) *AIHandler {
	return &AIHandler{sensitive: sensitive, cloud: cloud, contacts: contacts, leads: leads, wa: wa}
}

// POST /api/v1/ai/score-lead/:id
func (h *AIHandler) ScoreLead(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), ollamaTimeout)
	defer cancel()

	lead, err := h.leads.GetByID(ctx, id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "lead not found"})
	}

	contact, err := h.contacts.GetByID(ctx, lead.ContactID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "contact not found"})
	}

	result, err := h.sensitive.ScoreLead(ctx, contact.FullName, lead.Notes, string(lead.Source))
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "AI service unavailable"})
	}

	return c.JSON(fiber.Map{"result": result})
}

// POST /api/v1/ai/score-contact/:id
// Scores a contact using their most recent lead, persists the score, and returns it.
func (h *AIHandler) ScoreContact(c *fiber.Ctx) error {
	contactID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), ollamaTimeout)
	defer cancel()

	contact, err := h.contacts.GetByID(ctx, contactID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "contact not found"})
	}

	leads, err := h.leads.List(ctx, repo.LeadFilter{ContactID: &contactID})
	if err != nil || len(leads) == 0 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "no leads found for this contact"})
	}
	lead := leads[0]

	raw, err := h.sensitive.ScoreLead(ctx, contact.FullName, lead.Notes, string(lead.Source))
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "AI service unavailable"})
	}

	var parsed struct {
		Score     int    `json:"score"`
		Reasoning string `json:"reasoning"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil || parsed.Score < 0 || parsed.Score > 100 {
		return c.JSON(fiber.Map{"result": raw, "score": nil})
	}

	if err := h.contacts.UpdateScore(ctx, contactID, parsed.Score); err != nil {
		log.Printf("score-contact: failed to persist score: %v", err)
	}

	return c.JSON(fiber.Map{
		"score":     parsed.Score,
		"reasoning": parsed.Reasoning,
	})
}

// POST /api/v1/ai/draft-reply/:thread_id
func (h *AIHandler) DraftReply(c *fiber.Ctx) error {
	threadID, err := uuid.Parse(c.Params("thread_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid thread_id"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), ollamaTimeout)
	defer cancel()

	msgs, err := h.wa.GetMessages(ctx, threadID, 20)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "thread not found"})
	}

	var bodies []string
	for _, m := range msgs {
		prefix := "Agent"
		if m.Direction == domain.DirectionInbound {
			prefix = "Customer"
		}
		bodies = append(bodies, prefix+": "+m.Body)
	}

	thread, err := h.wa.GetThread(ctx, threadID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "thread not found"})
	}

	summary, err := h.sensitive.SummarizeThread(ctx, bodies)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "AI service unavailable"})
	}

	contact, _ := h.contacts.GetByID(ctx, thread.ContactID)

	lang := "en"
	name := ""
	if contact != nil {
		lang = contact.Language
		name = contact.FullName
	}

	draft, err := h.sensitive.DraftReply(ctx, name, lang, summary)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "AI service unavailable"})
	}

	return c.JSON(fiber.Map{
		"draft":   draft,
		"summary": summary,
	})
}

// POST /api/v1/ai/extract-buyer-profile/:thread_id
func (h *AIHandler) ExtractBuyerProfile(c *fiber.Ctx) error {
	threadID, err := uuid.Parse(c.Params("thread_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid thread_id"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), ollamaTimeout)
	defer cancel()

	msgs, err := h.wa.GetMessages(ctx, threadID, 50)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "thread not found"})
	}

	var bodies []string
	for _, m := range msgs {
		bodies = append(bodies, m.Body)
	}

	raw, err := h.sensitive.ExtractBuyerProfile(ctx, bodies)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "AI service unavailable"})
	}

	var profile map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &profile); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to parse profile"})
	}

	return c.JSON(profile)
}

// SummarizeThread godoc
// @Summary      AI Summarize thread
// @Description  Uses Ollama (local LLM) to generate a summary of the last 50 messages in a WhatsApp thread.
// @Tags         AI
// @Produce      json
// @Param        thread_id  path      string  true  "Thread UUID"
// @Success      200        {object}  object{summary=string}
// @Failure      400        {object}  object{error=string}
// @Failure      503        {object}  object{error=string}  "Ollama unavailable"
// @Security     BearerAuth
// @Router       /ai/summarize/{thread_id} [post]
func (h *AIHandler) SummarizeThread(c *fiber.Ctx) error {
	threadID, err := uuid.Parse(c.Params("thread_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid thread_id"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), ollamaTimeout)
	defer cancel()

	msgs, err := h.wa.GetMessages(ctx, threadID, 50)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "thread not found"})
	}

	var bodies []string
	for _, m := range msgs {
		bodies = append(bodies, m.Body)
	}

	summary, err := h.sensitive.SummarizeThread(ctx, bodies)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "AI service unavailable"})
	}

	// Persist summary to the thread so it shows in the inbox list
	if err := h.wa.UpdateAISummary(ctx, threadID, summary); err != nil {
		log.Printf("summarize: failed to persist summary: %v", err)
	}

	return c.JSON(fiber.Map{"summary": summary})
}

// DescribePropertyListing godoc
// @Summary      AI property listing description
// @Description  Generates professional marketing copy for a property listing using public specs only (no customer PII). Uses Gemini if configured, falls back to Ollama.
// @Tags         AI
// @Accept       json
// @Produce      json
// @Param        body  body  object{}  true  "Property details: area, property_type, bedrooms, size_sqft, amenities, lang"
// @Success      200   {object}  object{description=string,provider=string}
// @Failure      400   {object}  object{error=string}
// @Failure      503   {object}  object{error=string}
// @Security     BearerAuth
// @Router       /ai/describe-listing [post]
func (h *AIHandler) DescribePropertyListing(c *fiber.Ctx) error {
	if h.cloud == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "AI service unavailable"})
	}

	var req struct {
		Area         string   `json:"area"`
		PropertyType string   `json:"property_type"`
		Bedrooms     string   `json:"bedrooms"`
		SizeSqft     int      `json:"size_sqft"`
		Amenities    []string `json:"amenities"`
		Lang         string   `json:"lang"`
	}
	if err := c.BodyParser(&req); err != nil || req.Area == "" || req.PropertyType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "area and property_type are required"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), cloudTimeout)
	defer cancel()

	description, err := h.cloud.DescribePropertyListing(ctx, req.Area, req.PropertyType, req.Bedrooms, req.SizeSqft, req.Amenities, req.Lang)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "AI service unavailable"})
	}

	return c.JSON(fiber.Map{"description": description})
}
