// repclient is the repbin client to send and receive files.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/repbin/repbin/cmd/repclient/client"
	log "github.com/repbin/repbin/deferconsole"
	"github.com/repbin/repbin/hashcash"
	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils"
	"github.com/repbin/repbin/utils/keyauth"
	"github.com/repbin/repbin/utils/repproto"
	"github.com/repbin/repbin/utils/repproto/structs"
)

// Version of main
const Version = "0.0.1 very alpha"

const defaultBootStrapPeer = "http://mmb4alp3d2vc55op.onion/"

const (
	cmdGenKey = iota
	cmdGenTempKey
	cmdEncrypt
	cmdDecrypt
	cmdPost
	cmdGet
	cmdRewrap
	cmdIndex
	cmdShowConfig
	cmdPeerList
	cmdSTM
	cmdHelp
	cmdExpert
	cmdGuru
	cmdVersion
	cmdMax
)

var options client.Options

var globalconfig = client.ConfigVariables{
	BodyLength:    message.DefaultTotalLength,
	PadToLength:   message.DefaultPadToLength,
	MinHashCash:   24,
	KeyDir:        "",
	SocksServer:   "socks5://127.0.0.1:9050",
	BootStrapPeer: "",
	PasteServers:  []string{},
}
var commands [cmdMax]*bool
var callFunc [cmdMax]func() int

func init() {
	for k := range commands {
		commands[k] = new(bool)
	}
	flag.BoolVar(commands[cmdShowConfig], "defaultconfig", false, "Print default configfile")
	callFunc[cmdShowConfig] = client.CmdShowConfig

	flag.BoolVar(commands[cmdGenKey], "genkey", false, "Generate a long-term key")
	callFunc[cmdGenKey] = client.CmdGenKey

	flag.BoolVar(commands[cmdGenTempKey], "gentemp", false, "Generate a temporary key")
	callFunc[cmdGenTempKey] = client.CmdGenTempKey

	flag.BoolVar(commands[cmdEncrypt], "encrypt", false, "Encrypt data")
	callFunc[cmdEncrypt] = client.CmdEncrypt

	flag.BoolVar(commands[cmdDecrypt], "decrypt", false, "Decrypt data")
	callFunc[cmdDecrypt] = client.CmdDecrypt

	flag.BoolVar(commands[cmdPost], "post", false, "Post message")
	callFunc[cmdPost] = client.CmdPost

	flag.BoolVar(commands[cmdGet], "get", false, "Get message")
	callFunc[cmdGet] = client.CmdGet

	flag.BoolVar(commands[cmdSTM], "stm", false, "STM post run")
	callFunc[cmdSTM] = client.CmdSTM

	flag.BoolVar(commands[cmdIndex], "index", false, "List index")
	callFunc[cmdIndex] = client.CmdIndex

	flag.BoolVar(commands[cmdPeerList], "peerlist", false, "Update configuration with new peerlist")
	callFunc[cmdPeerList] = client.CmdPeerList

	flag.BoolVar(commands[cmdVersion], "version", false, "Show version information")
	callFunc[cmdVersion] = CmdVersion
	flag.BoolVar(commands[cmdHelp], "help", false, "Show help")
	flag.BoolVar(commands[cmdHelp], "h", false, "Show help")
	callFunc[cmdHelp] = CmdHelp

	flag.BoolVar(commands[cmdExpert], "expert", false, "Show expert help")
	callFunc[cmdExpert] = CmdExpert

	flag.BoolVar(commands[cmdGuru], "guru", false, "Show guru help")
	callFunc[cmdGuru] = CmdGuru

	flag.IntVar(&options.Start, "start", 0, "index start position")
	flag.IntVar(&options.Count, "count", 10, "index count")
	flag.StringVar(&options.Outdir, "outdir", "", "Index batch download directory")

	flag.BoolVar(&options.Verbose, "verbose", false, "be verbose")
	flag.BoolVar(&options.KEYVERB, "KEYVERB", false, "show secrets during calculation")
	flag.BoolVar(&options.Appdata, "appdata", false, "show application integration data")

	flag.BoolVar(&options.Sync, "sync", true, "sync messages for key")
	flag.BoolVar(&options.Hidden, "hidden", false, "hide messages for key")

	flag.StringVar(&options.Privkey, "privkey", "", "private key")

	flag.StringVar(&options.Infile, "in", "", "Read data from file (can be stdin: -)")
	flag.StringVar(&options.Outfile, "out", "", "Write data to file (can be stdout: -)")
	flag.StringVar(&options.Stmdir, "stmdir", "", "STM Directory")

	flag.StringVar(&options.Senderkey, "senderPubKey", "", "Public key of sender")

	flag.StringVar(&options.Recipientkey, "recipientPubKey", "", "Public key of recipient")
	flag.BoolVar(&options.Embedkey, "embedReply", false, "Embed reply keys.")

	flag.BoolVar(&options.Notrace, "notrace", false, "Generate new embedded keypair.")
	flag.BoolVar(&options.Anonymous, "anonymous", false, "Do not use private key or signer keypair.")

	flag.IntVar(&options.MessageType, "messageType", message.MsgTypeBlob, "Message type")
	flag.IntVar(&options.Keymgt, "keymgt", -1, "Key management file descriptor")

	flag.IntVar(&options.Mindelay, "minDelay", 0, "Minimum repost delay")
	flag.IntVar(&options.Maxdelay, "maxDelay", 0, "Maximum repost delay")

	flag.BoolVar(&options.Repost, "repost", false, "Create a repost message.")

	flag.StringVar(&options.Socksserver, "socksserver", "socks5://127.0.0.1:9050", "Socks server URL")
	flag.StringVar(&options.Server, "server", "", "Repbin server")

	flag.StringVar(&options.Signkey, "signkey", "", "Post signature file")
	flag.StringVar(&options.Signdir, "signdir", "", "Post signature directory")
	flag.StringVar(&options.Configfile, "config", "", "Load configuration from file")

	flag.StringVar(&globalconfig.BootStrapPeer, "bootstrap", defaultBootStrapPeer, "Bootstrap-Server URL")
	flag.Parse()
}

