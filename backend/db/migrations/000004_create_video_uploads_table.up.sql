-- Create video_upload_status enum type
CREATE TYPE video_upload_status AS ENUM ('uploaded', 'processed', 'paid');

-- Create video_uploads table
CREATE TABLE video_uploads (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    dropbox_url TEXT NOT NULL,
    file_name TEXT NOT NULL,
    status video_upload_status NOT NULL DEFAULT 'uploaded',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_user
        FOREIGN KEY(user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

-- Create index on user_id for faster lookups
CREATE INDEX idx_video_uploads_user_id ON video_uploads(user_id);

-- Create index on status for faster filtering
CREATE INDEX idx_video_uploads_status ON video_uploads(status); 