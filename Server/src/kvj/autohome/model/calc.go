package model

import (
//"log"
)

type Calculator interface {
	Calculate(message *MeasureMessage) *MeasureMessage
}

type AverageCalculator struct {
	Max   int
	times int
	value float32
}

func (self *AverageCalculator) Calculate(message *MeasureMessage) *MeasureMessage {
	self.value += message.Value
	self.times++
	// log.Printf("Average %v %v:", self, message)
	if self.times >= self.Max {
		message.Value = self.value / float32(self.times)
		self.times = 0
		self.value = 0
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
