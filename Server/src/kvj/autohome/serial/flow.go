package serial

import (
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
		}
	}
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
	c := connection.Connect()
	go func() {
		for message := range c {
			message = model.InvokeCalculator(index, message)
			if message == nil {
				continue
			}
			log.Printf("Message: %v %v", index, message)
			err := self.db.AddMeasure(index, message)
			if err != nil {
				log.Printf("Error: %v", err)
			}
		}
	}()
}
