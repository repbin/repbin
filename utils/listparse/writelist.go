package listparse

import (
	"bytes"
	"github.com/repbin/repbin/utils/repproto/structs"
	"errors"
	"io"
)

var (
	// ErrInMessage is returned if the message contained an error string
	ErrInMessage = errors.New("listparse: Error returned")
	// ErrFormat is returned if the format is corrupt
	ErrFormat = errors.New("listparse: Format malformed")
	// ErrNoEntries is returned if no entries are found in the list
	ErrNoEntries = errors.New("listparse: No entries in list")
	// ErrSomeErrors is returned if some entries were corrupts but we were able to read at least some
	ErrSomeErrors = errors.New("listparse: List contains some errors, recovered")
)

// WriteMessageList formats a message list into a list of newline separated lines with space separated values
func WriteMessageList(messages []*structs.MessageStruct, w io.Writer) {
	for _, msg := range messages {
		line := []byte("IDX: ")
		line = append(line, msg.Encode()...)
		w.Write(line)
		w.Write([]byte("\n"))
	}
}

func lineToMsgStruct(d []byte) *structs.MessageStruct {
	return structs.MessageStructDecode(d)
}

// ReadMessageList reads a message list from reader and outputs a list of MessageStructs
// Callers must verify that messages is nil or that err!=ErrSomeErrors
func ReadMessageList(d []byte) (messages []*structs.MessageStruct, lastline []byte, err error) {
	var countErrors, countLines int
	idxm := []byte("IDX:")
	sucm := []byte("SUCC")
	errm := []byte("ERRO")
	cmdm := []byte("CMD:")
	lines := bytes.Split(d, []byte("\n"))
	ret := make([]*structs.MessageStruct, 0, len(lines))
parseLoop:
	for _, l := range lines {
		l = bytes.Trim(l, " \t\r\n")
		if len(l) == 0 {
			continue
		}
		lastline = l
		if bytes.Equal(l[0:4], idxm) {
			s := lineToMsgStruct(l[5:])
			if s != nil {
				ret = append(ret, s)
				countLines++
			} else {
				countErrors++
			}
		} else if bytes.Equal(l[0:4], cmdm) {
			// end of list
			break parseLoop
		} else if bytes.Equal(l[0:4], errm) {
			// Error, return
			return nil, l, ErrInMessage
		} else if !bytes.Equal(l[0:4], sucm) {
			return nil, l, ErrFormat
		}
	}
	if countLines == 0 {
		return nil, lastline, ErrNoEntries
	}
	if countErrors > 0 {
		return ret, lastline, ErrSomeErrors
	}
	return ret, lastline, nil
}
