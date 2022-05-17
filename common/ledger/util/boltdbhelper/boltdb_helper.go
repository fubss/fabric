package boltdbhelper

import (
	"fmt"
	"sync"
	"syscall"
	"time"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/internal/fileutil"
	"github.com/pkg/errors"

	bolt "go.etcd.io/bbolt"
)

var logger = flogging.MustGetLogger("boltdbhelper")

type dbState int32

const (
	closed dbState = iota
	opened
)

// DB - a wrapper on an actual store
type DB struct {
	conf     *Conf
	db       *bolt.DB
	dbState  dbState
	mutex    sync.RWMutex
	activeTx []*bolt.Tx //for active boltdb transactions
	//readOpts        *opt.ReadOptions
	//writeOptsNoSync *opt.WriteOptions
	//writeOptsSync   *opt.WriteOptions
}

// CreateDB constructs a `DB`
func CreateDB(conf *Conf) *DB {
	//readOpts := &opt.ReadOptions{}
	//writeOptsNoSync := &opt.WriteOptions{}
	//writeOptsSync := &opt.WriteOptions{}
	//writeOptsSync.Sync = true

	return &DB{
		conf:    conf,
		dbState: closed,
		//readOpts:        readOpts,
		//writeOptsNoSync: writeOptsNoSync,
		//writeOptsSync:   writeOptsSync,
	}
}

// Open opens the underlying db
func (dbInst *DB) Open() {
	dbInst.mutex.Lock()
	defer dbInst.mutex.Unlock()
	if dbInst.dbState == opened {
		return
	}
	//dbOpts := &opt.Options{}
	dbPath := dbInst.conf.DBPath
	var err error

	if _, err := fileutil.CreateDirIfMissing(dbPath); err != nil {
		panic(fmt.Sprintf("Error creating dir if missing: %s", err))
	}
	//db will be created if it doesn't exist.
	dbInst.db, err = bolt.Open(dbPath+"/bolt.db", 0666, nil)
	if err != nil {
		panic(fmt.Sprintf("Error opening boltdb: %s", err))
	}
	dbInst.dbState = opened
	logger.Debugf("dbpath is: %s/bolt.db", dbPath)
	// We have to create global bucket for being able to check db with function IsEmpty().
	// TODO: to decide what is better - one large global bucket or many different buckets
	// (e.g. personal bucket for every chaincode's data).
	err = dbInst.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("fabric"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		panic(fmt.Sprintf("%s", err))
	}
}

// IsEmpty returns whether or not a database is empty
func (dbInst *DB) IsEmpty() (bool, error) {
	var hasItems bool
	dbInst.mutex.RLock()
	defer dbInst.mutex.RUnlock()
	err := dbInst.db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("fabric"))

		c := b.Cursor()

		k, v := c.First()
		// If the bucket is empty then a nil key and value are returned by First()
		if k == nil && v == nil {
			hasItems = true
		} else {
			hasItems = false
		}
		return nil
	})

	return hasItems,
		errors.Wrapf(err, "error while trying to see if the boltdb at path [%s] is empty", dbInst.conf.DBPath)
}

// Close closes the underlying db
func (dbInst *DB) Close() {
	dbInst.mutex.Lock()
	defer dbInst.mutex.Unlock()
	if dbInst.dbState == closed {
		return
	}
	if err := dbInst.db.Close(); err != nil {
		logger.Errorf("Error closing boltdb: %s", err)
	}
	dbInst.dbState = closed
}

// Get returns the value for the given key
func (dbInst *DB) Get(key []byte) ([]byte, error) {
	var value []byte
	dbInst.mutex.RLock()
	defer dbInst.mutex.RUnlock()

	err := dbInst.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("fabric"))
		v := b.Get(key)
		var tmpv = make([]byte, len(v))
		copy(tmpv, v)
		value = tmpv
		return nil
	})
	if err != nil {
		logger.Errorf("Error retrieving boltdb key [%#v]: %s", key, err)
		return nil, errors.Wrapf(err, "error retrieving boltdb key [%#v]", key)
	}
	if value == nil { //likewise leveldb.ErrNotFound
		value = nil
		err = nil
	}

	return value, nil
}

// Put saves the key/value
func (dbInst *DB) Put(key []byte, value []byte, sync bool) error {
	dbInst.mutex.RLock()
	defer dbInst.mutex.RUnlock()
	if !sync {
		dbInst.setSyncOffOnce()
	}

	// Create a bucket and a key.
	err := dbInst.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("fabric"))
		if err != nil {
			return err
		}
		if err := b.Put(key, value); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		logger.Errorf("Error writing boltdb key [%#v]", key)
		return errors.Wrapf(err, "error writing boltdb key [%#v]", key)
	}
	return nil
}

