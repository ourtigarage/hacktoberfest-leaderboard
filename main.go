package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/joho/godotenv"
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
	PARTICIPANTS_FILE = "https://raw.githubusercontent.com/ourtigarage/hacktoberfest-summit/master/participants/2022.md"
	// EVENT_DATE        = ">=2005"
	EVENT_DATE = "2022-10"
	GH_TOKEN   = ""
)

func LookupEnvDefault(key string, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

func loadConfig() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file to load")
	} else {
		fmt.Println("Loaded .env file")
	}
	PORT = LookupEnvDefault("PORT", PORT)
	PARTICIPANTS_FILE = LookupEnvDefault("PARTICIPANTS_FILE", PARTICIPANTS_FILE)
	EVENT_DATE = LookupEnvDefault("EVENT_DATE", EVENT_DATE)
	GH_TOKEN = LookupEnvDefault("GH_TOKEN", GH_TOKEN)
}

func main() {
	loadConfig()

	lb := NewLeaderboard(EVENT_DATE, PARTICIPANTS_FILE)
	hdl := handlers.CombinedLoggingHandler(os.Stdout, routes(lb))
	svcc := make(chan error, 1)
	server := http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%s", PORT),
		Handler: hdl,
	}
	go func() {
		defer close(svcc)
		fmt.Println("Listenning on", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			svcc <- err
			return
		}
		fmt.Println("Server closed")
	}()

	sigc := make(chan os.Signal, 128)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	select {
	case err := <-svcc:
		fmt.Fprintln(os.Stderr, "Server exitted:", err)
		lb.Shutdown()
	case <-lb.C:
		fmt.Fprintln(os.Stderr, "Stats fetcher exitted")
		_ = server.Shutdown(context.Background())
	case sig := <-sigc:
		fmt.Printf("Signal %s received. Shutting down server\n", sig.String())
		lb.Shutdown()
		_ = server.Shutdown(context.Background())
		fmt.Println("Graceful shutdown completed")
	}
}
