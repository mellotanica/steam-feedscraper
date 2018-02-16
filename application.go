package main

import (
	"log"
	"feedscraper/services"
	"time"
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
	go keepCacheUpdated()
	services.Start_service(8080)
}