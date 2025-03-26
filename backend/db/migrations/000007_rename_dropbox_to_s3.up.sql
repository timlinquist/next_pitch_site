-- Rename dropbox_url column to s3_url
ALTER TABLE video_uploads RENAME COLUMN dropbox_url TO s3_url; 