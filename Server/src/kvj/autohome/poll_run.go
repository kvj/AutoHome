package main

import (
	"kvj/autohome/data"
	"kvj/autohome/serial"
	"log"
)

func main() {
	log.Printf("Will test read")
	db, err := data.OpenDB("postgres://arduino:arduino@localhost/arduino")
	if err != nil {
		log.Fatal("Failed to open DB connection: %v", err)
	}
	talker := serial.NewTalker(db)
	talker.AddDevice(&serial.SerialConnection{
		Device: "/dev/rfcomm0",
		Index:  0,
	})
	talker.AddDevice(&serial.SerialConnection{
		Device: "/dev/rfcomm1",
		Index:  1,
	})
	talker.Start()
}
