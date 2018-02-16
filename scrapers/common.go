package scrapers

import (
	"github.com/texttheater/golang-levenshtein/levenshtein"
	"fmt"
	"feedscraper/games_cache"
	"regexp"
	"net/url"
	"strings"
	"github.com/PuerkitoBio/goquery"
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
