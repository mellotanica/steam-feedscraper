package main

import (
	"feedscraper/games_cache"
	"github.com/PuerkitoBio/goquery"
	"log"
	"fmt"
	"strings"
	"net/url"
	"github.com/mmcdole/gofeed"
	"regexp"
	"github.com/texttheater/golang-levenshtein/levenshtein"
)

type ScraperError struct {
	What string
}

func (e ScraperError) Error() string {
	return fmt.Sprintf("ScraperError: %s", e.What)
}

type ScraperWarning struct {
	What string
}

func (e ScraperWarning) Error() string {
	return fmt.Sprintf("ScraperWarning: %s", e.What)
}

var cleanSpacesRe = regexp.MustCompile(`\s+`)

func gameFromLinkAndName(link string, name string, genre string) (*games_cache.Game, error) {
	gid := ""

	if len(name) <= 0 {
		return nil, ScraperError{fmt.Sprintf("missing game name! (link: %s, genre: %s)", link, genre)}
	}

	if len(link) > 0 {
		url, err := url.Parse(link)
		if err != nil {
			return nil, ScraperError{fmt.Sprintf("invalid link: %s (name: %s, genre: %s)", link, name, genre)}
		}

		toks := strings.Split(url.Path, "/")
		if len(toks) > 1 {
			if len(toks[0]) > 0 {
				gid = fmt.Sprintf("%s/%s", toks[0], toks[1])
			} else {
				gid = fmt.Sprintf("%s/%s", toks[1], toks[2])
			}
		} else {
			return nil, ScraperError{fmt.Sprintf("invalid link: %s (name: %s, genre: %s)", link, name, genre)}
		}
	} else {
		cname := cleanSpacesRe.ReplaceAllLiteralString(name, " ")
		if len(cname) > len("Standard Edition") && cname[len(cname) - len("Standard Edition"):] == "Standard Edition" {
			cname = strings.TrimSpace(strings.TrimRight(cname, "Standard Edition"))
		}

		url := "http://store.steampowered.com/search/?term=" + url.PathEscape(cname)

		doc, err := goquery.NewDocument(url)
		if err != nil {
			return nil, ScraperError{fmt.Sprintf("unable to find game {} ({}) on steam", name, genre)}
		}

		link := ""
		minedit := -1

		as := doc.Find("div#search_result_container a")
		for i := range as.Nodes {
			a := as.Eq(i)
			gref := a.Find("div.col.search_name.ellipsis span.title")
			if gref != nil {
				game_name := cleanSpacesRe.ReplaceAllLiteralString(gref.Text(), " ")

				distance := levenshtein.DistanceForStrings([]rune(name), []rune(game_name), levenshtein.DefaultOptions)
				if minedit < 0 || distance < minedit {
					clink, exists := a.Attr("href")
					if exists {
						link = clink
						minedit = distance
					}
				}
				if distance == 0 {
					break
				}
			}
		}

		if len(link) > 0 {
			game, err := gameFromLinkAndName(link, name, genre)
			if err == nil && minedit > 4 {
				return game, ScraperWarning{fmt.Sprintf("search on steam uncertain.. name: %s, link: %s, distance: %d", name, link, minedit)}
			}
			return game, err
		}

		return &games_cache.Game{name, "", "", genre}, ScraperWarning{fmt.Sprintf("no link found: %s (%s)", name, genre)}
	}

	switch {
	case len(link) <= 0:
		return nil, ScraperError{"unable to retrieve game link"}
	case len(gid) <= 0:
		return nil, ScraperError{"unbale to retrieve game id"}
	default:
		return &games_cache.Game{name, gid, link, genre}, nil
	}
}

func parseSkidRowReloaded(item *gofeed.Item) (*games_cache.Game, error){
	name := ""
	genre := ""
	link := ""

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(item.Content))
	if err != nil {
		return nil, ScraperError{"unable to parse item: no content found"}
	}

	ps := doc.Find("p")
	for i := range ps.Nodes {
		p := ps.Eq(i)
		if strings.Contains(p.Text(), "Title:") || strings.Contains(p.Text(), "Genre:") {
			for _, l := range strings.Split(p.Text(), "\n") {
				if strings.Contains(l,"Title:") {
					name = strings.TrimSpace(l[strings.Index(l, "Title:")+len("Title:"):])
				} else if strings.Contains(l, "Genre:") {
					genre = strings.TrimSpace(l[strings.Index(l, "Genre:")+len("Genre:"):])
				}
			}
		}
		if len(name) > 0 && len(genre) > 0 {
			break
		}
	}

	as := doc.Find("a")
	for i := range as.Nodes {
		a := as.Eq(i)
		ref, existst := a.Attr("href")
		if existst && strings.Contains(ref, "store.steampowered.com"){
			link = ref
			break
		}
	}

	return gameFromLinkAndName(link, name, genre)
}

