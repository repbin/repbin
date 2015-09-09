package handlers

import (
	"time"

	"github.com/repbin/repbin/cmd/repserver/stat"
	log "github.com/repbin/repbin/deferconsole"
)

// FetchRun is called when it is time to update peer information and load messages.
func (ms MessageServer) FetchRun() {
	ms.LoadPeers()
	ms.FetchPeers()
}

func (ms MessageServer) notifyWatch() {
	var lastNotify, lastMessage int64
	notifyTick := time.Tick(time.Duration(ms.NotifyDuration) * time.Second)
	fetchTick := time.Tick(time.Duration(ms.FetchDuration) * time.Second)
	expireTick := time.Tick(time.Duration(ms.ExpireDuration) * time.Second)
	expireFSTick := time.Tick(time.Duration(ms.ExpireFSDuration) * time.Second)
	var statTick <-chan time.Time
	if ms.Stat {
		// show statistics once a minute
		statTick = time.Tick(60 * time.Second)
	}
	for {
		select {
		case <-notifyTick:
			log.Debugs("Check notification.\n")
			if lastNotify < lastMessage { // both are zero when started
				log.Debugs("Notify run started.\n")
				lastNotify = CurrentTime()
				ms.NotifyPeers()
			}
		case <-fetchTick:
			log.Debugs("Fetch run started.\n")
			ms.FetchRun()
		case <-expireTick:
			log.Debugs("Expire run started.\n")
			ms.DB.ExpireFromIndex(2) // we go back at most 2 cycles
		case <-expireFSTick:
			log.Debugs("ExpireFS run started.\n")
			ms.DB.ExpireFromFS()
		case <-statTick:
			stat.Input <- stat.Show
		case <-ms.notifyChan:
			log.Debugs("Notification reason\n")
			lastMessage = CurrentTime()
		}
	}
}
