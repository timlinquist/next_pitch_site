CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add user_id to schedule_entries table as nullable first
ALTER TABLE schedule_entries 
ADD COLUMN user_id INTEGER NULL;

-- Add foreign key constraint
ALTER TABLE schedule_entries
ADD CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id);

-- Add index on email for faster lookups
CREATE INDEX idx_users_email ON users(email); 