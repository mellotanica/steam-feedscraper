package scrapers

import (
	"strings"
	"feedscraper/games_cache"
	"github.com/PuerkitoBio/goquery"
	"log"
	"github.com/mmcdole/gofeed"
)

func ParseSkidRowCrack(item *gofeed.Item) (*games_cache.Game, error){
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
