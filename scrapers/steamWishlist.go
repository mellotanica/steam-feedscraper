package scrapers

import (
	"feedscraper/games_cache"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"strings"
	"io/ioutil"
	"net/http"
	"encoding/json"
)

const wishlistUrl = "http://store.steampowered.com/wishlist/id/"

type steamwishgame struct {
	Name string `json:"name"`
}

func ScrapeWishlist(checked *games_cache.Cache, steamUsername string) {
	steamUrl := fmt.Sprintf("%s%s", wishlistUrl, steamUsername)

	doc, err := goquery.NewDocument(steamUrl)
	if err != nil {
		log.Printf("ERROR: unable to get user wishlist at url: '%s'", steamUrl)
		return
	}

	doc.Find("script").Each(func(i int, selection *goquery.Selection) {
		if strings.Contains(selection.Text(), "var g_rgAppInfo") {
			firstJson := selection.Text()[strings.Index(selection.Text(), "var g_rgAppInfo"):]
			firstJson = firstJson[:strings.Index(firstJson, "\n")]
			firstJson = firstJson[strings.IndexAny(firstJson, "[{"):strings.LastIndexAny(firstJson, "]}")]

			secondUrl := selection.Text()[strings.Index(selection.Text(), "var g_strWishlistBaseURL"):]
			secondUrl = secondUrl[:strings.Index(secondUrl, "\n")]
			secondUrl = secondUrl[strings.Index(secondUrl,"\"")+1:strings.LastIndex(secondUrl,"\"")]
			secondUrl = fmt.Sprintf("%swishlistdata", strings.Replace(secondUrl, "\\/", "/", -1))

			resp, err := http.Get(secondUrl)
			if err != nil {
				log.Printf("ERROR: loading second json for steam wishlist: %s", err.Error())
				return
			}

			rd, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("ERROR: reading second json for steam wishlist: %s", err.Error())
				return
			}

			secondJson := string(rd)[1:]

			data := make(map[string]steamwishgame)
			err = json.Unmarshal([]byte(fmt.Sprintf("%s,%s", firstJson, secondJson)), &data)
			if err != nil {
				log.Printf("ERROR: unmarshalling json data for steam wishlist: %s", err.Error())
				return
			}

			for gid, g := range data {
				checked.AppendElements(games_cache.Game{g.Name, fmt.Sprintf("app/%s", gid), "", ""})
			}
		}
	})
}
