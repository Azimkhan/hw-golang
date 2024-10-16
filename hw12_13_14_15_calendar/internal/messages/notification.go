package messages

import "time"

type Notification struct {
	EventID   string    `json:"eventId"`
	Title     string    `json:"title"`
	UserID    string    `json:"userId"`
	StartTime time.Time `json:"startTime"`
}
