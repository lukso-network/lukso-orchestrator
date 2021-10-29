package pandorachain

import (
	"context"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/lukso-network/lukso-orchestrator/shared/fork"
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

	err := fork.GuardAllUnsupportedPandoraForks(header.Hash(), panExtraDataWithSig.Slot)

	if nil != err {
		log.Error(err)
		// Fallback to nil because it will close the connection of the stream, which shouldn't happen
		return nil
	}

	log.WithField("slot", panExtraDataWithSig.Slot).
		WithField("blockNumber", header.Number.Uint64()).
		WithField("headerHash", header.Hash()).
		Info("New pandora header info has arrived")

	s.pandoraHeaderInfoFeed.Send(&types.PandoraHeaderInfo{
		Header: header,
		Slot:   panExtraDataWithSig.Slot,
	})
	return nil
}
