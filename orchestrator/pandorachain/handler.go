package pandorachain

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

var (
	unsupportedForkL15Prod = map[uint64]common.Hash{
		5723: common.HexToHash("0x92dd18789abb7fbb3a9b7e7b2eee3ee99f6717068a9db6da2f18327f75b10304"),
		5274: common.HexToHash("0x479088bbb0f4a1577f2528fbea41ff3d99373f323769375888ac0bdf9fe506e8"),
		5275: common.HexToHash("0x138d937d02c386e1c47d5550a87deff8652f20683dfd771afa39e3553f596490"),
		5276: common.HexToHash("0xa137cb9c9a22f678994b82c09f31de79994222470c1a77dfc7c78268bb5d0bbc"),
		5278: common.HexToHash("0xcbd44e37125c599ff218b966658d720665e1dfa9fd3db0230ad8743e754495d5"),
		5279: common.HexToHash("0x0acf03ae123dc232e181d3273114b4fc1ae570f469c64655ccb7bc8c6b6aaa28"),
	}
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

	err := s.guardUnsupportedForks(header, panExtraDataWithSig, unsupportedForkL15Prod)

	if nil != err {
		return err
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

func (s *Service) guardUnsupportedForks(
	header *eth1Types.Header,
	panExtraDataWithSig types.PanExtraDataWithBLSSig,
	forks ...map[uint64]common.Hash,
) (err error) {
mainLoop:
	for _, fork := range forks {
		for slot, hash := range fork {
			if header.Hash() == hash && panExtraDataWithSig.Slot == slot {
				err = fmt.Errorf(
					"unsupported fork pair. Hash: %s, slot: %d, blockNumber: %d",
					hash,
					slot,
					header.Number.Uint64(),
				)

				log.Error(err)

				break mainLoop
			}
		}
	}

	return
}
