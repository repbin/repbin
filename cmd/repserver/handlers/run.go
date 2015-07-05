package handlers

import (
	"crypto/rand"
	"math/big"
	"net/http"
	"os"
	"strconv"

	log "github.com/repbin/repbin/deferconsole"
)

// RunServer starts the HTTP handlers.
func (ms MessageServer) RunServer() {
	ts, err := rand.Int(rand.Reader, big.NewInt(ms.MaxTimeSkew))
	if err != nil {
		panic(err)
	}
	ms.TimeSkew = ts.Int64()
	log.Debugf("TimeSkew: %d\n", ms.TimeSkew)

	// Peering
	ms.notifyChan = make(chan bool, 3)
	// Load peers
	ms.LoadPeers()
	// Start timers
	go ms.notifyWatch()
	// static file handler
	http.Handle("/", http.FileServer(http.Dir(ms.path+string(os.PathSeparator)+"static")))
	http.HandleFunc("/id", ms.ServeID)
	if !ms.HubOnly {
		http.HandleFunc("/keyindex", ms.GetKeyIndex)
		http.HandleFunc("/post", ms.GenPostHandler(false))
		if ms.EnableOneTimeHandler {
			http.HandleFunc("/local/post", ms.GenPostHandler(true))
		}
		if ms.EnableDeleteHandler {
			http.HandleFunc("/delete", ms.Delete)
		}
	}
	http.HandleFunc("/globalindex", ms.GetGlobalIndex)
	http.HandleFunc("/fetch", ms.Fetch)
	http.HandleFunc("/notify", ms.GetNotify)
	listenURL := "127.0.0.1:" + strconv.Itoa(ms.ListenPort)
	http.ListenAndServe(listenURL, nil) // enable
}
