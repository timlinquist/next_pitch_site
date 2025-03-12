-- Drop the foreign key constraint first
ALTER TABLE schedule_entries DROP CONSTRAINT fk_user;

-- Drop the user_id column
ALTER TABLE schedule_entries DROP COLUMN user_id;

-- Drop the index
DROP INDEX IF EXISTS idx_users_email;

-- Drop the users table
DROP TABLE IF EXISTS users; 