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

func chr2byte(buffer []byte, index int) byte {
	return buffer[index] - 0x40 + ((buffer[index+1] - 0x40) << 4)
}

func byte2chr(value byte, buffer []byte, index int) {
	buffer[index] = (value & 15) + 0x40
	buffer[index+1] = ((value >> 4) & 15) + 0x40
}

func (self *SerialConnection) listen() {
	c := &serial.Config{Name: self.Device, Baud: 9600}
	handleMessage := func(buffer []byte) {
		// log.Printf("Message %v: %q %v", self.Index, buffer, len(buffer))
		if len(buffer) == 0 {
			return
		}
		// Some test
		if buffer[0] == Message_Measure {
			for i := 1; i < len(buffer)-4; i += 5 {
				var value int
				value = int(buffer[i+4])
				value = (value << 8) + int(buffer[i+3])
				message := &model.MeasureMessage{
					Type:    int(buffer[i]),
					Sensor:  int(buffer[i+1]),
					Measure: int(buffer[i+2]),
					Value:   float64(value),
				}
				// log.Printf("Message prepared:", message)
				self.queue <- message
			}
			return
		}
		log.Printf("Unknown message %v: %q %v", self.Index, buffer, len(buffer))
	}
	message := &model.MeasureMessage{
		Type:    100,
		Sensor:  0,
		Measure: 0,
		Value:   0,
	}
	self.queue <- message
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
		message := &model.MeasureMessage{
			Type:    100,
			Sensor:  0,
			Measure: 0,
			Value:   1,
		}
		self.queue <- message
		for {
			// Listen for messages
			sizeBuf := make([]byte, 1)
			bytesRead, err := io.ReadFull(s, sizeBuf)
			if err != nil || bytesRead != 1 {
				log.Printf("Error reading size %v %v", err, bytesRead)
				break
			}
			// log.Printf("Size came: %v: %v %v", self.Index, bytesRead, int(sizeBuf[0]))
			var inBuf []byte
			if sizeBuf[0] == 0 {
				sizeBuf = make([]byte, 2)
				bytesRead, err = io.ReadFull(s, sizeBuf)
				if err != nil {
					log.Printf("Error reading size2 %v %v", err, bytesRead)
					break
				}
				bytes := chr2byte(sizeBuf, 0)
				// log.Printf("Size2: %v %v", sizeBuf, bytes)
				inBuf = make([]byte, bytes)
				inDoubleBuf := make([]byte, 2*bytes)
				bytesRead, err = io.ReadFull(s, inDoubleBuf)
				// log.Printf("Incoming size: %v: %v - %v: %v", self.Index, int(sizeBuf[0]), bytesRead, err)
				if err != nil {
					log.Printf("Error reading data %v", err)
					break
				}
				for i := 0; i < int(bytes); i++ {
					inBuf[i] = chr2byte(inDoubleBuf, 2*i)
				}
				// log.Printf("Read: %v %v %v", bytesRead, inBuf, inDoubleBuf)
			} else {
				inBuf = make([]byte, sizeBuf[0])
				bytesRead, err = io.ReadFull(s, inBuf)
				// log.Printf("Incoming size: %v: %v - %v: %v", self.Index, int(sizeBuf[0]), bytesRead, err)
				if err != nil {
					log.Printf("Error reading data %v", err)
					break
				}
			}
			handleMessage(inBuf)
		}
		s.Close()
		message = &model.MeasureMessage{
			Type:    100,
			Sensor:  0,
			Measure: 0,
			Value:   0,
		}
		self.queue <- message
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
	buf := make([]byte, 3+2*len(buffer))
	buf[0] = 0
	byte2chr(byte(len(buffer)), buf, 1)
	for index, i := 3, 0; i < len(buffer); i++ {
		byte2chr(buffer[i], buf, index)
		index += 2
	}
	// log.Printf("Sending: %v", buf)
	_, err := self.current.Write(buf)
	return err
}
