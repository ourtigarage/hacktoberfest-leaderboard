package main

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jinzhu/copier"
)

type Leaderboard struct {
	lock      *sync.RWMutex
	collector *StatsCollector
	players   Players
	sorted    []*Player
	ready     int32
	cancel    func()
	C         chan struct{}
}

func NewLeaderboard(eventDate, playerFile string) *Leaderboard {
	lb := &Leaderboard{
		lock:      new(sync.RWMutex),
		collector: NewStatsCollector(eventDate, playerFile),
		players:   Players{},
		sorted:    []*Player{},
		ready:     0,
		C:         make(chan struct{}),
	}
	var ctx context.Context
	ctx, lb.cancel = context.WithCancel(context.Background())
	lb.start(ctx)
	return lb
}

func (lb *Leaderboard) Shutdown() {
	lb.cancel()
	<-lb.C
}

func (lb *Leaderboard) Ready() bool {
	return atomic.LoadInt32(&lb.ready) > 0
}

func (lb *Leaderboard) PlayersSorted() ([]*Player, error) {
	lb.lock.RLock()
	defer lb.lock.RUnlock()
	pls := []*Player{}
	return pls, copier.Copy(&pls, lb.sorted)
}

func (lb *Leaderboard) Player(username string) (*Player, error) {
	lb.lock.RLock()
	p, ok := lb.players[username]
	if !ok {
		lb.lock.RUnlock()
		return nil, fmt.Errorf("User \"%s\" not found", username)
	}
	lb.lock.RUnlock()
	res := new(Player)
	return res, copier.Copy(res, p)
}

func (lb *Leaderboard) PlayerNames() []string {
	names := make([]string, 0, len(lb.players))
	lb.lock.RLock()
	for k := range lb.players {
		names = append(names, k)
	}
	lb.lock.RUnlock()
	return names
}

func (lb *Leaderboard) start(ctx context.Context) {
	fmt.Println("Starting background stats collector")
	go func() {
		defer close(lb.C)
		for {
			lb.update(ctx)
			t := time.NewTimer(COLLECT_PERIOD)
			select {
			case <-ctx.Done():
				fmt.Println("Background stats collector shutted down")
				if !t.Stop() {
					<-t.C
				}
				return
			case <-t.C: // NOP
			}
		}
	}()
}

func (lb *Leaderboard) update(ctx context.Context) {
	fmt.Println("Collecting data")
	start := time.Now()
	players, err := lb.collector.Players(ctx)
	if err != nil && err != context.Canceled {
		fmt.Println("[ERROR]", err)
	} else if err != context.Canceled {
		duration := time.Since(start)
		fmt.Println("Collection completed in", duration.String())
		lb.lock.Lock()
		lb.players = players
		lb.sort()
		lb.lock.Unlock()
		atomic.StoreInt32(&lb.ready, 1)
	}
}

func (lb *Leaderboard) sort() {
	players := make([]*Player, 0, len(lb.players))
	for _, p := range lb.players {
		players = append(players, p)
	}
	sort.Slice(players, func(i, j int) bool {
		ci := players[i].ContributionCount()
		cj := players[j].ContributionCount()
		if ci != cj {
			return ci > cj
		} else if len(players[i].Merged) == len(players[j].Merged) {
			return players[i].LastMergeAt.Before(players[j].LastMergeAt)
		}
		return len(players[i].Merged) > len(players[j].Merged)
	})
	lb.sorted = players
}
