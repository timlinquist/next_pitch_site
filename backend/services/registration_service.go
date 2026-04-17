package services

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/webhook"
	"nextpitch.com/backend/models"
)

type RegistrationService struct {
	db              DB
	campService     *CampService
	paypalToken     string
	paypalTokenExp  time.Time
	paypalTokenLock sync.Mutex
}

func NewRegistrationService(db DB, campService *CampService) *RegistrationService {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	return &RegistrationService{
		db:          db,
		campService: campService,
	}
}

func (s *RegistrationService) CreateAthleteAndRegistration(athlete *models.Athlete, campID int, paymentMethod models.PaymentMethod) (*models.CampRegistration, error) {
	// Validate the camp exists and is active
	camp, err := s.campService.GetCampByID(campID)
	if err != nil {
		return nil, fmt.Errorf("camp not found: %w", err)
	}
	if !camp.IsActive {
		return nil, errors.New("camp is not active")
	}

	// Check capacity
	if camp.MaxCapacity != nil {
		count, err := s.campService.GetCampRegistrationCount(campID)
		if err != nil {
			return nil, fmt.Errorf("failed to check capacity: %w", err)
		}
		if count >= *camp.MaxCapacity {
			return nil, errors.New("camp is at full capacity")
		}
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()

	// Insert athlete
	var athleteID int
	err = tx.QueryRow(`
		INSERT INTO athletes (name, age, years_played, position, parent_email, parent_phone, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		RETURNING id
	`,
		athlete.Name,
		athlete.Age,
		athlete.YearsPlayed,
		athlete.Position,
		athlete.ParentEmail,
		athlete.ParentPhone,
		now,
	).Scan(&athleteID)
	if err != nil {
		return nil, fmt.Errorf("failed to create athlete: %w", err)
	}

	athlete.ID = athleteID

	// Insert pending registration
	reg := &models.CampRegistration{
		AthleteID:     athleteID,
		CampID:        campID,
		PaymentStatus: models.PaymentStatusPending,
		PaymentMethod: paymentMethod,
		AmountCents:   camp.PriceCents,
		ParentEmail:   athlete.ParentEmail,
	}

	err = tx.QueryRow(`
		INSERT INTO camp_registrations (athlete_id, camp_id, payment_status, payment_method, amount_cents, parent_email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		RETURNING id
	`,
		reg.AthleteID,
		reg.CampID,
		reg.PaymentStatus,
		reg.PaymentMethod,
		reg.AmountCents,
		reg.ParentEmail,
		now,
	).Scan(&reg.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create registration: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	reg.CreatedAt = now
	reg.UpdatedAt = now
	return reg, nil
}

// Stripe methods

func (s *RegistrationService) InitiateStripePayment(registrationID int) (string, error) {
	reg, err := s.GetRegistrationByID(registrationID)
	if err != nil {
		return "", err
	}

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(reg.AmountCents)),
		Currency: stripe.String("usd"),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}
	params.AddMetadata("registration_id", fmt.Sprintf("%d", reg.ID))
	params.AddMetadata("camp_id", fmt.Sprintf("%d", reg.CampID))
	params.AddMetadata("athlete_id", fmt.Sprintf("%d", reg.AthleteID))

	pi, err := paymentintent.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create payment intent: %w", err)
	}

	// Store the payment intent ID on the registration
	_, err = s.db.Exec(`
		UPDATE camp_registrations
		SET stripe_payment_intent_id = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, pi.ID, registrationID)
	if err != nil {
		return "", fmt.Errorf("failed to update registration with payment intent: %w", err)
	}

	return pi.ClientSecret, nil
}

func (s *RegistrationService) ConfirmStripePayment(registrationID int) error {
	reg, err := s.GetRegistrationByID(registrationID)
	if err != nil {
		return err
	}

	if reg.PaymentStatus == models.PaymentStatusPaid {
		return nil // Already paid, idempotent
	}

	if reg.StripePaymentIntentID == nil {
		return errors.New("no stripe payment intent found for this registration")
	}

	// Verify with Stripe API
	pi, err := paymentintent.Get(*reg.StripePaymentIntentID, nil)
	if err != nil {
		return fmt.Errorf("failed to verify payment intent: %w", err)
	}

	if pi.Status != stripe.PaymentIntentStatusSucceeded {
		return fmt.Errorf("payment not yet succeeded, status: %s", pi.Status)
	}

	_, err = s.db.Exec(`
		UPDATE camp_registrations
		SET payment_status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, models.PaymentStatusPaid, registrationID)
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	return nil
}

