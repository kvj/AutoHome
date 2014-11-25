package serial

import (
	"encoding/json"
	"kvj/autohome/data"
	"kvj/autohome/model"
	"log"
	"os"
	"os/signal"
	"time"
)

type arduinoTalker struct {
	devices []*SerialConnection
	ticker  *time.Ticker
	db      *data.DBProvider
}

func NewTalker(db *data.DBProvider) *arduinoTalker {
	talker := &arduinoTalker{}
	talker.devices = make([]*SerialConnection, 0)
	talker.db = db
	return talker
}

func (self *arduinoTalker) close() {
	self.ticker.Stop()
	for _, device := range self.devices {
		device.Close()
	}
}

func (self *arduinoTalker) poll() {
	// log.Printf("Time to poll")
	buf := []byte{Message_Measure}
	for _, device := range self.devices {
		go func(device *SerialConnection) {
			err := device.Send(buf)
			if err != nil {
				log.Printf("Sent: %v", err)
			}
		}(device)
	}
}

func (self *arduinoTalker) Start() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
	log.Printf("Main thread")
	self.ticker = time.NewTicker(10 * time.Second)
	for {
		select {
		case signal := <-signals:
			log.Printf("Received signal: %v", signal)
			self.close()
			return
		case <-self.ticker.C:
			self.poll()
			// case n := <-self.db.SensorChannel:
			//	log.Printf("Here is Sensor message: %v", n)
		}
	}
}

func (self *arduinoTalker) notifyWithMessage(index int, message *model.MeasureMessage) {
	m := &model.MeasureNotification{
		Device:  index,
		Type:    message.Type,
		Index:   message.Sensor,
		Measure: message.Measure,
		Value:   message.Value,
	}
	bodyOutBytes, err := json.Marshal(m)
	if err == nil {
		self.db.Notify("sensor", string(bodyOutBytes))
	} else {
		log.Printf("JSON error: %v", err)
	}

}

func (self *arduinoTalker) AddMessageProvider(index int, ch model.MMChannel, forecast model.MMsChannel) {
	go func() {
		for message := range ch {
			// log.Printf("Message came: %v %v", index, message)
			log.Printf("Message[P]: %v %v", index, message)
			err := self.db.AddMeasure(index, message)
			if err != nil {
				log.Printf("Error: %v", err)
				continue
			}
			self.notifyWithMessage(index, message)
		}
	}()
	go func() {
		for messages := range forecast {
			// log.Printf("Message came: %v %v", index, message)
			if len(messages) == 0 {
				continue
			}
			log.Printf("Forecast: %v %v", index, len(messages))
			err := self.db.AddMeasures(data.TypeForecast, index, messages)
			if err != nil {
				log.Printf("Forecast add failed: %v", err)
				continue
			}
			log.Printf("Forecast added: %v", len(messages))
		}
	}()
}

func (self *arduinoTalker) AddDevice(connection *SerialConnection) {
	index := len(self.devices)
	self.devices = append(self.devices, connection)
	model.AddCalculator(index, 0, 0, &model.AverageCalculator{
		Max: 20,
	})
	model.AddCalculator(index, 0, 1, &model.AverageCalculator{
		Max: 10,
	})
	model.AddCalculator(index, 0, 2, &model.AverageCalculator{
		Max: 10,
	})
	model.AddCalculator(index, 2, 0, &model.AverageLimitCalculator{
		AverageCalculator: model.AverageCalculator{
			Max: 30,
		},
		Limit: 10,
	})
	c := connection.Connect()
	go func() {
		for message := range c {
			// log.Printf("Message came: %v %v", index, message)
			message = model.InvokeCalculator(index, message)
			if message == nil {
				continue
			}
			log.Printf("Message: %v %v", index, message)
			err := self.db.AddMeasure(index, message)
			if err != nil {
				log.Printf("Error: %v", err)
				continue
			}
			self.notifyWithMessage(index, message)
		}
	}()
}
