package services

import (
	"net/http"
	"feedscraper/games_cache"
)

type wishlistResult struct {
	Success bool `json:"success"`
}

func store_caches(caches ...*games_cache.Cache) {
	for _, cache := range caches {
		cache.Store()
	}
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
