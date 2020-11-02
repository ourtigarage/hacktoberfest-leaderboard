package main

import (
	"html/template"
)

type Badge struct {
	Short       string
	Title       string
	Description template.HTML
	f           func(player *Player) int
}

func (b *Badge) EarnedBy(player *Player) bool {
	return b.TotalEarn(player) > 0
}

func (b *Badge) TotalEarn(player *Player) int {
	return b.f(player)
}

var BADGES = []Badge{
	{
		"hacktoberfest",
		"Hacktoberfest completed",
		"Completed the hacktoberfest challenge by submitting enough pull requests",
		func(p *Player) int {
			if len(p.Contributions()) >= TARGET_OBJECTIVE {
				return 1
			}
			return 0
		},
	},
	{
		"leaderboard",
		"The leaderboard contributor",
		"Submitted 1 Pull Request to this leaderboard's code repository",
		func(p *Player) int {
			c := 0
			for _, contrib := range p.Contributions() {
				if contrib.Repo.URL == LEADERBOARD_URL {
					c += 1
				}
			}
			return c
		},
	},
	{
		"snake",
		"The snake charmer",
		"Submitted 1 Pull Request to the <a href=\"https://ourtigarage.github.io/web-snake/\">snake game</a>'s code repository",
		func(p *Player) int {
			c := 0
			for _, contrib := range p.Contributions() {
				if contrib.Repo.URL == SNAKE_URL {
					c += 1
				}
			}
			return c
		},
	},
	{
		"10-contributions",
		"The Pull Request champion",
		"Submitted more than 10 Pull requests",
		func(p *Player) int {
			return p.ContributionCount() / 10
		},
	},
	{
		"crap",
		"The smelly code",
		"Has a Pull Request marked as invalid. Probably some bad smelling code",
		func(p *Player) int {
			return len(p.Invalids)
		},
	},
	{
		"novelist",
		"The novelist",
		"Wrote more than 100 words in a Pull Request's description",
		func(p *Player) int {
			c := 0
			for _, contrib := range p.Contributions() {
				if len(contrib.Description) >= 100 {
					c += 1
				}
			}
			return c
		},
	},
	{
		"taciturn",
		"The taciturn",
		"Submitted a Pull Request with no description",
		func(p *Player) int {
			c := 0
			for _, contrib := range p.Contributions() {
				if len(contrib.Description) == 0 {
					c += 1
				}
			}
			return c
		},
	},
	{
		"pirate",
		"The mighty pirate",
		"A lawless pirate who submitted Pull Requests to his own repositories. Cheater...",
		func(p *Player) int {
			c := 0
			for _, contrib := range p.Contributions() {
				if contrib.Repo.Owner == p.Username {
					c += 1
				}
			}
			return c
		},
	},
	{
		"adventure",
		"The adventurer",
		"Submitted 1 Pull Request to a repository he does not own, out of <a href=\"https://github.com/ourtigarage\">ourtigarage</a> organisation",
		func(p *Player) int {
			c := 0
			for _, contrib := range p.Contributions() {
				if contrib.Repo.Owner != p.Username && contrib.Repo.Owner != "ourtigarage" {
					c += 1
				}
			}
			return c
		},
	},
	{
		"narcissistic",
		"The narcissistic",
		"Self assigned & merged one of his own tremendous pull request",
		func(p *Player) int {
			c := 0
			for _, contrib := range p.Merged {
				if contrib.MergedBy == p.Username && contrib.Assignee == p.Username {
					c += 1
				}
			}
			return c
		},
	},
	{
		"first-merge",
		"First merge",
		"Got his first Pull Request merged",
		func(p *Player) int {
			if len(p.Merged) > 0 {
				return 1
			}
			return 0
		},
	},
}
