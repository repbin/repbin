package fileback

import (
	"math/rand"
	"os"
	"sync"
	"time"
)

// PagedList is a paged list implementation
type PagedList struct {
	Dir        string        // The directory containing the list
	PageSize   int64         // The size per entry
	pageSize   int64         // true pagesize
	fill       []byte        // Template for entry
	index      []byte        // the index in operation
	dir        string        // the directory containing the file
	name       string        // the filename containing the pages
	numWorkers int           // number of active workers
	workerID   *sync.Mutex   // the assigned worker
	workers    []*sync.Mutex // worker channels
	headerSize int64         // Size of header, if any
	maxEntries int64         // Maximum entries per file
	listType   byte
}

// Index returns the index within a list. the index is converted to a path
// indices are not additive. There's exactly one level only
func (list PagedList) Index(index []byte) PagedList {
	var dir string
	if index == nil || len(index) == 0 {
		panic("No index")
	}
	pl := PagedList{
		Dir:        list.Dir,
		PageSize:   list.PageSize,
		pageSize:   list.pageSize,
		fill:       list.fill,
		index:      index,
		numWorkers: list.numWorkers,
		workerID:   list.workers[readerID(index, list.numWorkers)],
		workers:    list.workers,
		headerSize: list.headerSize,
		maxEntries: list.maxEntries,
		listType:   list.listType,
	}
	dir, pl.name = indexToPath(index)
	pl.dir = list.Dir + dir
	return pl
}

// CreateAppend appends data to at index, creates index if necessary
func (list PagedList) CreateAppend(data []byte) error {
	list.workerID.Lock()
	defer list.workerID.Unlock()
	if !list.Exists() {
		return list.create(data)
	}
	return list.append(data)
}

// Exists tests if a list exists
func (list PagedList) Exists() bool {
	// NO mutex
	return list.exists()
}

func (list PagedList) exists() bool {
	if list.listType == listTypeRolling {
		return list.rollingExists()
	}
	if _, err := os.Stat(list.Filename(0)); os.IsNotExist(err) {
		return false
	}
	return true
}

// Filename returns the filename of the file containing page pos
func (list PagedList) Filename(pos int64) string {
	// NO mutex
	if list.name == "" {
		panic("No index")
	}
	if list.listType == listTypeRolling {
		// pos < 0 returns last/cur file for rolling lists
		return list.rollingFileName(pos)
	}
	return list.dir + list.name + listExtension
}

// Truncate the list to zero and make it unusable
func (list PagedList) Truncate() {
	list.workerID.Lock()
	defer list.workerID.Unlock()
	list.truncate()
}

func (list PagedList) truncate() {
	if !list.exists() {
		return
	}
	if list.listType == listTypeRolling {
		list.rollingTruncate()
		return
	}
	os.Truncate(list.Filename(0), 0)
}

// Delete the database file(s)
func (list PagedList) Delete() {
	list.workerID.Lock()
	defer list.workerID.Unlock()
	list.delete()
}

func (list PagedList) delete() {
	if !list.exists() {
		return
	}
	if list.listType == listTypeRolling {
		list.rollingDelete()
		return
	}
	os.Remove(list.Filename(0))
}

// Create a new list at index if it does not exist yet
func (list PagedList) Create(data []byte) error {
	list.workerID.Lock()
	defer list.workerID.Unlock()
	return list.create(data)
}

// Write the position header in a roundRobin list
func (list PagedList) writeHeader(pos int64, file *os.File) error {
	if list.headerSize > 0 && list.listType == listTypeRoundRobin {
		// page := createPage(list.PageSize, list.headerSize, list.fill, []byte(uintToHex(uint64(pos%list.maxEntries))))
		page := append([]byte(uintToHex(uint64(pos%list.maxEntries))), byte('\n'))
		_, err := file.Seek(0, os.SEEK_SET)
		if err != nil {
			return err
		}
		_, err = file.Write(page)
		if err != nil {
			return err
		}
	}
	return nil
}

// readHeader reads the header of the file and returns the encoded last writing position
func (list PagedList) readHeader(file *os.File) (int64, error) {
	if list.headerSize > 0 && list.listType == listTypeRoundRobin {
		_, err := file.Seek(0, os.SEEK_SET)
		if err != nil {
			return -1, err
		}
		headerPos := make([]byte, 16) // Encoded header position is no more than 16 bytes
		_, err = file.Read(headerPos)
		if err != nil {
			return -1, err
		}
		pos, err := strToUint(string(headerPos))
		if err != nil {
			return -1, err
		}
		return int64(pos), nil
	}
	return -1, ErrNoHeader
}

