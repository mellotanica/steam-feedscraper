package scraper

import (
	"feedscraper/games_cache"
	"feedscraper/scraper/scrapers"
	"log"
	"strings"
	"github.com/mmcdole/gofeed"
	"feedscraper/configs"
)

type dataSource struct {
	url string
	parser func (*gofeed.Item) (*games_cache.Game, error)
}

// ############################
// # global scraper variables #
// ############################

var sources = []dataSource {
	dataSource{"http://feeds.feedburner.com/SkidrowReloadedGames", scrapers.ParseSkidRowReloaded},
	dataSource{"https://feeds.feedburner.com/skidrowgamesfeed", scrapers.ParseSkidRowReloaded},
	dataSource{"https://feeds.feedburner.com/skidrowgames", scrapers.ParseSkidRowCrack},
	dataSource{"http://feeds.feedburner.com/skidrowcrack", scrapers.ParseSkidRowCrack},
	dataSource{"http://fitgirl-repacks.site/feed/", scrapers.ParseFitGirlRepack},
}


// ##################
// # scraper engine #
// ##################

func scrapeSource(source dataSource) (*[]games_cache.Game, *[]games_cache.Game) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(source.url)
	if err != nil {
		log.Printf("unable to parse feed %s: %s\n", source.url, err.Error())
		return nil, nil
	}

	var games = make([]games_cache.Game, 0, games_cache.MediumCacheSize)
	var dubious_games = make([]games_cache.Game, 0, games_cache.MediumCacheSize)

	for _, i := range feed.Items {
		game, err := source.parser(i)
		if err != nil {
			switch err.(type){
			case scrapers.ScraperWarning:
				dubious_games = append(dubious_games, (*game))
			default:
				log.Printf("error parsing game (%s): %s\n", source.url, err.Error())
			}
		} else if game != nil {
			games = append(games, (*game))
		}
	}

	return &games, &dubious_games
}

func cleanList(list []games_cache.Game, excludes... *games_cache.Cache) (clean_list []games_cache.Game) {
	clean_list = make([]games_cache.Game, 0, len(list))

	config := configs.GetConfigs()

	for _, g := range list {
		skip := false
		for _, e := range excludes {
			if e.GameInList(g) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		if config.Blacklist != nil && len(config.Blacklist) > 0 {
			genre := strings.ToLower(g.Genre)
			for _, b := range config.Blacklist {
				bl := strings.ToLower(b)
				if strings.Contains(genre, bl) {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
		}
		clean_list = append(clean_list, g)
	}

	return
}



func updateCache(pending, dubious, checked *games_cache.Cache, scraped_list, scraped_dubious []games_cache.Game) {
	scraped_list = cleanList(scraped_list, checked)
	pending.AppendElements(scraped_list...)
	pending.Store()

	scraped_dubious = cleanList(scraped_dubious, pending, checked)
	dubious.AppendElements(scraped_dubious...)
	dubious.Store()
}


func Update_all() {
	pending_cache := games_cache.LoadCache(games_cache.GamesCachePendingFile)
	dubious_cache := games_cache.LoadCache(games_cache.GamesCacheDubiousFile)
	checked_cache := games_cache.LoadCache(games_cache.GamesCacheCheckedFile)

	config := configs.GetConfigs()
	if len(config.SteamUser) > 0 {
		scrapers.ScrapeSteam(checked_cache, config.SteamUser)
		checked_cache.Store()
	}

	for _, source := range sources {
		list, dubious := scrapeSource(source)
		if list != nil || dubious != nil {
			updateCache(pending_cache, dubious_cache, checked_cache, *list, *dubious)
		}
	}

	dubious_cache.CleanDuplicates(checked_cache)
	dubious_cache.Store()

	pending_cache.CleanDuplicates(checked_cache)
	pending_cache.Store()
}

