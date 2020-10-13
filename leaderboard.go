package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/gregjones/httpcache"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/jinzhu/copier"
	"golang.org/x/oauth2"
)

type Leaderboard struct {
	ParticipantFile string
	EventDate       string
	GitHub          *github.Client
}

func NewLeaderboard(eventDate, playerFile string) *Leaderboard {
	var tksrc oauth2.TokenSource
	if GH_TOKEN != "" {
		tksrc = oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: GH_TOKEN,
		})
	}
	cl := oauth2.NewClient(context.Background(), tksrc)

	cachedTransport := httpcache.NewMemoryCacheTransport()
	// cachedTransport.Transport = &retryablehttp.RoundTripper{Client: retryCl}
	cachedTransport.Transport = cl.Transport
	retryCl := retryablehttp.NewClient()
	retryCl.Logger = nil
	retryCl.HTTPClient = cachedTransport.Client()
	gh := github.NewClient(retryCl.StandardClient())
	return &Leaderboard{
		ParticipantFile: playerFile,
		EventDate:       eventDate,
		GitHub:          gh,
	}
}

var reUsername = regexp.MustCompile(`^\* .*@([a-zA-Z0-9]+).*$`)

func (lb *Leaderboard) Players() (Players, error) {
	usernames, err := lb.PlayerNames()
	if err != nil {
		return nil, err
	}
	return lb.queryUserData(usernames...)
}

func (lb *Leaderboard) PlayersSorted() ([]*Player, error) {
	playerMap, err := lb.Players()
	if err != nil {
		return nil, err
	}
	players := make([]*Player, 0, len(playerMap))
	for _, p := range playerMap {
		players = append(players, p)
	}
	sort.SliceStable(players, func(i, j int) bool {
		return players[i].ContributionCount() > players[j].ContributionCount()
	})
	return players, nil
}

func (lb *Leaderboard) Player(username string) (*Player, error) {
	players, err := lb.queryUserData(username)
	if err != nil {
		return nil, err
	}
	if len(players) == 0 {
		return nil, errors.New("Player not found")
	}
	return players[username], nil
}

func (lb *Leaderboard) PlayerNames() ([]string, error) {
	res, err := http.Get(lb.ParticipantFile)
	if err != nil {
		return nil, err
	}
	names := []string{}
	r := bufio.NewScanner(res.Body)
	for r.Scan() {
		line := strings.TrimSpace(r.Text())
		m := reUsername.FindStringSubmatch(line)
		if len(m) == 0 || m[1] == "username" {
			continue
		}
		names = append(names, m[1])
	}
	return names, nil
}

func searchQuery(usernames []string, eventDate string) string {
	b := strings.Builder{}
	for _, u := range usernames {
		b.WriteString("author:")
		b.WriteString(u)
		b.WriteString(" ")
	}
	b.WriteString("created:")
	b.WriteString(eventDate)
	return b.String()
}

func (lb *Leaderboard) queryUserData(usernames ...string) (Players, error) {
	players := Players{}
	for _, u := range usernames {
		user, _, err := lb.GitHub.Users.Get(context.TODO(), u)
		if err != nil {
			return nil, err
		}
		players[user.GetLogin()] = NewPlayer(user, lb.GitHub)
	}
	query := searchQuery(usernames, lb.EventDate)
	opt := github.SearchOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
			Page:    0,
		},
	}
	for {
		res, r, err := lb.GitHub.Search.Issues(context.TODO(), query, &opt)
		if err != nil {
			return players, err
		}
		opt.ListOptions.Page = r.NextPage
		for _, issue := range res.Issues {
			p := players[issue.User.GetLogin()]
			if err := p.AddContrib(issue); err != nil {
				return players, err
			}
		}
		if r.NextPage == 0 {
			break
		}
	}
	return players, nil
}

type BackgroundLeaderboard struct {
	lock    *sync.RWMutex
	inner   *Leaderboard
	players Players
	sorted  []*Player
}

func NewBackgroundLeaderboard(eventDate, playerFile string) *BackgroundLeaderboard {
	lb := &BackgroundLeaderboard{
		lock:    new(sync.RWMutex),
		inner:   NewLeaderboard(eventDate, playerFile),
		players: Players{},
		sorted:  []*Player{},
	}
	lb.start()
	return lb
}

// func (lb *BackgroundLeaderboard) Players() Players {
// 	lb.lock.RLock()
// 	defer lb.lock.RUnlock()
// 	return lb.players
// }

func (lb *BackgroundLeaderboard) PlayersSorted() []*Player {
	lb.lock.RLock()
	defer lb.lock.RUnlock()
	pls := []*Player{}
	copier.Copy(&pls, lb.sorted)
	return lb.sorted
}

func (lb *BackgroundLeaderboard) Player(username string) (*Player, error) {
	lb.lock.RLock()
	defer lb.lock.RUnlock()
	p, ok := lb.players[username]
	if !ok {
		return nil, errors.New("User not found")
	}
	res := new(Player)
	return res, copier.Copy(res, p)
}

func (lb *BackgroundLeaderboard) PlayerNames() []string {
	lb.lock.RLock()
	defer lb.lock.RUnlock()
	names := make([]string, 0, len(lb.players))
	for k := range lb.players {
		names = append(names, k)
	}
	return names
}

func (lb *BackgroundLeaderboard) start() {
	fmt.Println("Starting background updater")
	go func() {
		for {
			lb.update()
			time.Sleep(1 * time.Minute)
		}
	}()
}

func (lb *BackgroundLeaderboard) update() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Background routine panicked:", err)
		}
	}()
	fmt.Println("Collecting data")
	players, err := lb.inner.Players()
	if err != nil {
		fmt.Println("[ERROR]", err)
	} else {
		lb.lock.Lock()
		lb.players = players
		lb.sort()
		lb.lock.Unlock()
		fmt.Println("Collection completed")
	}
}

func (lb *BackgroundLeaderboard) sort() {
	players := make([]*Player, 0, len(lb.players))
	for _, p := range lb.players {
		players = append(players, p)
	}
	sort.SliceStable(players, func(i, j int) bool {
		return players[i].ContributionCount() > players[j].ContributionCount()
	})
	lb.sorted = players
}
