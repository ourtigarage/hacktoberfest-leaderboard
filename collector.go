package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/go-github/v32/github"
	"github.com/gregjones/httpcache"
	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/oauth2"
)

var reUsername = regexp.MustCompile(`^\* .*@([a-zA-Z0-9]+).*$`)

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

type StatsCollector struct {
	ParticipantFile string
	EventDate       string
	GitHub          *github.Client
}

func NewStatsCollector(eventDate, playerFile string) *StatsCollector {
	var tksrc oauth2.TokenSource
	if GH_TOKEN != "" {
		tksrc = oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: GH_TOKEN,
		})
	}
	cl := oauth2.NewClient(context.Background(), tksrc)

	cachedTransport := httpcache.NewMemoryCacheTransport()
	cachedTransport.Transport = cl.Transport
	retryCl := retryablehttp.NewClient()
	retryCl.Logger = nil
	retryCl.HTTPClient = cachedTransport.Client()
	gh := github.NewClient(retryCl.StandardClient())
	return &StatsCollector{
		ParticipantFile: playerFile,
		EventDate:       eventDate,
		GitHub:          gh,
	}
}

func (sc *StatsCollector) PlayerNames(ctx context.Context) ([]string, error) {
	req, err := retryablehttp.NewRequest(http.MethodGet, sc.ParticipantFile, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	res, err := retryablehttp.NewClient().Do(req)
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

func (sc *StatsCollector) Players(ctx context.Context) (Players, error) {
	usernames, err := sc.PlayerNames(ctx)
	if err != nil {
		return nil, err
	}
	return sc.queryUserData(ctx, usernames...)
}

func (sc *StatsCollector) queryUserData(ctx context.Context, usernames ...string) (Players, error) {
	players := Players{}
	for _, u := range usernames {
		user, _, err := sc.GitHub.Users.Get(ctx, u)
		if err != nil {
			return nil, err
		}
		players[user.GetLogin()] = NewPlayer(user)
	}
	query := searchQuery(usernames, sc.EventDate)
	fmt.Println("Search query:", query)
	opt := github.SearchOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
			Page:    0,
		},
	}
	repos := map[string]*Repo{}
	for {
		res, r, err := sc.GitHub.Search.Issues(ctx, query, &opt)
		if err != nil {
			return players, err
		}
		opt.ListOptions.Page = r.NextPage
		for _, issue := range res.Issues {
			p := players[issue.User.GetLogin()]
			if p == nil { // Skip if player does not exist
				continue
			}
			repo, ok := repos[issue.GetRepositoryURL()]
			if !ok {
				var err error
				repo, err = NewRepoFromURL(ctx, sc.GitHub, issue.GetRepositoryURL())
				if err != nil {
					return players, err
				}
				repos[issue.GetRepositoryURL()] = repo
			}
			if issue.IsPullRequest() {
				pullreq, _, err := sc.GitHub.PullRequests.Get(ctx, repo.Owner, repo.Name, issue.GetNumber())
				if err != nil {
					return nil, err
				}
				pr, err := NewPullRequest(pullreq, issue, repo)
				if err != nil {
					return players, err
				}
				p.AddContrib(pr)
			} else {
				p.AddIssue(NewIssue(issue, repo))
			}
		}
		if r.NextPage == 0 {
			break
		}
	}
	return players, nil
}
