package scrapers

import (
	"github.com/mmcdole/gofeed"
	"github.com/PuerkitoBio/goquery"
	"strings"
	"feedscraper/games_cache"
)

func ParseSkidRowReloaded(item *gofeed.Item) (*games_cache.Game, error){
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
