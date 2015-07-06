// repmbox handles repbin mailboxes.
//
// Release 20150606.
// You can reach me with
// repclient --recipientPubKey 7VW3oPLzQc7VS2anLyDtrdARDdSwa7QTF7h3N2t6J2VN_DTshmJFEDa7XM2w9nHBq4CgtvK4kYdBp8G3wFPkYcGd1
// Don't forget to put your own key into your message! (create with repmbox -add)
package main

import (
	"bufio"
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

type sender struct {
	Sender     string
	PublicKey  string
	PrivateKey string
}

type config struct {
	PrivateKey string
	Server     string
	Start      int
	Count      int
	Sender     []sender
}

type senderKey struct {
	user    string
	privKey string
}

func genConfig() (*config, error) {
	// generate private key with repclient
	cmd := exec.Command("repclient", "--genkey", "--appdata")
	var out bytes.Buffer
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	var cfg config
	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) >= 2 && parts[0] == "STATUS (PrivateKey):" {
			cfg.PrivateKey = parts[1]
			break
		}
	}
	if cfg.PrivateKey == "" {
		return nil, fmt.Errorf("could not parse private key from output:\n%s", out.String())
	}
	cfg.Start = -1
	cfg.Count = 100
	return &cfg, nil
}

func loadConfig(configFile, configDir string) (*config, error) {
	if _, err := os.Stat(configFile); err != nil {
		return nil, fmt.Errorf("config file '%s' doesn't exist, generate with:\n"+
			"mkdir -p %s && repmbox --genconfig > %s", configFile, configDir, configFile)
	}
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	var cfg config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	// check server entry
	if cfg.Server == "" {
		return nil, fmt.Errorf("specify Server entry in config file '%s'", configFile)
	}
	return &cfg, nil
}

func (cfg *config) save(configFile string) error {
	jsn, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(configFile, jsn, 0600); err != nil {
		return err
	}
	return nil
}

func (cfg *config) show() error {
	jsn, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsn))
	return nil
}

func (cfg *config) addSender(addSender, configFile string) error {
	var snd sender
	snd.Sender = addSender
	// generate key pair for sender
	cmd := exec.Command("repclient", "--gentemp",
		"--privkey", cfg.PrivateKey,
		"--appdata")
	var out bytes.Buffer
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return err
	}
	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) >= 2 && parts[0] == "STATUS (PrivateKey):" {
			snd.PrivateKey = parts[1]
		}
		if len(parts) >= 2 && parts[0] == "STATUS (PublicKey):" {
			snd.PublicKey = parts[1]
			break
		}
	}
	if snd.PrivateKey == "" {
		return fmt.Errorf("could not parse private key from output:\n%s", out.String())
	}
	if snd.PublicKey == "" {
		return fmt.Errorf("could not parse public key from output:\n%s", out.String())
	}
	cfg.Sender = append(cfg.Sender, snd)
	if err := cfg.save(configFile); err != nil {
		return err
	}
	fmt.Printf("mailbox for %s added:\n", addSender)
	fmt.Println(snd.PublicKey)
	return nil
}

func (cfg *config) getList() ([]string, bool, error) {
	cmd := exec.Command("repclient",
		"--index",
		"--server", cfg.Server,
		"--privkey", cfg.PrivateKey,
		"--start", strconv.Itoa(cfg.Start+1),
		"--count", strconv.Itoa(cfg.Count),
		"--appdata")
	var out bytes.Buffer
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		// check if error indicates no new messages
		lines := strings.Split(out.String(), "\n")
		for _, line := range lines {
			if line == "FATAL: List error: listparse: No entries in list" {
				// return empty list
				return nil, false, nil
			}
		}
		// show and return error
		fmt.Print(out.String())
		return nil, false, err
	}
	var list []string
	var more bool
	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) >= 2 {
			if parts[0] == "STATUS (MessageList):" {
				particles := strings.Split(parts[1], " ")
				if len(particles) >= 2 {
					list = append(list, particles[1])
				} else {
					return nil, false, fmt.Errorf("could not parse line: %s", line)
				}
			} else if parts[0] == "STATUS (ListResult):" {
				particles := strings.Split(parts[1], " ")
				if len(particles) >= 3 {
					var err error
					more, err = strconv.ParseBool(particles[2])
					if err != nil {
						return nil, false, err
					}
					break
				}
			}
		}
	}
	return list, more, nil
}

