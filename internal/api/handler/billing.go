package handler

import (
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/maidulcu/masaar-crm/internal/api/middleware"
	"github.com/maidulcu/masaar-crm/internal/billing"
	"github.com/maidulcu/masaar-crm/internal/repo"
	"github.com/stripe/stripe-go/v76"
)

type BillingHandler struct {
	billingRepo     *repo.BillingRepo
	settingsRepo    *repo.CompanySettingsRepo
	stripeConfig    *billing.StripeConfig
}

func NewBillingHandler(
	billingRepo *repo.BillingRepo,
	settingsRepo *repo.CompanySettingsRepo,
	stripeCfg *billing.StripeConfig,
) *BillingHandler {
	return &BillingHandler{
		billingRepo:  billingRepo,
		settingsRepo: settingsRepo,
		stripeConfig: stripeCfg,
	}
}

// GetBilling godoc
// @Summary      Current plan and usage
// @Description  Returns the company's active plan, monthly usage counters, and available plan options.
// @Tags         Billing
// @Produce      json
// @Success      200  {object}  object{}
// @Security     BearerAuth
// @Router       /billing [get]
func (h *BillingHandler) GetBilling(c *fiber.Ctx) error {
	ctx := c.Context()

	companyPlan, err := h.billingRepo.GetPlan(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load plan"})
	}

	companyIDStr, _ := c.Locals("company_id").(string)
	companyID, _ := uuid.Parse(companyIDStr)

	usage, err := h.billingRepo.GetUsage(ctx, companyID)
	if err != nil {
		usage = map[string]int{"bos24": 0, "ai": 0, "pdf": 0}
	}

	plan := billing.Get(companyPlan.Plan)

	// Build quota display (limit -1 shown as "unlimited")
	quotaDisplay := func(used, limit int) fiber.Map {
		if limit == 0 {
			return fiber.Map{"used": 0, "limit": 0, "available": false}
		}
		if limit == -1 {
			return fiber.Map{"used": used, "limit": -1, "available": true, "unlimited": true}
		}
		return fiber.Map{"used": used, "limit": limit, "available": true, "pct": used * 100 / limit}
	}

	// Build plans list for upgrade UI
	planList := make([]fiber.Map, 0, len(billing.Plans))
	for _, p := range billing.Plans {
		planList = append(planList, fiber.Map{
			"id":       p.ID,
			"name":     p.Name,
			"price":    p.PriceUSDMonth,
			"current":  p.ID == plan.ID,
			"features": p.Features,
			"quotas": fiber.Map{
				"bos24_monthly": p.Quotas.BOS24Monthly,
				"ai_monthly":    p.Quotas.AIMonthly,
				"pdf_monthly":   p.Quotas.PDFMonthly,
			},
		})
	}

	return c.JSON(fiber.Map{
		"plan": fiber.Map{
			"id":             plan.ID,
			"name":           plan.Name,
			"price":          plan.PriceUSDMonth,
			"started_at":     companyPlan.PlanStartedAt,
			"expires_at":     companyPlan.PlanExpiresAt,
			"stripe_enabled": h.stripeConfig.IsEnabled(),
			"has_sub":        companyPlan.StripeSubID != "",
		},
		"usage": fiber.Map{
			"bos24": quotaDisplay(usage["bos24"], plan.Quotas.BOS24Monthly),
			"ai":    quotaDisplay(usage["ai"], plan.Quotas.AIMonthly),
			"pdf":   quotaDisplay(usage["pdf"], plan.Quotas.PDFMonthly),
			"reset": nextMonthReset(),
		},
		"plans": planList,
	})
}

// CreateCheckout godoc
// @Summary      Start plan upgrade (Stripe Checkout)
// @Description  Creates a Stripe Checkout session and returns the redirect URL.
// @Tags         Billing
// @Accept       json
// @Produce      json
// @Param        body  body  object{plan=string}  true  "Target plan ID"
// @Success      200   {object}  object{checkout_url=string}
// @Failure      400   {object}  object{error=string}
// @Failure      503   {object}  object{error=string}
// @Security     BearerAuth
// @Router       /billing/checkout [post]
func (h *BillingHandler) CreateCheckout(c *fiber.Ctx) error {
	if !h.stripeConfig.IsEnabled() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "payment processing not configured — contact your administrator",
		})
	}

	var body struct {
		Plan string `json:"plan"`
	}
	if err := c.BodyParser(&body); err != nil || body.Plan == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "plan is required"})
	}

	targetPlan := billing.Get(body.Plan)
	if targetPlan.ID == billing.PlanCommunity {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "community plan is free — no checkout needed"})
	}

	ctx := c.Context()

	companyPlan, err := h.billingRepo.GetPlan(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load billing state"})
	}

	settings, err := h.settingsRepo.Get(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load company settings"})
	}

	// Get or create Stripe customer
	customerID, err := billing.GetOrCreateCustomer(settings.Name, settings.BusinessEmail, companyPlan.StripeCustomer)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "payment setup failed"})
	}

	companyIDStr, _ := c.Locals("company_id").(string)
	url, err := billing.CreateCheckoutSession(h.stripeConfig, targetPlan, companyIDStr, customerID)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "failed to create checkout session"})
	}

	return c.JSON(fiber.Map{"checkout_url": url})
}

