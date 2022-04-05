package rocksdbhelper

import (
	"fmt"
	"sync"
	"syscall"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/internal/fileutil"
	rocksdb "github.com/linxGnu/grocksdb"
	"github.com/pkg/errors"
)

var logger = flogging.MustGetLogger("rocksdbdbhelper")

type dbState int32

const (
	closed dbState = iota
	opened
)

// DB - a wrapper on an actual store
type DB struct {
	conf    *Conf
	db      *rocksdb.DB
	dbState dbState
	mutex   sync.RWMutex

	readOpts *rocksdb.ReadOptions
	/*writeOptsNoSync *rocksdb.WriteOptions
	writeOptsSync   *rocksdb.WriteOptions*/
}

// CreateDB constructs a `DB`
func CreateDB(conf *Conf) *DB {
	logger.Debugf("RocksDB constructing...")
	logger.Debugf("RocksDB constructing successfully finished")
	return &DB{
		conf:     conf,
		dbState:  closed,
		readOpts: rocksdb.NewDefaultReadOptions(),
	}
}

// Open opens the underlying db
func (dbInst *DB) Open() {
	logger.Debugf("Opening DB...")
	dbInst.mutex.Lock()
	defer dbInst.mutex.Unlock()
	if dbInst.dbState == opened {
		return
	}
	//block based table from the example
	//bbto := rocksdb.NewDefaultBlockBasedTableOptions()
	//bbto.SetBlockCache(rocksdb.NewLRUCache(3 << 30)) //3 GB

	dbOpts := rocksdb.NewDefaultOptions()
	//dbOpts.SetBlockBasedTableFactory(bbto)

	dbPath := dbInst.conf.DBPath
	var err error
	logger.Debugf("ParanoidChecks is: %t, (default value is false, so it will be set to false)", dbOpts.ParanoidChecks())

	isDirEmpty, err := fileutil.CreateDirIfMissing(dbPath)
	if err != nil {
		panic(fmt.Sprintf("Error creating dir if missing: %s", err))
	}
	dbOpts.SetParanoidChecks(false) //docs says that default value is false
	dbOpts.SetCreateIfMissing(isDirEmpty)
	if dbInst.db, err = rocksdb.OpenDb(dbOpts, dbPath); err != nil {
		panic(fmt.Sprintf("Error opening rocksdb: %s", err))
	}
	logger.Debugf("DB RocksDB was successfully opened in path: [ %s ]", dbPath)
	dbInst.dbState = opened
}

// IsEmpty returns whether or not a database is empty
func (dbInst *DB) IsEmpty() (bool, error) {
	logger.Debugf("Checkin if DB is empty...")
	if dbInst.dbState == closed {
		return false, errors.New("RocksDB is closed.")
	}
	dbInst.mutex.RLock()
	defer dbInst.mutex.RUnlock()
	itr := dbInst.db.NewIterator(rocksdb.NewDefaultReadOptions())
	defer itr.Close()
	itr.SeekToFirst()
	hasItems := itr.Valid()
	logger.Debugf("Checking for emptiness has finished")
	return !hasItems,
		errors.Wrapf(itr.Err(), "error while trying to see if the rocksdb at path [%s] is empty", dbInst.conf.DBPath)
}

// Close closes the underlying db
func (dbInst *DB) Close() {
	dbInst.mutex.Lock()
	defer dbInst.mutex.Unlock()
	if dbInst.dbState == closed {
		return
	}
	logger.Debugf("Closing db...")
	dbInst.db.Close() //TODO: should we check if db have been already closed here?
	dbInst.dbState = closed
}

// Get returns the value for the given key
func (dbInst *DB) Get(key []byte) ([]byte, error) {
	logger.Debugf("Getting key [%s] from RocksDB...", key)
	dbInst.mutex.RLock()
	defer dbInst.mutex.RUnlock()
	//TODO: delete commened code below, if we never have to uncomment it
	/*if dbInst.dbState == closed {
		logger.Errorf("Error retrieving rocksdb key [%#v]: rocksdb is closed", key)
		return nil, errors.Errorf("Error retrieving rocksdb key [%#v]: rocksdb is closed", key)
	}*/
	rocksdbValue, err := dbInst.db.Get(dbInst.readOpts, key)
	if err != nil {
		logger.Errorf("Error retrieving rocksdb key [%#v]: %s", key, err)
		return nil, errors.Wrapf(err, "error retrieving rocksdb key [%#v]", key)
	}
	valueData := rocksdbValue.Data()
	value := make([]byte, len(valueData))
	copy(value, valueData)
	rocksdbValue.Free()
	logger.Debugf("got data [%s]", value)
	if len(value) == 0 {
		return nil, nil //module require asserts nil for empty byte slices
	}
	return value, nil
}

// Put saves the key/value
func (dbInst *DB) Put(key []byte, value []byte, sync bool) error {
	dbInst.mutex.RLock()
	defer dbInst.mutex.RUnlock()
	wo := rocksdb.NewDefaultWriteOptions()
	if sync {
		wo.SetSync(true)
	}
	err := dbInst.db.Put(wo, key, value)
	if err != nil {
		logger.Errorf("Error writing rocksdb key [%#v]", key)
		return errors.Wrapf(err, "error writing rocksdb key [%#v]", key)
	}
	return nil
}

