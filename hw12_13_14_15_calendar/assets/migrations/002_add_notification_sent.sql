-- Add notification_sent column to events table
ALTER TABLE events
ADD COLUMN notification_sent BOOLEAN NOT NULL DEFAULT FALSE;

-- Create an index on notification_sent for efficient querying
CREATE INDEX idx_events_notification_sent ON events(notification_sent);

---- create above / drop below ----

-- Drop the index
DROP INDEX IF EXISTS idx_events_notification_sent;

-- Remove the notification_sent column
ALTER TABLE events
DROP COLUMN IF EXISTS notification_sent;