func (s *RegistrationService) HandleStripeWebhook(payload []byte, signature string) (*models.CampRegistration, error) {
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	event, err := webhook.ConstructEvent(payload, signature, webhookSecret)
	if err != nil {
		return nil, fmt.Errorf("webhook signature verification failed: %w", err)
	}

	switch event.Type {
	case "payment_intent.succeeded":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			return nil, fmt.Errorf("failed to parse payment intent: %w", err)
		}
		return s.updateRegistrationByStripePI(pi.ID, models.PaymentStatusPaid)

	case "payment_intent.payment_failed":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			return nil, fmt.Errorf("failed to parse payment intent: %w", err)
		}
		return s.updateRegistrationByStripePI(pi.ID, models.PaymentStatusFailed)
	}

	return nil, nil
}

func (s *RegistrationService) updateRegistrationByStripePI(paymentIntentID string, status models.PaymentStatus) (*models.CampRegistration, error) {
	var reg models.CampRegistration
	err := s.db.QueryRow(`
		UPDATE camp_registrations
		SET payment_status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE stripe_payment_intent_id = $2 AND payment_status != 'paid'
		RETURNING id, athlete_id, camp_id, payment_status, payment_method,
		          stripe_payment_intent_id, paypal_order_id, amount_cents,
		          parent_email, created_at, updated_at
	`, status, paymentIntentID).Scan(
		&reg.ID, &reg.AthleteID, &reg.CampID, &reg.PaymentStatus, &reg.PaymentMethod,
		&reg.StripePaymentIntentID, &reg.PaypalOrderID, &reg.AmountCents,
		&reg.ParentEmail, &reg.CreatedAt, &reg.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // Already processed or not found, idempotent
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update registration: %w", err)
	}
	return &reg, nil
}

// PayPal methods

func (s *RegistrationService) getPayPalAccessToken() (string, error) {
	s.paypalTokenLock.Lock()
	defer s.paypalTokenLock.Unlock()

	if s.paypalToken != "" && time.Now().Before(s.paypalTokenExp) {
		return s.paypalToken, nil
	}

	clientID := os.Getenv("PAYPAL_CLIENT_ID")
	clientSecret := os.Getenv("PAYPAL_CLIENT_SECRET")
	apiBase := os.Getenv("PAYPAL_API_BASE")
	if apiBase == "" {
		apiBase = "https://api-m.sandbox.paypal.com"
	}

	req, err := http.NewRequest("POST", apiBase+"/v1/oauth2/token", bytes.NewBufferString("grant_type=client_credentials"))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("paypal token request failed: %s", string(body))
	}

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	s.paypalToken = result.AccessToken
	s.paypalTokenExp = time.Now().Add(time.Duration(result.ExpiresIn-60) * time.Second)
	return s.paypalToken, nil
}

