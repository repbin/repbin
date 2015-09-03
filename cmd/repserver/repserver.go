// repserver is the repbin server that makes posts available to users.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/repbin/repbin/cmd/repserver/handlers"
	"github.com/repbin/repbin/cmd/repserver/messagestore"
	"github.com/repbin/repbin/cmd/repserver/messagestore/sql"
	log "github.com/repbin/repbin/deferconsole"
	"github.com/repbin/repbin/hashcash"
	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils"
	"github.com/repbin/repbin/utils/keyauth"
	"github.com/repbin/repbin/utils/keyproof"
	"github.com/repbin/repbin/utils/repproto"
	"github.com/repbin/repbin/utils/repproto/structs"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

// Version of this release
const Version = "0.0.2 very alpha"

var (
	start *bool
	stat  *bool
)

func init() {
	version := flag.Bool("version", false, "Print version information")
	showconfig := flag.Bool("showconfig", false, "Print config file template")
	configfile := flag.String("configfile", "", "Path to configuration file")
	verbose := flag.Bool("verbose", false, "Show some verbose output")
	start = flag.Bool("start", false, "Start server")
	stat = flag.Bool("stat", false, "Enable usage statistics")
	flag.Parse()
	if *version {
		fmt.Printf("Repserver: %s\n", Version)
		fmt.Printf("MessageStore: %s\n", messagestore.Version)
		fmt.Printf("SQL Store: %s\n", sql.Version)
		fmt.Printf("Handlers: %s\n", handlers.Version)
		fmt.Printf("HashCash: %s\n", hashcash.Version)
		fmt.Printf("Utils: %s\n", utils.Version)
		fmt.Printf("KeyAuth: %s\n", keyauth.Version)
		fmt.Printf("KeyProof: %s\n", keyproof.Version)
		fmt.Printf("Protocol: %s\n", repproto.Version)
		fmt.Printf("Protocol Structures: %s\n", structs.Version)
		fmt.Printf("Message: %s\n", message.VersionID)
		fmt.Printf("Message Format: %d\n", message.Version)
		os.Exit(0)
	}
	if *showconfig {
		showConfig()
		os.Exit(0)
	}
	if *configfile == "" || *configfile == "/" || len(*configfile) < 4 {
		fmt.Println("No configuration file found. Specify with --configfile=FILE")
		os.Exit(0)
	}
	if *verbose {
		log.SetMinLevel(log.LevelDebug)
	}
	if *stat && !*verbose {
		fmt.Println("Error: option --stat requires option --verbose")
		os.Exit(1)
	}
	err := loadConfig(*configfile)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}

func main() {
	var pubKey, privKey []byte
	pubKey, privKey = nil, nil

	if defaultSettings.StoragePath == "" {
		fmt.Println("Error: Storage path not configured")
		os.Exit(1)
	}

	if defaultSettings.PeeringPublicKey != "" && defaultSettings.PeeringPrivateKey != "" {
		pubKey = utils.B58decode(defaultSettings.PeeringPublicKey)
		privKey = utils.B58decode(defaultSettings.PeeringPrivateKey)
	}

	runtime.GOMAXPROCS(runtime.NumCPU() - 1)

	ms, err := handlers.New(defaultSettings.DBDriver, defaultSettings.DBURL, defaultSettings.StoragePath, pubKey, privKey)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	applyConfig(ms)
	ms.Stat = *stat
	if *start {
		ms.RunServer()
	} else {
		fmt.Println("Server not started. Enable with --start")
		os.Exit(1)
	}
	os.Exit(0)
}
