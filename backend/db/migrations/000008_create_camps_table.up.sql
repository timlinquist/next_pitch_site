CREATE TABLE IF NOT EXISTS camps (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    price_cents INTEGER NOT NULL,
    max_capacity INTEGER,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER update_camps_updated_at
    BEFORE UPDATE ON camps
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
