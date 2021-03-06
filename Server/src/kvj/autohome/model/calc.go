package model

import (
	// "log"
	"math"
)

type ValueNormalizer func(value float64) float64

func TempNormalizer(value float64) float64 {
	return value/10 - 50
}

func LightNormalizer(value float64) float64 {
	return value / 4
}

type Calculator interface {
	Calculate(message *MeasureMessage) *MeasureMessage
}

type AverageCalculator struct {
	Max        int
	Normalizer ValueNormalizer
	times      int
	value      float64
}

func (self *AverageCalculator) reset() {
	self.times = 0
	self.value = 0
}

func (self *AverageCalculator) Calculate(message *MeasureMessage) *MeasureMessage {
	self.value += message.Value
	self.times++
	// log.Printf("Average %v %v:", self, message)
	if self.times >= self.Max {
		message.Value = self.value / float64(self.times)
		if self.Normalizer != nil {
			message.Value = self.Normalizer(message.Value)
		}
		self.reset()
		return message
	}
	return nil
}

type AverageLimitCalculator struct {
	AverageCalculator
	Limit float64
	prev  float64
}

func (self *AverageLimitCalculator) Calculate(message *MeasureMessage) *MeasureMessage {
	prev := self.prev
	self.prev = message.Value
	if math.Abs(prev-message.Value) > self.Limit {
		self.reset()
		if self.Normalizer != nil {
			message.Value = self.Normalizer(message.Value)
		}
		return message // Dropped over limit
	}
	self.value += message.Value
	self.times++
	// log.Printf("Average %v %v:", self, message)
	if self.times >= self.Max {
		message.Value = self.value / float64(self.times)
		if self.Normalizer != nil {
			message.Value = self.Normalizer(message.Value)
		}
		self.reset()
		return message
	}
	return nil
}

type NopeCalculator struct {
}

func (self *NopeCalculator) Calculate(message *MeasureMessage) *MeasureMessage {
	return nil // Ignore all values
}

type calculatorPair struct {
	index      int
	stype      int
	measure    int
	calculator Calculator
}

var calculators []*calculatorPair

func AddCalculator(index int, stype int, measure int, calculator Calculator) {
	calculators = append(calculators, &calculatorPair{
		index:      index,
		stype:      stype,
		measure:    measure,
		calculator: calculator,
	})
}

func InvokeCalculator(index int, message *MeasureMessage) *MeasureMessage {
	for _, pair := range calculators {
		if pair.index != -1 && index != pair.index {
			continue
		}
		if pair.stype != -1 && message.Type != pair.stype {
			continue
		}
		if pair.measure != -1 && message.Measure != pair.measure {
			continue
		}
		return pair.calculator.Calculate(message)
	}
	return message // Not found
}
