package main

import (
	"kvj/autohome/data"
	"kvj/autohome/internet"
	"kvj/autohome/serial"
	"log"
)

func main() {
	log.Printf("Internet crawlers test")
	db, err := data.OpenDB()
	if err != nil {
		log.Fatal("Failed to open DB connection: %v", err)
	}
	talker := serial.NewTalker(db)
	talker.AddMessageProvider(10, internet.StartWeatherCrawler(0, "locid:JATY0021"))
	talker.Start()
}
