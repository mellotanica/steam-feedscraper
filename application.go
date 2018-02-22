package main

import (
	"log"
	"feedscraper/services"
	"time"
	"os"
	"strconv"
)

var cacheUpdateDelay = "1h"

func keepCacheUpdated() {
	sleepDuration, err := time.ParseDuration(cacheUpdateDelay)
	if err != nil {
		sleepDuration = time.Hour
	}
	for true {
		log.Print("Updating cache...")
		services.Update_all()
		log.Print("Cache updated.")

		time.Sleep(sleepDuration)
	}
}

func main()  {
	var err error
	port := 8080

	if len(os.Args) > 1 {
		port, err = strconv.Atoi(os.Args[1])
		if err != nil || port < 1 || port > 65535 {
			log.Fatalf("ERROR: %s is not a valid port number!", os.Args[1])
		}
	}

	go keepCacheUpdated()
	services.StartService(port)
}