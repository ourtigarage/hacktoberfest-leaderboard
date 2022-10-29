package main

import (
	"fmt"
	"io"
	"leaderboard/static"
	"leaderboard/views"
	"net/http"

	"github.com/gorilla/mux"
)

func routes(lb *Leaderboard) http.Handler {
	router := mux.NewRouter()

	// HealthCheck
	router.Methods("GET").Path("/health/alive").HandlerFunc(healthAlive)
	router.Methods("GET").Path("/health/ready").HandlerFunc(healthReady(lb))

	// Badges
	router.Methods("GET").Path("/badges").HandlerFunc(badges())

	sr := router.PathPrefix("/").Subrouter()
	sr.Use(serverReady(lb))
	sr.Methods("GET").Path("/").HandlerFunc(index(lb))
	sr.Methods("GET").Path("/player/{username}").HandlerFunc(player(lb))

	// All other static files
	router.Methods("GET").PathPrefix("/").Handler(
		http.FileServer(http.FS(static.Files)),
	)
	return router
}

func healthAlive(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "OK")
}

func healthReady(lb *Leaderboard) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if !lb.Ready() {
			w.WriteHeader(http.StatusServiceUnavailable)
			io.WriteString(w, "NOK")
			return
		}
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "OK")
	}
}

func serverReady(lb *Leaderboard) mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !lb.Ready() {
				views.View(w, "not_ready", views.Data{Refresh: 15})
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}

func index(lb *Leaderboard) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		views.View(w, "index", views.Data{Refresh: int(COLLECT_PERIOD.Seconds()), Data: lb})
	}
}

func badges() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		views.View(w, "badges", views.Data{Refresh: 0, Data: BADGES})
	}
}

func player(lb *Leaderboard) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := mux.Vars(r)["username"]
		player, err := lb.Player(username)
		if err != nil {
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		views.View(w, "player", views.Data{Refresh: int(COLLECT_PERIOD.Seconds()), Data: player})
	}
}
