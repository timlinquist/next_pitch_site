ALTER TABLE camp_age_groups ADD COLUMN price_cents INTEGER;
UPDATE camp_age_groups SET price_cents = (SELECT price_cents FROM camps WHERE camps.id = camp_age_groups.camp_id);
ALTER TABLE camp_age_groups ALTER COLUMN price_cents SET NOT NULL;
ALTER TABLE camp_age_groups ADD CONSTRAINT chk_price_cents_positive CHECK (price_cents > 0);

ALTER TABLE camps ALTER COLUMN price_cents DROP NOT NULL;
