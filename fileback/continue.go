package fileback

// NewContinue returns a new continuous paged list
// dir is the path to the directory containing it, pagesize is the size of each page
// pages are filled with fill and terminated with end. end is not counted in the page size
// workers is the number of worker routines to use
func NewContinue(dir string, pageSize, maxEntries int64, fill byte, end []byte, workers int) PagedList {
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
		listType:   listTypeContinue,
		maxEntries: maxEntries,
	}
	return pl
}