func (list PagedList) create(data []byte) error {
	fileName := list.Filename(0)
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(list.dir, 0700) // create missing dirs
			if err != nil {
				return err
			}
			file, err = os.OpenFile(fileName, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600) // try again
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	defer file.Sync()
	defer file.Close()
	if list.listType == listTypeRolling {
		// Rolling, create cur/last link
		err = os.Symlink(fileName, list.dir+list.name+lastExtension)
		if err != nil {
			return nil
		}
		// Create index file
		err = list.writeRollingIndex(0, 0)
		if err != nil {
			return nil
		}
	} else {
		// Never write header for rolling
		err = list.writeHeader(0, file)
		if err != nil {
			return err
		}
	}
	_, err = file.Write(createPage(list.PageSize, list.pageSize, list.fill, data))
	if err != nil {
		return err
	}
	return nil
}

// GetLast returns the last page or nil.
func (list PagedList) GetLast() []byte {
	// No: Mutex
	return list.getLast()
}

func (list PagedList) getLast() []byte {
	if !list.exists() {
		return nil
	}
	file, err := os.OpenFile(list.Filename(lastFile), os.O_RDONLY, 0600)
	if err != nil {
		return nil
	}
	defer file.Close()
	if list.listType == listTypeRoundRobin {
		lastPos, err := list.readHeader(file)
		if err != nil {
			return nil
		}
		_, err = file.Seek(list.headerSize+(lastPos*list.pageSize), os.SEEK_SET)
		if err != nil {
			return nil
		}
	} else {
		_, err = file.Seek(-list.pageSize, os.SEEK_END)
		if err != nil {
			return nil
		}
	}
	page := make([]byte, list.PageSize)
	_, err = file.Read(page)
	if err != nil {
		return nil
	}
	return page
}

// Append data to end of list
func (list PagedList) Append(data []byte) error {
	list.workerID.Lock()
	defer list.workerID.Unlock()
	return list.append(data)
}

func (list PagedList) append(data []byte) error {
	if !list.exists() {
		return ErrExists
	}
	if list.maxEntries == 1 {
		return ErrMaxEntries
	}
	if list.listType == listTypeContinue && list.Entries() >= list.maxEntries {
		return ErrMaxEntries
	}
	if list.listType == listTypeRolling {
		// Roll over if necessary
		err := list.rollover()
		if err != nil {
			return err
		}
	}
	mode := os.O_WRONLY
	if list.listType == listTypeRoundRobin {
		mode = os.O_RDWR
	}
	file, err := os.OpenFile(list.Filename(appendFile), mode, 0600)
	if err != nil {
		// RollOver: File might not exist,create
		if os.IsNotExist(err) {
			mode = mode | os.O_CREATE
			file, err = os.OpenFile(list.Filename(appendFile), mode, 0600)
			if err != nil {
				return err
			}
			// .idx file written by rollover
		} else {
			return err
		}
	}
	defer file.Close()
	// RoundRobin is different. Has to look up last position
	if list.listType == listTypeRoundRobin {
		lastPos, err := list.readHeader(file)
		if err != nil {
			return err
		}
		// Update header
		err = list.writeHeader(lastPos+1, file)
		if err != nil {
			return err
		}
		if lastPos+1 >= list.maxEntries {
			// go to beginning
			_, err = file.Seek(list.headerSize, os.SEEK_SET)
		} else {
			// go to end of last written
			_, err = file.Seek(list.headerSize+((lastPos+1)*list.pageSize), os.SEEK_SET)
		}
		if err != nil {
			return err
		}
	} else {
		_, err = file.Seek(0, os.SEEK_END)
		if err != nil {
			return err
		}
	}
	_, err = file.Write(createPage(list.PageSize, list.pageSize, list.fill, data))
	return err
}

// ReadEntry at pos
func (list PagedList) ReadEntry(pos int64) ([]byte, error) {
	// No: Mutex
	return list.readEntry(pos, 0, list.PageSize)
}

// ReadEntryPart reads entry at pos, starting with byte start and returning at most size
func (list PagedList) ReadEntryPart(pos, start, size int64) ([]byte, error) {
	return list.readEntry(pos, start, size)
}

