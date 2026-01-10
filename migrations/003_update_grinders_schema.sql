-- SQLite doesn't support DROP COLUMN, so we need to:
-- 1. Create a new table with the desired schema
-- 2. Copy data from old table
-- 3. Drop old table
-- 4. Rename new table

-- Create new grinders table with updated schema
CREATE TABLE grinders_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    grinder_type TEXT NOT NULL CHECK(grinder_type IN ('Hand', 'Electric', 'Portable Electric')),
    burr_type TEXT CHECK(burr_type IN ('Conical', 'Flat', '')),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Copy existing data from old table (map 'type' to 'grinder_type')
-- Since we don't know the old type values, we'll default to 'Hand'
INSERT INTO grinders_new (id, name, grinder_type, burr_type, notes, created_at)
SELECT id, name, 
    CASE 
        WHEN type LIKE '%electric%' OR type LIKE '%Electric%' THEN 'Electric'
        WHEN type LIKE '%hand%' OR type LIKE '%Hand%' OR type LIKE '%manual%' OR type LIKE '%Manual%' THEN 'Hand'
        ELSE 'Hand'
    END as grinder_type,
    CASE
        WHEN type LIKE '%flat%' OR type LIKE '%Flat%' THEN 'Flat'
        WHEN type LIKE '%conical%' OR type LIKE '%Conical%' THEN 'Conical'
        ELSE ''
    END as burr_type,
    notes, 
    created_at
FROM grinders;

-- Drop old table
DROP TABLE grinders;

-- Rename new table
ALTER TABLE grinders_new RENAME TO grinders;
