package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"math/big"
	"github.com/repbin/repbin/message"
	"strings"
	"time"
)

var (
	// ErrNoList is returned if list verification failed
	ErrNoList = errors.New("utils: No list")
)

// RepostHeaderSize is the length of the repost header containg repad key, min-delay and max-delay setting
const RepostHeaderSize = message.PadKeySize + 4 + 4

// RunPeriod is the period in which the STM delivery is started
var RunPeriod = int64(300) // 5min

// EncodeEmbedded converts two public keys into a byte slice
func EncodeEmbedded(k1, k2 *message.Curve25519Key) (ret []byte) {
	ret = make([]byte, message.Curve25519KeySize*2)
	if k1 != nil && k2 != nil {
		copy(ret[:message.Curve25519KeySize], k1[:])
		copy(ret[message.Curve25519KeySize:message.Curve25519KeySize*2], k2[:])
	}
	return ret
}

// DecodeEmbedded decodes an embedded reply public key pair
func DecodeEmbedded(d []byte) (k1, k2 *message.Curve25519Key) {
	if len(d) < message.Curve25519KeySize*2 {
		return nil, nil
	}
	k1, k2 = new(message.Curve25519Key), new(message.Curve25519Key)
	copy(k1[:], d[0:message.Curve25519KeySize])
	copy(k2[:], d[message.Curve25519KeySize:message.Curve25519KeySize*2])
	return k1, k2
}

// ParseKeyPair parses a public key pair as given on commandline
func ParseKeyPair(str string) (k1, k2 *message.Curve25519Key) {
	if str == "" || len(str) < 32 {
		return
	}
	split := strings.SplitN(str, "_", 2)
	if len(split) == 0 {
		return
	}
	if len(split) >= 1 {
		k1 = new(message.Curve25519Key)
		copy(k1[:], B58decode(split[0]))
	}
	if len(split) == 2 {
		k2 = new(message.Curve25519Key)
		copy(k2[:], B58decode(split[1]))
	}
	return k1, k2
}

// VerifyListContent verifies that data is in list format
func VerifyListContent(d []byte) error {
	lines := bytes.Split(d, []byte("\n"))
	if len(lines) > 0 {
		for _, l := range lines {
			// Entries consist of three fields: server address, messageID, key. server and key can be NULL
			parts := bytes.Split(l, []byte(" "))
			if len(parts) != 3 {
				return ErrNoList
			}
			messageID := B58decode(string(parts[1]))
			if len(messageID) != message.MessageIDSize {
				return ErrNoList
			}
			key := string(parts[2])
			if key != "NULL" {
				keyDec := B58decode(key)
				if len(keyDec) != message.Curve25519KeySize {
					return ErrNoList
				}
			}
		}
	}
	return nil
}

// EncodeRepostHeader encodes the padkey, mindelay and maxdelay into a byte array
func EncodeRepostHeader(padkey *[message.PadKeySize]byte, minDelay uint32, maxDelay uint32) (ret [RepostHeaderSize]byte) {
	copy(ret[:message.PadKeySize], padkey[:])
	binary.BigEndian.PutUint32(ret[message.PadKeySize:message.PadKeySize+4], minDelay)
	binary.BigEndian.PutUint32(ret[message.PadKeySize+4:message.PadKeySize+8], maxDelay)
	return
}

// DecodeRepostHeader decoes a repost header
func DecodeRepostHeader(d []byte) (padkey *[message.PadKeySize]byte, minDelay uint32, maxDelay uint32) {
	padkey = new([message.PadKeySize]byte)
	copy(padkey[:], d[:message.PadKeySize])
	minDelay = binary.BigEndian.Uint32(d[message.PadKeySize : message.PadKeySize+4])
	maxDelay = binary.BigEndian.Uint32(d[message.PadKeySize+4 : message.PadKeySize+8])
	return padkey, minDelay, maxDelay
}

// STM calculates the STM time
func STM(minDelay, maxDelay int) int64 {
	if minDelay < 0 {
		minDelay = 0
	}
	if maxDelay <= 0 {
		maxDelay = 1
	}
	if maxDelay < minDelay {
		t := maxDelay
		maxDelay = minDelay
		minDelay = t
	}
	diff := maxDelay - minDelay
	p, err := rand.Int(rand.Reader, big.NewInt(int64(diff)))
	if err != nil {
		return 3600
	}
	t := ((time.Now().Unix()+p.Int64())/RunPeriod)*RunPeriod + RunPeriod // Next 5min run
	return t
}
