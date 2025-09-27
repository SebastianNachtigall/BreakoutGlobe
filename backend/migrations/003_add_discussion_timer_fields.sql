-- Add discussion timer fields to POIs table
-- Migration: 003_add_discussion_timer_fields.sql

-- Add discussion timer fields
ALTER TABLE pois ADD COLUMN discussion_start_time TIMESTAMP NULL;
ALTER TABLE pois ADD COLUMN is_discussion_active BOOLEAN DEFAULT FALSE NOT NULL;
ALTER TABLE pois ADD COLUMN discussion_duration INTEGER DEFAULT 0 NOT NULL;

-- Add index on discussion_start_time for performance
CREATE INDEX idx_pois_discussion_start_time ON pois(discussion_start_time);

-- Add index on is_discussion_active for filtering active discussions
CREATE INDEX idx_pois_is_discussion_active ON pois(is_discussion_active);