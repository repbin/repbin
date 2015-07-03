package fileback

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
)

// makeFill creates the fill template of size pageSize+len(end) containing fill and ending in end
func makeFill(pageSize int64, fill byte, end []byte) []byte {
	length := pageSize + int64(len(end))
	f := bytes.Repeat([]byte{fill}, int(length)) // dangerous. limits page sizes to maxInt
	if pageSize < length {
		copy(f[pageSize:length], end[:])
	}
	return f
}

func cleanpath(dir string) string {
	return strings.TrimRight(strings.Trim(dir, " \t\n"), ". "+string(os.PathSeparator)) + string(os.PathSeparator)
}

func indexToPath(index []byte) (string, string) {
	indexHex := hex.EncodeToString(index)
	return string(os.PathSeparator) + indexHex[0:3] + string(os.PathSeparator) + indexHex[3:6] + string(os.PathSeparator), indexHex[6:]
}

func readerID(index []byte, numWorkers int) int {
	q := make([]byte, len(index)+4)
	copy(q, index)
	copy(q[len(index):], secret)
	s := sha256.Sum224(q)
	s1 := binary.BigEndian.Uint64(s[0:8])
	return int(s1 % uint64(numWorkers))
}

// createPage gets the input data and fills it
func createPage(pageSizeInlet, pageSizeTotal int64, fill, data []byte) []byte {
	dataLen := int64(len(data))
	if pageSizeInlet == pageSizeTotal && pageSizeInlet == dataLen { // fixed page size without end matching
		return data
	}
	if dataLen > pageSizeInlet {
		dataLen = pageSizeInlet
	}
	page := make([]byte, pageSizeTotal)
	copy(page, data[:dataLen])                  // Copy data to page
	if pageSizeTotal > dataLen && fill != nil { // Add whatever fill is missing
		copy(page[dataLen:], fill[dataLen:])
	}
	return page
}

func createMutexes(workers int) []*sync.Mutex {
	ret := make([]*sync.Mutex, workers)
	for i := 0; i < workers; i++ {
		ret[i] = new(sync.Mutex)
	}
	return ret
}

func uintToHex(i uint64) string {
	t := make([]byte, 8) // uint64 is 8 bytes
	binary.BigEndian.PutUint64(t, i)
	return hex.EncodeToString(t)
}

func strToUint(s string) (uint64, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(b), nil
}

// listDir is a helper function to list files in a directory and return them has hex-encoded indices
func listDir(parent, directory string, depths int) ([]string, error) {
	var uniq map[string]bool
	if depths == 0 {
		return nil, fmt.Errorf("Depths reached %d", depths)
	}
	stat, err := os.Lstat(parent + string(os.PathSeparator) + directory)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("Not a directory: %s", stat.Name())
	}
	file, err := os.Open(parent + string(os.PathSeparator) + directory)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	uniq = make(map[string]bool)
	for {
		entries, err := file.Readdir(10)
		if entries == nil {
			break
		}
		if err != nil {
			break
		}
		for _, entry := range entries {
			if entry.IsDir() {
				list, err := listDir(parent, directory+string(os.PathSeparator)+entry.Name(), depths-1)
				if err != nil {
					return nil, err
				}
				for _, v := range list {
					uniq[v] = true
				}
			} else {
				e := path.Clean(directory + string(os.PathSeparator) + entry.Name())
				ext := path.Ext(e)
				e = e[:len(e)-len(ext)]
				e = strings.Replace(e, string(os.PathSeparator), "", -1)
				uniq[e] = true
			}
		}
	}
	ret := make([]string, 0, len(uniq))
	for k := range uniq {
		ret = append(ret, k)
	}
	return ret, nil
}

func stringListToByte(i []string) [][]byte {
	if i == nil || len(i) == 0 {
		return nil
	}
	ret := make([][]byte, 0, len(i))
	for _, e := range i {
		t, err := hex.DecodeString(e)
		if err == nil {
			ret = append(ret, t)
		}
	}
	return ret[:len(ret)]
}
