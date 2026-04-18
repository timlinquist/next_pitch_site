ALTER TABLE camp_age_groups DROP CONSTRAINT IF EXISTS chk_price_cents_positive;
ALTER TABLE camp_age_groups DROP COLUMN IF EXISTS price_cents;

UPDATE camps SET price_cents = 0 WHERE price_cents IS NULL;
ALTER TABLE camps ALTER COLUMN price_cents SET NOT NULL;
