package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/maidulcu/masaar-crm/internal/ai"
	"github.com/maidulcu/masaar-crm/internal/domain"
	"github.com/maidulcu/masaar-crm/internal/repo"
)

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

	lead, err := h.leads.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "lead not found"})
	}

	contact, err := h.contacts.GetByID(c.Context(), lead.ContactID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "contact not found"})
	}

	result, err := h.ollama.ScoreLead(c.Context(), contact.FullName, lead.Notes, string(lead.Source))
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

	msgs, err := h.wa.GetMessages(c.Context(), threadID, 20)
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

	threads, err := h.wa.ListThreads(c.Context(), "", 1, 1)
	if err != nil || len(threads) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "thread not found"})
	}

	summary, err := h.ollama.SummarizeThread(c.Context(), bodies)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "AI service unavailable"})
	}

	contact, _ := h.contacts.GetByID(c.Context(), threads[0].ContactID)

	lang := "en"
	name := ""
	if contact != nil {
		lang = contact.Language
		name = contact.FullName
	}

	draft, err := h.ollama.DraftReply(c.Context(), name, lang, summary)
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

	msgs, err := h.wa.GetMessages(c.Context(), threadID, 50)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "thread not found"})
	}

	var bodies []string
	for _, m := range msgs {
		bodies = append(bodies, m.Body)
	}

	summary, err := h.ollama.SummarizeThread(c.Context(), bodies)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "AI service unavailable"})
	}

	return c.JSON(fiber.Map{"summary": summary})
}
