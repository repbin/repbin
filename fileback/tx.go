package fileback

// Tx provides function access during transactions
type Tx struct {
	pagedList PagedList
	PageSize  int64 // Maximum inlay page
}

// Update executes updateFunc while the index is locked.
func (list PagedList) Update(updateFunc func(tx Tx) error) error {
	list.workerID.Lock()
	defer list.workerID.Unlock()
	tx := Tx{
		pagedList: list,
		PageSize:  list.PageSize,
	}
	return updateFunc(tx)
}

// ===============================================================================================

// Exists tests if a list exists
func (tx Tx) Exists() bool {
	// Exists tests if a list exists
	return tx.pagedList.exists()
}

// Filename returns the filename of the file containing page pos
func (tx Tx) Filename(pos int64) string {
	return tx.pagedList.Filename(pos)
}

// Truncate the list to zero
func (tx Tx) Truncate() {
	tx.pagedList.truncate()
}

// Delete the database file(s)
func (tx Tx) Delete() {
	tx.pagedList.delete()
}

// Create a new list at index if it does not exist yet
func (tx Tx) Create(data []byte) error {
	return tx.pagedList.create(data)
}

// GetLast returns the last page or nil.
func (tx Tx) GetLast() []byte {
	return tx.pagedList.getLast()
}

// Append data to end of list
func (tx Tx) Append(data []byte) error {
	return tx.pagedList.append(data)
}

// ReadEntry at pos
func (tx Tx) ReadEntry(pos int64) ([]byte, error) {
	return tx.pagedList.readEntry(pos, 0, tx.pagedList.PageSize)
}

// ReadEntryPart reads entry at pos, starting with byte start and returning at most size
func (tx Tx) ReadEntryPart(pos, start, size int64) ([]byte, error) {
	return tx.pagedList.readEntry(pos, start, size)
}

// Change page at position
func (tx Tx) Change(pos int64, data []byte) error {
	return tx.pagedList.change(pos, data)
}

// ChangePart changes part of page at position pos starting at start
func (tx Tx) ChangePart(pos int64, data []byte, start int64) error {
	return tx.pagedList.changePart(pos, data, start)
}

// Entries returns the number of pages in a file
func (tx Tx) Entries() int64 {
	return tx.pagedList.entries()
}

// ReadRandom returns a random entry
func (tx Tx) ReadRandom() (int64, []byte) {
	return tx.pagedList.readRandom()
}

// ReadRange returns a range of count entries starting with pos
func (tx Tx) ReadRange(pos, count int64) ([][]byte, error) {
	return tx.pagedList.readRange(pos, count)
}
