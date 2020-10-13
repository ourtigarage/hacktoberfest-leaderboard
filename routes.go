package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func routes(lb *BackgroundLeaderboard) http.Handler {
	router := mux.NewRouter()
	router.Methods("GET").Path("/").HandlerFunc(index(lb))
	router.Methods("GET").Path("/badges").HandlerFunc(badges(lb))
	router.Methods("GET").Path("/player/{username}").HandlerFunc(player(lb))
	router.Methods("GET").PathPrefix("/").Handler(
		http.FileServer(http.Dir("./static")),
	)
	return router
}

func index(lb *BackgroundLeaderboard) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := tmplIndex.ExecuteTemplate(w, "main", lb); err != nil {
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func badges(lb *BackgroundLeaderboard) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := tmplBadges.ExecuteTemplate(w, "main", BADGES); err != nil {
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func player(lb *BackgroundLeaderboard) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := mux.Vars(r)["username"]
		player, err := lb.Player(username)
		if err != nil {
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := tmplPlayer.ExecuteTemplate(w, "main", player); err != nil {
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
