package fork

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	log "github.com/sirupsen/logrus"
)

var (
	UnsupportedForkL15PandoraProd = map[uint64]common.Hash{
		5273: common.HexToHash("0x92dd18789abb7fbb3a9b7e7b2eee3ee99f6717068a9db6da2f18327f75b10304"),
		5274: common.HexToHash("0x479088bbb0f4a1577f2528fbea41ff3d99373f323769375888ac0bdf9fe506e8"),
		5275: common.HexToHash("0x138d937d02c386e1c47d5550a87deff8652f20683dfd771afa39e3553f596490"),
		5276: common.HexToHash("0xa137cb9c9a22f678994b82c09f31de79994222470c1a77dfc7c78268bb5d0bbc"),
		5278: common.HexToHash("0xcbd44e37125c599ff218b966658d720665e1dfa9fd3db0230ad8743e754495d5"),
		5279: common.HexToHash("0x0acf03ae123dc232e181d3273114b4fc1ae570f469c64655ccb7bc8c6b6aaa28"),
	}
)

func GuardAllUnsupportedPandoraForks(headerHash common.Hash, receivedSlot uint64) (err error) {
	err = guardUnsupportedForks(headerHash, receivedSlot, UnsupportedForkL15PandoraProd)

	return
}

func guardUnsupportedForks(
	headerHash common.Hash,
	receivedSlot uint64,
	forks ...map[uint64]common.Hash,
) (err error) {
mainLoop:
	for _, fork := range forks {
		for slot, hash := range fork {
			if headerHash == hash && receivedSlot == slot {
				err = fmt.Errorf(
					"unsupported fork pair. Hash: %s, slot: %d",
					hash,
					slot,
				)

				log.Error(err)

				break mainLoop
			}
		}
	}

	return
}
