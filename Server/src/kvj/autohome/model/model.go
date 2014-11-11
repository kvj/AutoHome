package model

import (
	"time"
)

type MeasureMessage struct {
	Type, Sensor, Measure int
	Value                 float64
	Time                  time.Time
}

type MeasureMessages []*MeasureMessage

type MMChannel chan *MeasureMessage
type MMsChannel chan MeasureMessages
