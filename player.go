package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v32/github"
)

type Player struct {
	gh                                                *github.Client
	Username                                          string
	Avatar                                            string
	Profile                                           string
	Contributions, Pending, Ignored, Invalids, Issues []*github.Issue
	Repos                                             map[string]*Repo
}

type Players map[string]*Player

func NewPlayer(user *github.User, gh *github.Client) *Player {
	return &Player{
		gh:       gh,
		Username: user.GetLogin(),
		Avatar:   user.GetAvatarURL(),
		Profile:  user.GetHTMLURL(),
		Repos:    make(map[string]*Repo),
	}
}

func (p *Player) IsChallengeComplete() bool {
	return p.ContributionCount() >= TARGET_OBJECTIVE
}

func (p *Player) ContributionCount() int {
	return len(p.Contributions)
}

func (_ *Player) Objective() int {
	return TARGET_OBJECTIVE
}

func (p *Player) Badges() []Badge {
	badges := []Badge{}
	for _, b := range BADGES {
		if b.EarnedBy(p) {
			badges = append(badges, b)
		}
	}
	return badges
}

func (p *Player) ChallengeCompletion() int {
	percent := 100 * p.ContributionCount() / TARGET_OBJECTIVE
	if percent > 100 {
		percent = 100
	}
	return percent
}

var date2020, _ = time.Parse("2006-01-02", "2020-10-03")

func (p *Player) AddContrib(contrib *github.Issue) error {
	repo := repoFromUrl(p.gh, contrib.GetRepositoryURL())
	// reponame := fmt.Sprintf("%s/%s", repo.Owner.GetLogin(), repo.GetFullName())
	p.Repos[repo.FullName()] = repo
	if !contrib.IsPullRequest() {
		p.Issues = append(p.Issues, contrib)
	} else if labelsContainsOneOf(contrib.Labels, "invalid", "spam") {
		p.Invalids = append(p.Invalids, contrib)
	} else if labelsContainsOneOf(contrib.Labels, "hacktoberfest-accepted") || contrib.CreatedAt.Before(date2020) {
		p.Contributions = append(p.Contributions, contrib)
	} else {
		ok, err := repo.HasTopic("hacktoberfest")
		if err != nil {
			return err
		}
		if ok {
			merged, err := repo.IsMerged(contrib.GetNumber())
			if err != nil {
				return err
			}
			if merged {
				p.Contributions = append(p.Contributions, contrib)
			} else if contrib.GetState() == "closed" {
				p.Invalids = append(p.Invalids, contrib)
			} else {
				p.Pending = append(p.Pending, contrib)
			}
		} else {
			p.Ignored = append(p.Ignored, contrib)
		}
	}
	return nil
}

func (p *Player) AddContribs(contribs []*github.Issue) error {
	for _, contrib := range contribs {
		if err := p.AddContrib(contrib); err != nil {
			return err
		}
	}
	return nil
}

func labelsContainsOneOf(labels []*github.Label, l ...string) bool {
	for _, label := range labels {
		for _, other := range l {
			if label.GetName() == other {
				return true
			}
		}
	}
	return false
}

type Repo struct {
	gh     *github.Client
	Owner  string
	Name   string
	URL    string
	topics []string
	merged map[int]*bool
}

func (r *Repo) FullName() string {
	return fmt.Sprintf("%s/%s", r.Owner, r.Name)
}

func (r *Repo) Topics() ([]string, error) {
	if r.topics == nil {
		topics, _, err := r.gh.Repositories.ListAllTopics(context.TODO(), r.Owner, r.Name)
		if err != nil {
			return nil, err
		}
		r.topics = topics
	}
	return r.topics, nil
}

func (r *Repo) HasTopic(topic string) (bool, error) {
	topics, err := r.Topics()
	if err != nil {
		return false, err
	}
	for _, t := range topics {
		if t == topic {
			return true, nil
		}
	}
	return false, nil
}

func (r *Repo) IsMerged(number int) (bool, error) {
	if _, ok := r.merged[number]; !ok {
		merged, _, err := r.gh.PullRequests.IsMerged(context.TODO(), r.Owner, r.Name, number)
		if err != nil {
			return false, err
		}
		r.merged[number] = &merged
	}
	return *r.merged[number], nil
}

func repoFromUrl(gh *github.Client, url string) *Repo {
	s := strings.Split(url, "/")
	s = s[len(s)-2:]
	return &Repo{gh, s[0], s[1], url, nil, map[int]*bool{}}
}
