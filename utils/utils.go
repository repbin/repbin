// Package utils implements a few utilities for repbin commands.
package utils

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
	"math/big"
	mathrand "math/rand"
	"os"
	"time"
)

// Version of this release
const Version = "0.0.1 very alpha"

var (
	// ErrMaxBytes is returned if too many bytes are available
	ErrMaxBytes = errors.New("utils: Too many bytes")
	// ErrNoFiles is returned if no files could be found
	ErrNoFiles = errors.New("utils: No files found")
)
var (
	secret [8]byte
)

func init() {
	rand.Read(secret[:])
}

// MaxRead reads n bytes from r. If more bytes are available, return ErrMaxBytes
func MaxRead(n int64, r io.Reader) ([]byte, error) {
	limitReader := io.LimitReader(r, n)
	ret, err := ioutil.ReadAll(limitReader)
	if err != nil {
		return nil, err
	}
	nr, err := r.Read(make([]byte, 1))
	if nr > 0 || err == nil {
		return nil, ErrMaxBytes
	}
	return ret, nil
}

// MaxReadFile reads a file into []byte
func MaxReadFile(n int64, filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	d, err := MaxRead(n, f)
	if err != nil {
		return nil, err
	}
	return d, nil
}

// MaxStdinRead reads n bytes from stdin. If more bytes are available, return ErrMaxBytes
func MaxStdinRead(n int64) ([]byte, error) {
	return MaxRead(n, os.Stdin)
}

// WriteStdout writes b to stdout
func WriteStdout(b []byte) error {
	_, err := os.Stdout.Write(b)
	os.Stdout.Sync()
	return err
}

// OverWriteFile overwrites an existing file
func OverWriteFile(f string, b []byte) error {
	file, err := os.OpenFile(f, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(b)
	return err
}

// WriteNewFile writes b to f IF f does not exist
func WriteNewFile(f string, b []byte) error {
	file, err := os.OpenFile(f, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(b)
	return err

}

// randFile returns a random os.FileInfo from slice of os.FileInfo
func randFile(s []os.FileInfo) (os.FileInfo, error) {
	x := int64(len(s))
	if x == 0 {
		return nil, ErrNoFiles
	}
	if x == 1 {
		return s[0], nil
	}
	r, err := rand.Int(rand.Reader, big.NewInt(x-1))
	if err != nil {
		return nil, ErrNoFiles
	}
	return s[r.Int64()], nil
}

// OpenRandomFile opens a random file from dirname
func OpenRandomFile(dirname string) (*os.File, error) {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, err
	}
	nfiles := make([]os.FileInfo, 0, len(files))
	for _, x := range files {
		if !x.IsDir() {
			nfiles = append(nfiles, x)
		}
	}
	finfo, err := randFile(nfiles)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(dirname+string(os.PathSeparator)+finfo.Name(), os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// ReadRandomFile reads a random file from dirname and deletes it
func ReadRandomFile(dirname string, n int64) ([]byte, string, error) {
	f, err := OpenRandomFile(dirname)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()
	b, err := MaxRead(n, f)
	if err != nil {
		return nil, "", err
	}
	return b, f.Name(), nil
}

// MakeDirMany creates many directories
func MakeDirMany(dir ...string) error {
	for _, d := range dir {
		os.Mkdir(d, 0750)
	}
	return nil
}

// EncodeUInt64 encodes a uint64 to byte[]
func EncodeUInt64(i uint64) []byte {
	r := make([]byte, 8)
	binary.BigEndian.PutUint64(r, i)
	return r
}

// DecodeUInt64 does a safe decode of byte[] to uint64
func DecodeUInt64(i []byte) uint64 {
	j := make([]byte, len(i))
	copy(j, i)
	return binary.BigEndian.Uint64(j)
}

// RandomUInt64 creates a random uint64
func RandomUInt64() []byte {
	mathrand.Seed(time.Now().Unix())
	t := mathrand.Int63()
	r := make([]byte, 8)
	binary.BigEndian.PutUint64(r, uint64(t))
	rand.Read(secret[:])
	r[0] ^= secret[0]
	r[1] ^= secret[1]
	r[2] ^= secret[2]
	r[3] ^= secret[3]
	r[4] ^= secret[4]
	r[5] ^= secret[5]
	r[6] ^= secret[6]
	r[7] ^= secret[7]
	return r
}
