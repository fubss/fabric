package rocksdbhelper

import (
	"fmt"
	"sync"

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

	readOpts        *rocksdb.ReadOptions
	writeOptsNoSync *rocksdb.WriteOptions
	writeOptsSync   *rocksdb.WriteOptions
}

// CreateDB constructs a `DB`
func CreateDB(conf *Conf) *DB {
	logger.Debugf("RocksDB constructing...")
	readOpts := rocksdb.NewDefaultReadOptions()
	writeOptsNoSync := rocksdb.NewDefaultWriteOptions()
	writeOptsSync := rocksdb.NewDefaultWriteOptions()
	writeOptsSync.SetSync(true)
	logger.Debugf("RocksDB constructing successfully finished")
	return &DB{
		conf:            conf,
		dbState:         closed,
		readOpts:        readOpts,
		writeOptsNoSync: writeOptsNoSync,
		writeOptsSync:   writeOptsSync,
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
	dbOpts := rocksdb.NewDefaultOptions()
	dbPath := dbInst.conf.DBPath
	dbOpts.SetCreateIfMissing(true)
	var err error
	if _, err = fileutil.CreateDirIfMissing(dbPath); err != nil {
		panic(fmt.Sprintf("Error creating dir if missing: %s", err))
	}
	if dbInst.db, err = rocksdb.OpenDb(dbOpts, dbPath); err != nil {
		panic(fmt.Sprintf("Error opening rocksdb: %s", err))
	}
	logger.Debugf("DB was successfully opened")
	dbInst.dbState = opened
}

// IsEmpty returns whether or not a database is empty
func (dbInst *DB) IsEmpty() (bool, error) {
	logger.Debugf("Checkin if DB is empty...")
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
	dbInst.db.Close() //TODO: should we check if db closed here?
	dbInst.dbState = closed
}

// Get returns the value for the given key
func (dbInst *DB) Get(key []byte) ([]byte, error) {
	dbInst.mutex.RLock()
	defer dbInst.mutex.RUnlock()
	value, err := dbInst.db.Get(dbInst.readOpts, key)
	if err != nil {
		logger.Errorf("Error retrieving rocksdb key [%#v]: %s", key, err)
		return nil, errors.Wrapf(err, "error retrieving rocksdb key [%#v]", key)
	}
	return value.Data(), nil
}

// Put saves the key/value
func (dbInst *DB) Put(key []byte, value []byte, sync bool) error {
	dbInst.mutex.RLock()
	defer dbInst.mutex.RUnlock()
	wo := dbInst.writeOptsNoSync
	if sync {
		wo = dbInst.writeOptsSync
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
	wo := dbInst.writeOptsNoSync
	if sync {
		wo = dbInst.writeOptsSync
	}
	err := dbInst.db.Delete(wo, key)
	if err != nil {
		logger.Errorf("Error deleting rocksdb key [%#v]", key)
		return errors.Wrapf(err, "error deleting rocksdb key [%#v]", key)
	}
	return nil
}

// GetIterator returns an iterator over key-value store. The iterator should be released after the use.
// The resultset contains all the keys that are present in the db between the startKey (inclusive) and the endKey (exclusive).
// A nil startKey represents the first available key and a nil endKey represent a logical key after the last available key
func (dbInst *DB) GetIterator(startKey []byte, endKey []byte) *rocksdb.Iterator {
	dbInst.mutex.RLock()
	defer dbInst.mutex.RUnlock()
	ro := dbInst.readOpts
	ro.SetIterateLowerBound(startKey)
	ro.SetIterateUpperBound(endKey)
	return dbInst.db.NewIterator(ro)
}

// WriteBatch writes a batch
func (dbInst *DB) WriteBatch(batch *rocksdb.WriteBatch, sync bool) error {
	dbInst.mutex.RLock()
	defer dbInst.mutex.RUnlock()
	wo := dbInst.writeOptsNoSync
	if sync {
		wo = dbInst.writeOptsSync
	}
	if err := dbInst.db.Write(wo, batch); err != nil {
		return errors.Wrap(err, "error writing batch to rocksdb")
	}
	return nil
}
