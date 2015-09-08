package handlers

import (
	"crypto/rand"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/repbin/repbin/cmd/repserver/stat"
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
	// Start statistics goroutine
	if ms.Stat {
		go stat.Run()
	}
	// Start timers
	go ms.notifyWatch()
	// static file handler
	httpHandlers := http.NewServeMux()

	httpHandlers.Handle("/", http.FileServer(http.Dir(ms.path+string(os.PathSeparator)+"static")))
	httpHandlers.HandleFunc("/id", ms.ServeID)
	if !ms.HubOnly {
		httpHandlers.HandleFunc("/keyindex", ms.GetKeyIndex)
		httpHandlers.HandleFunc("/post", ms.GenPostHandler(false))
		if ms.EnableOneTimeHandler {
			httpHandlers.HandleFunc("/local/post", ms.GenPostHandler(true))
		}
		if ms.EnableDeleteHandler {
			httpHandlers.HandleFunc("/delete", ms.Delete)
		}
	}
	httpHandlers.HandleFunc("/globalindex", ms.GetGlobalIndex)
	httpHandlers.HandleFunc("/fetch", ms.Fetch)
	httpHandlers.HandleFunc("/notify", ms.GetNotify)
	httpServer := &http.Server{
		Addr:           "127.0.0.1:" + strconv.Itoa(ms.ListenPort),
		Handler:        httpHandlers,
		ReadTimeout:    90 * time.Second,
		WriteTimeout:   90 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	httpServer.ListenAndServe()
}
