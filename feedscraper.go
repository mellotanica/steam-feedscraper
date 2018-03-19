package main

import (
	"log"
	"feedscraper/webservice"
	"time"
	"feedscraper/scraper"
)

var cacheUpdateDelay = "1h"

func keepCacheUpdated() {
	sleepDuration, err := time.ParseDuration(cacheUpdateDelay)
	if err != nil {
		sleepDuration = time.Hour
	}
	for true {
		log.Print("Updating cache...")
		scraper.Update_all()
		log.Print("Cache updated.")

		time.Sleep(sleepDuration)
	}
}

func main()  {
	go keepCacheUpdated()
	webservice.StartService()
}