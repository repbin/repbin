// reptoken is the repbin tool to generate hashcash tokens that are required for posting to repservers.
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/agl/ed25519"
	"github.com/repbin/repbin/hashcash"
	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils"
)

var minbits = flag.Int("minBits", 24, "Generate token of minBits value")
var outDir = flag.String("outDir", "", "Write tokens to directory. Generate many tokens.")
var outFile = flag.String("outFile", "", "Write one token to outfile.")
var contCalc = flag.Bool("continue", false, "Continue producing more bits for one token.")

func main() {
	flag.Parse()
	if *contCalc && *outFile != "" {
		// Read token
		d, err := utils.MaxReadFile(2048, *outFile)
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}
		kpt := new(message.SignKeyPair)
		kp, err := kpt.Unmarshal(d)
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}
		fmt.Print("Continue")
		start := hashcash.NonceToUInt64(kp.Nonce[:])
		_, startBits := hashcash.TestNonce(kp.PublicKey[:], kp.Nonce[:], 0)
		if len(d) == (ed25519.PublicKeySize + ed25519.PrivateKeySize + hashcash.NonceSize + 1 + 8) {
			fmt.Print(" from state ")
			start = hashcash.NonceToUInt64(d[ed25519.PublicKeySize+ed25519.PrivateKeySize+hashcash.NonceSize+1 : ed25519.PublicKeySize+ed25519.PrivateKeySize+hashcash.NonceSize+1+8])
		}
		fmt.Printf(" (%d, bits: %d) ", start, startBits)
		for {
			startBits++
			nonce, _ := hashcash.ComputeNonceSelect(kp.PublicKey[:], startBits, start, start+hashcash.Steps*8)
			start = hashcash.NonceToUInt64(nonce[:])
			_, nextBits := hashcash.TestNonce(kp.PublicKey[:], nonce, 0)
			if nextBits > startBits {
				fmt.Printf("(%d)", nextBits)
				copy(kp.Nonce[:], nonce)
				kp.Bits = nextBits
				startBits = nextBits
			}
			fmt.Print(".")
			newData := append(kp.Marshal(), nonce[:]...)
			err := utils.OverWriteFile(*outFile, newData)
			if err != nil {
				fmt.Printf("%s\n", err)
				os.Exit(1)
			}
			if nextBits > byte(*minbits) {
				fmt.Printf("\n%d reached. Finish.\n", nextBits)
				os.Exit(0)
			}
		}
	} else if *outDir == "" {
		keypair, err := message.GenKey(byte(*minbits))
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}
		b := keypair.Marshal()
		if *outFile != "" {
			err := utils.WriteNewFile(*outFile, b)
			if err != nil {
				fmt.Printf("%s\n", err)
				os.Exit(1)
			}
		} else {
			os.Stdout.Write(b)
			os.Stdout.Sync()
			os.Exit(0)
		}
	} else {
		for {
			keypair, err := message.GenKey(byte(*minbits))
			if err != nil {
				fmt.Printf("%s\n", err)
				os.Exit(1)
			}
			b := keypair.Marshal()
			err = utils.WriteNewFile(*outDir+string(os.PathSeparator)+strconv.Itoa(int(time.Now().Unix()))+strconv.Itoa(int(time.Now().Nanosecond()))+".hashcash", b)
			if err != nil {
				fmt.Printf("%s\n", err)
				os.Exit(1)
			}
		}
	}
}
