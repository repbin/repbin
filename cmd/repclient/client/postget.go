package client

import (
	"flag"

	log "github.com/repbin/repbin/deferconsole"
	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils/repproto"
)

// CmdPost posts a message
func CmdPost() int {
	var inData []byte
	var err error
	if OptionsVar.Server == "" {
		log.Fatal("Server must be specified: --server")
		return 1
	}

	maxInData := int64(GlobalConfigVar.BodyLength-(message.Curve25519KeySize*2)) * 5
	inData, err = inputData(OptionsVar.Infile, maxInData)
	if err != nil {
		log.Fatalf("No input data: %s\n", err)
		return 1
	}
	log.Datas("STATUS (Process):\tPOST\n")
	proto := repproto.New(OptionsVar.Socksserver, OptionsVar.Server)
	err = proto.PostSpecific(OptionsVar.Server, inData)
	if err != nil {
		log.Fatalf("Output failed: %s\n", err)
		log.Datas("STATUS (Result):\tFAIL\n")
	} else {
		log.Datas("STATUS (Result):\tDONE\n")
	}

	log.Sync()
	return 0
}

// CmdGet gets a file
func CmdGet() int {
	var err error
	// write to "-","",file
	// server must be specified
	// command line must be parseable
	args := flag.Args()
	if len(args) == 0 {
		log.Fatal("URL missing")
		return 1
	}
	server, messageID, _ := cmdlineURLparse(args...)
	if server == "" {
		// if server is missing, we need --server
		if OptionsVar.Server == "" {
			log.Fatal("Server must be specified: --server")
			return 1
		}
		server = OptionsVar.Server
	}
	if OptionsVar.Server != "" {
		// overwrite server if --server is given
		server = OptionsVar.Server
	}
	err = loadStoreMessage(server, messageID, OptionsVar.Outfile)
	if err != nil {
		return 1
	}
	log.Dataf("STATUS (Process):\tDONE\n")
	return 0
}

func loadStoreMessage(server string, messageID []byte, outfile string) error {
	proto := repproto.New(OptionsVar.Socksserver, server)
	log.Dataf("STATUS (Process):\tFETCH\n")
	inData, err := proto.GetSpecific(server, messageID)
	if err != nil {
		log.Dataf("STATUS (Process):\tFAIL\n")
		log.Fatalf("Fetch error: %s\n", err)
		return err
	}
	err = outputData(outfile, inData)
	if err != nil {
		log.Dataf("STATUS (Process):\tFAIL\n")
		log.Fatalf("Output failed: %s\n", err)
		return err
	}
	return nil
}
