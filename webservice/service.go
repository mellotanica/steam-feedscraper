package webservice

import (
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"
	"feedscraper/configs"
)

var templates = template.Must(template.ParseFiles("templates/no_files.html", "templates/steamapp_overrides.html"))

const redirPath = "/redir/?"
const steamWishlistApi = "http://store.steampowered.com/api/addtowishlist"

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

func StartService() {
	router := mux.NewRouter().StrictSlash(true)

	configs := configs.GetConfigs()

	if !configs.RestOnly {
		log.Print("Enabling all web services")
		router.HandleFunc("/review", reviewHandler).Methods("GET")
		router.HandleFunc("/wishlist", wishlistHandler).Methods("POST")
		router.HandleFunc("/checked", checkedHandler).Methods("POST")
		router.HandleFunc(fmt.Sprintf("%s{link}", redirPath), redirHandler)
	} else {
		log.Print("Enabling REST services only")
	}
	router.HandleFunc("/getItem", getItemGETHandler).Methods("GET")
	router.HandleFunc("/getItem", getItemPOSTHandler).Methods("POST")
	router.HandleFunc("/checkGet", checkGetHandler).Methods("POST")
	router.HandleFunc("/doubt", doubtHandler).Methods("POST")

	log.Printf("Starting service on port %d\n", configs.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", configs.Port), Rewriter(router)))
}
