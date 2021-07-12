package pandorachain

import (
	"context"
	"fmt"

	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

// OnNewPendingHeader :
//	- cache and store header and header hash with status
//  - send to consensus service for checking header with vanguard header for confirmation
func (s *Service) OnNewPendingHeader(ctx context.Context, header *eth1Types.Header) error {
	var panExtraDataWithSig types.PanExtraDataWithBLSSig
	if err := rlp.DecodeBytes(header.Extra, &panExtraDataWithSig); err != nil {
		log.WithError(err).Error("Failed to decode extra data fields")
		return err
	}
	log.WithField("slot", panExtraDataWithSig.Slot).Debug("Got new pan header")
	// Catch for possible reorg
	latestRealmVerifiedSlot := s.db.LatestVerifiedRealmSlot()

	if latestRealmVerifiedSlot > panExtraDataWithSig.Slot {
		err := fmt.Errorf("reorgs not supported")
		log.WithField("extraData", panExtraDataWithSig.Slot).
			WithField("latestRealmVerifiedSlot", latestRealmVerifiedSlot).Error(err.Error())

		return err
	}

	if err := s.cache.Put(ctx, panExtraDataWithSig.Slot, header); err != nil {
		log.WithError(err).Error("Failed to cache header")
		return err
	}

	pandoraHeaderHash := &types.HeaderHash{
		HeaderHash:       header.Hash(),
		Status:           types.Pending,
		PandoraShardHash: header.Hash(),
		Signature:        panExtraDataWithSig.BlsSignatureBytes.Bytes(),
	}
	if err := s.db.SavePandoraHeaderHash(panExtraDataWithSig.Slot, pandoraHeaderHash); err != nil {
		log.WithError(err).Error("Failed to store pandora header hash into db")
		return err
	}

	log.WithField("headerHash", pandoraHeaderHash).
		WithField("slot", panExtraDataWithSig.Slot).Trace("Successfully inserted pandora hash to db")

	s.pendingWorkChannel <- pandoraHeaderHash

	return nil
}
