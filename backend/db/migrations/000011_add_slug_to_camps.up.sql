ALTER TABLE camps ADD COLUMN slug VARCHAR(255);
CREATE UNIQUE INDEX idx_camps_slug ON camps(slug);
