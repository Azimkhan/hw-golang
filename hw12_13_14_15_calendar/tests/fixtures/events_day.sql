-- Insert 2 events for 2024-10-13 and one event for 2024-10-14
INSERT INTO events (id, title, start_time, end_time, user_id, notify_delta, notification_sent)
VALUES
    ('event1', 'Event 1', '2024-10-13 10:00:00', '2024-10-13 11:00:00', 'user1', 10, false),
    ('event2', 'Event 2', '2024-10-13 12:00:00', '2024-10-13 13:00:00', 'user2', 10, false),
    ('event3', 'Event 3', '2024-10-14 10:00:00', '2024-10-14 11:00:00', 'user1', 10, false);
