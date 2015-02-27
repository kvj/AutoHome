package main

import (
	"kvj/autohome/api"
	"kvj/autohome/data"
	// "kvj/autohome/internet"
	"log"
)

func main() {
	log.Printf("Starting server...")
	config := data.MakeConfig()
	if config == nil {
		return
	}
	db := data.OpenDB(config)
	// _ = internet.StartWeatherNotifier(db, 10, 0)
	api.StartServer(config, db)
}
