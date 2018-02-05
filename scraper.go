package main

import (
	"feedscraper/games_cache"
	"github.com/PuerkitoBio/goquery"
	"log"
	"fmt"
	"strings"
	"net/url"
	"github.com/mmcdole/gofeed"
)

const mediumCacheSize = 30

type ScraperError struct {
	What string
}

func (e ScraperError) Error() string {
	return fmt.Sprintf("ScraperError: %s", e.What)
}

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
		return nil, ScraperError{fmt.Sprintf("search on steam: %s (%s)", name, genre)}
		//TODO query steam for gid and link
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

func parseSkidRowReloaded(item *goquery.Selection) (*games_cache.Game, error){
	name := ""
	genre := ""
	link := ""

	tcont := strings.TrimSpace(item.Text())
	lcont := strings.ToLower(tcont)

	titi := strings.Index(lcont, "title:")
	if titi >= 0 {
		name = tcont[titi+len("title:"):]
		name = strings.TrimSpace( name[:strings.Index(name, "\n")])
	}

	geni := strings.Index(lcont, "genre:")
	if geni >= 0 {
		genre = tcont[geni+len("genre:"):]
		genre = strings.TrimSpace(genre[:strings.Index(genre, "\n")])
	}

	lnki := strings.Index(lcont, "store.steampowered.com")
	if lnki >= 0 {
		spci := strings.LastIndex(lcont[:lnki], " ")
		stri := strings.LastIndex(lcont[:lnki], "\"")
		lni := strings.LastIndex(lcont[:lnki], "\n")
		beg := spci
		if beg < stri {
			beg = stri
		}
		if beg < lni {
			beg = lni
		}
		link = tcont[beg + 1:]

		spci = strings.Index(link, " ")
		stri = strings.Index(link, "\"")
		lni = strings.Index(link, "\n")
		end := lni
		if spci >= 0 && end > spci {
			end = spci
		}
		if stri >= 0 && end > stri {
			end = stri
		}
		if end >= 0 {
			link = link[:end]
		} else {
			link = ""
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
	//dataSource{"http://feeds.feedburner.com/SkidrowReloadedGames", parseSkidRowReloaded},
	//dataSource{"https://feeds.feedburner.com/skidrowgamesfeed", parseSkidRowReloaded},
	dataSource{"https://feeds.feedburner.com/skidrowgames", parseSkidRowCrack},
	dataSource{"http://feeds.feedburner.com/skidrowcrack", parseSkidRowCrack},
	//dataSource{"http://fitgirl-repacks.com/feed/", parseFitGirlRepack},
}

//func scrapeSource(source dataSource) (*[]games_cache.Game) {
//	doc, err := goquery.NewDocument(source.url)
//	if err != nil {
//		log.Printf("feed source %s is unreachable: %s\n", source.url, err.Error())
//		return nil
//	}
//
//	var games = make([]games_cache.Game, 0, mediumCacheSize)
//
//	doc.Find("item").Each(func(i int, selection *goquery.Selection) {
//		game, err := source.parser(selection)
//		if err != nil {
//			log.Printf("error parsing game: %s\n", err.Error())
//		} else if game != nil {
//			games = append(games, (*game))
//		}
//	})
//	return &games
//}

func scrapeSource(source dataSource) (*[]games_cache.Game) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(source.url)
	if err != nil {
		log.Printf("unable to parse feed %s: %s\n", source.url, err.Error())
		return nil
	}

	var games = make([]games_cache.Game, 0, mediumCacheSize)

	for _, i := range feed.Items {
		game, err := source.parser(i)
		if err != nil {
			log.Printf("error parsing game: %s\n", err.Error())
		} else if game != nil {
			games = append(games, (*game))
		}
	}

	return &games
}

func main() {
	var cache = make([]games_cache.Game, 0, mediumCacheSize*len(sources))
	for _, source := range sources {
		games := scrapeSource(source)
		if games != nil{
			for j := 0; j < len(*games); j++{
				cache = append(cache, (*games)[j])
			}
		}
	}

	for i := 0; i < len(cache); i++ {
		log.Println(cache[i])
	}
}