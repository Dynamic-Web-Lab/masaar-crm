package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/maidulcu/masaar-crm/internal/ai"
	"github.com/maidulcu/masaar-crm/internal/domain"
	"github.com/maidulcu/masaar-crm/internal/repo"
	"github.com/maidulcu/masaar-crm/internal/ws"
)

type LeadHandler struct {
	leads          *repo.LeadRepo
	contacts       *repo.ContactRepo
	commHistRepo   *repo.CommunicationHistoryRepo
	scoringService *ai.ScoringService
	hub            *ws.Hub
	audit          *repo.AuditLogRepo
}

func NewLeadHandler(leads *repo.LeadRepo, contacts *repo.ContactRepo, commHistRepo *repo.CommunicationHistoryRepo, scoringService *ai.ScoringService, hub *ws.Hub, audit *repo.AuditLogRepo) *LeadHandler {
	return &LeadHandler{leads: leads, contacts: contacts, commHistRepo: commHistRepo, scoringService: scoringService, hub: hub, audit: audit}
}

// KanbanBoard godoc
// @Summary      Get Kanban board
// @Description  Returns all leads grouped by stage for the Kanban pipeline view.
// @Tags         Leads
// @Produce      json
// @Success      200  {object}  object  "Map of stage → []Lead"
// @Security     BearerAuth
// @Router       /leads [get]
func (h *LeadHandler) KanbanBoard(c *fiber.Ctx) error {
	board, err := h.leads.KanbanBoard(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(board)
}

// Create godoc
// @Summary      Create lead
// @Description  Creates a new lead and broadcasts a lead.created WebSocket event.
// @Tags         Leads
// @Accept       json
// @Produce      json
// @Param        body  body      domain.Lead  true  "Lead payload (contact_id required)"
// @Success      201   {object}  domain.Lead
// @Failure      400   {object}  object{error=string}
// @Security     BearerAuth
// @Router       /leads [post]
func (h *LeadHandler) Create(c *fiber.Ctx) error {
	var lead domain.Lead
	if err := c.BodyParser(&lead); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if lead.ContactID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "contact_id is required"})
	}
	if lead.Stage == "" {
		lead.Stage = domain.StageNew
	}
	if lead.Currency == "" {
		lead.Currency = "AED"
	}

	if err := h.leads.Create(c.Context(), &lead); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	h.hub.Broadcast(ws.Event{
		Type:    "lead.created",
		Payload: lead,
	})

	actorID := c.Locals("user_id").(uuid.UUID)
	h.audit.Log(c.Context(), actorID, repo.AuditCreate, repo.AuditLead, lead.ID, lead)
	return c.Status(fiber.StatusCreated).JSON(lead)
}

// UpdateStage godoc
// @Summary      Move lead to stage
// @Description  Updates the pipeline stage of a lead (drag-drop). Broadcasts lead.stage_changed event.
// @Tags         Leads
// @Accept       json
// @Produce      json
// @Param        id    path      string                    true  "Lead UUID"
// @Param        body  body      object{stage=string}      true  "New stage: new|contacted|qualified|proposal|won|lost"
// @Success      200   {object}  object{lead_id=string,stage=string}
// @Failure      400   {object}  object{error=string}
// @Security     BearerAuth
// @Router       /leads/{id}/stage [patch]
func (h *LeadHandler) UpdateStage(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	var body struct {
		Stage domain.LeadStage `json:"stage"`
	}
	if err := c.BodyParser(&body); err != nil || body.Stage == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "stage is required"})
	}

	validStages := map[domain.LeadStage]bool{
		domain.StageNew:       true,
		domain.StageContacted: true,
		domain.StageQualified: true,
		domain.StageProposal:  true,
		domain.StageWon:       true,
		domain.StageLost:      true,
	}
	if !validStages[body.Stage] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid stage"})
	}

	if err := h.leads.UpdateStage(c.Context(), id, body.Stage); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Update lead score based on stage progression
	if h.scoringService != nil {
		h.scoringService.UpdateScoreOnStageChange(c.Context(), id, body.Stage)
	}

	h.hub.Broadcast(ws.Event{
		Type: "lead.stage_changed",
		Payload: fiber.Map{
			"lead_id": id,
			"stage":   body.Stage,
		},
	})

	return c.JSON(fiber.Map{"lead_id": id, "stage": body.Stage})
}

// UpdateNotes godoc
// @Summary      Update lead notes
// @Description  Replaces the notes text on a lead. Agents and admins only.
// @Tags         Leads
// @Accept       json
// @Produce      json
// @Param        id    path      string               true  "Lead UUID"
// @Param        body  body      object{notes=string} true  "Notes"
// @Success      200   {object}  object{lead_id=string,notes=string}
// @Failure      400   {object}  object{error=string}
// @Security     BearerAuth
// @Router       /leads/{id}/notes [patch]
func (h *LeadHandler) UpdateNotes(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	var body struct {
		Notes string `json:"notes"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	if err := h.leads.UpdateNotes(c.Context(), id, body.Notes); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"lead_id": id, "notes": body.Notes})
}

// Get godoc
// @Summary      Get lead
// @Description  Returns a single lead with its contact joined.
// @Tags         Leads
// @Produce      json
// @Param        id  path      string  true  "Lead UUID"
// @Success      200  {object}  domain.Lead
// @Failure      400  {object}  object{error=string}
// @Failure      404  {object}  object{error=string}
// @Security     BearerAuth
// @Router       /leads/{id} [get]
func (h *LeadHandler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	lead, err := h.leads.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "lead not found"})
	}

	// Join contact
	contact, _ := h.contacts.GetByID(c.Context(), lead.ContactID)
	lead.Contact = contact

	return c.JSON(lead)
}

// GetCommunications godoc
// @Summary      Get lead communications
// @Description  Returns all communications (WhatsApp, email, calls) for a lead, ordered by date.
// @Tags         Leads
// @Produce      json
// @Param        id    path      string  true  "Lead UUID"
// @Param        limit query     int     false  "Max communications to return (default 100)"
// @Success      200   {array}   domain.CommunicationHistory
// @Failure      400   {object}  object{error=string}
// @Security     BearerAuth
// @Router       /leads/{id}/communications [get]
func (h *LeadHandler) GetCommunications(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	comms, err := h.commHistRepo.GetByLead(c.Context(), id, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(comms)
}
