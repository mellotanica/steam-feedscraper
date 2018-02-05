package main

import (
	"html/template"
	"regexp"
	"net/http"
	"feedscraper/games_cache"
)

var validReview = regexp.MustCompile("^/review/$")
var validChecked = regexp.MustCompile("^/checked/$")

var templates = template.Must(template.ParseFiles("review.html", "no_files.html"))

var cachePrefix = "feeds_cache_"
var pendingFile = cachePrefix + "pending"
var checkedFile = cachePrefix + "checked"

func reviewHandler(w http.ResponseWriter, r *http.Request, tokens []string) {
	pending, err := games_cache.GetList(pendingFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(*pending) <= 0 {
		err = templates.ExecuteTemplate(w, "no_files.html", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	err = templates.ExecuteTemplate(w, "review.html", (*pending)[0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func checkedHandler(w http.ResponseWriter, r *http.Request, token []string) {

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
	http.HandleFunc("/review/", makeHandler(reviewHandler, validReview))
	http.HandleFunc("/checked/", makeHandler(checkedHandler, validChecked))
	http.ListenAndServe(":8080", nil)
}
