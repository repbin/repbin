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
			removeOld(postList1, t-1*60)
			removeOld(postList5, t-5*60)
			removeOld(postList60, t-60*60)
			removeOld(postList1440, t-1440*60)
			removeOld(fetchList1, t-1*60)
			removeOld(fetchList5, t-5*60)
			removeOld(fetchList60, t-60*60)
			removeOld(fetchList1440, t-1440*60)
			log.Debugf("Posts/minute:   %8.3f (1m) %8.3f (5m) %8.3f (1h) %8.3f (24h)\n",
				float64(postList1.Len()),
				float64(postList5.Len())/5.0,
				float64(postList60.Len())/60.0,
				float64(postList1440.Len())/1440.0)
			log.Debugf("Fetches/minute: %8.3f (1m) %8.3f (5m) %8.3f (1h) %8.3f (24h)\n",
				float64(fetchList1.Len()),
				float64(fetchList5.Len())/5.0,
				float64(fetchList60.Len())/60.0,
				float64(fetchList1440.Len())/1440.0)
		}
	}
}
