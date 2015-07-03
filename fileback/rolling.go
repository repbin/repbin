package fileback

import (
	"bytes"
	"os"
)

// NewRoll returns a rolling list
// RollList: index is path
// 		Files roll over into new files. Filename ends with index / maxentrycount
// 		Has a single file containing first + last entry

// NewRolling returns a rolling list
func NewRolling(dir string, pageSize, maxEntries int64, fill byte, end []byte, workers int) PagedList {
	if pageSize > MaxPage || pageSize < 1 {
		panic("Pagesize > UInt32")
	}
	pl := PagedList{
		Dir:        dir,
		PageSize:   pageSize,
		fill:       makeFill(pageSize, fill, end),
		pageSize:   int64(len(end)) + pageSize,
		numWorkers: workers,
		workers:    createMutexes(workers),
		listType:   listTypeRolling,
		maxEntries: maxEntries,
	}
	return pl
}

func (list PagedList) rollingEntries() int64 {
	begin, end, err := list.readRollingIndex()
	if err != nil {
		return 0
	}
	count1 := (end - begin) * list.maxEntries
	count2 := list.entriesInFile(list.dir + list.name + lastExtension)
	return count1 + count2
}

// rolling delete deletes all files belonging to the list
func (list PagedList) rollingDelete() {
	begin, end, err := list.readRollingIndex()
	if err != nil {
		return
	}
	for i := begin; i <= end; i++ {
		os.Remove(list.rollingFileName(i * list.maxEntries))
	}
	os.Remove(list.dir + list.name + indexExtension)
	os.Remove(list.dir + list.name + lastExtension)
}

// rollingTruncate implements the truncation operation for rolling lists.
// it deletes all files except the first (which is truncated) and the last link (which will dangle)
func (list PagedList) rollingTruncate() {
	begin, end, err := list.readRollingIndex()
	if err != nil {
		return
	}
	for i := begin + 1; i <= end; i++ {
		os.Remove(list.rollingFileName(i * list.maxEntries))
	}
	firstfile := list.rollingFileName(begin)
	os.Truncate(firstfile, 0)
	list.writeRollingIndex(0, 0)
	os.Remove(list.dir + list.name + lastExtension)
	os.Symlink(firstfile, list.dir+list.name+lastExtension)
}

// DeleteBefore deletes files before pos, not including the file containing pos
func (list PagedList) DeleteBefore(pos int64) error {
	if list.listType != listTypeRolling {
		return nil
	}
	_, end, err := list.readRollingIndex()
	if err != nil {
		return err
	}
	myfile := list.rollingFileName(pos)
	for i := pos; i > 0; i = i - list.maxEntries {
		dfile := list.rollingFileName(i)
		if dfile != myfile {
			os.Remove(dfile) // ignore errors
		}
	}
	return list.writeRollingIndex(pos%list.maxEntries, end)
}

// writeRollingIndex writes the index file of a rolling list
func (list PagedList) writeRollingIndex(begin, end int64) error {
	// Index file contains two entries (hex encoded):
	// firstExtension lastExtension
	// extensions of files. That way we can easily calculate the difference/size
	entry := uintToHex(uint64(begin)) + " " + uintToHex(uint64(end)) + " \n"
	filename := list.dir + list.name + indexExtension
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write([]byte(entry))
	if err != nil {
		return err
	}
	return nil
}

// readRollingIndex reads the index file of a rolling list
func (list PagedList) readRollingIndex() (begin, end int64, err error) {
	filename := list.dir + list.name + indexExtension
	file, errf := os.OpenFile(filename, os.O_RDONLY, 0600)
	if errf != nil {
		return 0, 0, errf
	}
	defer file.Close()
	input := make([]byte, 16+16+1) // two hex uint plus space
	_, err = file.Read(input)
	if err != nil {
		return 0, 0, err
	}
	fields := bytes.Split(input, []byte(" ")) // split
	if len(fields) < 2 {
		return 0, 0, ErrListCorrupt
	}
	beginI, err := strToUint(string(fields[0]))
	if err != nil {
		return 0, 0, err
	}
	endI, err := strToUint(string(fields[1]))
	if err != nil {
		return 0, 0, err
	}
	return int64(beginI), int64(endI), nil
}

// rollingExists implements exists() for rolling lists
func (list PagedList) rollingExists() bool {
	_, err1 := os.Stat(list.dir + list.name + lastExtension)
	_, err2 := os.Stat(list.dir + list.name + indexExtension)
	if os.IsNotExist(err1) || os.IsNotExist(err2) {
		return false
	}
	return true
}

// rollingFileName implements Filename(pos) for rolling lists
func (list PagedList) rollingFileName(pos int64) string {
	if pos < 0 {
		return list.dir + list.name + lastExtension
	}
	fn := list.dir + list.name + "." + uintToHex(uint64(pos/list.maxEntries)) + listExtension
	return fn
}

// rollover tests if .last has more than maxEntries entries, if yes, it creates a new file and moves the link
func (list PagedList) rollover() error {
	lastFile := list.dir + list.name + lastExtension
	countEntries := list.entriesInFile(lastFile)
	if countEntries < list.maxEntries {
		return nil
	}
	begin, end, err := list.readRollingIndex()
	if err != nil {
		return err
	}
	end++
	endS := uintToHex(uint64(end))
	// Create the next file
	newFile := list.dir + list.name + "." + endS + listExtension
	file, err := os.OpenFile(newFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	// Update index
	err = list.writeRollingIndex(begin, end)
	if err != nil {
		return err
	}
	// remove old link
	err = os.Remove(lastFile)
	if err != nil {
		return err
	}
	// create new link
	err = os.Symlink(newFile, lastFile)
	if err != nil {
		return err
	}
	return nil
}
