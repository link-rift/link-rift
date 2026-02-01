# Billing & Subscriptions

> Last Updated: 2025-01-24

Linkrift implements a comprehensive billing system using Stripe for payment processing, subscription management, usage tracking, and automated overage handling.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Stripe Integration](#stripe-integration)
  - [Products and Prices](#products-and-prices)
  - [Customer Management](#customer-management)
- [Subscription Lifecycle](#subscription-lifecycle)
  - [Creating Subscriptions](#creating-subscriptions)
  - [Subscription States](#subscription-states)
  - [Plan Changes](#plan-changes)
- [Usage Tracking](#usage-tracking)
- [Overage Handling](#overage-handling)
- [Webhook Handling](#webhook-handling)
- [API Endpoints](#api-endpoints)
- [React Components](#react-components)

---

## Overview

The billing system provides:

- **Stripe-powered payments** for secure, reliable payment processing
- **Flexible subscription plans** with monthly and annual billing
- **Usage-based billing** for metered features like clicks and links
- **Automatic overage charges** when usage exceeds plan limits
- **Self-service portal** for managing billing and invoices
- **Webhook handling** for real-time subscription updates

## Architecture

```
                                    ┌─────────────────────────────────────────┐
                                    │           Linkrift Frontend              │
                                    │        (Pricing, Checkout, Portal)       │
                                    └──────────────────┬──────────────────────┘
                                                       │
                                                       ▼
                                    ┌─────────────────────────────────────────┐
                                    │           Billing Service                │
                                    │     (Subscription Management API)        │
                                    └──────────────────┬──────────────────────┘
                                                       │
                              ┌────────────────────────┼────────────────────────┐
                              │                        │                        │
                              ▼                        ▼                        ▼
               ┌──────────────────────┐  ┌──────────────────────┐  ┌──────────────────────┐
               │    Stripe API        │  │    PostgreSQL        │  │    Usage Tracker     │
               │  (Payments, Subs)    │  │  (Plans, Features)   │  │   (Click Counts)     │
               └──────────────────────┘  └──────────────────────┘  └──────────────────────┘
                              │
                              ▼
               ┌──────────────────────┐
               │   Stripe Webhooks    │
               │   (Event Handler)    │
               └──────────────────────┘
```

---

## Stripe Integration

### Products and Prices

```go
// internal/billing/plans.go
package billing

import (
	"github.com/stripe/stripe-go/v76"
)

// Plan represents a subscription plan
type Plan struct {
	ID                string        `json:"id" db:"id"`
	Name              string        `json:"name" db:"name"`
	Description       string        `json:"description" db:"description"`
	StripeProductID   string        `json:"-" db:"stripe_product_id"`
	StripePriceIDMonthly string     `json:"-" db:"stripe_price_id_monthly"`
	StripePriceIDYearly  string     `json:"-" db:"stripe_price_id_yearly"`
	PriceMonthly      int64         `json:"price_monthly" db:"price_monthly"`
	PriceYearly       int64         `json:"price_yearly" db:"price_yearly"`
	Features          PlanFeatures  `json:"features" db:"features"`
	Limits            PlanLimits    `json:"limits" db:"limits"`
	IsActive          bool          `json:"is_active" db:"is_active"`
	SortOrder         int           `json:"sort_order" db:"sort_order"`
}

// PlanFeatures defines what features are included
type PlanFeatures struct {
	CustomDomains     bool `json:"custom_domains"`
	BioPages          bool `json:"bio_pages"`
	QRCodes           bool `json:"qr_codes"`
	APIAccess         bool `json:"api_access"`
	TeamMembers       bool `json:"team_members"`
	AdvancedAnalytics bool `json:"advanced_analytics"`
	BulkOperations    bool `json:"bulk_operations"`
	Webhooks          bool `json:"webhooks"`
	PrioritySupport   bool `json:"priority_support"`
	WhiteLabel        bool `json:"white_label"`
}

// PlanLimits defines usage limits
type PlanLimits struct {
	MonthlyClicks     int64 `json:"monthly_clicks"`      // -1 = unlimited
	TotalLinks        int64 `json:"total_links"`         // -1 = unlimited
	CustomDomains     int   `json:"custom_domains"`
	BioPages          int   `json:"bio_pages"`
	TeamMembers       int   `json:"team_members"`
	APIRequestsPerDay int64 `json:"api_requests_per_day"`
	LinkRetention     int   `json:"link_retention_days"` // -1 = forever
}

// Predefined plans
var Plans = map[string]Plan{
	"free": {
		ID:           "free",
		Name:         "Free",
		Description:  "Get started with basic link shortening",
		PriceMonthly: 0,
		PriceYearly:  0,
		Features: PlanFeatures{
			CustomDomains:     false,
			BioPages:          false,
			QRCodes:           true,
			APIAccess:         false,
			TeamMembers:       false,
			AdvancedAnalytics: false,
		},
		Limits: PlanLimits{
			MonthlyClicks:     1000,
			TotalLinks:        25,
			CustomDomains:     0,
			BioPages:          0,
			TeamMembers:       1,
			APIRequestsPerDay: 0,
			LinkRetention:     30,
		},
	},
	"starter": {
		ID:           "starter",
		Name:         "Starter",
		Description:  "For individuals and small projects",
		PriceMonthly: 900,  // $9.00
		PriceYearly:  7900, // $79.00
		Features: PlanFeatures{
			CustomDomains:     true,
			BioPages:          true,
			QRCodes:           true,
			APIAccess:         true,
			TeamMembers:       false,
			AdvancedAnalytics: false,
		},
		Limits: PlanLimits{
			MonthlyClicks:     10000,
			TotalLinks:        500,
			CustomDomains:     1,
			BioPages:          3,
			TeamMembers:       1,
			APIRequestsPerDay: 1000,
			LinkRetention:     -1,
		},
	},
	"pro": {
		ID:           "pro",
		Name:         "Pro",
		Description:  "For growing businesses",
		PriceMonthly: 2900,  // $29.00
		PriceYearly:  24900, // $249.00
		Features: PlanFeatures{
			CustomDomains:     true,
			BioPages:          true,
			QRCodes:           true,
			APIAccess:         true,
			TeamMembers:       true,
			AdvancedAnalytics: true,
			BulkOperations:    true,
			Webhooks:          true,
		},
		Limits: PlanLimits{
			MonthlyClicks:     100000,
			TotalLinks:        5000,
			CustomDomains:     5,
			BioPages:          10,
			TeamMembers:       5,
			APIRequestsPerDay: 10000,
			LinkRetention:     -1,
		},
	},
	"business": {
		ID:           "business",
		Name:         "Business",
		Description:  "For large teams and enterprises",
		PriceMonthly: 9900,  // $99.00
		PriceYearly:  99900, // $999.00
		Features: PlanFeatures{
			CustomDomains:     true,
			BioPages:          true,
			QRCodes:           true,
			APIAccess:         true,
			TeamMembers:       true,
			AdvancedAnalytics: true,
			BulkOperations:    true,
			Webhooks:          true,
			PrioritySupport:   true,
			WhiteLabel:        true,
		},
		Limits: PlanLimits{
			MonthlyClicks:     -1, // Unlimited
			TotalLinks:        -1, // Unlimited
			CustomDomains:     20,
			BioPages:          50,
			TeamMembers:       25,
			APIRequestsPerDay: 100000,
			LinkRetention:     -1,
		},
	},
}
```

### Customer Management

```go
// internal/billing/customer.go
package billing

import (
	"context"
	"fmt"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/paymentmethod"
	"github.com/link-rift/link-rift/internal/db"
)

// Customer represents a billing customer
type Customer struct {
	ID               string    `json:"id" db:"id"`
	WorkspaceID      string    `json:"workspace_id" db:"workspace_id"`
	StripeCustomerID string    `json:"-" db:"stripe_customer_id"`
	Email            string    `json:"email" db:"email"`
	Name             string    `json:"name" db:"name"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

// CustomerService handles customer operations
type CustomerService struct {
	repo *db.CustomerRepository
}

// NewCustomerService creates a new customer service
func NewCustomerService(repo *db.CustomerRepository) *CustomerService {
	return &CustomerService{repo: repo}
}

// CreateOrGetCustomer creates a Stripe customer or returns existing one
func (cs *CustomerService) CreateOrGetCustomer(ctx context.Context, workspaceID, email, name string) (*Customer, error) {
	// Check if customer already exists
	existing, err := cs.repo.GetByWorkspace(ctx, workspaceID)
	if err == nil {
		return existing, nil
	}

	// Create Stripe customer
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(name),
		Metadata: map[string]string{
			"workspace_id": workspaceID,
		},
	}

	stripeCustomer, err := customer.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe customer: %w", err)
	}

	// Store in database
	cust := &Customer{
		WorkspaceID:      workspaceID,
		StripeCustomerID: stripeCustomer.ID,
		Email:            email,
		Name:             name,
		CreatedAt:        time.Now(),
	}

	if err := cs.repo.Create(ctx, cust); err != nil {
		return nil, err
	}

	return cust, nil
}

// GetPaymentMethods retrieves customer's payment methods
func (cs *CustomerService) GetPaymentMethods(ctx context.Context, workspaceID string) ([]*PaymentMethod, error) {
	cust, err := cs.repo.GetByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(cust.StripeCustomerID),
		Type:     stripe.String("card"),
	}

	iter := paymentmethod.List(params)

	var methods []*PaymentMethod
	for iter.Next() {
		pm := iter.PaymentMethod()
		methods = append(methods, &PaymentMethod{
			ID:        pm.ID,
			Brand:     string(pm.Card.Brand),
			Last4:     pm.Card.Last4,
			ExpMonth:  int(pm.Card.ExpMonth),
			ExpYear:   int(pm.Card.ExpYear),
			IsDefault: pm.ID == cust.DefaultPaymentMethodID,
		})
	}

	return methods, iter.Err()
}

// AddPaymentMethod attaches a payment method to the customer
func (cs *CustomerService) AddPaymentMethod(ctx context.Context, workspaceID, paymentMethodID string) error {
	cust, err := cs.repo.GetByWorkspace(ctx, workspaceID)
	if err != nil {
		return err
	}

	params := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(cust.StripeCustomerID),
	}

	_, err = paymentmethod.Attach(paymentMethodID, params)
	if err != nil {
		return fmt.Errorf("failed to attach payment method: %w", err)
	}

	return nil
}

// PaymentMethod represents a stored payment method
type PaymentMethod struct {
	ID        string `json:"id"`
	Brand     string `json:"brand"`
	Last4     string `json:"last4"`
	ExpMonth  int    `json:"exp_month"`
	ExpYear   int    `json:"exp_year"`
	IsDefault bool   `json:"is_default"`
}
```

---

## Subscription Lifecycle

### Creating Subscriptions

```go
// internal/billing/subscription.go
package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/subscription"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/link-rift/link-rift/internal/db"
)

// Subscription represents a subscription
type Subscription struct {
	ID                   string             `json:"id" db:"id"`
	WorkspaceID          string             `json:"workspace_id" db:"workspace_id"`
	StripeSubscriptionID string             `json:"-" db:"stripe_subscription_id"`
	PlanID               string             `json:"plan_id" db:"plan_id"`
	Status               SubscriptionStatus `json:"status" db:"status"`
	BillingCycle         string             `json:"billing_cycle" db:"billing_cycle"` // monthly, yearly
	CurrentPeriodStart   time.Time          `json:"current_period_start" db:"current_period_start"`
	CurrentPeriodEnd     time.Time          `json:"current_period_end" db:"current_period_end"`
	CancelAtPeriodEnd    bool               `json:"cancel_at_period_end" db:"cancel_at_period_end"`
	CanceledAt           *time.Time         `json:"canceled_at,omitempty" db:"canceled_at"`
	CreatedAt            time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at" db:"updated_at"`
}

// SubscriptionStatus defines subscription states
type SubscriptionStatus string

const (
	StatusTrialing        SubscriptionStatus = "trialing"
	StatusActive          SubscriptionStatus = "active"
	StatusPastDue         SubscriptionStatus = "past_due"
	StatusCanceled        SubscriptionStatus = "canceled"
	StatusIncomplete      SubscriptionStatus = "incomplete"
	StatusIncompleteExpired SubscriptionStatus = "incomplete_expired"
	StatusUnpaid          SubscriptionStatus = "unpaid"
)

// SubscriptionService handles subscription operations
type SubscriptionService struct {
	repo           *db.SubscriptionRepository
	customerSvc    *CustomerService
	usageTracker   *UsageTracker
	successURL     string
	cancelURL      string
}

// NewSubscriptionService creates a new subscription service
func NewSubscriptionService(
	repo *db.SubscriptionRepository,
	customerSvc *CustomerService,
	usageTracker *UsageTracker,
	successURL, cancelURL string,
) *SubscriptionService {
	return &SubscriptionService{
		repo:         repo,
		customerSvc:  customerSvc,
		usageTracker: usageTracker,
		successURL:   successURL,
		cancelURL:    cancelURL,
	}
}

// CreateCheckoutSession creates a Stripe Checkout session
func (ss *SubscriptionService) CreateCheckoutSession(
	ctx context.Context,
	workspaceID string,
	planID string,
	billingCycle string,
) (*CheckoutSession, error) {
	plan, ok := Plans[planID]
	if !ok {
		return nil, ErrPlanNotFound
	}

	cust, err := ss.customerSvc.GetCustomer(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// Determine price ID
	priceID := plan.StripePriceIDMonthly
	if billingCycle == "yearly" {
		priceID = plan.StripePriceIDYearly
	}

	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(cust.StripeCustomerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(ss.successURL + "?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String(ss.cancelURL),
		Metadata: map[string]string{
			"workspace_id":  workspaceID,
			"plan_id":       planID,
			"billing_cycle": billingCycle,
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"workspace_id": workspaceID,
				"plan_id":      planID,
			},
		},
		AllowPromotionCodes: stripe.Bool(true),
	}

	sess, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkout session: %w", err)
	}

	return &CheckoutSession{
		ID:  sess.ID,
		URL: sess.URL,
	}, nil
}

// CreateSubscription creates a subscription directly (with existing payment method)
func (ss *SubscriptionService) CreateSubscription(
	ctx context.Context,
	workspaceID string,
	planID string,
	billingCycle string,
) (*Subscription, error) {
	plan, ok := Plans[planID]
	if !ok {
		return nil, ErrPlanNotFound
	}

	cust, err := ss.customerSvc.GetCustomer(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	priceID := plan.StripePriceIDMonthly
	if billingCycle == "yearly" {
		priceID = plan.StripePriceIDYearly
	}

	params := &stripe.SubscriptionParams{
		Customer: stripe.String(cust.StripeCustomerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(priceID),
			},
		},
		Metadata: map[string]string{
			"workspace_id": workspaceID,
			"plan_id":      planID,
		},
		PaymentBehavior: stripe.String("default_incomplete"),
		PaymentSettings: &stripe.SubscriptionPaymentSettingsParams{
			SaveDefaultPaymentMethod: stripe.String("on_subscription"),
		},
	}

	stripeSub, err := subscription.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	sub := &Subscription{
		WorkspaceID:          workspaceID,
		StripeSubscriptionID: stripeSub.ID,
		PlanID:               planID,
		Status:               SubscriptionStatus(stripeSub.Status),
		BillingCycle:         billingCycle,
		CurrentPeriodStart:   time.Unix(stripeSub.CurrentPeriodStart, 0),
		CurrentPeriodEnd:     time.Unix(stripeSub.CurrentPeriodEnd, 0),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	if err := ss.repo.Create(ctx, sub); err != nil {
		return nil, err
	}

	return sub, nil
}

// CheckoutSession represents a checkout session response
type CheckoutSession struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}
```

### Subscription States

```go
// internal/billing/states.go
package billing

import (
	"context"
	"time"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/subscription"
)

// CancelSubscription cancels a subscription at period end
func (ss *SubscriptionService) CancelSubscription(ctx context.Context, workspaceID string) (*Subscription, error) {
	sub, err := ss.repo.GetByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	}

	_, err = subscription.Update(sub.StripeSubscriptionID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel subscription: %w", err)
	}

	sub.CancelAtPeriodEnd = true
	now := time.Now()
	sub.CanceledAt = &now
	sub.UpdatedAt = now

	if err := ss.repo.Update(ctx, sub); err != nil {
		return nil, err
	}

	return sub, nil
}

// ReactivateSubscription removes cancellation
func (ss *SubscriptionService) ReactivateSubscription(ctx context.Context, workspaceID string) (*Subscription, error) {
	sub, err := ss.repo.GetByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(false),
	}

	_, err = subscription.Update(sub.StripeSubscriptionID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to reactivate subscription: %w", err)
	}

	sub.CancelAtPeriodEnd = false
	sub.CanceledAt = nil
	sub.UpdatedAt = time.Now()

	if err := ss.repo.Update(ctx, sub); err != nil {
		return nil, err
	}

	return sub, nil
}

// CancelImmediately cancels subscription immediately
func (ss *SubscriptionService) CancelImmediately(ctx context.Context, workspaceID string) error {
	sub, err := ss.repo.GetByWorkspace(ctx, workspaceID)
	if err != nil {
		return err
	}

	_, err = subscription.Cancel(sub.StripeSubscriptionID, nil)
	if err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	sub.Status = StatusCanceled
	now := time.Now()
	sub.CanceledAt = &now
	sub.UpdatedAt = now

	return ss.repo.Update(ctx, sub)
}
```

### Plan Changes

```go
// internal/billing/changes.go
package billing

import (
	"context"
	"fmt"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/subscription"
)

// ChangePlan upgrades or downgrades the subscription
func (ss *SubscriptionService) ChangePlan(
	ctx context.Context,
	workspaceID string,
	newPlanID string,
	billingCycle string,
) (*Subscription, error) {
	sub, err := ss.repo.GetByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	newPlan, ok := Plans[newPlanID]
	if !ok {
		return nil, ErrPlanNotFound
	}

	// Get the current Stripe subscription to find the item ID
	stripeSub, err := subscription.Get(sub.StripeSubscriptionID, nil)
	if err != nil {
		return nil, err
	}

	if len(stripeSub.Items.Data) == 0 {
		return nil, fmt.Errorf("no subscription items found")
	}

	itemID := stripeSub.Items.Data[0].ID

	priceID := newPlan.StripePriceIDMonthly
	if billingCycle == "yearly" {
		priceID = newPlan.StripePriceIDYearly
	}

	// Determine proration behavior
	// Upgrade: prorate immediately
	// Downgrade: apply at period end
	currentPlan := Plans[sub.PlanID]
	isUpgrade := newPlan.PriceMonthly > currentPlan.PriceMonthly

	prorationBehavior := "create_prorations"
	if !isUpgrade {
		prorationBehavior = "none"
	}

	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:    stripe.String(itemID),
				Price: stripe.String(priceID),
			},
		},
		ProrationBehavior: stripe.String(prorationBehavior),
		Metadata: map[string]string{
			"workspace_id": workspaceID,
			"plan_id":      newPlanID,
		},
	}

	updatedSub, err := subscription.Update(sub.StripeSubscriptionID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to change plan: %w", err)
	}

	// Update local record
	sub.PlanID = newPlanID
	sub.BillingCycle = billingCycle
	sub.CurrentPeriodStart = time.Unix(updatedSub.CurrentPeriodStart, 0)
	sub.CurrentPeriodEnd = time.Unix(updatedSub.CurrentPeriodEnd, 0)
	sub.UpdatedAt = time.Now()

	if err := ss.repo.Update(ctx, sub); err != nil {
		return nil, err
	}

	return sub, nil
}

// PreviewPlanChange shows what a plan change would cost
func (ss *SubscriptionService) PreviewPlanChange(
	ctx context.Context,
	workspaceID string,
	newPlanID string,
	billingCycle string,
) (*PlanChangePreview, error) {
	sub, err := ss.repo.GetByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	newPlan, ok := Plans[newPlanID]
	if !ok {
		return nil, ErrPlanNotFound
	}

	stripeSub, err := subscription.Get(sub.StripeSubscriptionID, nil)
	if err != nil {
		return nil, err
	}

	priceID := newPlan.StripePriceIDMonthly
	if billingCycle == "yearly" {
		priceID = newPlan.StripePriceIDYearly
	}

	// Get upcoming invoice preview
	params := &stripe.InvoiceParams{
		Customer:     stripe.String(stripeSub.Customer.ID),
		Subscription: stripe.String(sub.StripeSubscriptionID),
		SubscriptionItems: []*stripe.SubscriptionItemsParams{
			{
				ID:    stripe.String(stripeSub.Items.Data[0].ID),
				Price: stripe.String(priceID),
			},
		},
		SubscriptionProrationBehavior: stripe.String("create_prorations"),
	}

	invoice, err := invoice.GetNext(params)
	if err != nil {
		return nil, err
	}

	return &PlanChangePreview{
		CurrentPlan:     sub.PlanID,
		NewPlan:         newPlanID,
		ImmediateCharge: invoice.AmountDue,
		NextBillDate:    time.Unix(invoice.NextPaymentAttempt, 0),
		NextBillAmount:  newPlan.PriceMonthly,
		ProrationItems:  ss.extractProrationItems(invoice),
	}, nil
}

// PlanChangePreview shows the cost impact of a plan change
type PlanChangePreview struct {
	CurrentPlan     string          `json:"current_plan"`
	NewPlan         string          `json:"new_plan"`
	ImmediateCharge int64           `json:"immediate_charge"`
	NextBillDate    time.Time       `json:"next_bill_date"`
	NextBillAmount  int64           `json:"next_bill_amount"`
	ProrationItems  []ProrationItem `json:"proration_items"`
}

type ProrationItem struct {
	Description string `json:"description"`
	Amount      int64  `json:"amount"`
}

func (ss *SubscriptionService) extractProrationItems(invoice *stripe.Invoice) []ProrationItem {
	var items []ProrationItem
	for _, line := range invoice.Lines.Data {
		if line.Proration {
			items = append(items, ProrationItem{
				Description: line.Description,
				Amount:      line.Amount,
			})
		}
	}
	return items
}
```

---

## Usage Tracking

```go
// internal/billing/usage.go
package billing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// UsageTracker tracks workspace usage for billing
type UsageTracker struct {
	redis  *redis.Client
	repo   *db.UsageRepository
	mu     sync.RWMutex
}

// NewUsageTracker creates a new usage tracker
func NewUsageTracker(redis *redis.Client, repo *db.UsageRepository) *UsageTracker {
	return &UsageTracker{
		redis: redis,
		repo:  repo,
	}
}

// Usage represents current usage for a workspace
type Usage struct {
	WorkspaceID   string    `json:"workspace_id"`
	Period        string    `json:"period"` // YYYY-MM format
	Clicks        int64     `json:"clicks"`
	Links         int64     `json:"links"`
	APIRequests   int64     `json:"api_requests"`
	CustomDomains int       `json:"custom_domains"`
	BioPages      int       `json:"bio_pages"`
	TeamMembers   int       `json:"team_members"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// IncrementClicks increments click count for a workspace
func (ut *UsageTracker) IncrementClicks(ctx context.Context, workspaceID string, count int64) error {
	period := time.Now().Format("2006-01")
	key := fmt.Sprintf("usage:clicks:%s:%s", workspaceID, period)

	return ut.redis.IncrBy(ctx, key, count).Err()
}

// IncrementAPIRequests increments API request count
func (ut *UsageTracker) IncrementAPIRequests(ctx context.Context, workspaceID string) error {
	period := time.Now().Format("2006-01")
	key := fmt.Sprintf("usage:api:%s:%s", workspaceID, period)

	// Also track daily for rate limiting
	dailyKey := fmt.Sprintf("usage:api:daily:%s:%s", workspaceID, time.Now().Format("2006-01-02"))

	pipe := ut.redis.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Incr(ctx, dailyKey)
	pipe.Expire(ctx, dailyKey, 48*time.Hour)

	_, err := pipe.Exec(ctx)
	return err
}

// GetUsage retrieves current usage for a workspace
func (ut *UsageTracker) GetUsage(ctx context.Context, workspaceID string) (*Usage, error) {
	period := time.Now().Format("2006-01")

	// Get click count from Redis
	clickKey := fmt.Sprintf("usage:clicks:%s:%s", workspaceID, period)
	clicks, _ := ut.redis.Get(ctx, clickKey).Int64()

	// Get API request count
	apiKey := fmt.Sprintf("usage:api:%s:%s", workspaceID, period)
	apiRequests, _ := ut.redis.Get(ctx, apiKey).Int64()

	// Get link count from database
	links, _ := ut.repo.GetLinkCount(ctx, workspaceID)

	// Get other counts from database
	customDomains, _ := ut.repo.GetCustomDomainCount(ctx, workspaceID)
	bioPages, _ := ut.repo.GetBioPageCount(ctx, workspaceID)
	teamMembers, _ := ut.repo.GetTeamMemberCount(ctx, workspaceID)

	return &Usage{
		WorkspaceID:   workspaceID,
		Period:        period,
		Clicks:        clicks,
		Links:         links,
		APIRequests:   apiRequests,
		CustomDomains: customDomains,
		BioPages:      bioPages,
		TeamMembers:   teamMembers,
		UpdatedAt:     time.Now(),
	}, nil
}

// CheckLimits verifies if workspace is within plan limits
func (ut *UsageTracker) CheckLimits(ctx context.Context, workspaceID string, planID string) (*LimitCheck, error) {
	plan, ok := Plans[planID]
	if !ok {
		return nil, ErrPlanNotFound
	}

	usage, err := ut.GetUsage(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	limits := plan.Limits

	check := &LimitCheck{
		WorkspaceID: workspaceID,
		PlanID:      planID,
		Limits:      make(map[string]LimitStatus),
	}

	// Check each limit
	check.Limits["clicks"] = ut.checkLimit(usage.Clicks, limits.MonthlyClicks)
	check.Limits["links"] = ut.checkLimit(usage.Links, limits.TotalLinks)
	check.Limits["custom_domains"] = ut.checkLimit(int64(usage.CustomDomains), int64(limits.CustomDomains))
	check.Limits["bio_pages"] = ut.checkLimit(int64(usage.BioPages), int64(limits.BioPages))
	check.Limits["team_members"] = ut.checkLimit(int64(usage.TeamMembers), int64(limits.TeamMembers))
	check.Limits["api_requests"] = ut.checkDailyAPILimit(ctx, workspaceID, limits.APIRequestsPerDay)

	// Determine overall status
	check.IsWithinLimits = true
	for _, status := range check.Limits {
		if status.IsExceeded {
			check.IsWithinLimits = false
			break
		}
	}

	return check, nil
}

func (ut *UsageTracker) checkLimit(current, limit int64) LimitStatus {
	if limit == -1 {
		return LimitStatus{
			Current:    current,
			Limit:      -1,
			IsUnlimited: true,
			IsExceeded: false,
			Percentage: 0,
		}
	}

	percentage := float64(current) / float64(limit) * 100
	if percentage > 100 {
		percentage = 100
	}

	return LimitStatus{
		Current:    current,
		Limit:      limit,
		IsUnlimited: false,
		IsExceeded: current >= limit,
		Percentage: percentage,
	}
}

func (ut *UsageTracker) checkDailyAPILimit(ctx context.Context, workspaceID string, limit int64) LimitStatus {
	if limit == -1 || limit == 0 {
		return LimitStatus{IsUnlimited: limit == -1, Limit: limit}
	}

	dailyKey := fmt.Sprintf("usage:api:daily:%s:%s", workspaceID, time.Now().Format("2006-01-02"))
	current, _ := ut.redis.Get(ctx, dailyKey).Int64()

	return ut.checkLimit(current, limit)
}

// LimitCheck represents the result of a limit check
type LimitCheck struct {
	WorkspaceID    string                 `json:"workspace_id"`
	PlanID         string                 `json:"plan_id"`
	Limits         map[string]LimitStatus `json:"limits"`
	IsWithinLimits bool                   `json:"is_within_limits"`
}

// LimitStatus represents the status of a single limit
type LimitStatus struct {
	Current     int64   `json:"current"`
	Limit       int64   `json:"limit"`
	IsUnlimited bool    `json:"is_unlimited"`
	IsExceeded  bool    `json:"is_exceeded"`
	Percentage  float64 `json:"percentage"`
}
```

---

## Overage Handling

```go
// internal/billing/overage.go
package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/usagerecord"
)

// OverageConfig defines overage pricing
type OverageConfig struct {
	ClicksPerUnit     int64 `json:"clicks_per_unit"`     // e.g., 1000 clicks per unit
	PricePerUnit      int64 `json:"price_per_unit"`      // e.g., 100 cents = $1.00
	MaxOverageUnits   int64 `json:"max_overage_units"`   // Maximum overage allowed
}

var DefaultOverageConfig = OverageConfig{
	ClicksPerUnit:   1000,
	PricePerUnit:    100, // $1.00 per 1000 clicks
	MaxOverageUnits: 100, // Max 100,000 extra clicks
}

// OverageHandler manages usage-based overage billing
type OverageHandler struct {
	usageTracker    *UsageTracker
	subRepo         *db.SubscriptionRepository
	overageConfig   OverageConfig
	meteredPriceID  string // Stripe metered price ID for overages
}

// NewOverageHandler creates a new overage handler
func NewOverageHandler(
	tracker *UsageTracker,
	repo *db.SubscriptionRepository,
	config OverageConfig,
	meteredPriceID string,
) *OverageHandler {
	return &OverageHandler{
		usageTracker:   tracker,
		subRepo:        repo,
		overageConfig:  config,
		meteredPriceID: meteredPriceID,
	}
}

// ProcessOverage calculates and reports overage to Stripe
func (oh *OverageHandler) ProcessOverage(ctx context.Context, workspaceID string) error {
	sub, err := oh.subRepo.GetByWorkspace(ctx, workspaceID)
	if err != nil {
		return err
	}

	plan, ok := Plans[sub.PlanID]
	if !ok {
		return ErrPlanNotFound
	}

	// Get current usage
	usage, err := oh.usageTracker.GetUsage(ctx, workspaceID)
	if err != nil {
		return err
	}

	// Calculate overage clicks
	if plan.Limits.MonthlyClicks == -1 {
		// Unlimited plan, no overage
		return nil
	}

	overageClicks := usage.Clicks - plan.Limits.MonthlyClicks
	if overageClicks <= 0 {
		return nil
	}

	// Calculate overage units
	overageUnits := overageClicks / oh.overageConfig.ClicksPerUnit
	if overageClicks%oh.overageConfig.ClicksPerUnit > 0 {
		overageUnits++ // Round up
	}

	// Cap overage
	if overageUnits > oh.overageConfig.MaxOverageUnits {
		overageUnits = oh.overageConfig.MaxOverageUnits
	}

	// Report to Stripe
	_, err = usagerecord.New(&stripe.UsageRecordParams{
		SubscriptionItem: stripe.String(oh.getMeteredItemID(sub)),
		Quantity:         stripe.Int64(overageUnits),
		Timestamp:        stripe.Int64(time.Now().Unix()),
		Action:           stripe.String("set"), // Set total, not increment
	})

	return err
}

func (oh *OverageHandler) getMeteredItemID(sub *Subscription) string {
	// In a real implementation, this would be stored with the subscription
	// when the metered item is added
	return sub.MeteredItemID
}

// GetOverageSummary returns current overage information
func (oh *OverageHandler) GetOverageSummary(ctx context.Context, workspaceID string) (*OverageSummary, error) {
	sub, err := oh.subRepo.GetByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	plan, ok := Plans[sub.PlanID]
	if !ok {
		return nil, ErrPlanNotFound
	}

	usage, err := oh.usageTracker.GetUsage(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	summary := &OverageSummary{
		WorkspaceID:    workspaceID,
		Period:         usage.Period,
		IncludedClicks: plan.Limits.MonthlyClicks,
		UsedClicks:     usage.Clicks,
	}

	if plan.Limits.MonthlyClicks == -1 {
		summary.IsUnlimited = true
		return summary, nil
	}

	if usage.Clicks > plan.Limits.MonthlyClicks {
		summary.OverageClicks = usage.Clicks - plan.Limits.MonthlyClicks
		summary.OverageUnits = summary.OverageClicks / oh.overageConfig.ClicksPerUnit
		if summary.OverageClicks%oh.overageConfig.ClicksPerUnit > 0 {
			summary.OverageUnits++
		}
		summary.OverageCost = summary.OverageUnits * oh.overageConfig.PricePerUnit
	}

	return summary, nil
}

// OverageSummary contains overage information
type OverageSummary struct {
	WorkspaceID    string `json:"workspace_id"`
	Period         string `json:"period"`
	IncludedClicks int64  `json:"included_clicks"`
	UsedClicks     int64  `json:"used_clicks"`
	OverageClicks  int64  `json:"overage_clicks"`
	OverageUnits   int64  `json:"overage_units"`
	OverageCost    int64  `json:"overage_cost"` // in cents
	IsUnlimited    bool   `json:"is_unlimited"`
}
```

---

## Webhook Handling

```go
// internal/billing/webhooks.go
package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/webhook"
	"github.com/link-rift/link-rift/internal/db"
)

// WebhookHandler handles Stripe webhooks
type WebhookHandler struct {
	endpointSecret string
	subRepo        *db.SubscriptionRepository
	customerRepo   *db.CustomerRepository
	eventHandlers  map[string]EventHandler
}

// EventHandler is a function that handles a specific event type
type EventHandler func(ctx context.Context, event stripe.Event) error

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(
	endpointSecret string,
	subRepo *db.SubscriptionRepository,
	customerRepo *db.CustomerRepository,
) *WebhookHandler {
	wh := &WebhookHandler{
		endpointSecret: endpointSecret,
		subRepo:        subRepo,
		customerRepo:   customerRepo,
		eventHandlers:  make(map[string]EventHandler),
	}

	// Register handlers
	wh.eventHandlers["customer.subscription.created"] = wh.handleSubscriptionCreated
	wh.eventHandlers["customer.subscription.updated"] = wh.handleSubscriptionUpdated
	wh.eventHandlers["customer.subscription.deleted"] = wh.handleSubscriptionDeleted
	wh.eventHandlers["invoice.paid"] = wh.handleInvoicePaid
	wh.eventHandlers["invoice.payment_failed"] = wh.handleInvoicePaymentFailed
	wh.eventHandlers["customer.updated"] = wh.handleCustomerUpdated
	wh.eventHandlers["checkout.session.completed"] = wh.handleCheckoutCompleted

	return wh
}

// HandleWebhook processes incoming Stripe webhooks
func (wh *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	// Verify webhook signature
	event, err := webhook.ConstructEvent(body, r.Header.Get("Stripe-Signature"), wh.endpointSecret)
	if err != nil {
		http.Error(w, "Invalid signature", http.StatusBadRequest)
		return
	}

	// Handle the event
	ctx := r.Context()
	handler, ok := wh.eventHandlers[string(event.Type)]
	if ok {
		if err := handler(ctx, event); err != nil {
			// Log error but return 200 to prevent retries for handled errors
			fmt.Printf("Error handling event %s: %v\n", event.Type, err)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (wh *WebhookHandler) handleSubscriptionCreated(ctx context.Context, event stripe.Event) error {
	var stripeSub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &stripeSub); err != nil {
		return err
	}

	workspaceID := stripeSub.Metadata["workspace_id"]
	planID := stripeSub.Metadata["plan_id"]

	sub := &Subscription{
		WorkspaceID:          workspaceID,
		StripeSubscriptionID: stripeSub.ID,
		PlanID:               planID,
		Status:               SubscriptionStatus(stripeSub.Status),
		CurrentPeriodStart:   time.Unix(stripeSub.CurrentPeriodStart, 0),
		CurrentPeriodEnd:     time.Unix(stripeSub.CurrentPeriodEnd, 0),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	return wh.subRepo.Create(ctx, sub)
}

func (wh *WebhookHandler) handleSubscriptionUpdated(ctx context.Context, event stripe.Event) error {
	var stripeSub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &stripeSub); err != nil {
		return err
	}

	sub, err := wh.subRepo.GetByStripeID(ctx, stripeSub.ID)
	if err != nil {
		return err
	}

	sub.Status = SubscriptionStatus(stripeSub.Status)
	sub.CurrentPeriodStart = time.Unix(stripeSub.CurrentPeriodStart, 0)
	sub.CurrentPeriodEnd = time.Unix(stripeSub.CurrentPeriodEnd, 0)
	sub.CancelAtPeriodEnd = stripeSub.CancelAtPeriodEnd

	if stripeSub.CanceledAt > 0 {
		canceledAt := time.Unix(stripeSub.CanceledAt, 0)
		sub.CanceledAt = &canceledAt
	}

	// Check if plan changed
	if planID, ok := stripeSub.Metadata["plan_id"]; ok && planID != sub.PlanID {
		sub.PlanID = planID
	}

	sub.UpdatedAt = time.Now()

	return wh.subRepo.Update(ctx, sub)
}

func (wh *WebhookHandler) handleSubscriptionDeleted(ctx context.Context, event stripe.Event) error {
	var stripeSub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &stripeSub); err != nil {
		return err
	}

	sub, err := wh.subRepo.GetByStripeID(ctx, stripeSub.ID)
	if err != nil {
		return err
	}

	sub.Status = StatusCanceled
	now := time.Now()
	sub.CanceledAt = &now
	sub.UpdatedAt = now

	return wh.subRepo.Update(ctx, sub)
}

func (wh *WebhookHandler) handleInvoicePaid(ctx context.Context, event stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return err
	}

	// Update subscription status to active if it was past_due
	if invoice.Subscription != nil {
		sub, err := wh.subRepo.GetByStripeID(ctx, invoice.Subscription.ID)
		if err != nil {
			return err
		}

		if sub.Status == StatusPastDue {
			sub.Status = StatusActive
			sub.UpdatedAt = time.Now()
			return wh.subRepo.Update(ctx, sub)
		}
	}

	return nil
}

func (wh *WebhookHandler) handleInvoicePaymentFailed(ctx context.Context, event stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return err
	}

	if invoice.Subscription != nil {
		sub, err := wh.subRepo.GetByStripeID(ctx, invoice.Subscription.ID)
		if err != nil {
			return err
		}

		sub.Status = StatusPastDue
		sub.UpdatedAt = time.Now()

		// Send notification to workspace owner
		// notificationService.SendPaymentFailedEmail(ctx, sub.WorkspaceID)

		return wh.subRepo.Update(ctx, sub)
	}

	return nil
}

func (wh *WebhookHandler) handleCheckoutCompleted(ctx context.Context, event stripe.Event) error {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		return err
	}

	// The subscription is created via the subscription.created webhook
	// This handler can be used for additional logic like:
	// - Sending welcome emails
	// - Analytics tracking
	// - Onboarding flow triggers

	return nil
}

func (wh *WebhookHandler) handleCustomerUpdated(ctx context.Context, event stripe.Event) error {
	var customer stripe.Customer
	if err := json.Unmarshal(event.Data.Raw, &customer); err != nil {
		return err
	}

	cust, err := wh.customerRepo.GetByStripeID(ctx, customer.ID)
	if err != nil {
		return err
	}

	cust.Email = customer.Email
	cust.Name = customer.Name

	return wh.customerRepo.Update(ctx, cust)
}
```

---

## API Endpoints

```go
// internal/api/handlers/billing.go
package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/link-rift/link-rift/internal/billing"
)

// BillingHandler handles billing API requests
type BillingHandler struct {
	subscriptionSvc *billing.SubscriptionService
	customerSvc     *billing.CustomerService
	usageTracker    *billing.UsageTracker
	overageHandler  *billing.OverageHandler
}

// RegisterRoutes registers billing routes
func (h *BillingHandler) RegisterRoutes(app *fiber.App) {
	b := app.Group("/api/v1/billing")

	// Plans
	b.Get("/plans", h.ListPlans)

	// Subscriptions
	b.Get("/subscription", h.GetSubscription)
	b.Post("/subscription/checkout", h.CreateCheckout)
	b.Post("/subscription/cancel", h.CancelSubscription)
	b.Post("/subscription/reactivate", h.ReactivateSubscription)
	b.Post("/subscription/change-plan", h.ChangePlan)
	b.Post("/subscription/preview-change", h.PreviewPlanChange)

	// Usage
	b.Get("/usage", h.GetUsage)
	b.Get("/limits", h.GetLimits)
	b.Get("/overage", h.GetOverage)

	// Payment methods
	b.Get("/payment-methods", h.ListPaymentMethods)
	b.Post("/payment-methods", h.AddPaymentMethod)
	b.Delete("/payment-methods/:id", h.RemovePaymentMethod)
	b.Post("/payment-methods/:id/default", h.SetDefaultPaymentMethod)

	// Invoices
	b.Get("/invoices", h.ListInvoices)
	b.Get("/invoices/:id", h.GetInvoice)

	// Portal
	b.Post("/portal", h.CreatePortalSession)

	// Webhooks (no auth)
	app.Post("/webhooks/stripe", h.HandleStripeWebhook)
}

// CreateCheckout creates a Stripe checkout session
func (h *BillingHandler) CreateCheckout(c *fiber.Ctx) error {
	var req CreateCheckoutRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	workspaceID := c.Locals("workspaceID").(string)

	session, err := h.subscriptionSvc.CreateCheckoutSession(
		c.Context(),
		workspaceID,
		req.PlanID,
		req.BillingCycle,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(session)
}

// GetUsage returns current usage statistics
func (h *BillingHandler) GetUsage(c *fiber.Ctx) error {
	workspaceID := c.Locals("workspaceID").(string)

	usage, err := h.usageTracker.GetUsage(c.Context(), workspaceID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get usage",
		})
	}

	return c.JSON(usage)
}

// GetLimits checks if workspace is within plan limits
func (h *BillingHandler) GetLimits(c *fiber.Ctx) error {
	workspaceID := c.Locals("workspaceID").(string)

	sub, err := h.subscriptionSvc.GetSubscription(c.Context(), workspaceID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Subscription not found",
		})
	}

	limits, err := h.usageTracker.CheckLimits(c.Context(), workspaceID, sub.PlanID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check limits",
		})
	}

	return c.JSON(limits)
}

// CreatePortalSession creates a Stripe billing portal session
func (h *BillingHandler) CreatePortalSession(c *fiber.Ctx) error {
	workspaceID := c.Locals("workspaceID").(string)

	session, err := h.customerSvc.CreatePortalSession(c.Context(), workspaceID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create portal session",
		})
	}

	return c.JSON(fiber.Map{
		"url": session.URL,
	})
}

// Request types
type CreateCheckoutRequest struct {
	PlanID       string `json:"plan_id" validate:"required"`
	BillingCycle string `json:"billing_cycle" validate:"required,oneof=monthly yearly"`
}

type ChangePlanRequest struct {
	PlanID       string `json:"plan_id" validate:"required"`
	BillingCycle string `json:"billing_cycle" validate:"required,oneof=monthly yearly"`
}
```

---

## React Components

### Pricing Component

```typescript
// src/components/billing/PricingPlans.tsx
import React, { useState } from 'react';
import { useQuery, useMutation } from '@tanstack/react-query';
import { billingApi, Plan } from '@/api/billing';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Switch } from '@/components/ui/switch';
import { Badge } from '@/components/ui/badge';
import { Check, X } from 'lucide-react';
import { loadStripe } from '@stripe/stripe-js';

const stripePromise = loadStripe(import.meta.env.VITE_STRIPE_PUBLIC_KEY);

interface PricingPlansProps {
  currentPlanId?: string;
}

export const PricingPlans: React.FC<PricingPlansProps> = ({ currentPlanId }) => {
  const [billingCycle, setBillingCycle] = useState<'monthly' | 'yearly'>('monthly');

  const { data: plans } = useQuery({
    queryKey: ['plans'],
    queryFn: billingApi.getPlans,
  });

  const checkoutMutation = useMutation({
    mutationFn: billingApi.createCheckout,
    onSuccess: async (data) => {
      const stripe = await stripePromise;
      if (stripe) {
        await stripe.redirectToCheckout({ sessionId: data.id });
      }
    },
  });

  const handleSelectPlan = (planId: string) => {
    if (planId === 'free') return;
    checkoutMutation.mutate({ plan_id: planId, billing_cycle: billingCycle });
  };

  const yearlyDiscount = 20; // 20% off annual

  return (
    <div className="space-y-8">
      {/* Billing Toggle */}
      <div className="flex items-center justify-center gap-4">
        <span className={billingCycle === 'monthly' ? 'font-medium' : 'text-muted-foreground'}>
          Monthly
        </span>
        <Switch
          checked={billingCycle === 'yearly'}
          onCheckedChange={(checked) => setBillingCycle(checked ? 'yearly' : 'monthly')}
        />
        <span className={billingCycle === 'yearly' ? 'font-medium' : 'text-muted-foreground'}>
          Yearly
          <Badge variant="secondary" className="ml-2">
            Save {yearlyDiscount}%
          </Badge>
        </span>
      </div>

      {/* Plans Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {plans?.map((plan) => {
          const isCurrentPlan = plan.id === currentPlanId;
          const price = billingCycle === 'yearly' ? plan.price_yearly : plan.price_monthly;
          const monthlyPrice = billingCycle === 'yearly'
            ? Math.round(plan.price_yearly / 12)
            : plan.price_monthly;

          return (
            <Card
              key={plan.id}
              className={`relative ${plan.id === 'pro' ? 'border-primary shadow-lg' : ''}`}
            >
              {plan.id === 'pro' && (
                <Badge className="absolute -top-3 left-1/2 -translate-x-1/2">
                  Most Popular
                </Badge>
              )}

              <CardHeader>
                <CardTitle>{plan.name}</CardTitle>
                <CardDescription>{plan.description}</CardDescription>
              </CardHeader>

              <CardContent className="space-y-6">
                {/* Price */}
                <div>
                  <span className="text-4xl font-bold">
                    ${(monthlyPrice / 100).toFixed(0)}
                  </span>
                  <span className="text-muted-foreground">/month</span>
                  {billingCycle === 'yearly' && price > 0 && (
                    <p className="text-sm text-muted-foreground">
                      Billed ${(price / 100).toFixed(0)} annually
                    </p>
                  )}
                </div>

                {/* Features */}
                <ul className="space-y-2">
                  <LimitItem
                    label="Monthly clicks"
                    value={plan.limits.monthly_clicks}
                  />
                  <LimitItem
                    label="Total links"
                    value={plan.limits.total_links}
                  />
                  <LimitItem
                    label="Custom domains"
                    value={plan.limits.custom_domains}
                  />
                  <LimitItem
                    label="Team members"
                    value={plan.limits.team_members}
                  />
                  <FeatureItem
                    label="Bio pages"
                    enabled={plan.features.bio_pages}
                  />
                  <FeatureItem
                    label="Advanced analytics"
                    enabled={plan.features.advanced_analytics}
                  />
                  <FeatureItem
                    label="API access"
                    enabled={plan.features.api_access}
                  />
                  <FeatureItem
                    label="Webhooks"
                    enabled={plan.features.webhooks}
                  />
                </ul>

                {/* CTA */}
                <Button
                  className="w-full"
                  variant={plan.id === 'pro' ? 'default' : 'outline'}
                  disabled={isCurrentPlan || checkoutMutation.isPending}
                  onClick={() => handleSelectPlan(plan.id)}
                >
                  {isCurrentPlan ? 'Current Plan' : plan.id === 'free' ? 'Get Started' : 'Subscribe'}
                </Button>
              </CardContent>
            </Card>
          );
        })}
      </div>
    </div>
  );
};

const LimitItem: React.FC<{ label: string; value: number }> = ({ label, value }) => (
  <li className="flex items-center gap-2 text-sm">
    <Check className="w-4 h-4 text-primary" />
    <span>{label}:</span>
    <span className="font-medium">
      {value === -1 ? 'Unlimited' : value.toLocaleString()}
    </span>
  </li>
);

const FeatureItem: React.FC<{ label: string; enabled: boolean }> = ({ label, enabled }) => (
  <li className="flex items-center gap-2 text-sm">
    {enabled ? (
      <Check className="w-4 h-4 text-primary" />
    ) : (
      <X className="w-4 h-4 text-muted-foreground" />
    )}
    <span className={enabled ? '' : 'text-muted-foreground'}>{label}</span>
  </li>
);
```

### Usage Dashboard

```typescript
// src/components/billing/UsageDashboard.tsx
import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { billingApi, Usage, LimitCheck } from '@/api/billing';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { AlertTriangle } from 'lucide-react';

export const UsageDashboard: React.FC = () => {
  const { data: usage } = useQuery({
    queryKey: ['usage'],
    queryFn: billingApi.getUsage,
  });

  const { data: limits } = useQuery({
    queryKey: ['limits'],
    queryFn: billingApi.getLimits,
  });

  if (!usage || !limits) {
    return <div>Loading...</div>;
  }

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold">Usage This Month</h2>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <UsageCard
          title="Clicks"
          current={usage.clicks}
          limit={limits.limits.clicks}
        />
        <UsageCard
          title="Links"
          current={usage.links}
          limit={limits.limits.links}
        />
        <UsageCard
          title="API Requests"
          current={usage.api_requests}
          limit={limits.limits.api_requests}
          label="today"
        />
        <UsageCard
          title="Custom Domains"
          current={usage.custom_domains}
          limit={limits.limits.custom_domains}
        />
        <UsageCard
          title="Bio Pages"
          current={usage.bio_pages}
          limit={limits.limits.bio_pages}
        />
        <UsageCard
          title="Team Members"
          current={usage.team_members}
          limit={limits.limits.team_members}
        />
      </div>

      {!limits.is_within_limits && (
        <div className="bg-destructive/10 border border-destructive rounded-lg p-4 flex items-start gap-3">
          <AlertTriangle className="w-5 h-5 text-destructive mt-0.5" />
          <div>
            <h3 className="font-medium text-destructive">
              You have exceeded your plan limits
            </h3>
            <p className="text-sm text-muted-foreground mt-1">
              Some features may be restricted. Consider upgrading your plan to continue using all features.
            </p>
          </div>
        </div>
      )}
    </div>
  );
};

interface UsageCardProps {
  title: string;
  current: number;
  limit: {
    current: number;
    limit: number;
    is_unlimited: boolean;
    is_exceeded: boolean;
    percentage: number;
  };
  label?: string;
}

const UsageCard: React.FC<UsageCardProps> = ({ title, current, limit, label = 'this month' }) => {
  const getProgressColor = (percentage: number, exceeded: boolean) => {
    if (exceeded) return 'bg-destructive';
    if (percentage >= 80) return 'bg-yellow-500';
    return 'bg-primary';
  };

  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">
          {title}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          <div className="flex items-baseline gap-1">
            <span className="text-2xl font-bold">
              {current.toLocaleString()}
            </span>
            {!limit.is_unlimited && (
              <span className="text-muted-foreground">
                / {limit.limit.toLocaleString()}
              </span>
            )}
            {limit.is_unlimited && (
              <span className="text-muted-foreground">Unlimited</span>
            )}
          </div>

          {!limit.is_unlimited && (
            <>
              <Progress
                value={limit.percentage}
                className={`h-2 ${getProgressColor(limit.percentage, limit.is_exceeded)}`}
              />
              <p className="text-xs text-muted-foreground">
                {limit.percentage.toFixed(0)}% used {label}
              </p>
            </>
          )}
        </div>
      </CardContent>
    </Card>
  );
};
```
