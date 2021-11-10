package vanguardchain

import (
	"github.com/ethereum/go-ethereum/common"
	ethpb "github.com/prysmaticlabs/prysm/proto/eth/v1alpha1"
	"time"
)

func (s *Service) syncWithVanguardHead() {

	ticker := time.NewTicker(syncStatusPollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			status, err := s.vanGRPCClient.SyncStatus()
			if err != nil {
				log.WithError(err).Error("Could not fetch sync status")
				continue
			}
			if !status {
				log.Info("Waiting for vanguard node to be fully synced...")
				continue
			}

			log.Info("Vanguard node is fully synced now verifying orchestrator verified slot info db...")
			head, err := s.vanGRPCClient.ChainHead()
			if err != nil {
				log.WithError(err).Error("Could not fetch vanguard chain head, continuing...")
				continue
			}
			// trigger revert db if vanguard head slot block hash not match with latest verified slot van block hash
			if err := s.checkVerifiedSlotInfoHead(head); err != nil {
				log.WithError(err).Error("Could not revert db successfully.")
				continue
			}
		case <-s.ctx.Done():
			log.Debug("Context closed, exiting syncing vanguard chain go routine!")
			return
		}
	}
}

func (s *Service) checkVerifiedSlotInfoHead(head *ethpb.ChainHead) error {
	// update the latest finalized slot
	latestFinalizedSlot := s.orchestratorDB.LatestLatestFinalizedSlot()
	newFinalizedSlot := uint64(head.FinalizedSlot)
	if latestFinalizedSlot < newFinalizedSlot {
		if err := s.orchestratorDB.SaveLatestFinalizedSlot(newFinalizedSlot); err != nil {
			log.WithError(err).Error("Failed to store new finalized slot in db")
			return err
		}
	}

	// update the latest finalized epoch
	latestFinalizedEpoch := s.orchestratorDB.LatestLatestFinalizedEpoch()
	newFinalizedEpoch := uint64(head.FinalizedEpoch)
	if latestFinalizedEpoch < newFinalizedEpoch {
		if err := s.orchestratorDB.SaveLatestFinalizedEpoch(newFinalizedEpoch); err != nil {
			log.WithError(err).Error("Failed to store new finalized epoch in db")
			return err
		}
	}

	// trigger re-org if vanguard head slot block hash not match with latest verified slot van block hash
	latestVerifiedSlot := s.orchestratorDB.LatestSavedVerifiedSlot()
	latestVerifiedSlotInfo, err := s.orchestratorDB.VerifiedSlotInfo(latestVerifiedSlot)
	if err != nil {
		log.WithError(err).Error("Could not retrieve latest verified info")
		return err
	}

	canonicalBlockHash := common.BytesToHash(head.HeadBlockRoot)
	if latestVerifiedSlotInfo.VanguardBlockHash != canonicalBlockHash {

		log.WithField("vanCanonicalBlockHash", canonicalBlockHash).
			WithField("latestVerifiedVanHash", latestVerifiedSlotInfo.VanguardBlockHash).
			WithField("vanCanonicalHeadSlot", head.HeadSlot).
			WithField("latestVerifiedSlot", latestVerifiedSlot).
			Warn("Vanguard canonical head block hash does not match with latest verified block hash")

		revertSlot := latestFinalizedSlot
		if latestFinalizedSlot > 0 {
			revertSlot = latestFinalizedSlot + 1
		}
		s.orchestratorDB.RemoveRangeVerifiedInfo(revertSlot, 0)

		// Trigger to resubscribe van pending blocks stream and epoch info stream
		s.resubscribeEpochInfoCh <- struct{}{}
		s.resubscribePendingBlkCh <- struct{}{}
	}
	return nil
}
