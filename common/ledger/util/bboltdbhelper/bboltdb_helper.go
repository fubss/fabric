package boltdbhelper

import (
	"fmt"
	"sync"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/internal/fileutil"
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
	conf    *Conf
	db      *bolt.DB
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
	//dbOpts := &opt.Options{}
	dbPath := dbInst.conf.DBPath
	var err error

	if _, err = fileutil.CreateDirIfMissing(dbPath); err != nil {
		panic(fmt.Sprintf("Error creating dir if missing: %s", err))
	}
	//db will be created if it doesn't exist.
	dbInst.db, err = bolt.Open(dbPath, 0666, nil)
	if err != nil {
		panic(fmt.Sprintf("Error opening boltdb: %s", err))
	}
	dbInst.dbState = opened
}
