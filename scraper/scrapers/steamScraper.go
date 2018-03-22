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

const wishlistFmt = "http://store.steampowered.com/wishlist/id/%s"
var libraryFmt = "https://steamcommunity.com/id/%s/games/?tab=all"

type steamwishgame struct {
	Name string `json:"name"`
}

type steamlibgame struct {
	Name string `json:"name"`
	Gid int `json:"appid"`
}

func getJsonFromPage(urlFmt, username, varName string) (string, *goquery.Selection, error) {
	steamUrl := fmt.Sprintf(urlFmt, username)

	doc, err := goquery.NewDocument(steamUrl)
	if err != nil {
		return "", nil, err
	}

	json := ""
	var script *goquery.Selection = nil

	doc.Find("script").Each(func(i int, selection *goquery.Selection) {
		if strings.Contains(selection.Text(), varName) {
			script = selection
			json = selection.Text()[strings.Index(selection.Text(), varName):]
			json = json[:strings.Index(json, "\n")]
			json = strings.TrimSpace(json[strings.Index(json,"=")+1:strings.LastIndex(json, ";")])
		}
	})

	if len(json) <= 0 || script == nil {
		return "", nil, ScraperError{fmt.Sprintf("unable to find json in variable %s from page %s", varName, steamUrl)}
	}

	return json, script, nil
}

func scrapeWishlist(checked *games_cache.Cache, steamUsername string) {
	firstJson, script,err := getJsonFromPage(wishlistFmt, steamUsername, "var g_rgAppInfo")
	if err != nil {
		log.Printf("ERROR: getting game data for steam wishlist: %s", err.Error())
		return
	}
	firstJson = firstJson[strings.IndexAny(firstJson, "[{"):strings.LastIndexAny(firstJson, "]}")]

	secondUrl := script.Text()[strings.Index(script.Text(), "var g_strWishlistBaseURL"):]
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

func scrapeLibrary(checked *games_cache.Cache, steamUsername string) {
	jsonData, _, err := getJsonFromPage(libraryFmt, steamUsername, "var rgGames")
	if err != nil {
		log.Printf("ERROR: getting game data for steam library: %s", err.Error())
		return
	}

	data := make([]steamlibgame, 64)
	err = json.Unmarshal([]byte(jsonData), &data)
	if err != nil {
		log.Printf("ERROR: unmarshalling json data for steam library: %s", err.Error())
	}

	for _, g := range data {
		checked.AppendElements(games_cache.Game{g.Name, fmt.Sprintf("app/%d",g.Gid), "", ""})
	}
}

func ScrapeSteam(checked *games_cache.Cache, steamUsername string) {
	scrapeWishlist(checked, steamUsername)
	scrapeLibrary(checked, steamUsername)
}
