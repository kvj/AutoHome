package model

type MeasureMessage struct {
	Type, Sensor, Measure int
	Value                 float64
}

type MMChannel chan *MeasureMessage
