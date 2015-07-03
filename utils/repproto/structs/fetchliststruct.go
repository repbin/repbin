package structs

import (
	"bytes"
	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils"
	"strconv"
)

const (
	// FetchListStructSize is the maximum size of an encoded FetchListStruct
	FetchListStructSize = 66 // messageID 45 + uint 8 + fieldsep
	// FetchListStructMin is the minimum size of a fetchlist struct
	FetchListStructMin = 3
)

// FetchListStruct encodes data for the list of fetch jobs
type FetchListStruct struct {
	MessageID   [message.MessageIDSize]byte // The messageID
	TimeEntered uint64
}

// FetchListStructEncoded represents an encoded FetchList
type FetchListStructEncoded []byte

// Fill fills an encoded FetchList to maximum size
func (flse FetchListStructEncoded) Fill() FetchListStructEncoded {
	if len(flse) < FetchListStructSize {
		return append(flse, bytes.Repeat([]byte(" "), FetchListStructSize-len(flse))...)
	}
	return flse
}

// Encode a FetchListStruct to human readable format
func (fl FetchListStruct) Encode() FetchListStructEncoded {
	out := make([]byte, 0, FetchListStructSize)
	out = append(out, []byte(utils.B58encode(fl.MessageID[:])+" ")...)
	out = append(out, []byte(strconv.FormatUint(fl.TimeEntered, 10)+" ")...)
	return out[:len(out)]
}

// FetchListStructDecode decodes human readable to struct
func FetchListStructDecode(d FetchListStructEncoded) *FetchListStruct {
	if len(d) < FetchListStructMin {
		return nil
	}
	fields := bytes.Fields(d)
	if len(fields) != 2 {
		return nil
	}
	fs := new(FetchListStruct)
	cur := 0
	t := utils.B58decode(string(fields[cur]))
	copy(fs.MessageID[:], t)
	cur++
	fs.TimeEntered, _ = strconv.ParseUint(string(fields[cur]), 10, 64)
	return fs
}
