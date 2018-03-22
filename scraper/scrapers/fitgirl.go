package scrapers

import (
	"github.com/mmcdole/gofeed"
	"feedscraper/games_cache"
	"regexp"
	"github.com/PuerkitoBio/goquery"
	"strings"
)

var genreLabel = "Genres/Tags: "

var possibleMatchs = []*regexp.Regexp {
	regexp.MustCompile(" (-|–) v[0-9][.]?[0-9].*"),
	regexp.MustCompile(" [+] [^+]*DLC.*"),
	regexp.MustCompile(" [+] [^+]*Multiplayer.*"),
}

func cleanupName(name string) (clean string) {
	clean = name
	for _, re := range possibleMatchs {
		clean = re.ReplaceAllString(clean, "")
	}
	return
}

func ParseFitGirlRepack(item *gofeed.Item) (*games_cache.Game, error){
	genre := ""

	name := cleanupName(item.Title)

	if len(name) <= 0 {
		return nil, ScraperError{"unable to parse item: no name found"}
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(item.Content))
	if err != nil {
		return nil, ScraperError{"unbale to parse item: no content found"}
	}

	p := doc.Find("p")
	if len(p.Nodes) > 0 {
		for _, l := range strings.Split(p.First().Text(), "\n") {
			if strings.Contains(l, genreLabel) {
				genre = strings.TrimSpace(l[strings.Index(l, genreLabel) + len(genreLabel):])
			}
		}
	}

	if len(genre) <= 0 {
		return nil, ScraperError{"unable to parse item: does not seem a valid game"}
	}

	return gameFromLinkAndName("", name, genre)
}

