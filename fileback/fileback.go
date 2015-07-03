// Package fileback implements continuous, blob, round-robin and rolling lists.
// Lifted from DNCA. (c) 2015 DNCA. BSD 3 clause. thx!
package fileback

import (
	"crypto/rand"
	"errors"
)

// Version of this release
const Version = "0.0.1-repbinMod very alpha"

var (
	// ErrNoMore is returned if a range cannot be completed
	ErrNoMore = errors.New("fileback: No more entries")
	// ErrNoHeader is returned if a list has no headers
	ErrNoHeader = errors.New("fileback: No header")
	// ErrMaxEntries is returned if trying to add to a list that is filled
	ErrMaxEntries = errors.New("fileback: Cannot add more entries")
	// ErrListCorrupt is returned if the rolling list files are corrupt
	ErrListCorrupt = errors.New("fileback: List corrupt")
	// ErrOOB is returned when reading/writing a slice of a page that passes page borders
	ErrOOB = errors.New("fileback: access outside of page bounds")
	// ErrExists is returned if trying to operate on a non-existing index
	ErrExists = errors.New("fileback: index does not exist")
)

const (
	// MaxPage is the maximum page number
	MaxPage        = 4294967295
	lastFile       = -1
	appendFile     = -2
	listExtension  = ".list"
	lastExtension  = ".cur"
	indexExtension = ".idx"
)

const (
	listTypeBlob = iota
	listTypeContinue
	listTypeRoundRobin
	listTypeRolling
)

var secret = make([]byte, 4)

func init() {
	rand.Read(secret)
}

// Index is an index access to a list
type Index interface {
	Index(index []byte) ListIndex // Returns a new list with sub-index index
}

// ListIndex is a list implementation
type ListIndex interface {
	Append(data []byte) error                             // (go routine). Append to end of list
	Change(pos int64, data []byte) error                  // (overwrite at position, go routine)
	ChangePart(pos int64, data []byte, start int64) error // change entry at pos starting at byte start. Does not construct page
	Create(data []byte) error                             // create with one entry, if no entry exists yet. data can be nil
	CreateAppend(data []byte) error                       // create list, append if exists
	Delete()                                              // (not mediated,go routine))
	GetLast() []byte                                      // (not mediated). Returns nil if nothing/error
	ReadEntry(pos int64) ([]byte, error)                  // (not mediated). Returns nil if nothing/error
	ReadEntryPart(pos, start, size int64) ([]byte, error) // Read part of an entry starting at start with size bytes
	ReadRange(pos, count int64) ([][]byte, error)         // Read range starting with pos having count entries
	ReadRandom() (int64, []byte)                          // Return a random entry
	Truncate()                                            // Truncate list to zero so it will fail on create
	Exists() bool                                         // does the index exist
	EntryExists(pos int64) bool                           // Check if an entry exists
	Filename(pos int64) string                            // Returns the filename of the list
	Entries() int64                                       // Returns the number of entries
	Update(func(tx Tx) error) error                       // call update function while blocking other entries
	DeleteBefore(pos int64) error                         // delete files before pos, not including the file containing pos
	LastChange(pos int64) int64                           // return unixtime of last change to file containing pos
	Indices() [][]byte                                    // retun a list of indices
	MaxEntries() int64                                    // Returns maximum pages per object
}
