package models

import "time"

// GinLogEvent represents a parsed Gin log entry
type GinLogEvent struct {
	Level     string    `json:"level"`
	Timestamp time.Time `json:"timestamp"`
	Prefix    string    `json:"prefix"`
	Message   string    `json:"message"`
	Error     string    `json:"error,omitempty"`
}
