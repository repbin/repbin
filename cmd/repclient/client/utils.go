package client

import (
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	log "github.com/repbin/repbin/deferconsole"
	"github.com/repbin/repbin/utils"
	"strconv"
	"strings"
)

// selectPrivKey returns the selected private key as string.
// privkeyOpt is the data given on the commandline
// privkeyConfig is the data from the config file, which must be a key
// privkeyDefault is the default to use when the commandline option was empty
// supported commandline options are
// ""  : (empty string) use default
// "-" : read from stdin,
// number : read from filedescriptor number
// "tty" : query user on tty
// key : return the key
func selectPrivKey(privkeyOpt, privkeyConfig, privkeyDefault string) string {
	readFd := -1
	if privkeyOpt == "" {
		privkeyOpt = privkeyDefault
	}
	if privkeyOpt == "config" {
		return privkeyConfig
	}
	if privkeyOpt == "tty" {
		pass, err := readPassTTY("Private key(s): ")
		if err != nil {
			return ""
		}
		return string(pass)
	}
	if privkeyOpt == "-" {
		readFd = 0
	}
	fdI, err := strconv.Atoi(privkeyOpt)
	if err == nil {
		readFd = fdI
	}
	if readFd >= 0 {
		// Read from file descriptor
		log.Sync()
		fd := os.NewFile(uintptr(readFd), "fd/"+strconv.Itoa(readFd))
		defer fd.Close()
		b := make([]byte, 120)
		log.Datas("STATUS (KeyMGT):\tENTER KEY\n")
		log.Sync()
		n, _ := fd.Read(b)
		log.Datas("STATUS (KeyMGT):\tREAD DONE\n")
		if n == 0 {
			return ""
		}
		return strings.Trim(string(b[:n]), " \t\r\n")
	}
	return privkeyOpt
}

// readPassTTY reads a password from tty/terminal
func readPassTTY(prompt string) ([]byte, error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	defer tty.Close()
	fd := tty.Fd()
	if terminal.IsTerminal(int(fd)) {
		tty.WriteString(prompt)
		pass, err := terminal.ReadPassword(int(fd))
		tty.WriteString("\n")
		return pass, err
	}
	return nil, fmt.Errorf("No terminal")
}

// isDir checks if dirname exists and is a directory
func isDir(dirname string) bool {
	stat, err := os.Lstat(dirname)
	if err != nil {
		return false
	}
	return stat.IsDir()
}

// inputData reads data from filename. Filename can be empty or "-" for stdin, other for file
// if filename can be converted to int (decimal), then it is treated as file descriptor that has been
// opened by a parent process.
func inputData(filename string, maxData int64) ([]byte, error) {
	if filename == "" || filename == "-" {
		return utils.MaxStdinRead(maxData)
	}
	fdI, err := strconv.Atoi(filename)
	if err == nil {
		fd := os.NewFile(uintptr(fdI), "fd/"+filename)
		defer fd.Close()
		return utils.MaxRead(maxData, fd)
	}
	return utils.MaxReadFile(maxData, filename)
}

// outputData writes data to file or stdout.
// if filename can be converted to int (decimal), then it is treated as file descriptor that has been
// opened by a parent process.
func outputData(filename string, data []byte) error {
	if filename == "-" || filename == "" {
		return utils.WriteStdout(data)
	}
	fdI, err := strconv.Atoi(filename)
	if err == nil {
		fd := os.NewFile(uintptr(fdI), "fd/"+filename)
		defer fd.Close()
		_, err := fd.Write(data)
		return err
	}
	return utils.WriteNewFile(filename, data)
}

// CmdShowConfig prints current config to stdout
func CmdShowConfig() int {
	x, _ := json.MarshalIndent(GlobalConfigVar, "", "    ")
	utils.WriteStdout(x)
	utils.WriteStdout([]byte("\n"))
	return 0
}

// LoadConfigFile loads configuration from a file
func LoadConfigFile(file string) (conf ConfigVariables, err error) {
	data, err := utils.MaxReadFile(409600, file)
	if err != nil {
		log.Errorf("Cannot load config from file: %s\n", file)
		return conf, err
	}
	if err := json.Unmarshal(data, &conf); err != nil {
		return conf, err
	}
	return conf, nil
}

// WriteConfigFile writes the user's config file
func WriteConfigFile(conf ConfigVariables) error {
	configFile := UserConfigFile()
	if configFile == "" {
		return ErrNoConfig
	}
	dir := path.Dir(configFile)
	os.MkdirAll(dir, 0700)
	data, err := json.MarshalIndent(conf, "", "    ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(configFile, data, 0600)
	return err
}

// UserConfigFile returns the path to the user's config file, or an empty string on error
func UserConfigFile() string {
	uInfo, err := user.Current()
	if err != nil {
		return ""
	}
	return path.Join(uInfo.HomeDir, ".config", "repclient", "repclient.config")
}

// cmdlineURLparse parses the commandline into server, messageID and key
func cmdlineURLparse(args ...string) (string, []byte, string) {
	// server,messageID,key = server/messageID_key | server/messageID | messageID_key | messageID key | messageID
	var server, messageID, key string
	if len(args) == 2 { // messageid key
		return "", utils.B58decode(args[0]), args[1]
	}
	fsplit := strings.SplitN(args[0], "_", 2)
	if len(fsplit) == 2 {
		// last is key
		key = fsplit[1]
	}
	ssep := strings.LastIndex(fsplit[0], "/")
	if ssep == -1 {
		// No server data contained
		messageID = fsplit[0]
	} else {
		messageID = fsplit[0][ssep+1:] // last part is messageID
		server = fsplit[0][:ssep]      // first part is server
	}
	return server, utils.B58decode(messageID), key
}
