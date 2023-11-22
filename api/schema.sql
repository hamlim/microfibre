DROP TABLE IF EXISTS posts;
-- Create the posts table
CREATE TABLE IF NOT EXISTS posts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  body TEXT,
  location TEXT,
  created_time TIMESTAMP,
  created_timezone TEXT,
  updated_time TIMESTAMP,
  updated_timezone TEXT
);

DROP TABLE IF EXISTS media;
-- Create the media table
CREATE TABLE IF NOT EXISTS media (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  post_id INTEGER,
  url TEXT,
  alt_text TEXT,
  FOREIGN KEY (post_id) REFERENCES posts (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS secrets (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  secret TEXT
)
