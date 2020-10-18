package main

import (
	"time"

	"github.com/google/go-github/v32/github"
)

type Player struct {
	Username                                     string
	Avatar                                       string
	Profile                                      string
	Merged, Accepted, Pending, Ignored, Invalids []*PullRequest
	Issues                                       []*Issue
	Repos                                        map[string]*Repo
	LastMergeAt                                  time.Time
}

type Players map[string]*Player

func NewPlayer(user *github.User) *Player {
	return &Player{
		Username: user.GetLogin(),
		Avatar:   user.GetAvatarURL(),
		Profile:  user.GetHTMLURL(),
		Repos:    make(map[string]*Repo),
	}
}

func (p *Player) IsChallengeComplete() bool {
	return p.ContributionCount() >= TARGET_OBJECTIVE
}

func (p *Player) Contributions() []*PullRequest {
	prs := make([]*PullRequest, 0, p.ContributionCount())
	prs = append(prs, p.Merged...)
	prs = append(prs, p.Accepted...)
	return prs
}

func (p *Player) ContributionCount() int {
	return len(p.Merged) + len(p.Accepted)
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

func (p *Player) AddIssue(issue *Issue) {
	p.Issues = append(p.Issues, issue)
	p.Repos[issue.Repo.FullName()] = issue.Repo
}

// For all submissions after 3rd of October 2020, the rules have changed
var date2020, _ = time.Parse("2006-01-02", "2020-10-03")

func (p *Player) AddContrib(pr *PullRequest) {
	p.Repos[pr.Repo.FullName()] = pr.Repo
	if pr.HasOneOfLabels("invalid", "spam") {
		p.Invalids = append(p.Invalids, pr)
	} else if pr.HasOneOfLabels("hacktoberfest-accepted") || pr.CreatedAt.Before(date2020) {
		if pr.Merged {
			p.Merged = append(p.Merged, pr)
			if p.LastMergeAt.Before(pr.MergedAt) {
				p.LastMergeAt = pr.MergedAt
			}
		} else {
			p.Accepted = append(p.Accepted, pr)
		}
	} else {
		if pr.Repo.HasTopic("hacktoberfest") {
			if pr.Merged {
				p.Merged = append(p.Merged, pr)
				if p.LastMergeAt.Before(pr.MergedAt) {
					p.LastMergeAt = pr.MergedAt
				}
			} else if pr.State == "closed" {
				p.Invalids = append(p.Invalids, pr)
			} else {
				p.Pending = append(p.Pending, pr)
			}
		} else {
			p.Ignored = append(p.Ignored, pr)
		}
	}
}
