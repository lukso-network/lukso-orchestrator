package consensus

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

// onPandoraHeader
func (s *Service) onPandoraHeader(ctx context.Context, headerInfo *types.PandoraHeaderInfo) error {
	slot := headerInfo.Slot

	vanShardInfo, _ := s.vanguardPendingShardingCache.Get(ctx, slot)
	if vanShardInfo != nil {
		return s.verifyShardingInfo(slot, vanShardInfo, headerInfo)
	}

	log.WithField("slot", slot).Info("Vanguard sharding info does not come yet")
	return nil
}

// onVanguardShardInfo
func (s *Service) onVanguardShardInfo(ctx context.Context, vanShardInfo *types.VanguardShardInfo) error {
	slot := vanShardInfo.Slot

	headerInfo, _ := s.pandoraPendingHeaderCache.Get(ctx, slot)
	if headerInfo != nil {
		// TODO- compare shard info and header
		return s.verifyShardingInfo(slot, vanShardInfo, headerInfo)
	}

	log.WithField("slot", slot).Info("Pandora header does not come yet")
	return nil
}

// verifyShardingInfo
func (s *Service) verifyShardingInfo(slot uint64, vanShardInfo *types.VanguardShardInfo, headerInfo *types.PandoraHeaderInfo) error {
	slotInfo := new(types.SlotInfo)
	// TODO- compare shard info and header
	status := CompareShardingInfo(nil, nil)
	if status {
		log.WithField("slot", slot).Debug("Consensus is established between pandora header and shard info")

		slotInfo.PandoraHeaderHash = headerInfo.Header.Hash()
		slotInfo.VanguardBlockHash = common.BytesToHash(vanShardInfo.BlockHash[:])

		if err := s.verifiedSlotInfoDB.SaveVerifiedSlotInfo(slot, slotInfo); err != nil {
			log.WithField("slot", slot).WithField("slotInfo", fmt.Sprintf("%+v", slotInfo)).WithError(err).Error(
				"Failed to store verified slot info into verified slot info DB")
			return err
		}
	}

	if err := s.invalidSlotInfoDB.SaveInvalidSlotInfo(slot, slotInfo); err != nil {
		log.WithField("slot", slot).WithField("slotInfo", fmt.Sprintf("%+v", slotInfo)).WithError(err).Error(
			"Failed to store invalid slot info into invalid slot info DB")
		return err
	}

	// TODO- Need to delete previous slot from cache
	log.WithField("slot", slot).WithField("slotInfo", fmt.Sprintf("%+v", slotInfo)).Info("Successfully verified sharding info")
	return nil
}
