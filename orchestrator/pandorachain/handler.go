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
	log.WithField("slot", panExtraDataWithSig.Slot).Debug("Got new pan header")
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
