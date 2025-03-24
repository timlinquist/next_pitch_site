-- Add columns for recurring events
ALTER TABLE schedule_entries 
    ADD COLUMN recurrence_end_date TIMESTAMP WITH TIME ZONE,
    ADD COLUMN parent_event_id INTEGER REFERENCES schedule_entries(id);

-- Create an index on parent_event_id for faster lookups
CREATE INDEX idx_schedule_entries_parent_event_id ON schedule_entries(parent_event_id); 