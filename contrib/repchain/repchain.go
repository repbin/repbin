// repchain.go encodes repbin messages for repost chains.
// release 20150603
// you can reach me with
// repclient --recipientPubKey 7VW3oPLzQc7VS2anLyDtrdARDdSwa7QTF7h3N2t6J2VN_3xzmKrPZ1yDLWiWEsXLrA6vabESVLY2QEkscqDKU5U5f
// don't forget to put your own key into your message!
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strconv"
	"strings"
)

type config struct {
	STMKey []string
}

func genConfig() *config {
	var cfg config
	cfg.STMKey = append(cfg.STMKey, "REPLACE")
	return &cfg
}

func loadConfig(configFile, configDir string) (*config, error) {
	if _, err := os.Stat(configFile); err != nil {
		return nil, fmt.Errorf("config file '%s' doesn't exist, generate with:\n"+
			"mkdir -p %s && repchain --genconfig > %s", configFile, configDir, configFile)
	}
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	var cfg config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	// check STMKey entry
	if len(cfg.STMKey) == 0 || (len(cfg.STMKey) >= 1 && cfg.STMKey[0] == "REPLACE") {
		return nil, fmt.Errorf("specify STMKey(s) in config file '%s'", configFile)
	}
	return &cfg, nil
}

func (cfg *config) show() error {
	jsn, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsn))
	return nil
}

func readInput(in string) ([]byte, error) {
	var fpin *os.File
	var err error
	if in != "" {
		fpin, err = os.Open(in)
		if err != nil {
			return nil, err
		}
	} else {
		fpin = os.Stdin
	}
	input, err := ioutil.ReadAll(fpin)
	if err != nil {
		return nil, err
	}
	return input, nil
}

func inputMessage(recipient string, minDelay, maxDelay int, input []byte) (string, []byte, error) {
	args := []string{"--repost",
		"--appdata",
		"--minDelay", strconv.Itoa(minDelay),
		"--maxDelay", strconv.Itoa(maxDelay),
	}
	if recipient != "" {
		args = append(args, "--recipientPubKey")
		args = append(args, recipient)
	}
	cmd := exec.Command("repclient", args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", nil, err
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return "", nil, err
	}
	if _, err := stdin.Write(input); err != nil {
		return "", nil, err
	}
	if err := stdin.Close(); err != nil {
		return "", nil, err
	}
	if err := cmd.Wait(); err != nil {
		return "", nil, err
	}
	// process output
	var message string
	lines := strings.Split(stderr.String(), "\n")
	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) >= 2 && parts[0] == "STATUS (Message):" {
			message = parts[1]
			break
		}
	}
	return message, stdout.Bytes(), nil
}

func chainMessage(stmKey string, minDelay, maxDelay int, input []byte) ([]byte, error) {
	cmd := exec.Command("repclient",
		"--repost",
		"--minDelay", strconv.Itoa(minDelay),
		"--maxDelay", strconv.Itoa(maxDelay),
		"--messageType=3",
		"--recipientPubKey", stmKey)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	if _, err := stdin.Write(input); err != nil {
		return nil, err
	}
	if err := stdin.Close(); err != nil {
		return nil, err
	}
	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func outputMessage(stmKey string, input []byte) error {
	cmd := exec.Command("repclient",
		"--encrypt",
		"--messageType=3",
		"--recipientPubKey", stmKey)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	if _, err := stdin.Write(input); err != nil {
		return err
	}
	if err := stdin.Close(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func chain(recipient string, stmKeys []string, minDelay, maxDelay int, in string) error {
	input, err := readInput(in)
	if err != nil {
		return err
	}
	message, output, err := inputMessage(recipient, minDelay, maxDelay, input)
	if err != nil {
		return err
	}
	for i := 0; i+1 < len(stmKeys); i++ {
		output, err = chainMessage(stmKeys[i], minDelay, maxDelay, output)
	}
	if err := outputMessage(stmKeys[len(stmKeys)-1], output); err != nil {
		return err
	}
	fmt.Println(message)
	return nil
}

func appDataDir(appName string) (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}
	return path.Join(user.HomeDir, ".config", appName), nil
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "%s: error: %s\n", os.Args[0], err)
	os.Exit(1)
}

func main() {
	configDir, err := appDataDir("repchain")
	if err != nil {
		fatal(err)
	}
	// define options
	configFile := flag.String("configfile", path.Join(configDir, "repchain.config"), "Path to configuration file")
	in := flag.String("in", "", "Read data from file")
	minDelay := flag.Int("minDelay", 0, "Minimum repost delay")
	maxDelay := flag.Int("maxDelay", 0, "Maximum repost delay")
	genconfig := flag.Bool("genconfig", false, "Generate config file")
	recipient := flag.String("recipientPubKey", "", "Send to key")
	// parse options
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [STMKey ...]\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "If no STMKey(s) are defined on the command line the config file is loaded.\n")
	}
	flag.Parse()
	if *genconfig {
		cfg := genConfig()
		if err := cfg.show(); err != nil {
			fatal(err)
		}
		return
	}
	var stmKeys []string
	if flag.NArg() == 0 {
		// load config file
		cfg, err := loadConfig(*configFile, configDir)
		if err != nil {
			fatal(err)
		}
		stmKeys = cfg.STMKey
	} else {
		stmKeys = flag.Args()
	}
	if err := chain(*recipient, stmKeys, *minDelay, *maxDelay, *in); err != nil {
		fatal(err)
	}
}
