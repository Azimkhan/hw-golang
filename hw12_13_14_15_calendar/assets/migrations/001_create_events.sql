CREATE TABLE events
(
    id           VARCHAR(255) PRIMARY KEY,
    title        VARCHAR(255) NOT NULL,
    start_time   TIMESTAMP WITH TIME ZONE    NOT NULL,
    end_time     TIMESTAMP  WITH TIME ZONE    NOT NULL,
    user_id      VARCHAR(255) NOT NULL,
    notify_delta INTEGER      NOT NULL
);

-- Create an index on user_id for faster lookups
CREATE INDEX idx_events_user_id ON events (user_id);

-- Create an index on start_time for efficient date range queries
CREATE INDEX idx_events_time_range ON events(start_time, end_time);

---- create above / drop below ----

drop table events;
