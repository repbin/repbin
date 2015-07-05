// Package deferconsole writes data with level filters and defer to stdout/stderr.
package deferconsole

import (
	"fmt"
	"os"
)

// Message contains a message to write
type Message struct {
	Level    int
	Content  string
	Dest     int
	Minlevel int
	Sync     chan bool
}

const (
	// DstStdErr write to stderr
	DstStdErr = iota
	// DstStdOut write to stdout
	DstStdOut
	// dstnil drop
	dstnil
)

const (
	// LevelKeep don't change level
	LevelKeep = -3
	// LevelSilence supress all output
	LevelSilence = -2
	// LevelRestore disable silence
	LevelRestore = -1
	// LevelSecret prints secrets
	LevelSecret = 0
	// LevelTrace prints traces
	LevelTrace = 1
	// LevelDebug for debugging information
	LevelDebug = 2
	// LevelInfo for general information
	LevelInfo = 3
	// LevelError to print errors
	LevelError = 4
	// LevelData is for data
	LevelData = 5
	// LevelPrint is for normal output
	LevelPrint = 6
	// LevelFatal for fatal erros
	LevelFatal = 7
)

var consoleChannel chan Message
var defaultLevel = LevelError

// Replace the message channel to serve your own writer
func Replace(c chan Message) {
	close(consoleChannel)
	consoleChannel = c
}

func writer(c <-chan Message) {
	minlevel := defaultLevel
	minlevelPrev := defaultLevel
	for msg := range c {
		if msg.Minlevel != LevelKeep {
			if msg.Minlevel == LevelSilence {
				minlevelPrev = minlevel
				minlevel = msg.Minlevel
				defaultLevel = msg.Minlevel
			} else if msg.Minlevel == LevelRestore {
				minlevel = minlevelPrev
				defaultLevel = minlevelPrev
			} else {
				minlevel = msg.Minlevel
				defaultLevel = msg.Minlevel
			}
		}
		if (minlevel != LevelSilence && msg.Level >= minlevel) && len(msg.Content) > 0 {
			if msg.Dest == DstStdErr {
				fmt.Fprint(os.Stderr, msg.Content)
			} else if msg.Dest == DstStdOut {
				fmt.Fprint(os.Stdout, msg.Content)
			}
		}
		if msg.Sync != nil {
			msg.Sync <- true
		}
	}
}

// Sync output (block until message(s) are processed)
func Sync() {
	c := make(chan bool)
	consoleChannel <- Message{
		Minlevel: LevelKeep,
		Sync:     c,
	}
	<-c
}

// doWrite  writes to channel
func doWrite(level int, str string, dst int, minlevel int) {
	consoleChannel <- Message{
		Level:    level,
		Content:  str,
		Dest:     dst,
		Minlevel: minlevel,
	}
}

// Silence all output
func Silence() {
	SetMinLevel(LevelSilence)
}

// Voice output
func Voice() {
	SetMinLevel(LevelRestore)
}

// SetMinLevel to level
func SetMinLevel(level int) {
	doWrite(0, "", dstnil, level)
}

type directFunc func(string)
type stringFunc func(...interface{})
type formatFunc func(string, ...interface{})

func genDirectFunc(dest int, level int) directFunc {
	return func(str string) {
		doWrite(level, str, dest, LevelKeep)
	}
}

func genStringFunc(dest int, level int) stringFunc {
	return func(a ...interface{}) {
		doWrite(level, fmt.Sprintln(a...), dest, LevelKeep)
	}
}

func genFormatFunc(dest int, level int) formatFunc {
	return func(format string, a ...interface{}) {
		doWrite(level, fmt.Sprintf(format, a...), dest, LevelKeep)
	}
}

func genWrapperDirectFunc(level int, dest int, prestr string) directFunc {
	return func(str string) {
		if defaultLevel > level || defaultLevel == LevelSilence {
			return
		}
		doWrite(level, prestr+str, dest, LevelKeep)
	}
}

func genWrapperStringFunc(level int, dest int, prestr string) stringFunc {
	return func(a ...interface{}) {
		if defaultLevel > level || defaultLevel == LevelSilence {
			return
		}
		doWrite(level, prestr+fmt.Sprintln(a...), dest, LevelKeep)
	}
}

