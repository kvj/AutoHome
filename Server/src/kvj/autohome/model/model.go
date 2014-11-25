package model

import (
	"time"
)

type MeasureMessage struct {
	Type, Sensor, Measure int
	Value                 float64
	Time                  time.Time
}

type MeasureNotification struct {
	Device    int     `json:"device"`
	Type      int     `json:"type"`
	Index     int     `json:"index"`
	Measure   int     `json:"measure"`
	Value     float64 `json:"value"`
	Timestamp int64   `json:"ts"`
}

type MeasureMessages []*MeasureMessage

type MMChannel chan *MeasureMessage
type MMsChannel chan MeasureMessages
