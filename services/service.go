package services

import (
	"feedscraper/games_cache"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"
	"encoding/json"
	"compress/gzip"
	"io"
)

var templates = template.Must(template.ParseFiles("templates/no_files.html", "templates/steamapp_overrides.html"))

const redirPath = "/redir/?"
const steamWishlistApi = "http://store.steampowered.com/api/addtowishlist"

func store_caches(caches ...*games_cache.Cache) {
	for _, cache := range caches {
		cache.Store()
	}
}

type steamgame struct {
	Name    string
	Gid     string
	Link    string
	Genre   string
	Content string
}

type wishlistResult struct {
	Success bool `json:"success"`
}


func redirHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	link, err := url.PathUnescape(vars["link"])
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	ch := make(chan redirChan)
	go renderRedir(link, req, ch)
	response := <- ch
	if response.err != nil {
		http.Error(res, response.err.Error(), http.StatusInternalServerError)
		return
	}

	for k, v := range response.response.Header {
		if k != "Content-Length" && k != "Content-Encoding" {
			res.Header().Set(k, v[0])
			if len(v) > 1 {
				for _, iv := range v[1:] {
					res.Header().Add(k, iv)
				}
			}
		}
	}

	content := response.html
	if strings.HasPrefix(strings.TrimSpace(content), "<?xml") {
		content = content[strings.Index(content, "?>") + 2:]
	}
	res.Write([]byte(content))
}

func reviewHandler(res http.ResponseWriter, req *http.Request) {
	var err error

	pending := games_cache.LoadCache(games_cache.GamesCachePendingFile)

	if pending.Lenght() <= 0 {
		err = templates.ExecuteTemplate(res, "no_files.html", nil)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	g := pending.GetFirst()

	http.Redirect(res, req, redirPath+g.Link, http.StatusFound)
}

func getGamePostFields(req *http.Request) (name, gid string, err error) {
	err = nil
	if err = req.ParseForm(); err != nil {
		return
	}

	name = req.PostFormValue("name")
	gid = req.PostFormValue("gid")
	return
}

func checkGame(name, gid string, res http.ResponseWriter, req *http.Request) {
	pending := games_cache.LoadCache(games_cache.GamesCachePendingFile)
	checked := games_cache.LoadCache(games_cache.GamesCacheCheckedFile)

	err := pending.Migrate(checked, gid, name)
	if err != nil {
		http.Error(res, err.Error(), http.StatusNotAcceptable)
		return
	}

	go store_caches(pending, checked)

	http.Redirect(res, req, "/review", http.StatusFound)
}

func checkedHandler(res http.ResponseWriter, req *http.Request) {
	name, gid, err := getGamePostFields(req)
	if err != nil {
		http.Error(res, err.Error(), http.StatusNotAcceptable)
		return
	}

	checkGame(name, gid, res, req)
}

func wishlistHandler(res http.ResponseWriter, req *http.Request) {
	name, gid, err := getGamePostFields(req)
	if err != nil {
		http.Error(res, err.Error(), http.StatusNotAcceptable)
		return
	}

	data := url.Values{}

	cookie, err := req.Cookie("sessionid")
	if err != nil {
		http.Error(res, err.Error(), http.StatusNotAcceptable)
		return
	}
	data.Add("sessionid", cookie.Value)
	data.Add("appid", strings.TrimLeft(gid[strings.Index(gid,"/"):], "/"))

	pending := games_cache.LoadCache(games_cache.GamesCachePendingFile)
	game, ok := pending.GetElementById(gid)
	if !ok {
		http.Error(res, fmt.Sprintf("unable to find game %s (%s) in cache", name, gid), http.StatusNotAcceptable)
		return
	}

	postreq, err := http.NewRequest("POST", steamWishlistApi, strings.NewReader(data.Encode()))
	postreq.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	postreq.Header.Add("X-Requested-With", "XMLHttpRequest")
	postreq.Header.Add("Origin", "http://store.steampowered.com")
	postreq.Header.Add("Accept", "*/*")
	postreq.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.167 Safari/537.36")
	postreq.Header.Add("Referer", game.Link)
	postreq.Header.Add("Accept-Encoding", "gzip")
	postreq.Header.Add("Accept-Language", "en-US,en;q=0.9,it;q=0.8")

	for _, c := range req.Cookies() {
		postreq.AddCookie(c)
	}

	postres, err := http.DefaultClient.Do(postreq)
	if err != nil {
		http.Error(res, err.Error(), http.StatusNotAcceptable)
		return
	}
	defer postres.Body.Close()

	var responseReader io.Reader

	if postres.Header.Get("Content-Encoding") == "gzip" {
		responseReader, err = gzip.NewReader(postres.Body)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		responseReader = postres.Body
	}

	result := wishlistResult{}
	err = json.NewDecoder(responseReader).Decode(&result)
	if err != nil {
		log.Printf("ERROR reading wishlist post response for %s (%s):\n %s\n", name, gid, err.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	if !result.Success {
		http.Error(res, "Steam rejected wishlist request", http.StatusInternalServerError)
		return
	}

	checkGame(name, gid, res, req)
}

func Rewriter(h http.Handler) http.Handler {
	redirlen := len(redirPath)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pathReq := r.RequestURI
		if strings.HasPrefix(pathReq, redirPath) {
			pe := url.PathEscape(pathReq[redirlen:])
			r.URL.Path = pathReq[:redirlen] + pe
			r.URL.RawQuery = ""
		}

		h.ServeHTTP(w, r)
	})
}

func StartService(port int) {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/review", reviewHandler).Methods("GET")
	router.HandleFunc("/checked", checkedHandler).Methods("POST")
	router.HandleFunc("/wishlist", wishlistHandler).Methods("POST")
	router.HandleFunc(fmt.Sprintf("%s{link}", redirPath), redirHandler)

	log.Printf("Starting service on port %d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), Rewriter(router)))
}
