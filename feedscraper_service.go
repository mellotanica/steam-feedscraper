package main

import (
	"html/template"
	"regexp"
	"net/http"
	"feedscraper/games_cache"
	"log"
	"fmt"
)

var validReview = regexp.MustCompile("^/review/$")
var validChecked = regexp.MustCompile("^/checked/.*$")

var templates = template.Must(template.ParseFiles("review.html", "no_files.html"))

func reviewHandler(w http.ResponseWriter, r *http.Request, tokens []string) {
	var err error

	pending := games_cache.LoadCache(games_cache.GamesCachePendingFile)

	if pending.Lenght() <= 0 {
		err = templates.ExecuteTemplate(w, "no_files.html", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	err = templates.ExecuteTemplate(w, "review.html", pending.GetFirst())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func store_caches(caches... *games_cache.Cache) {
	for _, cache := range caches {
		cache.Store()
	}
}

func checkedHandler(w http.ResponseWriter, r *http.Request, token []string) {
	gid := r.URL.Path[len("/checked/"):]

	pending := games_cache.LoadCache(games_cache.GamesCachePendingFile)
	checked := games_cache.LoadCache(games_cache.GamesCacheCheckedFile)

	pending.Migrage(checked, gid)

	go store_caches(pending, checked)

	http.Redirect(w, r, "/review/", http.StatusFound)
}

func makeHandler(fn func (http.ResponseWriter, *http.Request, []string), validator *regexp.Regexp) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		m := validator.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m)
	}
}

func main() {
	port := 8080

	http.HandleFunc("/review/", makeHandler(reviewHandler, validReview))
	http.HandleFunc("/checked/", makeHandler(checkedHandler, validChecked))

	log.Printf("Starting service on port %d\n", port)

	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
