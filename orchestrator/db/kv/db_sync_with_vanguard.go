package kv

import (
	"github.com/boltdb/bolt"
	"github.com/lukso-network/lukso-orchestrator/shared/bytesutil"
)

func (s *Store) RemoveInfoFromAllDb(fromEpoch, toEpoch uint64) error {
	for i := fromEpoch; i <= toEpoch; i++ {
		err := s.removeConsensusInfoDb(i)
		if err != nil {
			return err
		}
	}
}