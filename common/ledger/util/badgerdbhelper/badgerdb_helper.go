/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package badgerdbhelper

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/internal/fileutil"
	"github.com/pkg/errors"
)

var logger = flogging.MustGetLogger("leveldbhelper")

type dbState int32

const (
	// ManifestFilename is the filename for the manifest file.
	ManifestFilename = "MANIFEST"

	closed dbState = iota
	opened
)

// DB - a wrapper on an actual store
type DB struct {
	conf    *Conf
	db      *badger.DB
	dbState dbState
	mutex   sync.RWMutex

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
	dbPath := dbInst.conf.DBPath
	var err error
	var dirEmpty bool
	if dirEmpty, err = fileutil.CreateDirIfMissing(dbPath); err != nil {
		panic(fmt.Sprintf("Error creating dir if missing: %s", err))
	}
	if !dirEmpty {
		if _, err := os.Stat(filepath.Join(dbPath, ManifestFilename)); err != nil {
			panic(fmt.Sprintf("Error opening badgerdb: %s", err))
		}
	}
	if dbInst.db, err = badger.Open(badger.DefaultOptions(dbPath)); err != nil {
		panic(fmt.Sprintf("Error opening badgerdb: %s", err))
	}
	dbInst.dbState = opened
}

// IsEmpty returns whether or not a database is empty
func (dbInst *DB) IsEmpty() (bool, error) {
	dbInst.mutex.RLock()
	defer dbInst.mutex.RUnlock()
	var hasItems bool
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	err := dbInst.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(opts)
		defer it.Close()
		it.Rewind()
		hasItems = it.Valid()
		return nil
	})
	return !hasItems,
		errors.Wrapf(err, "error while trying to see if the badgerdb at path [%s] is empty", dbInst.conf.DBPath)
}

// Close closes the underlying db
func (dbInst *DB) Close() {
	dbInst.mutex.Lock()
	defer dbInst.mutex.Unlock()
	if dbInst.dbState == closed {
		return
	}
	if err := dbInst.db.Close(); err != nil {
		logger.Errorf("Error closing badgerdb: %s", err)
	}
	dbInst.dbState = closed
}

// Get returns the value for the given key
func (dbInst *DB) Get(key []byte) ([]byte, error) {
	dbInst.mutex.RLock()
	defer dbInst.mutex.RUnlock()
	var value []byte
	err := dbInst.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		value, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}
		return nil
	})
	if errors.Is(err, badger.ErrKeyNotFound) {
		value = nil
		err = nil
	}
	if err != nil {
		logger.Errorf("Error retrieving badgerdb key [%#v]: %s", key, err)
		return nil, errors.Wrapf(err, "error retrieving badgerdb key [%#v]", key)
	}
	return value, nil
}

// Put saves the key/value
func (dbInst *DB) Put(key []byte, value []byte, sync bool) error {
	dbInst.mutex.RLock()
	defer dbInst.mutex.RUnlock()
	//wo := dbInst.writeOptsNoSync
	/*if sync {
		wo = dbInst.writeOptsSync
	}*/
	err := dbInst.db.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, value)
		return err
	})
	if err != nil {
		logger.Errorf("Error writing badgerdb key [%#v]", key)
		return errors.Wrapf(err, "error writing badgerdb key [%#v]", key)
	}
	return nil
}

// Delete deletes the given key
func (dbInst *DB) Delete(key []byte, sync bool) error {
	dbInst.mutex.RLock()
	defer dbInst.mutex.RUnlock()
	/*wo := dbInst.writeOptsNoSync
	if sync {
		wo = dbInst.writeOptsSync
	}*/
	err := dbInst.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete(key)
		return err
	})
	if err != nil {
		logger.Errorf("Error deleting badgerdb key [%#v]", key)
		return errors.Wrapf(err, "error deleting badgerdb key [%#v]", key)
	}
	return nil
}

type RangeIterator struct {
	iterator *badger.Iterator
	startKey []byte
	endKey   []byte
}

// Next() wraps Badger's functions to make function similar Leveldb Next
/*func (itr *RangeIterator) Next() bool {

	// Check does iterator need start from startKey
	if itr.justOpened {
		itr.iterator.Seek(itr.startKey)
		itr.justOpened = false
	}
	itr.iterator.Next()
	return itr.iterator.Valid()
}*/

// Key() wraps Badger's functions to make function similar Leveldb Key
func (itr *RangeIterator) Key() []byte {
	return itr.iterator.Item().Key()
}

// GetIterator returns an iterator over key-value store. The iterator should be released after the use.
// The resultset contains all the keys that are present in the db between the startKey (inclusive) and the endKey (exclusive).
// A nil startKey represents the first available key and a nil endKey represent a logical key after the last available key
func (dbInst *DB) GetIterator(startKey []byte, endKey []byte) RangeIterator {
	dbInst.mutex.RLock()
	defer dbInst.mutex.RUnlock()
	txn := dbInst.db.NewTransaction(true)
	defer txn.Discard()
	/*err := dbInst.db.View(func(txn *badger.Txn) error {
		i = txn.NewKeyIterator(startKey, badger.DefaultIteratorOptions)
		defer i.Close()
		return nil
	})
	if err != nil {
		logger.Errorf("Error getting badgerdb iterator with startKey [%#v] and endKey [%#v]", startKey, endKey)
		panic(errors.Wrapf(err, "error getting badgerdb iterator with startKey [%#v] and endKey [%#v]", startKey, endKey))
	}*/
	itr := txn.NewIterator(badger.DefaultIteratorOptions)
	defer itr.Close()
	return RangeIterator{
		iterator: itr,
		startKey: startKey,
		endKey:   endKey,
	}
}

// WriteBatch writes a batch
func (dbInst *DB) WriteBatch(batch *badger.WriteBatch, sync bool) error {
	dbInst.mutex.RLock()
	defer dbInst.mutex.RUnlock()
	/*wo := dbInst.writeOptsNoSync
	if sync {
		wo = dbInst.writeOptsSync
	}*/
	if err := batch.Flush(); err != nil {
		return errors.Wrap(err, "error writing batch to badgerdb")
	}
	return nil
}

// FileLock encapsulate the DB that holds the file lock.
// As the FileLock to be used by a single process/goroutine,
// there is no need for the semaphore to synchronize the
// FileLock usage.
type FileLock struct {
	db       *badger.DB
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
	//dbOpts := &opt.Options{}
	var err error
	var dirEmpty bool
	if dirEmpty, err = fileutil.CreateDirIfMissing(f.filePath); err != nil {
		panic(fmt.Sprintf("Error creating dir if missing: %s", err))
	}
	if !dirEmpty {
		if _, err := os.Stat(filepath.Join(f.filePath, ManifestFilename)); err != nil {
			panic(fmt.Sprintf("Error opening badgerdb: %s", err))
		}
	}
	db, err := badger.Open(badger.DefaultOptions(f.filePath))
	if fmt.Sprint(err) == fmt.Sprintf("Cannot acquire directory lock on \"%s\".  Another process is using this Badger database. error: resource temporarily unavailable", f.filePath) {
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
