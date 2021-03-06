package main

import (
	"kvj/autohome/api"
	"kvj/autohome/data"
	"log"
)

func main() {
	log.Printf("Starting server...")
	config := data.MakeConfig()
	if config == nil {
		return
	}
	db := data.OpenDB(config)
	api.StartServer(config, db)
}
