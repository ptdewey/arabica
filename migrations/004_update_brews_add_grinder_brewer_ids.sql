-- Update brews table to use grinder_id and brewer_id instead of strings

-- Add new columns for foreign keys
ALTER TABLE brews ADD COLUMN grinder_id INTEGER;
ALTER TABLE brews ADD COLUMN brewer_id INTEGER;

-- Add foreign key constraints (SQLite doesn't enforce these on ALTER TABLE, but good for documentation)
-- FOREIGN KEY (grinder_id) REFERENCES grinders(id)
-- FOREIGN KEY (brewer_id) REFERENCES brewers(id)

-- Note: We'll keep the old grinder and method columns for now to preserve existing data
-- They can be removed in a future migration if needed