func (cfg *config) getSenderKeys() map[string]senderKey {
	keys := make(map[string]senderKey)
	for _, sdr := range cfg.Sender {
		pubKeys := strings.Split(sdr.PublicKey, "_")
		privKeys := strings.Split(sdr.PrivateKey, "_")
		keys[pubKeys[0]] = senderKey{
			user:    sdr.Sender,
			privKey: privKeys[0],
		}
		keys[pubKeys[1]] = senderKey{
			user:    sdr.Sender,
			privKey: privKeys[1],
		}
	}
	return keys
}

func (cfg *config) getMessages(list []string, outdir, stmdir string, verbose bool) error {
	// make sure STM directory exists
	if err := os.MkdirAll(stmdir, 0700); err != nil {
		return err
	}
	for _, msg := range list {
		fmt.Printf("retrieving message %s\n", msg)
		cmd := exec.Command("repclient",
			"--decrypt",
			"--keymgt", "3",
			"--server", cfg.Server,
			"--stmdir", stmdir,
			"--appdata",
			msg)
		var out bytes.Buffer
		cmd.Stdout = &out
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return err
		}
		r, w, err := os.Pipe()
		if err != nil {
			return err
		}
		cmd.ExtraFiles = append(cmd.ExtraFiles, r)
		if err := cmd.Start(); err != nil {
			return err
		}
		var user string
		var stmfile string
		var keyNotAvailable bool
		keys := cfg.getSenderKeys()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			if verbose {
				fmt.Println(line)
			}
			parts := strings.Split(line, "\t")
			if len(parts) >= 2 {
				if parts[0] == "STATUS (KeyMGTRequest):" {
					key, ok := keys[parts[1]]
					if !ok {
						if _, err := w.WriteString("not available\n"); err != nil {
							return err
						}
						keyNotAvailable = true
						break
					} else {
						if _, err := w.WriteString(key.privKey + "\n"); err != nil {
							return err
						}
						user = key.user
					}
				} else if parts[0] == "STATUS (STMFile):" {
					stmfile = parts[1]
				}
			} else {
				return fmt.Errorf("could not parse repclient output: %s", line)
			}
		}
		if err := scanner.Err(); err != nil {
			return err
		}
		err = cmd.Wait()
		if !keyNotAvailable && err != nil {
			return fmt.Errorf("repclient call failed with: %s", err)
		}
		if stmfile != "" {
			fmt.Printf("new repost message from %s written to:\n%s\n", user, stmfile)
		} else if keyNotAvailable {
			fmt.Println("key not available, skipping message forever!")
		} else {
			// make sure output directory exists
			resdir := path.Join(outdir, user)
			if err := os.MkdirAll(resdir, 0700); err != nil {
				return err
			}
			// write message
			filename := path.Join(resdir, msg)
			if err := ioutil.WriteFile(path.Join(resdir, msg), out.Bytes(), 0600); err != nil {
				return err
			}
			fmt.Printf("new message from %s written to:\n%s\n", user, filename)
		}
	}
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
	configDir, err := appDataDir("repmbox")
	if err != nil {
		fatal(err)
	}
	// define options
	addSender := flag.String("add", "", "Add new mailbox for given sender name")
	configFile := flag.String("configfile", path.Join(configDir, "repmbox.config"), "Path to configuration file")
	genconfig := flag.Bool("genconfig", false, "Generate config file")
	listSender := flag.Bool("list", false, "List mailboxes")
	outdir := flag.String("outdir", path.Join(configDir, "messages"), "Directory for downloaded messages")
	stmdir := flag.String("stmdir", path.Join(configDir, "stmdir"), "Directory for STM messages")
	verbose := flag.Bool("v", false, "Be verbose")
	// parse options
	flag.Parse()
	if *genconfig {
		cfg, err := genConfig()
		if err != nil {
			fatal(err)
		}
		if err := cfg.show(); err != nil {
			fatal(err)
		}
		return
	}
	// load config file
	cfg, err := loadConfig(*configFile, configDir)
	if err != nil {
		fatal(err)
	}
	if *addSender != "" {
		if err := cfg.addSender(*addSender, *configFile); err != nil {
			fatal(err)
		}
	} else if *listSender {
		for _, snd := range cfg.Sender {
			fmt.Println(snd.Sender, snd.PublicKey)
		}
	} else {
		more := true
		for more {
			var list []string
			// download new messages
			list, more, err = cfg.getList()
			if err != nil {
				fatal(err)
			}
			if len(list) == 0 {
				fmt.Println("no new messages")
				return
			}
			// get new messages
			if err := cfg.getMessages(list, *outdir, *stmdir, *verbose); err != nil {
				fatal(err)
			}
			// increase start
			cfg.Start += len(list)
			if err := cfg.save(*configFile); err != nil {
				fatal(err)
			}
			if *verbose {
				fmt.Printf("more: %s\n", strconv.FormatBool(more))
			}
		}
	}
}
