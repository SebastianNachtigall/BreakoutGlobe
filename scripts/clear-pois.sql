-- Script to clear all POIs from the database
-- Run this with: docker compose exec postgres psql -U postgres -d breakoutglobe -f /scripts/clear-pois.sql

-- First, let's see what tables exist (uncomment to check)
-- \dt

-- Delete all POIs (assuming the table is named 'pois')
DELETE FROM pois;

-- If the table has a different name, try these alternatives:
-- DELETE FROM poi;
-- DELETE FROM points_of_interest;

-- Show count after deletion
SELECT COUNT(*) as remaining_pois FROM pois;

-- Optional: Reset the auto-increment sequence if using serial IDs
-- ALTER SEQUENCE pois_id_seq RESTART WITH 1;

COMMIT;