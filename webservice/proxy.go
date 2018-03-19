package webservice

import (
	"strings"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"fmt"
	"net/http"
	"time"
	"bytes"
	"feedscraper/games_cache"
)


type redirChan struct {
	html string
	response *http.Response
	err error
}

func createCookie(name, val string, maxage int) (*http.Cookie) {
	return & http.Cookie{
		name,
		val,
		"",
		"",
		time.Now(),
		"",
		maxage,
		false,
		false,
		"",
		nil,
	}
}

func renderRedir(target string, req *http.Request, response chan redirChan) {
	newReq, err := http.NewRequest(req.Method, target, req.Body)
	if err != nil {
		response <- redirChan{"", nil, err}
		return
	}

	for k, v := range req.Header {
		if k != "Accept-Encoding" {
			newReq.Header.Set(k, v[0])
			if len(v) > 1 {
				for _, iv := range v[1:] {
					newReq.Header.Add(k, iv)
				}
			}
		}
	}

	// skip age and mature content checks
	newReq.AddCookie(createCookie("mature_content", "1", -1))
	newReq.AddCookie(createCookie("lastagecheckage", "17-August-1982", -1))
	newReq.AddCookie(createCookie("birthtime", "398383201", -1))

	resp, err := http.DefaultClient.Do(newReq)
	if err != nil {
		response <- redirChan{"", nil, err}
		return
	}
	defer resp.Body.Close()

	showbuttons := false
	var body string

	if strings.Contains(target, "://store.steampowered.com/app/") {
		gid := target[strings.Index(target,"/app/")+len("/app/"):]
		end := strings.Index(gid, "/")
		if end < 0 {
			end = len(gid)
		}
		gid = "app/" + gid[:end]

		pending := games_cache.LoadCache(games_cache.GamesCachePendingFile)

		var game *games_cache.Game
		game, showbuttons = pending.GetElementById(gid)
		if showbuttons {
			doc, err := goquery.NewDocumentFromResponse(resp)
			if err != nil {
				response <- redirChan{"", nil, err}
				return
			}

			buff := bytes.NewBuffer(make([]byte, 512))
			err = templates.ExecuteTemplate(buff, "steamapp_overrides.html", game)
			if err != nil {
				response <- redirChan{"", nil, err}
				return
			}

			overrideDoc, err := goquery.NewDocumentFromReader(buff)
			if err != nil {
				response <- redirChan{"", nil, err}
				return
			}

			// if requested page is a game page inject next and wishlist buttons along with javascript functions

			head := doc.Find("head")
			overrideDoc.Find("head").Children().Each(func(i int, selection *goquery.Selection) {
				head.AppendSelection(selection.Clone())
			})

			sbody := doc.Find("body")
			obody := overrideDoc.Find("body")

			obody.Children().Each(func(i int, selection *goquery.Selection) {
				sbody.AppendSelection(selection.Clone())
			})

			for _, attr := range obody.Get(0).Attr {
				sattr, exists := sbody.Attr(attr.Key)
				if exists {
					sbody.SetAttr(attr.Key, fmt.Sprintf("%s; %s", attr.Val, sattr))
				} else {
					sbody.SetAttr(attr.Key, attr.Val)
				}
			}

			body, err = doc.Html()
			if err != nil {
				response <- redirChan{"", nil, err}
				return
			}
		}
	}

	if !showbuttons {
		bbody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			response <- redirChan{"", nil, err}
			return
		}

		body = string(bbody)
	}

	html := body
	html = strings.Replace(html, "https://store.steampowered.com/", fmt.Sprintf("%shttps://store.steampowered.com/", redirPath), -1)
	html = strings.Replace(html, "http://store.steampowered.com/", fmt.Sprintf("%shttp://store.steampowered.com/", redirPath), -1)


	response <- redirChan{html, resp, nil}
}

