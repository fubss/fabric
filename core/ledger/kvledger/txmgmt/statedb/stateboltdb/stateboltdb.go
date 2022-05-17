package stateboltdb

import (
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/common/ledger/dataformat"
	"github.com/hyperledger/fabric/common/ledger/util/boltdbhelper"
	"github.com/hyperledger/fabric/core/ledger/internal/version"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/statedb"
	kvdb "github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/statedb/stateleveldb"
	"github.com/pkg/errors"
)

var logger = flogging.MustGetLogger("stateboltdb")

var (
	dataKeyPrefix    = []byte{'d'}
	dataKeyStopper   = []byte{'e'}
	lastKeyIndicator = byte(0x01)

	savePointKey           = []byte{'s'}
	maxDataImportBatchSize = 4 * 1024 * 1024
)

// VersionedDBProvider implements interface VersionedDBProvider
type VersionedDBProvider struct {
	dbProvider *boltdbhelper.Provider
}

// NewVersionedDBProvider instantiates VersionedDBProvider
func NewVersionedDBProvider(dbPath string) (*VersionedDBProvider, error) {
	logger.Debugf("constructing VersionedDBProvider dbPath=%s", dbPath)
	dbProvider, err := boltdbhelper.NewProvider(
		&boltdbhelper.Conf{
			DBPath:         dbPath,
			ExpectedFormat: dataformat.CurrentFormat,
		})
	logger.Debugf("Provider was successfully created")
	if err != nil {
		return nil, err
	}
	return &VersionedDBProvider{dbProvider}, nil
}

// GetDBHandle gets the handle to a named database
func (provider *VersionedDBProvider) GetDBHandle(dbName string, namespaceProvider statedb.NamespaceProvider) (statedb.VersionedDB, error) {
	logger.Debugf("GetDBHandle for dbName=[%s]", dbName)
	return newVersionedDB(provider.dbProvider.GetDBHandle(dbName), dbName), nil
}

// ImportFromSnapshot loads the public state and pvtdata hashes from the snapshot files previously generated
func (provider *VersionedDBProvider) ImportFromSnapshot(
	dbName string,
	savepoint *version.Height,
	itr statedb.FullScanIterator,
) error {
	vdb := newVersionedDB(provider.dbProvider.GetDBHandle(dbName), dbName)
	return vdb.importState(itr, savepoint)
}

// BytesKeySupported returns true if a db created supports bytes as a key
func (provider *VersionedDBProvider) BytesKeySupported() bool {
	return true
}

// Close closes the underlying db
func (provider *VersionedDBProvider) Close() {
	provider.dbProvider.Close()
}

// Drop drops channel-specific data from the state leveldb.
// It is not an error if a database does not exist.
func (provider *VersionedDBProvider) Drop(dbName string) error {
	return provider.dbProvider.Drop(dbName)
}

// VersionedDB implements VersionedDB interface
type versionedDB struct {
	db     *boltdbhelper.DBHandle
	dbName string
}

// newVersionedDB constructs an instance of VersionedDB
func newVersionedDB(db *boltdbhelper.DBHandle, dbName string) *versionedDB {
	return &versionedDB{db, dbName}
}

// Open implements method in VersionedDB interface
func (vdb *versionedDB) Open() error {
	// do nothing because shared db is used
	return nil
}

// Close implements method in VersionedDB interface
func (vdb *versionedDB) Close() {
	// do nothing because shared db is used
}

// ValidateKeyValue implements method in VersionedDB interface
func (vdb *versionedDB) ValidateKeyValue(key string, value []byte) error {
	return nil
}

// BytesKeySupported implements method in VersionedDB interface
func (vdb *versionedDB) BytesKeySupported() bool {
	return true
}

// GetState implements method in VersionedDB interface
func (vdb *versionedDB) GetState(namespace string, key string) (*statedb.VersionedValue, error) {
	logger.Debugf("GetState(). ns=%s, key=%s", namespace, key)
	//TODO: add to kvdb-common-provider
	dbVal, err := vdb.db.Get(kvdb.EncodeDataKey(namespace, key))
	if err != nil {
		return nil, err
	}
	if dbVal == nil {
		return nil, nil
	}
	//TODO: add to kvdb-common-provider
	return kvdb.DecodeValue(dbVal)
}

