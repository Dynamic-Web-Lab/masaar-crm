package billing

import (
	"fmt"

	"github.com/stripe/stripe-go/v76"
	checkoutsession "github.com/stripe/stripe-go/v76/checkout/session"
	portalsession "github.com/stripe/stripe-go/v76/billingportal/session"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/webhook"
)

// StripeConfig holds keys and price IDs loaded from environment.
type StripeConfig struct {
	SecretKey        string
	WebhookSecret    string
	PriceIDStarter   string
	PriceIDPro       string
	PriceIDBusiness  string
	AppURL           string // e.g. https://crm.company.ae — for redirect URLs
}

// IsEnabled returns true when Stripe is configured.
func (c *StripeConfig) IsEnabled() bool {
	return c.SecretKey != ""
}

// SetupStripe initialises the global Stripe client and injects price IDs into Plans.
func SetupStripe(cfg *StripeConfig) {
	if !cfg.IsEnabled() {
		return
	}
	stripe.Key = cfg.SecretKey

	for i := range Plans {
		switch Plans[i].ID {
		case PlanStarter:
			Plans[i].StripePriceID = cfg.PriceIDStarter
		case PlanPro:
			Plans[i].StripePriceID = cfg.PriceIDPro
		case PlanBusiness:
			Plans[i].StripePriceID = cfg.PriceIDBusiness
		}
	}
}

// CreateCheckoutSession creates a Stripe Checkout session for a plan upgrade.
// companyID is stored in metadata so the webhook can identify the company.
func CreateCheckoutSession(cfg *StripeConfig, plan Plan, companyID, customerID string) (string, error) {
	if plan.StripePriceID == "" {
		return "", fmt.Errorf("no stripe price configured for plan %s", plan.ID)
	}

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(plan.StripePriceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(cfg.AppURL + "/settings/billing?success=1"),
		CancelURL:  stripe.String(cfg.AppURL + "/settings/billing?cancelled=1"),
		Metadata: map[string]string{
			"company_id": companyID,
			"plan":       plan.ID,
		},
	}

	if customerID != "" {
		params.Customer = stripe.String(customerID)
	}

	s, err := checkoutsession.New(params)
	if err != nil {
		return "", fmt.Errorf("stripe checkout: %w", err)
	}
	return s.URL, nil
}

// CreatePortalSession creates a Stripe Customer Portal session for billing management.
func CreatePortalSession(cfg *StripeConfig, customerID string) (string, error) {
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(cfg.AppURL + "/settings/billing"),
	}
	s, err := portalsession.New(params)
	if err != nil {
		return "", fmt.Errorf("stripe portal: %w", err)
	}
	return s.URL, nil
}

// GetOrCreateCustomer finds or creates a Stripe customer for a company.
func GetOrCreateCustomer(companyName, email, existingID string) (string, error) {
	if existingID != "" {
		return existingID, nil
	}
	params := &stripe.CustomerParams{
		Name:  stripe.String(companyName),
		Email: stripe.String(email),
	}
	c, err := customer.New(params)
	if err != nil {
		return "", fmt.Errorf("stripe customer: %w", err)
	}
	return c.ID, nil
}

// ParseWebhook validates and parses an incoming Stripe webhook.
func ParseWebhook(payload []byte, sig, secret string) (stripe.Event, error) {
	return webhook.ConstructEvent(payload, sig, secret)
}
