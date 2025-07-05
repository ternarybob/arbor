package models

import (
	"time"
	"github.com/phuslu/log"
)

type LogEvent struct {
	Index         uint64                 `json:"index"`
	Level         log.Level              `json:"level"`
	Timestamp     time.Time              `json:"time"`
	CorrelationID string                 `json:"correlationid"`
	Prefix        string                 `json:"prefix"`
	Message       string                 `json:"message"`
	Error         string                 `json:"error"`
	Function      string                 `json:"function"`
	Fields        map[string]interface{} `json:"fields"`
}
