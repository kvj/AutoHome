package main

import (
	"kvj/autohome/api"
	"kvj/autohome/data"
	_ "log"
	"os"
)

func main() {
	config := data.MakeStatusConfig()
	if config == nil {
		return
	}
	status, err := api.MakeStatus(config)
	if err != "" {
		// Error
		os.Stdout.Write([]byte("Err: " + err + "\n"))
	} else {
		os.Stdout.Write([]byte(status + "\n"))
	}
}
