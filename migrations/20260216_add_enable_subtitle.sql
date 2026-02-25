-- Add enable_subtitle column to video_generations table
-- SQLite doesn't support IF NOT EXISTS in ALTER TABLE, so we need to check first
-- This script will fail if the column already exists, which is acceptable

ALTER TABLE video_generations ADD COLUMN enable_subtitle BOOLEAN DEFAULT FALSE;
