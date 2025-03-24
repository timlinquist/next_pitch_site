-- Drop the index on parent_event_id
DROP INDEX IF EXISTS idx_schedule_entries_parent_event_id;

-- Remove the columns
ALTER TABLE schedule_entries 
    DROP COLUMN IF EXISTS recurrence_end_date,
    DROP COLUMN IF EXISTS parent_event_id; 