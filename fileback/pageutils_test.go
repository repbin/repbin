package fileback

import (
	"bytes"
	"testing"
)

func TestFill(t *testing.T) {
	pageSize := int64(20)
	pageSizeTotal := int64(22)
	td := []byte{32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 13, 10}
	fill := makeFill(pageSize, byte(' '), []byte("\r\n"))
	if !bytes.Equal(fill, td) {
		t.Error("fill is wrong")
	}

	page := createPage(pageSize, pageSizeTotal, fill, make([]byte, pageSize))
	if !bytes.Equal(page, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 13, 10}) {
		t.Errorf("createPage fail 1: %+v", page)
	}
	page = createPage(pageSize, pageSizeTotal, fill, make([]byte, pageSize+1))
	if !bytes.Equal(page, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 13, 10}) {
		t.Errorf("createPage fail 2: %+v", page)
	}
	page = createPage(pageSize, pageSizeTotal, fill, make([]byte, pageSize-1))
	if !bytes.Equal(page, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 13, 10}) {
		t.Errorf("createPage fail 3: %+v", page)
	}
	page = createPage(pageSizeTotal, pageSizeTotal, fill, make([]byte, pageSizeTotal))
	if !bytes.Equal(page, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}) {
		t.Errorf("createPage fail 4: %+v", page)
	}
	page = createPage(pageSizeTotal, pageSizeTotal, fill, make([]byte, pageSizeTotal-1))
	if !bytes.Equal(page, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10}) {
		t.Errorf("createPage fail 4: %+v", page)
	}
	page = createPage(pageSizeTotal, pageSizeTotal, fill, make([]byte, pageSizeTotal+1))
	if !bytes.Equal(page, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}) {
		t.Errorf("createPage fail 4: %+v", page)
	}
}

func TestCleanPath(t *testing.T) {
	if "/something/" != cleanpath("/something/") {
		t.Error("cleanpath test 1 failed")
	}
	if "/something/" != cleanpath("/something/./") {
		t.Error("cleanpath test 2 failed")
	}
	if "/something/" != cleanpath(" /something/ ../") {
		t.Error("cleanpath test 3 failed")
	}
	if "something/" != cleanpath(" something/../") {
		t.Error("cleanpath test 4 failed")
	}
}

func TestIndexToPath(t *testing.T) {
	td := []byte{0xff, 0xfa, 0xaa, 0xcc, 0xcc, 0xcc}
	a, b := indexToPath(td)
	if a+b != "/fff/aaa/cccccc" {
		t.Error("indexToPath failed")
	}
}

func TestReaderID(t *testing.T) {
	max := 10
	for i := 0; i < 100; i++ {
		v := readerID(make([]byte, 10), max)
		if v > max || v < 0 {
			t.Fatal("ReaderID bad number")
		}
	}
}