func testCommands() bool {
	i := 0
	for _, x := range commands {
		if *x {
			i++
		}
	}
	if i > 1 {
		log.Fatal("Only one command may be specified")
		log.Sync()
		return false
	}
	if i == 0 {
		if len(flag.Args()) > 0 {
			*commands[cmdDecrypt] = true
		} else {
			*commands[cmdEncrypt] = true
		}
	}
	return true
}

func runCommands() (retval int) {
	log.SetMinLevel(log.LevelPrint)
	if options.Appdata {
		log.SetMinLevel(log.LevelData)
	}
	if options.Verbose {
		log.SetMinLevel(log.LevelDebug)
	}
	if options.KEYVERB {
		log.SetMinLevel(log.LevelSecret)
	}
	usercfgfile := client.UserConfigFile()
	if options.Configfile != "" || usercfgfile != "" {
		configFile := usercfgfile
		if options.Configfile != "" {
			configFile = options.Configfile
		}
		globalconfigT, err := client.LoadConfigFile(configFile)
		if err == nil {
			globalconfig = globalconfigT
		} else {
			log.Errorf("Config file failed to load: %s\n", err)
		}
	}
	for k, v := range commands {
		if *v {
			if callFunc[k] != nil {
				client.OptionsVar = options
				client.GlobalConfigVar = globalconfig
				retval = callFunc[k]()
			} else {
				log.Fatal("Command not implemented")
				return 100
			}
			log.Sync()
			break
		}
	}
	return
}

