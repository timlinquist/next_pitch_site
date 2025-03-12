CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    phone_number VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add user_id to schedule_entries table as nullable first
ALTER TABLE schedule_entries 
ADD COLUMN user_id INTEGER NULL;

-- Add foreign key constraint
ALTER TABLE schedule_entries
ADD CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id);

-- Add index on email for faster lookups
CREATE INDEX idx_users_email ON users(email); 