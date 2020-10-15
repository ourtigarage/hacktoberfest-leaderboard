package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v32/github"
)

type Repo struct {
	Owner   string
	Name    string
	HTMLURL string
	URL     string
	Topics  []string
}

func (r *Repo) FullName() string {
	return fmt.Sprintf("%s/%s", r.Owner, r.Name)
}

func (r *Repo) HasTopic(topic string) bool {
	for _, t := range r.Topics {
		if t == topic {
			return true
		}
	}
	return false
}

func NewRepoFromURL(gh *github.Client, url string) (*Repo, error) {
	s := strings.Split(url, "/")
	s = s[len(s)-2:]
	repo, _, err := gh.Repositories.Get(context.TODO(), s[0], s[1])
	if err != nil {
		return nil, err
	}
	return &Repo{
		Owner:   repo.Owner.GetLogin(),
		Name:    repo.GetName(),
		HTMLURL: repo.GetHTMLURL(),
		URL:     repo.GetURL(),
		Topics:  repo.Topics,
	}, nil
}

type Issue struct {
	Title       string
	HTMLURL     string
	Description string
	Author      string
	Assignee    string
	CreatedAt   time.Time
	State       string
	Labels      []string
	Repo        *Repo
}

func NewIssue(issue *github.Issue, repo *Repo) *Issue {
	// r, _, _ := gh.Reactions.ListIssueReactions(context.TODO(), repo.Owner, repo.Name, issue.GetNumber(), nil)
	// r[0].
	iss := &Issue{
		Title:       issue.GetTitle(),
		HTMLURL:     issue.GetHTMLURL(),
		Description: strings.TrimSpace(issue.GetBody()),
		Author:      issue.GetUser().GetLogin(),
		State:       issue.GetState(),
		CreatedAt:   issue.GetCreatedAt(),
		Repo:        repo,
	}
	if assignee := issue.GetAssignee(); assignee != nil {
		iss.Assignee = assignee.GetLogin()
	}
	for _, label := range issue.Labels {
		iss.Labels = append(iss.Labels, label.GetName())
	}
	return iss
}

type PullRequest struct {
	*Issue
	Merged   bool
	MergedBy string
	MergedAt time.Time
}

func NewPullRequest(gh *github.Client, issue *github.Issue, repo *Repo) (*PullRequest, error) {
	pr, _, err := gh.PullRequests.Get(context.TODO(), repo.Owner, repo.Name, issue.GetNumber())
	if err != nil {
		return nil, err
	}
	p := &PullRequest{
		Issue:    NewIssue(issue, repo),
		Merged:   pr.GetMerged(),
		MergedAt: pr.GetMergedAt(),
	}
	if p.Merged {
		p.MergedBy = pr.GetMergedBy().GetLogin()
	}
	return p, nil
}

func (pr *PullRequest) HasOneOfLabels(labels ...string) bool {
	for _, label := range pr.Labels {
		for _, other := range labels {
			if label == other {
				return true
			}
		}
	}
	return false
}