// Delete deletes the given key
func (dbInst *DB) Delete(key []byte, sync bool) error {
	dbInst.mutex.RLock()
	defer dbInst.mutex.RUnlock()
	wo := rocksdb.NewDefaultWriteOptions()
	if sync {
		wo.SetSync(true)
	}
	err := dbInst.db.Delete(wo, key)
	if err != nil {
		logger.Errorf("Error deleting rocksdb key [%#v]", key)
		return errors.Wrapf(err, "error deleting rocksdb key [%#v]", key)
	}
	return nil
}

// IteratorHelper extends actual rocksdb iterator
type IteratorHelper struct {
	*rocksdb.Iterator
	ro *rocksdb.ReadOptions //we have to destroy them after iterator will be closed
}

// GetIterator returns an iterator over key-value store. The iterator should be closed after the use.
// The resultset contains all the keys that are present in the db between the startKey (inclusive) and the endKey (exclusive).
// A nil startKey represents the first available key and a nil endKey represent a logical key after the last available key
func (dbInst *DB) GetIterator(startKey []byte, endKey []byte) (*IteratorHelper, error) {
	logger.Debugf("Getting new RocksDB Iterator... for start key: [%s (%+v) and end key: [%s (%+v)]", startKey, startKey, endKey, endKey) //TODO: delete this
	if dbInst.dbState == closed {
		err := errors.New("error while obtaining db iterator: rocksdb: closed")
		logger.Debugf("itr.Err()=[%+v]. Impossible to create an iterator", err)
		return nil, err
	}
	//ro := dbInst.readOpts
	ro := rocksdb.NewDefaultReadOptions()
	// Docs says that If you want to avoid disturbing your live traffic
	// while doing the bulk read, be sure to call SetFillCache(false)
	// on the ReadOptions you use when creating the Iterator.
	ro.SetFillCache(false)
	ro.SetBackgroundPurgeOnIteratorCleanup(true)
	///	dbInst.mutex.RUnlock()
	if endKey != nil {
		logger.Debugf("if-case: endKey!=nil, UpperBound set")
		ro.SetIterateUpperBound(endKey)
	} else {
		logger.Debug("endKey is nil, no UpperBound would be set in the previous variant")
		ro.SetIterateUpperBound(endKey)

	}
	ni := dbInst.db.NewIterator(ro)
	ni.Seek(startKey)
	return &IteratorHelper{
			Iterator: ni,
			ro:       ro},
		nil
}

// WriteBatch writes a batch
func (dbInst *DB) WriteBatch(batch *rocksdb.WriteBatch, sync bool) error {
	logger.Debugf("WritingBatch.Count()=[%d]", batch.Count()) //TODO: delete this
	dbInst.mutex.RLock()
	defer dbInst.mutex.RUnlock()
	wo := rocksdb.NewDefaultWriteOptions()
	if sync {
		wo.SetSync(true)
	}
	if err := dbInst.db.Write(wo, batch); err != nil {
		return errors.Wrap(err, "error writing batch to rocksdb")
	}
	return nil
}

// FileLock encapsulate the DB that holds the file lock.
// As the FileLock to be used by a single process/goroutine,
// there is no need for the semaphore to synchronize the
// FileLock usage.
type FileLock struct {
	db       *rocksdb.DB
	filePath string
}

// NewFileLock returns a new file based lock manager.
func NewFileLock(filePath string) *FileLock {
	return &FileLock{
		filePath: filePath,
	}
}

// Lock acquire a file lock. We achieve this by opening
// a db for the given filePath. Internally, leveldb acquires a
// file lock while opening a db. If the db is opened again by the same or
// another process, error would be returned. When the db is closed
// or the owner process dies, the lock would be released and hence
// the other process can open the db. We exploit this leveldb
// functionality to acquire and release file lock as the leveldb
// supports this for Windows, Solaris, and Unix.
func (f *FileLock) Lock() error {
	dbOpts := rocksdb.NewDefaultOptions()
	var err error
	var dirEmpty bool
	var db *rocksdb.DB
	if f.IsLocked() {
		return errors.Errorf("lock is already acquired")
	}
	if dirEmpty, err = fileutil.CreateDirIfMissing(f.filePath); err != nil {
		panic(fmt.Sprintf("Error creating dir if missing: %s", err))
	}
	logger.Debugf("while Lock dirEmpty = %+v", dirEmpty)
	dbOpts.SetCreateIfMissing(dirEmpty)
	db, err = rocksdb.OpenDb(dbOpts, f.filePath)
	errString := fmt.Sprintln(err)
	//next if was added to pass the test TestFileLock from this package tests
	if err != nil && errString[:38] == "IO error: lock hold by current process" {
		return errors.Errorf("lock is already acquired on file %s", f.filePath)
	}
	if err != nil {
		panic(fmt.Sprintf("Error opening rocksdb: %s", err))
	}
	logger.Infof("RocksDB was successfully opened while Locking")
	if err != nil && err == syscall.EAGAIN {
		return errors.Errorf("lock is already acquired on file %s", f.filePath)
	}
	if err != nil {
		panic(fmt.Sprintf("Error acquiring lock on file %s: %s", f.filePath, err))
	}

	// only mutate the lock db reference AFTER validating that the lock was held.
	f.db = db

	return nil
}

// Determine if the lock is currently held open.
func (f *FileLock) IsLocked() bool {
	logger.Debugf("IsLocked = %+v", f.db != nil)
	return f.db != nil
}

// Unlock releases a previously acquired lock. We achieve this by closing
// the previously opened db. FileUnlock can be called multiple times.
func (f *FileLock) Unlock() {
	if f.db == nil {
		return
	}
	f.db.Close()
	f.db = nil
	logger.Debugf("FileLock successfully unlocked!")
}
