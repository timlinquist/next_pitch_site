-- Create the recurrence type enum
CREATE TYPE recurrence_type AS ENUM ('none', 'weekly', 'biweekly', 'monthly');

-- Add the recurrence column to schedule_entries table with default value of 'none'
ALTER TABLE schedule_entries ADD COLUMN recurrence recurrence_type NOT NULL DEFAULT 'none'; 