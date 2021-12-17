package rocksdbhelper

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/hyperledger/fabric/common/ledger/dataformat"
	rocksdb "github.com/line/gorocksdb"
	"github.com/pkg/errors"
)

const (
	// internalDBName is used to keep track of data related to internals such as data format
	// _ is used as name because this is not allowed as a channelname
	internalDBName = "_"
	// maxBatchSize limits the memory usage (1MB) for a batch. It is measured by the total number of bytes
	// of all the keys in a batch.
	maxBatchSize = 1000000
)

var (
	dbNameKeySep     = []byte{0x00} //TODO: should we change this to different from leveldb?
	lastKeyIndicator = byte(0x01)
	formatVersionKey = []byte{'f'} // a single key in db whose value indicates the version of the data format
)

// DBHandle is an handle to a named db
type DBHandle struct {
	dbName    string
	db        *DB
	closeFunc closeFunc
}

// closeFunc closes the db handle
type closeFunc func()

// Conf configuration for `Provider`
//
// `ExpectedFormat` is the expected value of the format key in the internal database.
// At the time of opening the db, A check is performed that
// either the db is empty (i.e., opening for the first time) or the value
// of the formatVersionKey is equal to `ExpectedFormat`. Otherwise, an error is returned.
// A nil value for ExpectedFormat indicates that the format is never set and hence there is no such record.
type Conf struct {
	DBPath         string
	ExpectedFormat string
}

// Provider enables to use a single rocksdb as multiple logical leveldbs
type Provider struct {
	db *DB

	mux       sync.Mutex
	dbHandles map[string]*DBHandle
}

// NewProvider constructs a Provider
func NewProvider(conf *Conf) (*Provider, error) {
	logger.Debugf("NewProvider intialization...")
	db, err := openDBAndCheckFormat(conf)
	if err != nil {
		return nil, err
	}
	return &Provider{
		db:        db,
		dbHandles: make(map[string]*DBHandle),
	}, nil
}

func openDBAndCheckFormat(conf *Conf) (d *DB, e error) {
	logger.Debugf("Opening DB and checking format...")
	db := CreateDB(conf)
	db.Open()

	defer func() {
		if e != nil {
			logger.Infof("Closing RocksDB...")
			db.Close()
		}
	}()

	internalDB := &DBHandle{
		db:     db,
		dbName: internalDBName,
	}

	dbEmpty, err := db.IsEmpty()
	if err != nil {
		return nil, err
	}
	logger.Infof("rocks db IsEmpty()=%t", dbEmpty) //TODO: delete this

	if dbEmpty && conf.ExpectedFormat != "" {
		logger.Infof("DB is empty Setting db format as %s", conf.ExpectedFormat)
		if err := internalDB.Put(formatVersionKey, []byte(conf.ExpectedFormat), true); err != nil {
			return nil, err
		}
		logger.Infof("formatVersionKey succesfully put into a rocksdb") //TODO: delete this
		return db, nil
	}

	formatVersion, err := internalDB.Get(formatVersionKey)
	if err != nil {
		return nil, err
	}
	logger.Debugf("Checking for db format at path [%s]", conf.DBPath)

	if !bytes.Equal(formatVersion, []byte(conf.ExpectedFormat)) {
		logger.Errorf("The db at path [%s] contains data in unexpected format. expected data format = [%s] (%#v), data format = [%s] (%#v).",
			conf.DBPath, conf.ExpectedFormat, []byte(conf.ExpectedFormat), formatVersion, formatVersion)
		return nil, &dataformat.ErrFormatMismatch{
			ExpectedFormat: conf.ExpectedFormat,
			Format:         string(formatVersion),
			DBInfo:         fmt.Sprintf("rocksdb at [%s]", conf.DBPath),
		}
	}
	logger.Debug("format is latest, nothing to do")
	return db, nil
}

// GetDataFormat returns the format of the data
func (p *Provider) GetDataFormat() (string, error) {
	f, err := p.GetDBHandle(internalDBName).Get(formatVersionKey)
	return string(f), err
}