// GetVersion implements method in VersionedDB interface
func (vdb *versionedDB) GetVersion(namespace string, key string) (*version.Height, error) {
	versionedValue, err := vdb.GetState(namespace, key)
	if err != nil {
		return nil, err
	}
	if versionedValue == nil {
		return nil, nil
	}
	return versionedValue.Version, nil
}

// GetStateMultipleKeys implements method in VersionedDB interface
func (vdb *versionedDB) GetStateMultipleKeys(namespace string, keys []string) ([]*statedb.VersionedValue, error) {
	vals := make([]*statedb.VersionedValue, len(keys))
	for i, key := range keys {
		val, err := vdb.GetState(namespace, key)
		if err != nil {
			return nil, err
		}
		vals[i] = val
	}
	return vals, nil
}

// GetStateRangeScanIterator implements method in VersionedDB interface
// startKey is inclusive
// endKey is exclusive
func (vdb *versionedDB) GetStateRangeScanIterator(namespace string, startKey string, endKey string) (statedb.ResultsIterator, error) {
	// pageSize = 0 denotes unlimited page size
	return vdb.GetStateRangeScanIteratorWithPagination(namespace, startKey, endKey, 0)
}

// GetStateRangeScanIteratorWithPagination implements method in VersionedDB interface
func (vdb *versionedDB) GetStateRangeScanIteratorWithPagination(namespace string, startKey string, endKey string, pageSize int32) (statedb.QueryResultsIterator, error) {
	dataStartKey := kvdb.EncodeDataKey(namespace, startKey)
	dataEndKey := kvdb.EncodeDataKey(namespace, endKey)
	if endKey == "" {
		logger.Debugf("endKey is empty string (but not nil)")
		dataEndKey[len(dataEndKey)-1] = lastKeyIndicator
	} else {
		logger.Debugf("endKey is not empty")
		dataEndKey = kvdb.EncodeDataKey(namespace, endKey)
	}
	dbItr, err := vdb.db.GetIterator(dataStartKey, dataEndKey)
	if err != nil {
		return nil, err
	}
	return newKVScanner(namespace, dbItr, pageSize), nil
}

// ExecuteQuery implements method in VersionedDB interface
func (vdb *versionedDB) ExecuteQuery(namespace, query string) (statedb.ResultsIterator, error) {
	return nil, errors.New("ExecuteQuery not supported for boltdb")
}

// ExecuteQueryWithPagination implements method in VersionedDB interface
func (vdb *versionedDB) ExecuteQueryWithPagination(namespace, query, bookmark string, pageSize int32) (statedb.QueryResultsIterator, error) {
	return nil, errors.New("ExecuteQueryWithMetadata not supported for boltdb")
}

// ApplyUpdates implements method in VersionedDB interface
func (vdb *versionedDB) ApplyUpdates(batch *statedb.UpdateBatch, height *version.Height) error {
	dbBatch := vdb.db.NewUpdateBatch()
	defer dbBatch.Rollback()
	namespaces := batch.GetUpdatedNamespaces()
	for _, ns := range namespaces {
		updates := batch.GetUpdates(ns)
		for k, vv := range updates {
			dataKey := kvdb.EncodeDataKey(ns, k)
			logger.Debugf("Channel [%s]: Applying key(string)=[%s] key(bytes)=[%#v]", vdb.dbName, string(dataKey), dataKey)

			if vv.Value == nil {
				logger.Debugf("deleting datakey [%s]", dataKey)
				dbBatch.Delete(dataKey)
			} else {
				encodedVal, err := kvdb.EncodeValue(vv)
				if err != nil {
					return err
				}
				logger.Debugf("dbBatch.Put(dataKey=[%#v (%s)], encodedVal=[%#v (%s)])", dataKey, dataKey, encodedVal, encodedVal)
				dbBatch.Put(dataKey, encodedVal)
			}
		}
	}
	// Record a savepoint at a given height
	// If a given height is nil, it denotes that we are committing pvt data of old blocks.
	// In this case, we should not store a savepoint for recovery. The lastUpdatedOldBlockList
	// in the pvtstore acts as a savepoint for pvt data.
	if height != nil {
		logger.Debugf("dbBatch.Put(savePointKey=[%#v (%s)], height.ToBytes()=[%#v (%s)])", savePointKey, savePointKey, height.ToBytes(), height.ToBytes())
		dbBatch.Put(savePointKey, height.ToBytes())
	}
	// Setting snyc to true as a precaution, false may be an ok optimization after further testing.
	logger.Debugf("WritingBatch...")
	return vdb.db.WriteBatch(dbBatch, true)
}