// CreatePortal godoc
// @Summary      Open Stripe billing portal
// @Description  Creates a Stripe Customer Portal session for managing subscription and invoices.
// @Tags         Billing
// @Produce      json
// @Success      200  {object}  object{portal_url=string}
// @Failure      503  {object}  object{error=string}
// @Security     BearerAuth
// @Router       /billing/portal [post]
func (h *BillingHandler) CreatePortal(c *fiber.Ctx) error {
	if !h.stripeConfig.IsEnabled() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "payment processing not configured",
		})
	}

	companyPlan, err := h.billingRepo.GetPlan(c.Context())
	if err != nil || companyPlan.StripeCustomer == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "no active subscription found"})
	}

	url, err := billing.CreatePortalSession(h.stripeConfig, companyPlan.StripeCustomer)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "failed to open billing portal"})
	}

	return c.JSON(fiber.Map{"portal_url": url})
}

// StripeWebhook handles Stripe subscription lifecycle events.
// @Summary      Stripe webhook receiver
// @Description  Receives and processes Stripe subscription events (internal use).
// @Tags         Billing
// @Router       /webhooks/stripe [post]
func (h *BillingHandler) StripeWebhook(c *fiber.Ctx) error {
	if !h.stripeConfig.IsEnabled() {
		return c.SendStatus(fiber.StatusOK)
	}

	payload, err := io.ReadAll(c.Request().BodyStream())
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "failed to read body"})
	}

	sig := c.Get("Stripe-Signature")
	event, err := billing.ParseWebhook(payload, sig, h.stripeConfig.WebhookSecret)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid webhook signature"})
	}

	ctx := c.Context()

	switch event.Type {
	case "checkout.session.completed":
		var sess stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
			return c.SendStatus(fiber.StatusOK)
		}
		planID := sess.Metadata["plan"]
		if planID == "" {
			return c.SendStatus(fiber.StatusOK)
		}
		customerID := ""
		if sess.Customer != nil {
			customerID = sess.Customer.ID
		}
		subID := ""
		if sess.Subscription != nil {
			subID = sess.Subscription.ID
		}
		exp := time.Now().AddDate(0, 1, 0)
		if err := h.billingRepo.SetPlan(ctx, planID, customerID, subID, &exp); err != nil {
			log.Printf("billing: SetPlan (checkout) failed: %v", err)
		}

	case "customer.subscription.updated":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			return c.SendStatus(fiber.StatusOK)
		}
		planID := sub.Metadata["plan"]
		if planID == "" {
			return c.SendStatus(fiber.StatusOK)
		}
		customerID := ""
		if sub.Customer != nil {
			customerID = sub.Customer.ID
		}
		exp := time.Unix(sub.CurrentPeriodEnd, 0)
		if err := h.billingRepo.SetPlan(ctx, planID, customerID, sub.ID, &exp); err != nil {
			log.Printf("billing: SetPlan (updated) failed: %v", err)
		}

	case "customer.subscription.deleted":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			return c.SendStatus(fiber.StatusOK)
		}
		customerID := ""
		if sub.Customer != nil {
			customerID = sub.Customer.ID
		}
		if err := h.billingRepo.SetPlan(ctx, billing.PlanCommunity, customerID, "", nil); err != nil {
			log.Printf("billing: SetPlan (deleted) failed: %v", err)
		}
	}

	return c.SendStatus(fiber.StatusOK)
}

func nextMonthReset() string {
	now := time.Now().UTC()
	first := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	return first.Format("2006-01-02")
}

// GetUsage returns current month usage for the authenticated user's company.
// Used by the frontend to show live counters without reloading the full billing page.
func (h *BillingHandler) GetUsage(c *fiber.Ctx) error {
	companyIDStr, _ := c.Locals("company_id").(string)
	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid company"})
	}

	companyPlan, _ := h.billingRepo.GetPlan(c.Context())
	plan := billing.Get(companyPlan.Plan)

	usage, err := h.billingRepo.GetUsage(c.Context(), companyID)
	if err != nil {
		usage = map[string]int{"bos24": 0, "ai": 0, "pdf": 0}
	}

	claims := middleware.ClaimsFromCtx(c)
	userID, _ := claims["sub"].(string)

	return c.JSON(fiber.Map{
		"plan":   plan.ID,
		"reset":  nextMonthReset(),
		"user":   userID,
		"bos24":  fiber.Map{"used": usage["bos24"], "limit": plan.Quotas.BOS24Monthly},
		"ai":     fiber.Map{"used": usage["ai"], "limit": plan.Quotas.AIMonthly},
		"pdf":    fiber.Map{"used": usage["pdf"], "limit": plan.Quotas.PDFMonthly},
	})
}