// GetDBHandle returns a handle to a named db
func (p *Provider) GetDBHandle(dbName string) *DBHandle {
	p.mux.Lock()
	defer p.mux.Unlock()
	dbHandle := p.dbHandles[dbName]
	if dbHandle == nil {
		closeFunc := func() {
			p.mux.Lock()
			defer p.mux.Unlock()
			delete(p.dbHandles, dbName)
		}
		dbHandle = &DBHandle{dbName, p.db, closeFunc}
		p.dbHandles[dbName] = dbHandle
	}
	return dbHandle
}

// Close closes the underlying leveldb
func (p *Provider) Close() {
	p.db.Close()
}

// Drop drops all the data for the given dbName
func (p *Provider) Drop(dbName string) error {
	dbHandle := p.GetDBHandle(dbName)
	defer dbHandle.Close()
	return dbHandle.deleteAll()
}

// Get returns the value for the given key
func (h *DBHandle) Get(key []byte) ([]byte, error) {
	return h.db.Get(constructRocksKey(h.dbName, key))
}

// Put saves the key/value
func (h *DBHandle) Put(key []byte, value []byte, sync bool) error {
	return h.db.Put(constructRocksKey(h.dbName, key), value, sync)
}

// Delete deletes the given key
func (h *DBHandle) Delete(key []byte, sync bool) error {
	return h.db.Delete(constructRocksKey(h.dbName, key), sync)
}

