package model

type MeasureMessage struct {
	Type, Sensor, Measure int
	Value                 float32
}
