package main

import (
	"kvj/autohome/data"
	"kvj/autohome/internet"
	"kvj/autohome/serial"
	"log"
)

func main() {
	log.Printf("Will test read")
	config := data.MakeConfig()
	if config == nil {
		return
	}
	db := data.OpenDB(config)
	talker := serial.NewTalker(db)
	talker.AddDevice(&serial.SerialConnection{
		Device: "/dev/rfcomm0",
		Index:  0,
	})
	talker.AddDevice(&serial.SerialConnection{
		Device: "/dev/rfcomm1",
		Index:  1,
	})
	talker.AddDevice(&serial.SerialConnection{
		Device: "/dev/rfcomm2",
		Index:  2,
	})
	talker.AddDevice(&serial.SerialConnection{
		Device: "/dev/rfcomm3",
		Index:  3,
	})
	mm, mms := internet.StartWeatherCrawler(0, "locid:JATY0021")
	talker.AddMessageProvider(10, mm, mms)
	go func() {
		internet.StartCameraProxy(config)
	}()
	talker.Start()
}
