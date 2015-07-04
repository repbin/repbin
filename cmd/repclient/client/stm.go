package client

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	log "github.com/repbin/repbin/deferconsole"
	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils"
	"github.com/repbin/repbin/utils/repproto"
)

// CmdSTM does an STM run to specific server and from specific stmdir
func CmdSTM() int {
	if OptionsVar.Server == "" {
		log.Fatal("Server must be specified: --server\n")
		return 1
	}
	if OptionsVar.Stmdir == "" {
		log.Fatal("Missing parameter: --stmdir\n")
		return 1
	}
	if !isDir(OptionsVar.Stmdir) {
		log.Fatalf("stmdir does not exist or is no directory: %s\n", OptionsVar.Stmdir)
		return 1
	}
	err := listSTM(OptionsVar.Stmdir) // Todo: Test if dir exists
	if err != nil {
		log.Errorf("STM Process errors: %s\n", err)
		return 1
	}
	return 0
}

func listSTM(dirname string) error {
	dirname = dirname + string(os.PathSeparator) //dirty
	now := strconv.Itoa(int(time.Now().Unix()))
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return err
	}
	sendFiles := make([]string, 0, 100)
	for _, x := range files {
		if !x.IsDir() {
			file := x.Name()
			if file <= now {
				log.Dataf("STATUS (STMSend):\t%s\n", file)
				sendFiles = append(sendFiles, dirname+file)
			} else {
				log.Dataf("STATUS (STMKeep):\t%s\n", file)
			}
		}
	}
	if len(sendFiles) == 0 {
		return fmt.Errorf("No files")
	}
	return sendSTM(sendFiles)
}

func sendSTM(files []string) error {
	var errCount int
	maxInData := int64(GlobalConfigVar.BodyLength-(message.Curve25519KeySize*2)) * 5
	files = utils.PermString(files)
	for _, file := range files {
		inData, err := utils.MaxReadFile(maxInData, file)
		if err != nil {
			errCount++
			log.Dataf("STATUS (STMErr):\t%s\t%s\n", file, err)
			continue
		}
		remove := false
		log.Dataf("STATUS (STMTrans):\t%s\n", file)
		proto := repproto.New(OptionsVar.Socksserver, OptionsVar.Server)
		err = proto.PostSpecific(OptionsVar.Server, inData)
		if err != nil {
			log.Dataf("STATUS (STMRes):\t%s\tFAIL\t%s\n", file, err)
			if err.Error() == "Server error: db: Duplicate" {
				remove = true
			} else if err.Error() == "Server error: Message too small" {
				remove = true
			} else {
				errCount++
				continue
			}
		} else {
			log.Dataf("STATUS (STMRes):\t%s\tDONE\t\n", file)
			remove = true
		}
		if remove {
			err := os.Remove(file)
			if err != nil {
				errCount++
				log.Dataf("STATUS (STMCleanErr):\t%s\n", err)
			}
		}
	}
	if errCount > 0 {
		return fmt.Errorf("Errors: %d", errCount)
	}
	return nil
}
