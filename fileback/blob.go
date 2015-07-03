package fileback

// NewBlob returns a new blobg list
// Blob:
// 		Blob is paged list that maps: Append,change -> Create, GetLast->ReadEntry
// 		Read on blob is NOT mediated
// func NewBlob(index []byte) Index {}

// NewBlob returns a new blob storage. Blobs are special since they are zero filled without end
// dir is the path to the directory containing it, pagesize is the size of each page
// pages are filled with fill and terminated with end. end is not counted in the page size
// workers is the number of worker routines to use
func NewBlob(dir string, pageSize int64, workers int) PagedList {
	if pageSize > MaxPage || pageSize < 1 {
		panic("Pagesize > UInt32")
	}
	pl := PagedList{
		Dir:        dir,
		PageSize:   pageSize,
		fill:       nil,
		pageSize:   pageSize,
		numWorkers: workers,
		workers:    createMutexes(workers),
		listType:   listTypeBlob,
		maxEntries: 1,
	}
	return pl
}
