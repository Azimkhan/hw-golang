-- Insert event for 2024-10-1, 2024-10-31, 2024-11-1
INSERT INTO events (id, title, start_time, end_time, user_id, notify_delta, notification_sent)
VALUES
    ('event101', 'Event 1', '2024-10-01 10:00:00', '2024-10-01 11:00:00', 'user1', 10, false),
    ('event107', 'Event 2', '2024-10-31 12:00:00', '2024-10-31 13:00:00', 'user2', 10, false),
    ('event111', 'Event 3', '2024-11-01 10:00:00', '2024-11-01 11:00:00', 'user1', 10, false);
