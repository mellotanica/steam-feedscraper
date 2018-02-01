package main

import (
	"html/template"
	"regexp"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

var valid_review = regexp.MustCompile("^/review/$")
var valid_checked = regexp.MustCompile("^/checked/$")

var templates = template.Must(template.ParseFiles("review.html", "no_files.html"))

var cache_prefix = "feeds_cache_"
var pending_file = cache_prefix + "pending"
var checked_file = cache_prefix + "checked"

type Game struct {
	Name string `json:"name"`
	Gid string `json:"gid"`
	Link string `json:"link"`
	Genre string `json:"genre"`
}

func get_list(fname string) (*[]Game, error) {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, err
	}

	var list []Game
	err = json.Unmarshal(data, &list)
	if err != nil {
		return nil, err
	}

	return &list, nil
}

func store_list(fname string, list *[]Game) error {
	data, err := json.Marshal(list)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(fname, data, 0600)
}

func reviewHandler(w http.ResponseWriter, r *http.Request, tokens []string) {
	pending, err := get_list(pending_file)
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
	http.HandleFunc("/review/", makeHandler(reviewHandler, valid_review))
	http.HandleFunc("/checked/", makeHandler(checkedHandler, valid_checked))
	http.ListenAndServe(":8080", nil)
}
