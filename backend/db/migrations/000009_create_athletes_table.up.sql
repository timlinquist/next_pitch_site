CREATE TABLE IF NOT EXISTS athletes (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    age INTEGER NOT NULL,
    years_played INTEGER NOT NULL DEFAULT 0,
    position VARCHAR(100),
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    parent_email VARCHAR(255) NOT NULL,
    parent_phone VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_athletes_user_id ON athletes(user_id);
CREATE INDEX idx_athletes_parent_email ON athletes(parent_email);

CREATE TRIGGER update_athletes_updated_at
    BEFORE UPDATE ON athletes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
