package pandorachain

import (
	"context"
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

	if slotInfo, _ := s.db.VerifiedSlotInfo(panExtraDataWithSig.Slot); slotInfo != nil {
		if slotInfo.PandoraHeaderHash == header.Hash() {
			log.WithField("slot", panExtraDataWithSig.Slot).
				WithField("headerHash", header.Hash()).
				Info("Pandora header is already in verified slot info db")
			return nil
		}
		// TODO: When pandora pushes new header info for old slot, then we should take take a rational decision for the header
		// TODO: We also need to have a fork choice mechanism in orchestrator client as well as pandora client
	}

	log.WithField("slot", panExtraDataWithSig.Slot).
		WithField("headerHash", header.Hash()).
		Info("New pandora header info has arrived")

	if err := s.cache.Put(ctx, panExtraDataWithSig.Slot, header); err != nil {
		log.WithError(err).Error("Failed to cache header")
		return err
	}

	nSent := s.pandoraHeaderInfoFeed.Send(&types.PandoraHeaderInfo{
		Header: header,
		Slot:   panExtraDataWithSig.Slot,
	})

	log.WithField("nSent", nSent).Trace("Header info pushed to consensus service")
	return nil
}