func (s *RegistrationService) CreatePayPalOrder(registrationID int) (string, error) {
	reg, err := s.GetRegistrationByID(registrationID)
	if err != nil {
		return "", err
	}

	token, err := s.getPayPalAccessToken()
	if err != nil {
		return "", fmt.Errorf("failed to get paypal access token: %w", err)
	}

	apiBase := os.Getenv("PAYPAL_API_BASE")
	if apiBase == "" {
		apiBase = "https://api-m.sandbox.paypal.com"
	}

	dollars := fmt.Sprintf("%.2f", float64(reg.AmountCents)/100)
	orderBody := map[string]interface{}{
		"intent": "CAPTURE",
		"purchase_units": []map[string]interface{}{
			{
				"amount": map[string]interface{}{
					"currency_code": "USD",
					"value":         dollars,
				},
				"custom_id": fmt.Sprintf("registration_%d", reg.ID),
			},
		},
	}

	body, err := json.Marshal(orderBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", apiBase+"/v2/checkout/orders", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("paypal order creation failed: %s", string(respBody))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	// Store paypal order ID on the registration
	_, err = s.db.Exec(`
		UPDATE camp_registrations
		SET paypal_order_id = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, result.ID, registrationID)
	if err != nil {
		return "", fmt.Errorf("failed to update registration with paypal order: %w", err)
	}

	return result.ID, nil
}

func (s *RegistrationService) CapturePayPalOrder(registrationID int, paypalOrderID string) error {
	reg, err := s.GetRegistrationByID(registrationID)
	if err != nil {
		return err
	}

	if reg.PaymentStatus == models.PaymentStatusPaid {
		return nil // Already paid, idempotent
	}

	if reg.PaypalOrderID == nil || *reg.PaypalOrderID != paypalOrderID {
		return errors.New("paypal order ID mismatch")
	}

	token, err := s.getPayPalAccessToken()
	if err != nil {
		return fmt.Errorf("failed to get paypal access token: %w", err)
	}

	apiBase := os.Getenv("PAYPAL_API_BASE")
	if apiBase == "" {
		apiBase = "https://api-m.sandbox.paypal.com"
	}

	req, err := http.NewRequest("POST", apiBase+"/v2/checkout/orders/"+paypalOrderID+"/capture", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("paypal capture failed: %s", string(respBody))
	}

	var result struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result.Status != "COMPLETED" {
		return fmt.Errorf("paypal capture status: %s", result.Status)
	}

	_, err = s.db.Exec(`
		UPDATE camp_registrations
		SET payment_status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, models.PaymentStatusPaid, registrationID)
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	return nil
}

// Query methods

func (s *RegistrationService) GetRegistrationByID(id int) (*models.CampRegistration, error) {
	var reg models.CampRegistration
	err := s.db.QueryRow(`
		SELECT id, athlete_id, camp_id, payment_status, payment_method,
		       stripe_payment_intent_id, paypal_order_id, amount_cents,
		       parent_email, created_at, updated_at
		FROM camp_registrations
		WHERE id = $1
	`, id).Scan(
		&reg.ID, &reg.AthleteID, &reg.CampID, &reg.PaymentStatus, &reg.PaymentMethod,
		&reg.StripePaymentIntentID, &reg.PaypalOrderID, &reg.AmountCents,
		&reg.ParentEmail, &reg.CreatedAt, &reg.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("registration not found")
	}
	if err != nil {
		return nil, err
	}

	return &reg, nil
}

type RegistrationWithAthlete struct {
	Registration models.CampRegistration `json:"registration"`
	Athlete      models.Athlete          `json:"athlete"`
}

func (s *RegistrationService) GetRegistrationsByCampID(campID int) ([]RegistrationWithAthlete, error) {
	rows, err := s.db.Query(`
		SELECT cr.id, cr.athlete_id, cr.camp_id, cr.payment_status, cr.payment_method,
		       cr.stripe_payment_intent_id, cr.paypal_order_id, cr.amount_cents,
		       cr.parent_email, cr.created_at, cr.updated_at,
		       a.id, a.name, a.age, a.years_played, a.position, a.user_id,
		       a.parent_email, a.parent_phone, a.created_at, a.updated_at
		FROM camp_registrations cr
		JOIN athletes a ON cr.athlete_id = a.id
		WHERE cr.camp_id = $1
		ORDER BY cr.created_at DESC
	`, campID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []RegistrationWithAthlete
	for rows.Next() {
		var r RegistrationWithAthlete
		err := rows.Scan(
			&r.Registration.ID, &r.Registration.AthleteID, &r.Registration.CampID,
			&r.Registration.PaymentStatus, &r.Registration.PaymentMethod,
			&r.Registration.StripePaymentIntentID, &r.Registration.PaypalOrderID,
			&r.Registration.AmountCents, &r.Registration.ParentEmail,
			&r.Registration.CreatedAt, &r.Registration.UpdatedAt,
			&r.Athlete.ID, &r.Athlete.Name, &r.Athlete.Age, &r.Athlete.YearsPlayed,
			&r.Athlete.Position, &r.Athlete.UserID, &r.Athlete.ParentEmail,
			&r.Athlete.ParentPhone, &r.Athlete.CreatedAt, &r.Athlete.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}

	return results, nil
}

func (s *RegistrationService) GetAthleteByID(id int) (*models.Athlete, error) {
	var a models.Athlete
	err := s.db.QueryRow(`
		SELECT id, name, age, years_played, position, user_id,
		       parent_email, parent_phone, created_at, updated_at
		FROM athletes
		WHERE id = $1
	`, id).Scan(
		&a.ID, &a.Name, &a.Age, &a.YearsPlayed, &a.Position, &a.UserID,
		&a.ParentEmail, &a.ParentPhone, &a.CreatedAt, &a.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("athlete not found")
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}