// DeleteAll deletes all the keys that belong to the channel (dbName).
func (h *DBHandle) deleteAll() error {
	iter, err := h.GetIterator(nil, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	// use leveldb iterator directly to be more efficient

	// This is common code shared by all the leveldb instances. Because each leveldb has its own key size pattern,
	// each batch is limited by memory usage instead of number of keys. Once the batch memory usage reaches maxBatchSize,
	// the batch will be committed.
	numKeys := 0
	batchSize := 0
	batch := rocksdb.NewWriteBatch()
	for dbIter := iter.Iterator; dbIter.Valid(); dbIter.Next() {
		if err := dbIter.Err(); err != nil {
			return errors.Wrap(err, "internal rocksdb error while retrieving data from db iterator")
		}
		rocksdbKey := dbIter.Key()
		key := rocksdbKey.Data()
		rocksdbKey.Free()
		numKeys++
		batchSize = batchSize + len(key)
		batch.Delete(key)
		if batchSize >= maxBatchSize {
			if err := h.db.WriteBatch(batch, true); err != nil {
				return err
			}
			logger.Infof("Have removed %d entries for channel %s in rocksdb %s", numKeys, h.dbName, h.db.conf.DBPath)
			batchSize = 0
			batch.Clear()
		}
	}
	if batch.Count() > 0 {
		return h.db.WriteBatch(batch, true)
	}
	return nil
}

// IsEmpty returns true if no data exists for the DBHandle
func (h *DBHandle) IsEmpty() (bool, error) {
	logger.Info("IsEmpty(), getting Iterator with nil start&end keys...")
	itr, err := h.GetIterator(nil, nil)
	if err != nil {
		return false, err
	}
	defer itr.Close()

	if err := itr.Err(); err != nil {
		return false, errors.WithMessagef(itr.Err(), "internal rocksdb error while obtaining next entry from iterator")
	}

	return !itr.Valid(), nil
}

// NewUpdateBatch returns a new UpdateBatch that can be used to update the db
func (h *DBHandle) NewUpdateBatch() *UpdateBatch {
	wb := rocksdb.NewWriteBatch()
	return &UpdateBatch{
		dbName:     h.dbName,
		WriteBatch: wb,
	}
}

// WriteBatch writes a batch in an atomic way
func (h *DBHandle) WriteBatch(batch *UpdateBatch, sync bool) error {
	if batch == nil || batch.Count() == 0 {
		return nil
	}
	if h.db.dbState == closed {
		return fmt.Errorf("error writing batch to rocksdb")
	}
	logger.Infof("WriteBatch()..., sync=[%+v]", sync)
	if err := h.db.WriteBatch(batch.WriteBatch, sync); err != nil {
		return err
	}
	return nil
}

// GetIterator gets an handle to iterator. The iterator should be released after the use.
// The resultset contains all the keys that are present in the db between the startKey (inclusive) and the endKey (exclusive).
// A nil startKey represents the first available key and a nil endKey represent a logical key after the last available key
func (h *DBHandle) GetIterator(startKey []byte, endKey []byte) (*Iterator, error) {
	eKey := constructRocksKey(h.dbName, endKey)
	sKey := constructRocksKey(h.dbName, startKey)
	if endKey == nil {
		logger.Info("endKey is nil")
		// replace the last byte 'dbNameKeySep' by 'lastKeyIndicator'
		eKey[len(eKey)-1] = lastKeyIndicator
		logger.Infof("endKey is nil: eKey=[%s(%#v)]", eKey, eKey)
	} else {
		logger.Infof("endKey is not nil: (%#v)", endKey)

	}
	logger.Infof("Constructing iterator with sKey=[%s(%#v)] and eKey=[%s(%#v)]", sKey, sKey, eKey, eKey)
	itr, err := h.db.GetIterator(sKey, eKey)
	if err != nil {
		logger.Infof("Error! Closing iterator...")
		return nil, err
	}
	if itr.Valid() {
		logger.Infof("itr is Valid")
	} else {
		logger.Infof("itr is not Valid")
	}
	return &Iterator{h.dbName, itr}, nil
}

// Close closes the DBHandle after its db data have been deleted
func (h *DBHandle) Close() {
	if h.closeFunc != nil {
		h.closeFunc()
	}
}

// UpdateBatch encloses the details of multiple `updates`
type UpdateBatch struct {
	*rocksdb.WriteBatch
	dbName string
}

// Put adds a KV
func (b *UpdateBatch) Put(key []byte, value []byte) {
	if value == nil {
		panic("Nil value not allowed")
	}
	b.WriteBatch.Put(constructRocksKey(b.dbName, key), value)
}

// Delete deletes a Key and associated value
func (b *UpdateBatch) Delete(key []byte) {
	b.WriteBatch.Delete(constructRocksKey(b.dbName, key))
}

// Iterator extends actual rocksdb iterator
type Iterator struct {
	dbName string
	*rocksdb.Iterator
}

//Next wraps actual rocksdb iterator method.
//It prevents a fatal error when Next() called after the last db key
//if iterator.Valid() == false
func (itr *Iterator) Next() {
	if itr.Iterator.Valid() {
		//itr.FreeKey()
		//itr.FreeValue()
		itr.Iterator.Next()
	} else {
		logger.Infof("iterator is not valid anymore")
	}
	if err := itr.Iterator.Err(); err != nil {
		logger.Infof("Error during iteration: %s", err)
	}
}

// Key wraps actual rocksdb iterator method
func (itr *Iterator) Key() []byte {
	rocksdbKey := itr.Iterator.Key()
	key := rocksdbKey.Data()
	rocksdbKey.Free()
	return retrieveAppKey(key)
}

// Key wraps actual rocksdb iterator method
func (itr *Iterator) Value() []byte {
	rocksdbValue := itr.Iterator.Value()
	value := rocksdbValue.Data()
	rocksdbValue.Free()
	return value
}

// FreeKey wraps actual freeing the Key slice data
/*func (itr *Iterator) FreeKey() {
	itr.Iterator.Key().Free()
}

// FreeValue wraps actual freeing the Value slice data
func (itr *Iterator) FreeValue() {
	itr.Iterator.Value().Free()
}*/

// Seek moves the iterator to the first key/value pair
// whose key is greater than or equal to the given key.
// It returns whether such pair exist.
func (itr *Iterator) Seek(key []byte) {
	rocksKey := constructRocksKey(itr.dbName, key)
	itr.Iterator.Seek(rocksKey)
}

//TODO: should we make the mechanism differ from level db one?
func constructRocksKey(dbName string, key []byte) []byte {
	return append(append([]byte(dbName), dbNameKeySep...), key...)
}

func retrieveAppKey(levelKey []byte) []byte {
	return bytes.SplitN(levelKey, dbNameKeySep, 2)[1]
}