// GetLatestSavePoint implements method in VersionedDB interface
func (vdb *versionedDB) GetLatestSavePoint() (*version.Height, error) {
	logger.Debugf("GetLatestSavePoint()...")
	versionBytes, err := vdb.db.Get(savePointKey)
	if err != nil {
		return nil, err
	}
	if versionBytes == nil {
		return nil, nil
	}
	version, _, err := version.NewHeightFromBytes(versionBytes)
	if err != nil {
		return nil, err
	}
	return version, nil
}

// GetFullScanIterator implements method in VersionedDB interface. 	This function returns a
// FullScanIterator that can be used to iterate over entire data in the statedb for a channel.
// `skipNamespace` parameter can be used to control if the consumer wants the FullScanIterator
// to skip one or more namespaces from the returned results. The intended use of this iterator
// is to generate the snapshot files for the stateleveldb
func (vdb *versionedDB) GetFullScanIterator(skipNamespace func(string) bool) (statedb.FullScanIterator, error) {
	return newFullDBScanner(vdb.db, skipNamespace)
}

// importState implements method in VersionedDB interface. The function is expected to be used
// for importing the state from a previously snapshotted state. The parameter itr provides access to
// the snapshotted state.
func (vdb *versionedDB) importState(itr statedb.FullScanIterator, savepoint *version.Height) error {
	logger.Debugf("State importing...")
	if itr == nil {
		return vdb.db.Put(savePointKey, savepoint.ToBytes(), true)
	}
	dbBatch := vdb.db.NewUpdateBatch()
	defer dbBatch.Rollback()
	for {
		versionedKV, err := itr.Next()
		if err != nil {
			return err
		}
		if versionedKV == nil {
			itr.Close() //TODO: it might be excess
			break
		}
		dbKey := kvdb.EncodeDataKey(versionedKV.Namespace, versionedKV.Key)
		dbValue, err := kvdb.EncodeValue(versionedKV.VersionedValue)
		if err != nil {
			return err
		}
		//batchSize += len(dbKey) + len(dbValue)
		dbBatch.Put(dbKey, dbValue)
		/*if batchSize >= maxDataImportBatchSize {
			if err := vdb.db.WriteBatch(dbBatch, true); err != nil {
				return err
			}
			batchSize = 0
			logger.Debugf("Clearing batch...")
			dbBatch.Tx.Rollback()
		}*/
	}
	dbBatch.Put(savePointKey, savepoint.ToBytes())
	return vdb.db.WriteBatch(dbBatch, true)
}

// IsEmpty return true if the statedb does not have any content
func (vdb *versionedDB) IsEmpty() (bool, error) {
	return vdb.db.IsEmpty()
}

type kvScanner struct {
	namespace            string
	dbItr                *boltdbhelper.Iterator
	requestedLimit       int32
	totalRecordsReturned int32
}

func newKVScanner(namespace string, dbItr *boltdbhelper.Iterator, requestedLimit int32) *kvScanner {
	return &kvScanner{namespace, dbItr, requestedLimit, 0}
}

