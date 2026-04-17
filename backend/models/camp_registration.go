package models

import (
	"time"
)

type PaymentStatus string
type PaymentMethod string

const (
	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusPaid     PaymentStatus = "paid"
	PaymentStatusFailed   PaymentStatus = "failed"
	PaymentStatusRefunded PaymentStatus = "refunded"
)

const (
	PaymentMethodStripe PaymentMethod = "stripe"
	PaymentMethodPaypal PaymentMethod = "paypal"
)

type CampRegistration struct {
	ID                    int           `json:"id"`
	AthleteID             int           `json:"athlete_id"`
	CampID                int           `json:"camp_id"`
	PaymentStatus         PaymentStatus `json:"payment_status"`
	PaymentMethod         PaymentMethod `json:"payment_method"`
	StripePaymentIntentID *string       `json:"stripe_payment_intent_id"`
	PaypalOrderID         *string       `json:"paypal_order_id"`
	AmountCents           int           `json:"amount_cents"`
	ParentEmail           string        `json:"parent_email"`
	CreatedAt             time.Time     `json:"created_at"`
	UpdatedAt             time.Time     `json:"updated_at"`
}
