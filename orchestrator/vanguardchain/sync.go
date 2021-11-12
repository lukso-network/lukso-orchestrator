package vanguardchain

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethpb "github.com/prysmaticlabs/prysm/proto/eth/v1alpha1"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) syncWithVanguardHead() {

	ticker := time.NewTicker(syncStatusPollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			syncStatus, err := s.nodeClient.GetSyncStatus(s.ctx, &emptypb.Empty{})
			if err != nil {
				log.WithError(err).Error("Could not fetch sync status")
				continue
			}
			if syncStatus.Syncing {
				log.Info("Waiting for vanguard node to be fully synced...")
				continue
			}
			log.Info("Vanguard node is fully synced now verifying orchestrator verified slot info db...")
			head, err := s.beaconClient.GetChainHead(s.ctx, &emptypb.Empty{})
			if err != nil {
				log.WithError(err).Error("Could not fetch vanguard chain head, continuing...")
				continue
			}
			// trigger revert db if vanguard head slot block hash not match with latest verified slot van block hash
			if err := s.revert(head); err != nil {
				log.WithError(err).Error("Could not revert db successfully.")
			}

		case <-s.ctx.Done():
			log.Debug("Context closed, exiting syncing vanguard chain go routine!")
			return
		}
	}
}

// revert
func (s *Service) revert(head *ethpb.ChainHead) error {
	// trigger re-org if vanguard head slot block hash not match with latest verified slot van block hash
	latestVerifiedSlot := s.db.LatestSavedVerifiedSlot()
	latestVerifiedSlotInfo, err := s.db.VerifiedSlotInfo(latestVerifiedSlot)
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

		// Stop subscription of vanguard new pending blocks
		s.stopSubscription()
		s.subscriptionShutdownFeed.Send(true) // shutdown pandora subscription
		// TODO- Stop pandora pending block subscription

		revertSlot := s.getFinalizedSlot()
		log.WithField("curFinalizedSlot", head.FinalizedSlot).WithField("revertSlot", revertSlot).
			WithField("latestVerifiedSlot", latestVerifiedSlot).Warn("Stop subscription and reverting orchestrator db from sync method")

		// Removing slot infos from verified slot info db
		if err := s.reorgDB(revertSlot); err != nil {
			log.WithError(err).Warn("Failed to revert verified info db")
			return err

		}
		// Re-subscribe vanguard new pending blocks
		go s.subscribeVanNewPendingBlockHash(s.ctx, revertSlot)
		//TODO- start pandora pending block subscription
		s.subscriptionShutdownFeed.Send(false)

		return nil
	}

	// If current orchestrator's finalize epoch is less than incoming finalized epoch, then update into db and in-memory
	if s.getFinalizedEpoch() < uint64(head.FinalizedEpoch) {
		newFS := uint64(head.FinalizedSlot)
		newFE := uint64(head.FinalizedEpoch)

		if err := s.updateFinalizedInfoInDB(newFS, newFE); err != nil {
			log.WithError(err).Warn("Failed to store new finalized info")
		}
		s.updateInMemoryFinalizedInfo(newFS, newFE)
	}

	return nil
}

// updateFinalizedInfoInDB stores new finalizedSlot and finalizedEpoch into db
func (s *Service) updateFinalizedInfoInDB(finalizedSlot, finalizedEpoch uint64) error {
	// store new finalized slot into db
	if err := s.db.SaveLatestFinalizedSlot(finalizedSlot); err != nil {
		log.WithError(err).Error("Failed to store new finalized slot in db")
		return err
	}

	// store new finalized epoch into db
	if err := s.db.SaveLatestFinalizedEpoch(finalizedEpoch); err != nil {
		log.WithError(err).Error("Failed to store new finalized epoch in db")
		return err
	}

	return nil
}

func (s *Service) reorgDB(revertSlot uint64) error {
	// Removing slot infos from verified slot info db
	if err := s.db.RemoveRangeVerifiedInfo(revertSlot+1, 0); err != nil {
		log.WithError(err).Error("found error while reverting orchestrator database")
		return err
	}

	//TODO: Updating latestVerifiedSlot and latestVerifiedHeaderHash
	if err := s.db.UpdateVerifiedSlotInfo(revertSlot); err != nil {
		log.WithError(err).Error("failed to update latest verified slot in db")
		return err
	}
	return nil
}

// updateInMemoryFinalizedInfo updates in-memory finalizedSlot and finalizedEpoch
func (s *Service) updateInMemoryFinalizedInfo(finalizedSlot, finalizedEpoch uint64) {
	s.processingLock.Lock()
	defer s.processingLock.Unlock()

	s.finalizedSlot = finalizedSlot
	s.finalizedEpoch = finalizedEpoch
}

// getFinalizedSlot
func (s *Service) getFinalizedSlot() uint64 {
	s.processingLock.Lock()
	defer s.processingLock.Unlock()

	return s.finalizedSlot
}

func (s *Service) getFinalizedEpoch() uint64 {
	s.processingLock.Lock()
	defer s.processingLock.Unlock()

	return s.finalizedEpoch
}

func (s *Service) stopSubscription() {
	s.processingLock.Lock()
	defer s.processingLock.Unlock()

	s.stopPendingBlkSubCh <- struct{}{}
}
