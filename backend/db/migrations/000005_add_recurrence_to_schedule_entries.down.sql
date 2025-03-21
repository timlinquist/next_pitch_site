-- Remove the recurrence column
ALTER TABLE schedule_entries DROP COLUMN recurrence;

-- Drop the recurrence type enum
DROP TYPE recurrence_type; 