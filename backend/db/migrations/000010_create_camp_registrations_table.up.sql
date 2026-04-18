CREATE TYPE payment_status AS ENUM ('pending', 'paid', 'failed', 'refunded');
CREATE TYPE payment_method AS ENUM ('stripe', 'paypal');

CREATE TABLE IF NOT EXISTS camp_registrations (
    id SERIAL PRIMARY KEY,
    athlete_id INTEGER NOT NULL REFERENCES athletes(id) ON DELETE CASCADE,
    camp_id INTEGER NOT NULL REFERENCES camps(id) ON DELETE CASCADE,
    payment_status payment_status NOT NULL DEFAULT 'pending',
    payment_method payment_method,
    stripe_payment_intent_id VARCHAR(255),
    paypal_order_id VARCHAR(255),
    amount_cents INTEGER NOT NULL,
    parent_email VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(athlete_id, camp_id)
);

CREATE INDEX idx_camp_reg_camp_id ON camp_registrations(camp_id);
CREATE INDEX idx_camp_reg_athlete_id ON camp_registrations(athlete_id);
CREATE INDEX idx_camp_reg_stripe_pi ON camp_registrations(stripe_payment_intent_id);

CREATE TRIGGER update_camp_registrations_updated_at
    BEFORE UPDATE ON camp_registrations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