// Delete deletes the given key
func (dbInst *DB) Delete(key []byte, sync bool) error {
	dbInst.mutex.RLock()
	defer dbInst.mutex.RUnlock()
	if !sync {
		dbInst.setSyncOffOnce()
	}
	err := dbInst.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("fabric")).Delete(key)
	})
	if err != nil {
		logger.Errorf("Error deleting boltdb key [%#v]", key)
		return errors.Wrapf(err, "error deleting boltdb key [%#v]", key)
	}
	return nil
}

// IteratorHelper extends actual rocksdb iterator
type IteratorHelper struct {
	*bolt.Cursor
}

// GetIterator returns an iterator over key-value store. The iterator should be released after the use.
// The resultset contains all the keys that are present in the db between the startKey (inclusive) and the endKey (exclusive).
// A nil startKey represents the first available key and a nil endKey represent a logical key after the last available key
// TODO: close iterator somewhere
func (dbInst *DB) GetIterator(startKey []byte, endKey []byte) (*bolt.Cursor, *bolt.Tx, []byte, []byte, error) {
	//logger.Debugf("Getting new BoltDB Iterator... for start key: [%s (%+v) and end key: [%s (%+v)]", startKey, startKey, endKey, endKey)
	dbInst.mutex.RLock()
	defer dbInst.mutex.RUnlock()
	tx, err := dbInst.db.Begin(false)
	if err != nil {
		logger.Errorf("Error after creating boltdb tx: %+v", err)
		return nil, tx, nil, nil, err
	}
	c := tx.Bucket([]byte("fabric")).Cursor()
	seekedKey, seekedValue := c.Seek(startKey)
	//return dbInst.db.NewIterator(&goleveldbutil.Range{Start: startKey, Limit: endKey}, dbInst.readOpts)
	//dbInst.activeTx = append()
	return c, tx, seekedKey, seekedValue, nil
}

// WriteBatch writes a boltdb tx, not a boltdb batch.
// Because batch in boltdb is only useful when
// several goroutines calling it.
func (dbInst *DB) WriteBatch(tx *bolt.Tx, sync bool) error {
	dbInst.mutex.RLock()
	defer dbInst.mutex.RUnlock()
	if !sync {
		dbInst.setSyncOffOnce()
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "error writing batch to boltdb")
	}
	return nil
}

func (dbInst *DB) setSyncOffOnce() {
	dbInst.db.NoSync = false
	//logger.Debugf("dbInst.db.NoSync = false")
	defer func() {
		dbInst.db.NoSync = true
		//logger.Debugf("dbInst.db.NoSync = true")
	}()
}

// FileLock encapsulate the DB that holds the file lock.
// As the FileLock to be used by a single process/goroutine,
// there is no need for the semaphore to synchronize the
// FileLock usage.
type FileLock struct {
	db       *bolt.DB
	filePath string
}

// NewFileLock returns a new file based lock manager.
func NewFileLock(filePath string) *FileLock {
	return &FileLock{
		filePath: filePath,
	}
}

// Lock acquire a file lock. We achieve this by opening
// a db for the given filePath. Internally, boltdb acquires a
// file lock while opening a db. If the db is opened again by the same or
// another process, error would be returned. When the db is closed
// or the owner process dies, the lock would be released and hence
// the other process can open the db. We exploit this boltdb
// functionality to acquire and release file lock as the boltdb
// supports this for Windows, Solaris, and Unix.
func (f *FileLock) Lock() error {
	//dbOpts := &opt.Options{}
	var err error
	//var dirEmpty bool
	if f.IsLocked() {
		return errors.Errorf("lock is already acquired")
	}
	if _, err = fileutil.CreateDirIfMissing(f.filePath); err != nil {
		panic(fmt.Sprintf("Error creating dir if missing: %s", err))
	}
	//db will be created if it doesn't exist.
	db, err := bolt.Open(f.filePath+"/lock_bolt.db", 0666, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return errors.Errorf("Error opening boltdb: %s", err)
	}
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
	return f.db != nil
}

// Unlock releases a previously acquired lock. We achieve this by closing
// the previously opened db. FileUnlock can be called multiple times.
func (f *FileLock) Unlock() {
	if f.db == nil {
		return
	}
	if err := f.db.Close(); err != nil {
		logger.Warningf("unable to release the lock on file %s: %s", f.filePath, err)
		return
	}
	f.db = nil
}
