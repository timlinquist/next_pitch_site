CREATE TABLE IF NOT EXISTS camp_age_groups (
    id SERIAL PRIMARY KEY,
    camp_id INTEGER NOT NULL REFERENCES camps(id) ON DELETE CASCADE,
    min_age INTEGER NOT NULL,
    max_age INTEGER NOT NULL,
    max_capacity INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_age_range CHECK (min_age <= max_age),
    CONSTRAINT chk_capacity_positive CHECK (max_capacity > 0)
);
CREATE INDEX idx_camp_age_groups_camp_id ON camp_age_groups(camp_id);

CREATE TRIGGER update_camp_age_groups_updated_at
    BEFORE UPDATE ON camp_age_groups
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
