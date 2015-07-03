package fileback

// NewRoundRobin returns a round-robin list
// RoundRobin: index is path
// 		File does not grow beyond maximum size
// 		First line contains page number "current entry" (last entry written) and last counter
// 		First line is read only partially (16byte + 16byte + one separator plus lineend)
// func NewRoundRobin(index []byte) Index {}

// NewRoundRobin returns a roundrobin list
func NewRoundRobin(dir string, pageSize, maxEntries int64, fill byte, end []byte, workers int) PagedList {
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
		maxEntries: maxEntries,
		listType:   listTypeRoundRobin,
		// headerSize: int64(len(end)) + pageSize,
		headerSize: 17, //int64(len(end)) + pageSize,
	}
	return pl
}
