-- Initial schema for Arabica coffee tracking

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS beans (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    origin TEXT NOT NULL,
    roast_level TEXT,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS roasters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    location TEXT,
    website TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS grinders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    type TEXT,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS brews (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL DEFAULT 1,
    bean_id INTEGER NOT NULL,
    roaster_id INTEGER NOT NULL,
    method TEXT NOT NULL,
    temperature REAL,
    time_seconds INTEGER,
    grind_size TEXT,
    grinder TEXT,
    tasting_notes TEXT,
    rating INTEGER CHECK(rating >= 1 AND rating <= 10),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (bean_id) REFERENCES beans(id),
    FOREIGN KEY (roaster_id) REFERENCES roasters(id)
);

-- Insert default user for single-user mode (ignore if exists)
INSERT OR IGNORE INTO users (username) VALUES ('default');

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_brews_user_id ON brews(user_id);
CREATE INDEX IF NOT EXISTS idx_brews_bean_id ON brews(bean_id);
CREATE INDEX IF NOT EXISTS idx_brews_roaster_id ON brews(roaster_id);
CREATE INDEX IF NOT EXISTS idx_brews_created_at ON brews(created_at DESC);
