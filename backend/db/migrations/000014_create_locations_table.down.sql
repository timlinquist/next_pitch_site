ALTER TABLE camps DROP COLUMN IF EXISTS location_id;
DROP TRIGGER IF EXISTS update_locations_updated_at ON locations;
DROP TABLE IF EXISTS locations;
