package kv

import (
	"context"
	"github.com/boltdb/bolt"
	"github.com/dgraph-io/ristretto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/lukso-network/lukso-orchestrator/shared/fileutil"
	"github.com/lukso-network/lukso-orchestrator/shared/params"
	"github.com/pkg/errors"
	"os"
	"path"
	"sync"
	"time"
)

const (
	// ConsensusInfosCacheSize with 1024 consensus infos will be 1.5MB.
	ConsensusInfosCacheSize = 1 << 10
	// HeaderHashesCacheSize with
	HeaderHashesCacheSize = 1 << 20
	// OrchestratorNodeDbDirName is the name of the directory containing the orchestrator node database.
	OrchestratorNodeDbDirName = "orchestrator"
	// DatabaseFileName is the name of the orchestrator node database.
	DatabaseFileName = "orchestrator.db"

	boltAllocSize = 8 * 1024 * 1024
)

// Config for the bolt db kv store.
type Config struct {
	InitialMMapSize int
}

type Store struct {
	ctx                   context.Context
	isRunning             bool
	db                    *bolt.DB
	databasePath          string
	consensusInfoCache    *ristretto.Cache
	panHeaderCache        *ristretto.Cache
	vanHeaderCache        *ristretto.Cache
	verifiedSlotInfoCache *ristretto.Cache
	invalidSlotInfoCache  *ristretto.Cache

	// Latest information need to be stored into db
	latestEpoch         uint64
	latestPanSlot       uint64
	latestPanHeaderHash common.Hash
	latestVanSlot       uint64
	latestVanHash       common.Hash
	// There should be mutex in store
	sync.Mutex
}

// NewKVStore initializes a new boltDB key-value store at the directory
// path specified, creates the kv-buckets based on the schema, and stores
// an open connection db object as a property of the Store struct.
func NewKVStore(ctx context.Context, dirPath string, config *Config) (*Store, error) {
	hasDir, err := fileutil.HasDir(dirPath)
	if err != nil {
		return nil, err
	}
	if !hasDir {
		if err := fileutil.MkdirAll(dirPath); err != nil {
			return nil, err
		}
	}
	datafile := path.Join(dirPath, DatabaseFileName)
	boltDB, err := bolt.Open(
		datafile,
		params.OrchestratorIoConfig().ReadWritePermissions,
		&bolt.Options{
			Timeout:         1 * time.Second,
			InitialMmapSize: config.InitialMMapSize,
		},
	)
	if err != nil {
		if errors.Is(err, bolt.ErrTimeout) {
			return nil, errors.New("cannot obtain database lock, database may be in use by another process")
		}
		return nil, err
	}
	boltDB.AllocSize = boltAllocSize
	consensusInfoCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1000,                    // number of keys to track frequency of (1000).
		MaxCost:     ConsensusInfosCacheSize, // maximum cost of cache (1000 consensus info).
		BufferItems: 64,                      // number of keys per Get buffer.
	})
	if err != nil {
		return nil, err
	}

	panHeaderCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1000,                  // number of keys to track frequency of (1000).
		MaxCost:     HeaderHashesCacheSize, // maximum cost of cache (1000 headers).
		BufferItems: 64,                    // number of keys per Get buffer.
	})
	if err != nil {
		return nil, err
	}

	vanBlockCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1000,                  // number of keys to track frequency of (1000).
		MaxCost:     HeaderHashesCacheSize, // maximum cost of cache (1000 headers).
		BufferItems: 64,                    // number of keys per Get buffer.
	})

	verifiedSlotInfoCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1000,                  // number of keys to track frequency of (1000).
		MaxCost:     HeaderHashesCacheSize, // maximum cost of cache (1000 headers).
		BufferItems: 64,                    // number of keys per Get buffer.
	})

	invalidSlotInfoCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1000,                  // number of keys to track frequency of (1000).
		MaxCost:     HeaderHashesCacheSize, // maximum cost of cache (1000 headers).
		BufferItems: 64,                    // number of keys per Get buffer.
	})

	if nil != err {
		return nil, err
	}

	kv := &Store{
		ctx:                   ctx,
		db:                    boltDB,
		databasePath:          dirPath,
		consensusInfoCache:    consensusInfoCache,
		panHeaderCache:        panHeaderCache,
		vanHeaderCache:        vanBlockCache,
		verifiedSlotInfoCache: verifiedSlotInfoCache,
		invalidSlotInfoCache:  invalidSlotInfoCache,
	}

	if err := kv.db.Update(func(tx *bolt.Tx) error {
		return createBuckets(
			tx,
			consensusInfosBucket,
			pandoraHeaderHashesBucket,
			vanguardHeaderHashesBucket,
			realmBucket,
			verifiedSlotInfosBucket,
			invalidSlotInfosBucket,
		)
	}); err != nil {
		return nil, err
	}
	// Retrieve initial data from DB
	kv.initLatestDataFromDB()

	return kv, err
}

// ClearDB removes the previously stored database in the data directory.
func (s *Store) ClearDB() error {
	if _, err := os.Stat(s.databasePath); os.IsNotExist(err) {
		return nil
	}
	if err := os.Remove(path.Join(s.databasePath, DatabaseFileName)); err != nil {
		return errors.Wrap(err, "could not remove database file")
	}
	return nil
}

// Close closes the underlying BoltDB database.
func (s *Store) Close() error {
	err := s.SaveLatestEpoch(s.ctx)

	if nil != err {
		return err
	}

	return s.db.Close()
}

// DatabasePath at which this database writes files.
func (s *Store) DatabasePath() string {
	return s.databasePath
}

// initLatestDataFromDB helps to retrieve initial data from DB
func (s *Store) initLatestDataFromDB() {
	// Retrieve latest saved epoch number from db
	s.latestEpoch = s.LatestSavedEpoch()
	// Retrieve latest saved pandora slot from db
	s.latestPanSlot = s.LatestSavedPandoraSlot()
	// Retrieve latest saved pandora header hash from db
	s.latestPanHeaderHash = s.LatestSavedPandoraHeaderHash()
	// Retrieve latest saved vanguard hash from db
	s.latestVanHash = s.LatestSavedVanguardHeaderHash()
	// Retrieve latest savend vanguard slot from db
	s.latestVanSlot = s.LatestSavedVanguardSlot()
}

// createBuckets
func createBuckets(tx *bolt.Tx, buckets ...[]byte) error {
	for _, bucket := range buckets {
		if _, err := tx.CreateBucketIfNotExists(bucket); err != nil {
			return err
		}
	}
	return nil
}