func parseSkidRowCrack(item *gofeed.Item) (*games_cache.Game, error){
	name := ""
	genre := ""
	link := ""

	title := strings.ToLower(item.Title)
	for _, c := range item.Categories {
		if len(c) <= len(title) && title[:len(c)] == strings.ToLower(c) && (len(name) <= 0 || len(c) < len(name)) {
			name = c
		}
	}

	if len(name) <= 0 {
		return nil, ScraperError{"unable to parse item: no name found"}
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(item.Content))
	if err != nil {
		log.Printf("unable to parse content, name: %s\n", name)
	} else {
		pre := doc.Find("pre")
		if len(pre.Nodes) > 0 {
			for _, l := range strings.Split(pre.First().Text(), "\n") {
				if strings.Contains(l, "store.steampowered.com") {
					for _, t := range strings.Split(l, " ") {
						if strings.Contains(t, "store.steampowered.com"){
							link = t
						}
					}
				}
				if strings.Contains(l, "Genre:") {
					genre = strings.TrimSpace(l[strings.Index(l, "Genre:")+len("Genre:"):])
				}
				if len(genre) > 0 && len(link) > 0 {
					break
				}
			}
		} else {
			ps := doc.Find("p")
			for i := range ps.Nodes {
				p := ps.Eq(i).Text()
				if strings.Contains(p, "Genre:"){
					for _, l := range strings.Split(p, "\n") {
						if strings.Contains(l, "Genre:") {
							genre = strings.TrimSpace(l[strings.Index(l, "Genre:")+len("Genre:"):])
							break
						}
					}
				}
				if len(genre) > 0{
					break
				}
			}
		}
	}

	return gameFromLinkAndName(link, name, genre)
}

func parseFitGirlRepack(item *gofeed.Item) (*games_cache.Game, error){
	name := ""
	genre := ""
	link := ""

	return gameFromLinkAndName(link, name, genre)
}


type dataSource struct {
	url string
	parser func (*gofeed.Item) (*games_cache.Game, error)
}

var sources = []dataSource{
	dataSource{"http://feeds.feedburner.com/SkidrowReloadedGames", parseSkidRowReloaded},
	dataSource{"https://feeds.feedburner.com/skidrowgamesfeed", parseSkidRowReloaded},
	dataSource{"https://feeds.feedburner.com/skidrowgames", parseSkidRowCrack},
	dataSource{"http://feeds.feedburner.com/skidrowcrack", parseSkidRowCrack},
	//dataSource{"http://fitgirl-repacks.com/feed/", parseFitGirlRepack},
}

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
			case ScraperWarning:
				dubious_games = append(dubious_games, (*game))
			default:
				log.Printf("error parsing game: %s\n", err.Error())
			}
		} else if game != nil {
			games = append(games, (*game))
		}
	}

	return &games, &dubious_games
}

func updateCache(pending, dubious, checked *games_cache.Cache, scraped_list, scraped_dubious []games_cache.Game) {
	// TODO clean cache as we update it (skip duplicated games and blacklisted ones)

	pending.AppendElements(scraped_list...)
	dubious.AppendElements(scraped_dubious...)

	pending.Store()
	dubious.Store()
}

func main() {
	var pending_cache = games_cache.LoadCache(games_cache.GamesCachePendingFile)
	var dubious_cache = games_cache.LoadCache(games_cache.GamesCacheDubiousFile)
	var checked_cache = games_cache.LoadCache(games_cache.GamesCacheCheckedFile)

	for _, source := range sources {
		list, dubious := scrapeSource(source)
		updateCache(pending_cache, dubious_cache, checked_cache, *list, *dubious)
	}


	for _, g := range pending_cache.GetContent() {
		log.Println(g)
	}
	log.Println("Dubious games (link is probably wrong: ")
	for _, g := range dubious_cache.GetContent() {
		log.Println(g)
	}
}