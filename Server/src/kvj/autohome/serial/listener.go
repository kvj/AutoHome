package serial

import (
	"github.com/tarm/goserial"
	"io"
	"kvj/autohome/model"
	"log"
	"time"
)

const (
	State_Not_Connected = iota
	State_Connected
	State_Closed

	Message_Measure = 0
)

type SerialConnection struct {
	Device  string
	Index   int
	state   int
	current io.WriteCloser
	queue   chan *model.MeasureMessage
}

func (self *SerialConnection) listen() {
	c := &serial.Config{Name: self.Device, Baud: 9600}
	handleMessage := func(buffer []byte) {
		if len(buffer) == 0 {
			return
		}
		// Some test
		// log.Printf("Message %v: %q %v", self.Index, buffer, len(buffer))
		if buffer[0] == Message_Measure {
			for i := 1; i < len(buffer)-3; i += 4 {
				message := &model.MeasureMessage{
					Type:    int(buffer[i]),
					Sensor:  int(buffer[i+1]),
					Measure: int(buffer[i+2]),
					Value:   float64(buffer[i+3]),
				}
				// log.Printf("Message prepared:", message)
				self.queue <- message
			}
			return
		}
		log.Printf("Unknown message %v: %q %v", self.Index, buffer, len(buffer))
	}
	for {
		if self.state == State_Closed {
			log.Printf("Terminating...")
			return
		}
		self.state = State_Not_Connected
		self.current = nil
		log.Printf("Re-connecting..,")
		s, err := serial.OpenPort(c)
		if err != nil {
			log.Printf("Failed to open port, will wait: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}
		self.state = State_Connected
		self.current = s
		log.Printf("Connected, waiting for message")
		for {
			// Listen for messages
			sizeBuf := make([]byte, 1)
			bytesRead, err := io.ReadFull(s, sizeBuf)
			if err != nil || bytesRead != 1 {
				log.Printf("Error reading size %v %v", err, bytesRead)
				break
			}
			// log.Printf("Size came: %v: %v %v", self.Index, bytesRead, int(sizeBuf[0]))
			if sizeBuf[0] == 0 {
				continue
			}
			inBuf := make([]byte, sizeBuf[0])
			bytesRead, err = io.ReadFull(s, inBuf)
			// log.Printf("Incoming size: %v: %v - %v: %v", self.Index, int(sizeBuf[0]), bytesRead, err)
			if err != nil {
				log.Printf("Error reading data %v", err)
				break
			}
			handleMessage(inBuf)
		}
		s.Close()
	}
}

func (self *SerialConnection) Connect() chan *model.MeasureMessage {
	self.state = State_Not_Connected
	self.queue = make(chan *model.MeasureMessage)
	go self.listen()
	return self.queue
}

func (self *SerialConnection) Close() {
	if self.state == State_Closed {
		return
	}
	log.Printf("Terminating connection: %v", self.Device)
	if self.state == State_Connected {
		self.state = State_Closed
		self.current.Close()
	}
	self.state = State_Closed
	close(self.queue)
}

type NotConnectedError struct{}

func (self *NotConnectedError) Error() string {
	return "Not connected"
}

func (self *SerialConnection) Send(buffer []byte) error {
	if self.state == State_Closed {
		return &NotConnectedError{}
	}
	if self.state == State_Not_Connected {
		return &NotConnectedError{}
	}
	// log.Printf("Send data: %v %v", len(buffer), buffer)
	buf := []byte{byte(len(buffer))}
	_, err := self.current.Write(buf)
	if err != nil {
		return err
	}
	_, err = self.current.Write(buffer)
	return err
}
