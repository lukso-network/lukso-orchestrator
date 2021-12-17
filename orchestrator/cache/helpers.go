package cache

import (
	"bytes"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	types2 "github.com/prysmaticlabs/eth2-types"
	"time"
)

func (pc *PandoraCache) VerifyPandoraCache(verifyParams *PanCacheInsertParams) error {
	if verifyParams == nil {
		return errInvalidElement
	}
	return pc.verifyPanCacheOrder(verifyParams.CurrentVerifiedHeader, verifyParams.LastVerifiedShardInfo)
}

func (pc *PandoraCache) verifyPanCacheOrder(currentHeader *eth1Types.Header, lastVerifiedShardInfo *types.MultiShardInfo) error {
	if currentHeader == nil {
		return errInvalidElement
	}

	keys := pc.cache.Keys()
	if len(keys) == 0 {
		// the cache has no element so compare with the latestVerifiedHeader
		if lastVerifiedShardInfo == nil || lastVerifiedShardInfo.IsNil() {
			// for fresh start there will be nothing inside database. So accept this header
			return nil
		}
		if bytes.Equal(currentHeader.ParentHash.Bytes(), lastVerifiedShardInfo.GetPanShardRootBytes()) {
			return nil
		}
		return errParentHashMismatch
	}
	// cache size is larger than 0, so compare with the latest element
	lastCachedElement, err := pc.stack.Top()
	if err != nil {
		return err
	}
	if bytes.Equal(currentHeader.ParentHash.Bytes(), lastCachedElement) {
		return nil
	}
	return errParentHashMismatch
}

func (vc *VanguardCache) VerifyVanguardCache(verifyParams *VanCacheInsertParams) error {
	if verifyParams == nil {
		return errInvalidElement
	}
	return vc.verifyVanCacheOrder(verifyParams.CurrentShardInfo, verifyParams.LastVerifiedShardInfo)
}

func (vc *VanguardCache) verifyVanCacheOrder(currentShard *types.VanguardShardInfo, lastVerifiedShardInfo *types.MultiShardInfo) error {
	if currentShard == nil {
		return errInvalidElement
	}

	keys := vc.cache.Keys()
	if len(keys) == 0 {
		if lastVerifiedShardInfo == nil || lastVerifiedShardInfo.IsNil() {
			// first element in the whole system. accept it
			return nil
		}
		// the cache has no element so compare with the latestVerifiedHeader
		if bytes.Equal(currentShard.ParentHash, lastVerifiedShardInfo.GetVanSlotRootBytes()) {
			return nil
		}
		return errParentHashMismatch
	}
	// cache size is larger than 0, so compare with the latest element
	lastCachedElement, err := vc.stack.Top()
	if err != nil {
		return err
	}
	if bytes.Equal(currentShard.ParentHash, lastCachedElement) {
		return nil
	}
	return errParentHashMismatch
}

func SlotStartTime(genesis uint64, slot types2.Slot, secondsPerSlot uint64) time.Time {
	duration := time.Second * time.Duration(slot.Mul(secondsPerSlot))
	startTime := time.Unix(int64(genesis), 0).Add(duration)
	return startTime
}
