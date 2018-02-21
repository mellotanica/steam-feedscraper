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
	"io/ioutil"
	"time"
)

var templates = template.Must(template.ParseFiles("templates/review.html", "templates/no_files.html"))

const redirPath = "/redir/"

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

func renderRedir(target string, req *http.Request, w http.ResponseWriter) (string, error) {
	newReq, err := http.NewRequest(req.Method, target, req.Body)
	if err != nil {
		return "", err
	}

	for _, c := range req.Cookies() {
		newReq.AddCookie(c)
	}
	// skip age and mature content checks
	newReq.AddCookie(createCookie("mature_content", "1", -1))
	newReq.AddCookie(createCookie("lastagecheckage", "17-August-1982", -1))
	newReq.AddCookie(createCookie("birthtime", "398383201", -1))

	resp, err := http.DefaultClient.Do(newReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	html := string(body)
	html = strings.Replace(html, "https://store.steampowered.com/", fmt.Sprintf("%shttps://store.steampowered.com/", redirPath), -1)
	html = strings.Replace(html, "http://store.steampowered.com/", fmt.Sprintf("%shttp://store.steampowered.com/", redirPath), -1)

	for _, c := range resp.Cookies() {
		http.SetCookie(w, c)
	}

	return html, nil
}

func redirHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	link, err := url.PathUnescape(vars["link"])
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	content, err := renderRedir(link, req, res)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
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
	content, err := renderRedir(g.Link, req, res)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	sg := steamgame{g.Name, g.Gid, g.Link, g.Genre, content}

	err = templates.ExecuteTemplate(res, "review.html", sg)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func checkedHandler(res http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		http.Error(res, err.Error(), http.StatusNotAcceptable)
		return
	}

	name := req.PostFormValue("name")
	gid := req.PostFormValue("gid")

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
	router.HandleFunc(fmt.Sprintf("%s{link}", redirPath), redirHandler)

	log.Printf("Starting service on port %d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), Rewriter(router)))
}
