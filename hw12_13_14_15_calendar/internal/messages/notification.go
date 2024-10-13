package messages

import "time"

type Notification struct {
	Title     string    `json:"title"`
	UserID    string    `json:"userId"`
	StartTime time.Time `json:"startTime"`
}
