package services

import (
	"html/template"
	"github.com/gorilla/mux"
	"log"
	"fmt"
	"net/http"
	"feedscraper/games_cache"
)

var templates = template.Must(template.ParseFiles("templates/review.html", "templates/no_files.html"))

func store_caches(caches... *games_cache.Cache) {
	for _, cache := range caches {
		cache.Store()
	}
}

func myReviewHandler(res http.ResponseWriter, req *http.Request) {
	var err error

	pending := games_cache.LoadCache(games_cache.GamesCachePendingFile)

	if pending.Lenght() <= 0 {
		err = templates.ExecuteTemplate(res, "no_files.html", nil)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	err = templates.ExecuteTemplate(res, "review.html", pending.GetFirst())
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func myCheckedHandler(res http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		http.Error(res, err.Error(), http.StatusNotAcceptable)
		return
	}

	name := req.PostFormValue("name")
	gid := req.PostFormValue("gid")

	pending := games_cache.LoadCache(games_cache.GamesCachePendingFile)
	checked := games_cache.LoadCache(games_cache.GamesCacheCheckedFile)

	err := pending.Migrage(checked, gid, name)
	if err != nil{
		http.Error(res, err.Error(), http.StatusNotAcceptable)
		return
	}

	go store_caches(pending, checked)

	http.Redirect(res, req, "/review", http.StatusFound)
}

func StartService(port int) {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/review", myReviewHandler).Methods("GET")
	router.HandleFunc("/checked", myCheckedHandler).Methods("POST")

	log.Printf("Starting service on port %d\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), router)
}
