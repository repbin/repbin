// repmulti.go sends and retrieves multi-part repbin messages.
// release 20150603
// you can reach me with
// repclient --recipientPubKey 7VW3oPLzQc7VS2anLyDtrdARDdSwa7QTF7h3N2t6J2VN_3kW1bXjj2WbaxBBRbtWCgjx1Qz9VfvDgyYnBFz5Ad9wL
// don't forget to put your own key into your message!
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

var (
	minSize = 65436
)

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
	if len(input) < minSize {
		return nil, fmt.Errorf("input too short (minimum size is %d bytes)", minSize)
	}
	return input, nil
}

func postInput(input []byte) ([]string, error) {
	var entries []string
	ctr := 1
	chunks := len(input) / minSize
	if len(input)%minSize > 0 {
		chunks++
	}
	for i := 0; i < len(input); i += minSize {
		cmd := exec.Command("repclient", "--appdata")
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return nil, err
		}
		var out bytes.Buffer
		cmd.Stderr = &out
		fmt.Fprintf(os.Stderr, "post chunk %d/%d\n", ctr, chunks)
		if err := cmd.Start(); err != nil {
			return nil, err
		}
		var end int
		if i+minSize > len(input) {
			end = len(input)
		} else {
			end = i + minSize
		}
		if _, err := stdin.Write(input[i:end]); err != nil {
			return nil, err
		}
		if err := stdin.Close(); err != nil {
			return nil, err
		}
		if err := cmd.Wait(); err != nil {
			return nil, err
		}
		// process output
		lines := strings.Split(out.String(), "\n")
		for _, line := range lines {
			parts := strings.Split(line, "\t")
			if len(parts) >= 2 && parts[0] == "STATUS (ListInput):" {
				entries = append(entries, parts[1])
				break
			}
		}
		ctr++
	}
	return entries, nil
}

func postList(entries []string) error {
	cmd := exec.Command("repclient", "--messageType=2")
	cmd.Stdin = strings.NewReader(strings.Join(entries, "\n"))
	cmd.Stdout = os.Stdout
	fmt.Fprintln(os.Stderr, "post list message")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("repclient call failed with: %s", err)
	}
	return nil
}

func encrypt(in string) error {
	input, err := readInput(in)
	if err != nil {
		return err
	}
	entries, err := postInput(input)
	if err != nil {
		return err
	}
	if err := postList(entries); err != nil {
		return err
	}
	return nil
}

func fetchChunk(buf *bytes.Buffer, server, privateKey string) error {
	var cmd *exec.Cmd
	if server != "" {
		cmd = exec.Command("repclient", "--server", server, privateKey)
	} else {
		cmd = exec.Command("repclient", privateKey)
	}
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("repclient call failed with: %s", err)
	}
	return nil
}

func writeOutput(outfile string, buf *bytes.Buffer) error {
	var fp *os.File
	var err error
	if outfile != "" {
		fp, err = os.Create(outfile)
		if err != nil {
			return err
		}
		defer fp.Close()
	} else {
		fp = os.Stdout
	}
	if _, err := fp.Write(buf.Bytes()); err != nil {
		return err
	}
	return nil
}

type chunk struct {
	server     string
	privateKey string
}

func decrypt(url, outfile string) error {
	cmd := exec.Command("repclient", "--appdata", url)
	var out bytes.Buffer
	cmd.Stderr = &out
	fmt.Fprintln(os.Stderr, "fetch list message")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("repclient call failed with: %s", err)
	}
	lines := strings.Split(out.String(), "\n")
	var buf bytes.Buffer
	var chunks []chunk
	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) >= 2 && parts[0] == "STATUS (ListItem):" {
			particles := strings.Split(parts[1], " ")
			if len(particles) != 3 {
				return fmt.Errorf("wrong ListItem format: %s", parts[1])
			}
			var srv string
			if particles[0] != "NULL" {
				srv = particles[0]
			}
			var pk string
			pk1 := particles[1]
			pk2 := particles[2]
			if pk2 != "NULL" {
				pk = pk1 + "_" + pk2
			} else {
				pk = pk1
			}
			chunks = append(chunks, chunk{
				server:     srv,
				privateKey: pk,
			})
		}
	}
	for i, chnk := range chunks {
		fmt.Fprintf(os.Stderr, "fetch chunk %d/%d\n", i+1, len(chunks))
		if err := fetchChunk(&buf, chnk.server, chnk.privateKey); err != nil {
			return err
		}
	}
	return writeOutput(outfile, &buf)
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "%s: error: %s\n", os.Args[0], err)
	os.Exit(1)
}

func main() {
	out := flag.String("out", "", "Write output data to file")
	in := flag.String("in", "", "Read data from file")
	flag.Parse()
	if flag.NArg() > 1 {
		fatal(fmt.Errorf("too many arguments"))
	} else if flag.NArg() > 0 {
		args := flag.Args()
		if *out != "" {
			if _, err := os.Stat(*out); err == nil {
				fatal(fmt.Errorf("output file exists: %s", *out))
			}
		}
		if err := decrypt(args[0], *out); err != nil {
			fatal(err)
		}
	} else {
		if err := encrypt(*in); err != nil {
			fatal(err)
		}
	}
}
