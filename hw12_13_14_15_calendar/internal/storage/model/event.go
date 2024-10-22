package model

import (
	"errors"
	"time"
)

var (
	ErrEventNotFound        = errors.New("event not found")
	ErrAlreadyExists        = errors.New("event already exists")
	ErrEmptyID              = errors.New("empty event id")
	ErrEventAlreadyNotified = errors.New("event notification already sent")
)

type Event struct {
	ID               string
	Title            string
	StartTime        time.Time
	EndTime          time.Time
	UserID           string
	NotifyDelta      int // in minutes
	NotificationSent bool
}