// CmdVersion returns version information
func CmdVersion() int {
	fmt.Printf("Repclient: %s\n", Version)
	fmt.Printf("Repclient functions: %s\n", client.Version)
	fmt.Printf("HashCash: %s\n", hashcash.Version)
	fmt.Printf("Utils: %s\n", utils.Version)
	fmt.Printf("KeyAuth: %s\n", keyauth.Version)
	fmt.Printf("Protocol: %s\n", repproto.Version)
	fmt.Printf("Protocol Structures: %s\n", structs.Version)
	fmt.Printf("Message: %s\n", message.VersionID)
	fmt.Printf("Message Format: %d\n", message.Version)
	return 0
}

// CmdHelp shows the help
func CmdHelp() int {
	fmt.Println(`Usage of repclient:

Standard usage
--------------
Post to server:
   cat FILE | repclient
 or
   repclient -in <FILE>         Post FILE to server

Fetch post from server:
   repclient <URL>              Fetch URL from server and display
 or
   repclient -out <FILE> <URL>  Fetch URL from server and write to FILE


General options:
  -out <FILE>          Write output data to FILE
  -in  <FILE>          Read data from FILE
  -server <URL>        Use repserver at URL
  -socksserver <URL>   Use Tor socks server at URL

Configuration file:
  Default configuration file is ~/.config/repclient/repclient.config
  -defaultconfig       Display default configuration
  -peerlist            Update the peerlist and display configuration
  -bootstrap <URL>     Use URL as bootstrap server instead of hardcoded default
  -config <FILE>       Load configuration from FILE

Help:
  --help       Show this help
  --version    Show version information
  --verbose    Display verbose output
  --expert     Show expert help
  --guru       Show guru help
`)
	return 0
}

// CmdExpert displays dieplay help
func CmdExpert() int {
	fmt.Println(`Expert help on repclient:

Encrypting, extra options:
  -encrypt         Encrypt data
  -privkey <KEY>   Private key. tty to read from tty
  -embedReply      Embed reply keys into message
  -notrace         Embedded keys are independet of private key
  -anonymous       Do not use any identifyable information
  -signdir <DIR>   Load signer from DIR and delete signer after use
  -signkey <FILE>  Load signer from FILE
  -recipientPubKey <KEY>  Send to <KEY>

Decrypting, extra options:
  -decrypt              Decrypt data
  -senderPubKey <KEY>   Verify sender's public key
  -privkey <KEY>        Use private key for decryption

Longterm key generation:
  -genkey          Generate a long-term key
  -hidden          Hide message index. Force authentication
  -sync false      Disable replication between peers

Temporary key generation:
  -gentemp         Generate a temporary key for longterm key
  -privkey <KEY>   Private longterm key. Can also be "-" to read from
                   stdin, or "tty" to read from tty

Post/Get message:
  -get             Get message. MessageID on commandline
  -post            Post message. Read from stdin.

Post-Box support:
  -index           List Post-Box index
  -privkey         Private key of post-box. Required
  -start <NUMBER>  Start at index NUMBER
  -count <NUMBER>  Return at most NUMBER posts
  -outdir <DIR>    Download messages to DIR
`)
	return 0
}

// CmdGuru displays guru help
func CmdGuru() int {
	fmt.Println(`Guru help: Do not use unless you really know what it all means.

Additional encryption options:
  -messageType <NUMBER>  Define Message-Type
  -repost                Create a repost message.
  -maxDelay <SECONDS>    Maximum repost delay
  -minDelay <SECONDS>    Minimum repost delay

Additional decryption options:
  -stmdir <DIR>          Store received repost messages in DIR

Application integration:
  -keymgt=<FD>    Enable key management callback
  -KEYVERB        Show secret information
  -appdata        Enable application integration output

STM/Repost server:
  -stm            Run STM server
  -stmdir         Directory from which to post
`)
	return 0
}

func main() {
	if testCommands() {
		os.Exit(runCommands())
	}
	os.Exit(1)
}