func (scanner *kvScanner) Next() (*statedb.VersionedKV, error) {
	logger.Debugf("kvScanner.Next()...")
	if scanner.requestedLimit > 0 && scanner.totalRecordsReturned >= scanner.requestedLimit {
		logger.Debugf("if-case scanner.requestedLimit=[%+v], scanner.totalRecordsReturned=[%+v]", scanner.requestedLimit, scanner.totalRecordsReturned) //TODO remove this
		return nil, nil
	}
	if !scanner.dbItr.Valid() {
		logger.Debugf("IF-CASE (TODO: DELETE IF NEVER HAPPENED IN TESTS) boltdb iterator is not valid")
		return nil, nil
	}
	scanner.dbItr.Next()
	if !scanner.dbItr.Valid() {
		logger.Debugf("if-case boltdb iterator is not valid after Next")
		return nil, nil
	}

	dbKey := scanner.dbItr.Key()
	dbVal := scanner.dbItr.Value()
	dbValCopy := make([]byte, len(dbVal))
	copy(dbValCopy, dbVal) //TODO: maybe this is not enough fast way of copying slice?
	_, key := kvdb.DecodeDataKey(dbKey)
	vv, err := kvdb.DecodeValue(dbValCopy)
	logger.Debugf("after iterator.Next(): dbKey=[%s (%#v)], key=[%s], vv=[%s (%#v)], dbVal=[%s]", dbKey, dbKey, key, vv, vv, dbValCopy)

	//scanner.dbItr.FreeKey()   //we have to free key & value,
	//scanner.dbItr.FreeValue() //otherwise iterator works incorrect
	if err != nil {
		return nil, err
	}

	scanner.totalRecordsReturned++
	return &statedb.VersionedKV{
		CompositeKey: &statedb.CompositeKey{
			Namespace: scanner.namespace,
			Key:       key,
		},
		VersionedValue: vv,
	}, nil
}

func (scanner *kvScanner) Close() {
	logger.Debugf("Closing boltdb iterator...")
	scanner.dbItr.Release()
	scanner.dbItr = nil
}

func (scanner *kvScanner) GetBookmarkAndClose() string {
	retval := ""
	scanner.dbItr.Next()
	if scanner.dbItr.Valid() {
		dbKey := scanner.dbItr.Key()
		_, key := kvdb.DecodeDataKey(dbKey)
		retval = key
	}
	scanner.Close()
	return retval
}

type fullDBScanner struct {
	db *boltdbhelper.DBHandle
	//TODO add to kv-common-provider or interface
	dbItr  *boltdbhelper.Iterator
	toSkip func(namespace string) bool
	closed bool
}

func newFullDBScanner(db *boltdbhelper.DBHandle, skipNamespace func(namespace string) bool) (*fullDBScanner, error) {
	dbItr, err := db.GetIterator(dataKeyPrefix, dataKeyStopper)
	if err != nil {
		return nil, err
	}
	return &fullDBScanner{
			db:     db,
			dbItr:  dbItr,
			toSkip: skipNamespace,
		},
		nil
}

// Next returns the key-values in the lexical order of <Namespace, key>
func (s *fullDBScanner) Next() (*statedb.VersionedKV, error) {
	logger.Debugf("fullDBScanner.Next()...")
	if s.closed {
		return nil, errors.Errorf("internal boltdb error while retrieving data from db iterator: iterator is not valid")
	}
	for {
		s.dbItr.Next()
		if s.dbItr.Valid() {
			logger.Debugf("itr is valid")
		} else {
			logger.Debugf("itr is not valid")
			break
		}
		ns, key := kvdb.DecodeDataKey(s.dbItr.Key())
		compositeKey := &statedb.CompositeKey{
			Namespace: ns,
			Key:       key,
		}
		logger.Debugf("CompKey: %s (%#v)", compositeKey, compositeKey)
		versionedVal, err := kvdb.DecodeValue(s.dbItr.Value())
		if err != nil {
			return nil, err
		}
		logger.Debugf("versionedVal: %s (%#v)", versionedVal, versionedVal)
		switch {
		case !s.toSkip(ns):
			return &statedb.VersionedKV{
				CompositeKey:   compositeKey,
				VersionedValue: versionedVal,
			}, nil
		default:
			s.dbItr.Seek(kvdb.DataKeyStarterForNextNamespace(ns))
			s.dbItr.Prev() // because 0-key of the ns1 is ommited if we call TestDataExportImport/test_case_2
		}
	}
	return nil, nil
}

func (s *fullDBScanner) Close() {
	if s == nil {
		return
	}
	s.dbItr.Release()
	s.closed = true
}
