-- Rename s3_url column back to dropbox_url
ALTER TABLE video_uploads RENAME COLUMN s3_url TO dropbox_url; 