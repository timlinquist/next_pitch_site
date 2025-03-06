DROP TRIGGER IF EXISTS update_schedule_entries_updated_at ON schedule_entries;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS schedule_entries; 