func (list PagedList) readEntry(pos, start, size int64) ([]byte, error) {
	if !list.exists() {
		return nil, ErrExists
	}
	file, err := os.OpenFile(list.Filename(pos), os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return list.readOneEntry(pos, file, start, size)
}

// Read partial. Start, end
func (list PagedList) readOneEntry(pos int64, file *os.File, start, size int64) ([]byte, error) {
	_, err := file.Seek(start+list.headerSize+((pos%list.maxEntries)*list.pageSize), os.SEEK_SET)
	if err != nil {
		return nil, err
	}
	if size <= 0 {
		size = list.PageSize
	}
	if start+size > list.PageSize {
		return nil, ErrOOB
	}
	page := make([]byte, size)
	_, err = file.Read(page)
	if err != nil {
		return nil, err
	}
	return page, nil
}

// Change page at positing
func (list PagedList) Change(pos int64, data []byte) error {
	list.workerID.Lock()
	defer list.workerID.Unlock()
	return list.change(pos, createPage(list.PageSize, list.pageSize, list.fill, data))
}

// ChangePart changes a part of an entry starting at start
func (list PagedList) ChangePart(pos int64, data []byte, start int64) error {
	list.workerID.Lock()
	defer list.workerID.Unlock()
	return list.changePart(pos, data, start)
}

// change a page completely (overwrite)
func (list PagedList) change(pos int64, data []byte) error {
	return list.changePart(pos, createPage(list.PageSize, list.pageSize, list.fill, data), 0)
}

// changePart changes a part of a page
func (list PagedList) changePart(pos int64, data []byte, start int64) error {
	if !list.exists() {
		return ErrExists
	}
	file, err := os.OpenFile(list.Filename(pos), os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	if start >= list.PageSize {
		return ErrOOB
	}
	defer file.Sync()
	defer file.Close()
	_, err = file.Seek(start+list.headerSize+((pos%list.maxEntries)*list.pageSize), os.SEEK_SET) // standard
	if err != nil {
		return err
	}
	// calculate end
	end := int64(len(data))
	if end > list.PageSize-start {
		end = list.PageSize - start
	}
	_, err = file.Write(data[:end])
	return err
}

// Entries returns the number of pages in a file
func (list PagedList) Entries() int64 {
	// No: Mutex
	return list.entries()
}

func (list PagedList) entriesInFile(filename string) int64 {
	stat, err := os.Stat(filename)
	if err != nil {
		return 0
	}
	size := stat.Size()
	if list.headerSize > 0 {
		if size <= list.headerSize {
			return 0
		}
	}
	return (size - list.headerSize) / list.pageSize
}

func (list PagedList) entries() int64 {
	if !list.exists() {
		return 0
	}
	if list.listType == listTypeRolling {
		return list.rollingEntries()
	}
	return list.entriesInFile(list.Filename(0))
}

// ReadRandom returns a random entry
func (list PagedList) ReadRandom() (int64, []byte) {
	// No: Mutex
	return list.readRandom()
}

func (list PagedList) readRandom() (int64, []byte) {
	if !list.exists() {
		return -1, nil
	}
	entries := list.entries()
	if entries == 0 {
		return -1, nil
	}
	rand.Seed(time.Now().UnixNano())
	entry := rand.Int63() % entries
	d, _ := list.readEntry(entry, 0, list.PageSize)
	return entry, d
}

// ReadRange returns a range of count entries starting with pos
func (list PagedList) ReadRange(pos, count int64) ([][]byte, error) {
	list.workerID.Lock()
	defer list.workerID.Unlock()
	return list.readRange(pos, count)
}

func (list PagedList) readRange(pos, count int64) ([][]byte, error) {
	var found int64
	var err error
	var ret [][]byte
	var filename string
	var curFile *os.File
	if !list.exists() {
		return nil, ErrExists
	}
	firstRead := true
	if count > list.maxEntries {
		count = list.maxEntries
	}
	for found < count {
		filenamePos := list.Filename(pos)
		if filenamePos != filename {
			if curFile != nil {
				curFile.Close()
			}
			curFile, err = os.OpenFile(filenamePos, os.O_RDONLY, 0600)
			if err != nil {
				return ret, ErrNoMore
			}
			// If roundrobin,
			if firstRead && list.listType == listTypeRoundRobin {
				last, err := list.readHeader(curFile)
				if err != nil {
					return ret, err
				}
				pos += last
				firstRead = false
			}
			filename = filenamePos
		}
		t, _ := list.readOneEntry(pos, curFile, 0, list.PageSize)
		if t != nil {
			ret = append(ret, t)
		} else {
			if curFile != nil {
				curFile.Close()
			}
			return ret, ErrNoMore
		}
		found++
		pos++
	}
	if curFile != nil {
		curFile.Close()
	}
	return ret, nil
}

// EntryExists returns true if an entry with position pos exists
func (list PagedList) EntryExists(pos int64) bool {
	if !list.exists() {
		return false
	}
	fnLen := list.entriesInFile(list.Filename(pos))
	first := (pos / list.maxEntries) * list.maxEntries
	last := ((pos / list.maxEntries) * list.maxEntries) + fnLen
	if first < pos && pos < last {
		return true
	}
	return false
}

// LastChange returns the unixtime at which the file containing pos was last changed. It returns 0 on error
func (list PagedList) LastChange(pos int64) int64 {
	if !list.exists() {
		return 0
	}
	fi, err := os.Stat(list.Filename(pos))
	if err != nil {
		return 0
	}
	return fi.ModTime().Unix()
}

// Indices returns a list of indices or nil
func (list PagedList) Indices() [][]byte {
	indicesHex, err := listDir(list.Dir, "", 4)
	if err != nil {
		return nil
	}
	return stringListToByte(indicesHex)
}

// MaxEntries returns the maximum entries per list object
func (list PagedList) MaxEntries() int64 {
	return list.maxEntries
}
