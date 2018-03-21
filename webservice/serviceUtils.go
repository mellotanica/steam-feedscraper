package webservice

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

func migrateGame(name, gid string, src, dest *games_cache.Cache, res http.ResponseWriter) {
	err := src.Migrate(dest, gid, name)
	if err != nil {
		http.Error(res, err.Error(), http.StatusNotAcceptable)
		return
	}

	go store_caches(src, dest)
}

func checkGame(name, gid string, res http.ResponseWriter) {
	pending := games_cache.LoadCache(games_cache.GamesCachePendingFile)
	checked := games_cache.LoadCache(games_cache.GamesCacheCheckedFile)

	migrateGame(name, gid, pending, checked, res)
}

func doubtGame(name, gid string, res http.ResponseWriter) {
	pending := games_cache.LoadCache(games_cache.GamesCachePendingFile)
	dubious := games_cache.LoadCache(games_cache.GamesCacheDubiousFile)

	migrateGame(name, gid, pending, dubious, res)
}

func moveHandler(res http.ResponseWriter, req *http.Request, function func(string, string, http.ResponseWriter)) {
	name, gid, err := getGamePostFields(req)
	if err != nil {
		http.Error(res, err.Error(), http.StatusNotAcceptable)
		return
	}

	function(name, gid, res)

	getItemGETHandler(res, req)
}