func genWrapperFormatFunc(level int, dest int, prestr string) formatFunc {
	return func(format string, a ...interface{}) {
		if defaultLevel > level || defaultLevel == LevelSilence {
			return
		}
		doWrite(level, prestr+fmt.Sprintf(format, a...), dest, LevelKeep)
	}
}

var (
	// ToErrors  writes string to stderr
	ToErrors directFunc
	// ToError writes elements to error
	ToError stringFunc
	// ToErrorf writes formatted elements to error
	ToErrorf formatFunc
	// Prints writes string to stdout
	Prints directFunc
	// Printf writes formatted elements to stdout
	Printf formatFunc
	// Print writes elements to stdout
	Print stringFunc
	// Infos is info
	Infos directFunc
	// Info is info
	Info stringFunc
	// Infof is info
	Infof formatFunc
	// Debugs is debug
	Debugs directFunc
	// Debug is debug
	Debug stringFunc
	// Debugf is debug
	Debugf formatFunc
	// Errors is error
	Errors directFunc
	// Error is error
	Error stringFunc
	// Errorf is error
	Errorf formatFunc
	// Fatals is fatal
	Fatals directFunc
	// Fatal is fatal
	Fatal stringFunc
	// Fatalf is fatal
	Fatalf formatFunc
	// Traces is trace
	Traces directFunc
	// Trace is trace
	Trace stringFunc
	// Tracef is trace
	Tracef formatFunc
	// Secrets is secret
	Secrets directFunc
	// Secret is secret
	Secret stringFunc
	// Secretf is secret
	Secretf formatFunc
	// Datas is data
	Datas directFunc
	// Data is data
	Data stringFunc
	// Dataf is data
	Dataf formatFunc
)

func init() {
	consoleChannel = make(chan Message, 100)

	ToErrors = genDirectFunc(DstStdErr, LevelPrint)
	ToError = genStringFunc(DstStdErr, LevelPrint)
	ToErrorf = genFormatFunc(DstStdErr, LevelPrint)

	Prints = genDirectFunc(DstStdOut, LevelPrint)
	Print = genStringFunc(DstStdOut, LevelPrint)
	Printf = genFormatFunc(DstStdOut, LevelPrint)

	Datas = genDirectFunc(DstStdErr, LevelData)
	Data = genStringFunc(DstStdErr, LevelData)
	Dataf = genFormatFunc(DstStdErr, LevelData)

	Infos = genWrapperDirectFunc(LevelInfo, DstStdErr, "INFO: ")
	Info = genWrapperStringFunc(LevelInfo, DstStdErr, "INFO: ")
	Infof = genWrapperFormatFunc(LevelInfo, DstStdErr, "INFO: ")
	Debugs = genWrapperDirectFunc(LevelDebug, DstStdErr, "DEBUG: ")
	Debug = genWrapperStringFunc(LevelDebug, DstStdErr, "DEBUG: ")
	Debugf = genWrapperFormatFunc(LevelDebug, DstStdErr, "DEBUG: ")
	Errors = genWrapperDirectFunc(LevelError, DstStdErr, "ERROR: ")
	Error = genWrapperStringFunc(LevelError, DstStdErr, "ERROR: ")
	Errorf = genWrapperFormatFunc(LevelError, DstStdErr, "ERROR: ")
	Fatals = genWrapperDirectFunc(LevelFatal, DstStdErr, "FATAL: ")
	Fatal = genWrapperStringFunc(LevelFatal, DstStdErr, "FATAL: ")
	Fatalf = genWrapperFormatFunc(LevelFatal, DstStdErr, "FATAL: ")
	Traces = genWrapperDirectFunc(LevelTrace, DstStdErr, "TRACE:  ")
	Trace = genWrapperStringFunc(LevelTrace, DstStdErr, "TRACE:  ")
	Tracef = genWrapperFormatFunc(LevelTrace, DstStdErr, "TRACE:  ")
	Secrets = genWrapperDirectFunc(LevelSecret, DstStdErr, "SECRET:  ")
	Secret = genWrapperStringFunc(LevelSecret, DstStdErr, "SECRET:  ")
	Secretf = genWrapperFormatFunc(LevelSecret, DstStdErr, "SECRET:  ")

	go writer(consoleChannel)
}
