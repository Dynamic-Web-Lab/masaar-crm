package handler

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/maidulcu/masaar-crm/internal/ai"
	"github.com/maidulcu/masaar-crm/internal/domain"
	"github.com/maidulcu/masaar-crm/internal/repo"
)

// ollamaTimeout caps Ollama inference calls. Model inference can legitimately
// take 30-60 s on CPU; 90 s is generous but prevents goroutine leaks.
const ollamaTimeout = 90 * time.Second

type AIHandler struct {
	ollama   *ai.Client
	contacts *repo.ContactRepo
	leads    *repo.LeadRepo
	wa       *repo.WhatsAppRepo
}

func NewAIHandler(ollama *ai.Client, contacts *repo.ContactRepo, leads *repo.LeadRepo, wa *repo.WhatsAppRepo) *AIHandler {
	return &AIHandler{ollama: ollama, contacts: contacts, leads: leads, wa: wa}
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

	result, err := h.ollama.ScoreLead(ctx, contact.FullName, lead.Notes, string(lead.Source))
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "AI service unavailable"})
	}

	return c.JSON(fiber.Map{"result": result})
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

	threads, err := h.wa.ListThreads(ctx, "", 1, 1)
	if err != nil || len(threads) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "thread not found"})
	}

	summary, err := h.ollama.SummarizeThread(ctx, bodies)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "AI service unavailable"})
	}

	contact, _ := h.contacts.GetByID(ctx, threads[0].ContactID)

	lang := "en"
	name := ""
	if contact != nil {
		lang = contact.Language
		name = contact.FullName
	}

	draft, err := h.ollama.DraftReply(ctx, name, lang, summary)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "AI service unavailable"})
	}

	return c.JSON(fiber.Map{
		"draft":   draft,
		"summary": summary,
	})
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

	summary, err := h.ollama.SummarizeThread(ctx, bodies)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "AI service unavailable"})
	}

	return c.JSON(fiber.Map{"summary": summary})
}

// POST /api/v1/ai/reply-suggestions/:thread_id
func (h *AIHandler) ReplySuggestions(c *fiber.Ctx) error {
	threadID, err := uuid.Parse(c.Params("thread_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid thread_id"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), ollamaTimeout)
	defer cancel()

	thread, err := h.wa.GetThread(ctx, threadID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "thread not found"})
	}

	msgs, err := h.wa.GetMessages(ctx, threadID, 20)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "messages not found"})
	}

	var bodies []string
	for _, m := range msgs {
		prefix := "Agent"
		if m.Direction == domain.DirectionInbound {
			prefix = "Customer"
		}
		bodies = append(bodies, prefix+": "+m.Body)
	}

	contactName := ""
	if thread.Contact != nil {
		contactName = thread.Contact.FullName
	}

	raw, err := h.ollama.ReplySuggestions(ctx, contactName, bodies)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "AI service unavailable"})
	}

	// Extract the JSON object from the response (model may add preamble text)
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start != -1 && end != -1 && end > start {
		raw = raw[start : end+1]
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return c.JSON(fiber.Map{"raw": raw})
	}
	return c.JSON(parsed)
}

// POST /api/v1/ai/extract-buyer-profile/:thread_id
func (h *AIHandler) ExtractBuyerProfile(c *fiber.Ctx) error {
	threadID, err := uuid.Parse(c.Params("thread_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid thread_id"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), ollamaTimeout)
	defer cancel()

	thread, err := h.wa.GetThread(ctx, threadID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "thread not found"})
	}

	msgs, err := h.wa.GetMessages(ctx, threadID, 40)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "messages not found"})
	}

	var bodies []string
	for _, m := range msgs {
		prefix := "Agent"
		if m.Direction == domain.DirectionInbound {
			prefix = "Customer"
		}
		bodies = append(bodies, prefix+": "+m.Body)
	}

	contactName := ""
	if thread.Contact != nil {
		contactName = thread.Contact.FullName
	}

	raw, err := h.ollama.ExtractBuyerProfile(ctx, contactName, bodies)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "AI service unavailable"})
	}

	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start != -1 && end != -1 && end > start {
		raw = raw[start : end+1]
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return c.JSON(fiber.Map{"raw": raw})
	}
	return c.JSON(parsed)
}

// POST /api/v1/ai/translate
func (h *AIHandler) Translate(c *fiber.Ctx) error {
	var body struct {
		Text string `json:"text"`
		To   string `json:"to"`
	}
	if err := c.BodyParser(&body); err != nil || body.Text == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "text and to (ar|en) are required"})
	}
	if body.To != "ar" && body.To != "en" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "to must be 'ar' or 'en'"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), ollamaTimeout)
	defer cancel()

	translated, err := h.ollama.Translate(ctx, body.Text, body.To)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "AI service unavailable"})
	}

	from := "en"
	if body.To == "en" {
		from = "ar"
	}
	return c.JSON(fiber.Map{
		"translated": strings.TrimSpace(translated),
		"from":       from,
		"to":         body.To,
	})
}
