package boltdbhelper

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/hyperledger/fabric/common/ledger/dataformat"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

const (
	// internalDBName is used to keep track of data related to internals such as data format
	// _ is used as name because this is not allowed as a channelname
	internalDBName = "_"
	// maxBatchSize limits the memory usage (1MB) for a batch. It is measured by the total number of bytes
	// of all the keys in a batch.
	//maxBatchSize = 1000000
)

var (
	dbNameKeySep     = []byte{0x00}
	lastKeyIndicator = byte(0x01)
	formatVersionKey = []byte{'f'} // a single key in db whose value indicates the version of the data format
)

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

// Provider enables to use a single leveldb as multiple logical leveldbs
type Provider struct {
	db *DB

	mux       sync.Mutex
	dbHandles map[string]*DBHandle
}

// NewProvider constructs a Provider
func NewProvider(conf *Conf) (*Provider, error) {
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
	db := CreateDB(conf)
	db.Open()

	defer func() {
		if e != nil {
			db.Close()
		}
	}()

	internalDB := &DBHandle{
		db:     db,
		dbName: internalDBName,
		DbType: "boltdb",
	}

	dbEmpty, err := db.IsEmpty()
	if err != nil {
		return nil, err
	}

	if dbEmpty && conf.ExpectedFormat != "" {
		logger.Infof("DB is empty Setting db format as %s", conf.ExpectedFormat)
		if err := internalDB.Put(formatVersionKey, []byte(conf.ExpectedFormat), true); err != nil {
			return nil, err
		}
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
			DBInfo:         fmt.Sprintf("leveldb at [%s]", conf.DBPath),
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
		dbHandle = &DBHandle{dbName, p.db, closeFunc, "boltdb"}
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

// DBHandle is an handle to a named db
type DBHandle struct {
	dbName    string
	db        *DB
	closeFunc closeFunc
	DbType    string // for testing. TODO: delete this
}

// Get returns the value for the given key
func (h *DBHandle) Get(key []byte) ([]byte, error) {
	val, err := h.db.Get(constructLevelKey(h.dbName, key))
	if len(val) == 0 {
		val = nil
	}
	return val, err
}

// Put saves the key/value
func (h *DBHandle) Put(key []byte, value []byte, sync bool) error {
	return h.db.Put(constructLevelKey(h.dbName, key), value, sync)
}

// Delete deletes the given key
func (h *DBHandle) Delete(key []byte, sync bool) error {
	return h.db.Delete(constructLevelKey(h.dbName, key), sync)
}

// createBatchTx creates a usual bboltdb tx
// as a replacement for a leveldb batch
func (h *DBHandle) createBatchTx() (*bolt.Tx, error) {
	// Start a writable transaction.
	tx, err := h.db.db.Begin(true)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// DeleteAll deletes all the keys that belong to the channel (dbName).
func (h *DBHandle) deleteAll() error {
	iter, err := h.GetIterator(nil, nil)
	if err != nil {
		return err
	}
	defer iter.Release()

	// This is common code shared by all the leveldb instances. Because each leveldb has its own key size pattern,
	// each batch is limited by memory usage instead of number of keys. Once the batch memory usage reaches maxBatchSize,
	// the batch will be committed.
	numKeys := 0
	batchSize := 0
	batchTx, err := h.createBatchTx()
	defer batchTx.Rollback()
	if err != nil {
		return err
	}
	iter.Seek(iter.lowerBound)
	iter.seekJustHappened = true
	for iter.Next() {
		numKeys++
		batchSize = batchSize + len(iter.key)
		batchTx.Bucket([]byte("fabric")).Delete(iter.key)
	}
	if err := h.db.WriteBatch(batchTx, true); err != nil {
		return err
	}
	// next leveldb part was commented because it's not clear
	// whether it is related to boltdb transactions
	/*if batchTx.Len() > 0 {
		return h.db.WriteBatch(batch, true)
	}*/
	return nil
}

// IsEmpty returns true if no data exists for the DBHandle
func (h *DBHandle) IsEmpty() (bool, error) {
	itr, err := h.GetIterator(nil, nil)
	if err != nil {
		return false, err
	}
	defer itr.Release()

	//it seems that no error possible
	/*if err := itr.Error(); err != nil {
		return false, errors.WithMessagef(itr.Error(), "internal leveldb error while obtaining next entry from iterator")
	}*/

	return !itr.Next(), nil
}

// NewUpdateBatch returns a new UpdateBatch that can be used to update the db
func (h *DBHandle) NewUpdateBatch() *UpdateBatch {
	batchTx, err := h.createBatchTx()
	if err != nil {
		logger.Errorf("Error creating batch (boltdb tx): %+v", err)
	}
	return &UpdateBatch{
		dbName: h.dbName,
		Tx:     batchTx,
	}
}

// WriteBatch writes a batch in an atomic way
func (h *DBHandle) WriteBatch(batch *UpdateBatch, sync bool) error {
	if batch == nil {
		return nil
	}
	if err := h.db.WriteBatch(batch.Tx, sync); err != nil {
		return err
	}
	return nil
}

// GetIterator gets an handle to iterator (actually bolt db cursor). The iterator should be released after the use.
// The resultset contains all the keys that are present in the db between the startKey (inclusive) and the endKey (exclusive).
// A nil startKey represents the first available key and a nil endKey represent a logical key after the last available key
func (h *DBHandle) GetIterator(startKey []byte, endKey []byte) (*Iterator, error) {
	sKey := constructLevelKey(h.dbName, startKey)
	eKey := constructLevelKey(h.dbName, endKey)
	if endKey == nil {
		// replace the last byte 'dbNameKeySep' by 'lastKeyIndicator'
		eKey[len(eKey)-1] = lastKeyIndicator
	}
	//logger.Debugf("Getting iterator for range [%#v] - [%#v]", sKey, eKey)
	c, tx, seekedKey, seekedValue, err := h.db.GetIterator(sKey, eKey)
	if err != nil {
		if tx != nil {
			tx.Rollback()
		}
		return nil, errors.Wrapf(err, "internal boltdb error while obtaining db iterator")
	}
	return &Iterator{
			dbName:           h.dbName,
			Cursor:           c,
			Tx:               tx,
			key:              seekedKey,
			value:            seekedValue,
			lowerBound:       sKey,
			upperBound:       eKey,
			seekJustHappened: true},
		nil
}

// Close closes the DBHandle after its db data have been deleted
func (h *DBHandle) Close() {
	if h.closeFunc != nil {
		h.closeFunc()
	}
}

// UpdateBatch encloses the details of multiple `updates`
type UpdateBatch struct {
	*bolt.Tx
	dbName string
}

// Put adds a KV
func (b *UpdateBatch) Put(key []byte, value []byte) {
	if value == nil {
		panic("Nil value not allowed")
	}
	b.Tx.Bucket([]byte("fabric")).Put(constructLevelKey(b.dbName, key), value)
}

// Delete deletes a Key and associated value
func (b *UpdateBatch) Delete(key []byte) {
	b.Tx.Bucket([]byte("fabric")).Delete(constructLevelKey(b.dbName, key))
}

// Iterator extends actual leveldb iterator
type Iterator struct {
	dbName string
	*bolt.Cursor
	*bolt.Tx
	key        []byte
	value      []byte
	lowerBound []byte // TODO test this
	upperBound []byte // TODO test this
	// In leveldb when Seek() is directly called or Prev() went over the boundaries,
	// then Next() does omit 0 key and value, where iterator had been pointing after
	// Seek() call.
	// Opposite when GetIterator() is called, Next() does not omit 0 key.
	// So let's make false for every time when we call Iterator.Seek()
	// as well as Iterator.First() and Iterator.Last()
	seekJustHappened bool
}

// Key wraps actual leveldb iterator method
func (itr *Iterator) Key() []byte {
	//TODO: decide whether it is better returning just itr.key
	var tmpKey = make([]byte, len(itr.key))
	copy(tmpKey, itr.key)
	return retrieveAppKey(tmpKey)
}

func (itr *Iterator) Value() []byte {
	//TODO: decide whether it is better returning just itr.value
	var tmpValue = make([]byte, len(itr.value))
	copy(tmpValue, itr.value)
	return tmpValue
}

// Seek moves the iterator to the first key/value pair
// whose key is greater than or equal to the given key.
// It returns whether such pair exist.
func (itr *Iterator) Seek(key []byte) bool {
	levelKey := constructLevelKey(itr.dbName, key)
	itr.key, itr.value = itr.Cursor.Seek(levelKey)
	itr.seekJustHappened = false // see comments to this flag
	if bytes.Compare(levelKey, itr.upperBound) < 0 {
		if bytes.Compare(levelKey, itr.lowerBound) < 0 {
			itr.key, itr.value = itr.Cursor.Seek(itr.lowerBound)
		} else if bytes.Compare(itr.key, itr.upperBound) > 0 {
			return false
		}
		return true
	} else if itr.key != nil || itr.value != nil {
		if bytes.Compare(itr.key, levelKey) >= 0 {
			return true
		}
	}
	return false
}

func (itr *Iterator) Release() {
	itr.Tx.Rollback()
	itr.key = nil
	itr.value = nil
	itr.lowerBound = nil
	itr.upperBound = nil
}

func (itr *Iterator) Next() bool {
	if itr.seekJustHappened { // to get 0 key in boltdb iterator
		itr.seekJustHappened = false
		return true
	}
	itr.key, itr.value = itr.Cursor.Next()
	if itr.key != nil && bytes.Compare(itr.key, itr.upperBound) < 0 && bytes.Compare(itr.key, itr.lowerBound) >= 0 {
		return true
	} else {
		itr.key = nil
		itr.value = nil
		return false
	}
}

func (itr *Iterator) First() bool {
	if itr.lowerBound != nil {
		itr.key, itr.value = itr.Cursor.Seek(itr.lowerBound)
		itr.seekJustHappened = false // see comments to this flag
	} else {
		itr.key, itr.value = itr.Cursor.First()
	}
	return true
}

func (itr *Iterator) Last() bool {
	if itr.upperBound != nil {
		itr.key, itr.value = itr.Cursor.Seek(itr.upperBound)
		itr.seekJustHappened = false // see comments to this flag
	} else {
		itr.key, itr.value = itr.Cursor.Last()
	}
	return true
}

func (itr *Iterator) Prev() bool {
	tmpKey, tmpValue := itr.Cursor.Prev()
	if tmpKey == nil {
		itr.key, itr.value = itr.Cursor.First()
		itr.seekJustHappened = true
		return false // over the begining of the cursor
	} else if itr.lowerBound != nil && bytes.Compare(tmpKey, itr.lowerBound) >= 0 {
		itr.key, itr.value = tmpKey, tmpValue
		return true // inside the boundary conditions
	} else if itr.lowerBound != nil && bytes.Compare(tmpKey, itr.lowerBound) < 0 {
		itr.key, itr.value = itr.Cursor.Seek(itr.lowerBound)
		itr.seekJustHappened = true
		return false // over the lower bound
	} else {
		itr.key, itr.value = tmpKey, tmpValue
		return true // inside the coursor
	}
}

// this was added because boltdb_provider_test.go is used it
func (itr *Iterator) Valid() bool {
	if itr.key != nil && bytes.Compare(itr.key, itr.upperBound) < 0 && bytes.Compare(itr.key, itr.lowerBound) >= 0 {
		return true
	} else {
		return false
	}
}

func constructLevelKey(dbName string, key []byte) []byte {
	return append(append([]byte(dbName), dbNameKeySep...), key...)
}

func retrieveAppKey(levelKey []byte) []byte {
	return bytes.SplitN(levelKey, dbNameKeySep, 2)[1]
}

/*func compareBounds() bool {
	if bytes.Compare(levelKey, itr.upperBound) <= 0 && bytes.Compare(levelKey, itr.lowerBound) >= 0  {
		return true
	} else {
		return false
	}
}*/
