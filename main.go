package main

import (
	"fmt"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/gorilla/handlers"
)

const (
	TARGET_OBJECTIVE = 4
	COLLECT_PERIOD   = 15 * time.Minute
	BASE_API_URL     = "https://api.github.com"
	BASE_REPOS_URL   = BASE_API_URL + "/repos"
	ORG_REPOS_URL    = BASE_REPOS_URL + "/ourtigarage"
	SNAKE_URL        = ORG_REPOS_URL + "/web-snake"
	LEADERBOARD_URL  = ORG_REPOS_URL + "/hacktoberfest-leaderboard"
)

var (
	PORT              = "8080"
	PARTICIPANTS_FILE = "https://raw.githubusercontent.com/ourtigarage/hacktoberfest-summit/master/participants/2020.md"
	// EVENT_DATE        = ">=2005"
	EVENT_DATE = "2020-10"
	GH_TOKEN   = ""
)

var (
	tmplBadges = template.Must(template.ParseFiles("./views/layouts/main.tmpl", "./views/badges.tmpl"))
	tmplPlayer = template.Must(template.ParseFiles("./views/layouts/main.tmpl", "./views/player.tmpl"))
	tmplIndex  = template.Must(template.ParseFiles("./views/layouts/main.tmpl", "./views/index.tmpl"))
)

func LookupEnvDefault(key string, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

func loadConfig() {
	PORT = LookupEnvDefault("PORT", PORT)
	PARTICIPANTS_FILE = LookupEnvDefault("PARTICIPANTS_FILE", PARTICIPANTS_FILE)
	EVENT_DATE = LookupEnvDefault("EVENT_DATE", EVENT_DATE)
	GH_TOKEN = LookupEnvDefault("GH_TOKEN", GH_TOKEN)
}

func main() {
	loadConfig()
	bindAddr := fmt.Sprintf("0.0.0.0:%s", PORT)

	lb := NewBackgroundLeaderboard(EVENT_DATE, PARTICIPANTS_FILE)
	hdl := handlers.CombinedLoggingHandler(os.Stdout, routes(lb))
	fmt.Println("Listenning on", bindAddr)
	if err := http.ListenAndServe(bindAddr, hdl); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
	fmt.Println("Shutting down")
}
