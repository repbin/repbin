// Package stat gathers and prints repbin server usage statistics.
// Post and fetch statistics are shown as minute-by-minute averages over the
// last minute, 5 minutes, 60 minutes and 24 hours.
package stat

import (
	"container/list"
	"time"

	log "github.com/repbin/repbin/deferconsole"
)

// Event describes the stat event send to the input channel.
type Event int

const (
	// Post adds a new post to the statistic.
	Post = iota
	// Fetch adds a new fetch to the statistic.
	Fetch
	// Show prints the statistics.
	Show
)

const (
	oneMin  = 1 * 60
	fiveMin = 5 * 60
	oneHour = 60 * 60
	oneDay  = 24 * 60 * 60
)

// Input is the repserver statistics input channel.
var Input = make(chan Event, 1024)

func now() int64 {
	return time.Now().UTC().Unix()
}

func removeOld(l *list.List, cutoff int64) {
	e := l.Front()
	for e != nil {
		if e.Value.(int64) <= cutoff {
			l.Remove(e)
			e = l.Front()
		} else {
			break
		}
	}
}

// Run starts the statistics handler.
func Run() {
	start := now()
	postList1 := list.New()
	postList5 := list.New()
	postList60 := list.New()
	postList1440 := list.New()
	fetchList1 := list.New()
	fetchList5 := list.New()
	fetchList60 := list.New()
	fetchList1440 := list.New()
	for {
		m := <-Input
		t := now()
		switch m {
		case Post:
			postList1.PushBack(t)
			postList5.PushBack(t)
			postList60.PushBack(t)
			postList1440.PushBack(t)
		case Fetch:
			fetchList1.PushBack(t)
			fetchList5.PushBack(t)
			fetchList60.PushBack(t)
			fetchList1440.PushBack(t)
		case Show:
			removeOld(postList1, t-oneMin)
			removeOld(postList5, t-fiveMin)
			removeOld(postList60, t-oneHour)
			removeOld(postList1440, t-oneDay)
			removeOld(fetchList1, t-oneMin)
			removeOld(fetchList5, t-fiveMin)
			removeOld(fetchList60, t-oneHour)
			removeOld(fetchList1440, t-oneDay)
			postOneMin := float64(postList1.Len())
			postFiveMin := float64(postList5.Len()) / 5.0
			postOneHour := float64(postList60.Len()) / 60.0
			postOneDay := float64(postList1440.Len()) / 1440.0
			fetchOneMin := float64(fetchList1.Len())
			fetchFiveMin := float64(fetchList5.Len()) / 5.0
			fetchOneHour := float64(fetchList60.Len()) / 60.0
			fetchOneDay := float64(fetchList1440.Len()) / 1440.0
			if t-oneDay >= start {
				log.Debugf("Posts/minute:   %8.3f (1m) %8.3f (5m) %8.3f (1h) %8.3f (24h)\n",
					postOneMin, postFiveMin, postOneHour, postOneDay)
				log.Debugf("Fetches/minute: %8.3f (1m) %8.3f (5m) %8.3f (1h) %8.3f (24h)\n",
					fetchOneMin, fetchFiveMin, fetchOneHour, fetchOneDay)
			} else if t-oneHour >= start {
				log.Debugf("Posts/minute:   %8.3f (1m) %8.3f (5m) %8.3f (1h)\n",
					postOneMin, postFiveMin, postOneHour)
				log.Debugf("Fetches/minute: %8.3f (1m) %8.3f (5m) %8.3f (1h)\n",
					fetchOneMin, fetchFiveMin, fetchOneHour)

			} else if t-fiveMin >= start {
				log.Debugf("Posts/minute:   %8.3f (1m) %8.3f (5m)\n",
					postOneMin, postFiveMin)
				log.Debugf("Fetches/minute: %8.3f (1m) %8.3f (5m)\n",
					fetchOneMin, fetchFiveMin)
			} else {
				log.Debugf("Posts/minute:   %8.3f (1m)\n",
					postOneMin)
				log.Debugf("Fetches/minute: %8.3f (1m)\n",
					fetchOneMin)
			}
		}
	}
}
