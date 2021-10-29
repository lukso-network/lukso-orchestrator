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
	SupportedForkL15PandoraProd = map[uint64]common.Hash{
		5280: common.HexToHash("0xd5aa89dff5365a87d6ed489a58c4e9d570e90bce327c2d51449f9e9e2917f588"),
	}
	SupportedL15HeadersJson = map[uint64][]byte{
		5280: []byte(`{
  "difficulty": "0x1",
  "extraData": "0xf869c68214a081a580b860aa86aafdf6f985ae5529f6ea49710c016f1262e9a51c535eb1522421d2e8dbe7ede52a0dbff3b020b45aa49bd59649fa14fb18187d0bee90b48ec4229b63d52b0945325b2b20b7264af4fc6ae81b78d26a71f1999c63c3fa59b55853151cbc0b",
  "gasLimit": "0x7A1200",
  "gasUsed": "0x0",
  "hash": "0xd5aa89dff5365a87d6ed489a58c4e9d570e90bce327c2d51449f9e9e2917f588",
  "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
  "miner": "0x1c8be7255e85e8a172b37086bef7f12f22410f93",
  "mixHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "nonce": "0x0000000000000000",
  "number": "0x1278",
  "parentHash": "0xfccd8d44c6554f390556e7a6d48670fa13147dade6824993725bdb27868f7e01",
  "receiptsRoot": "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
  "sha3Uncles": "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
  "size": "0x26A",
  "stateRoot": "0x0d3a89ab3d4d10bbd424902b176ef48c998b3f30656b904b428a46bc40e20acb",
  "timestamp": "0x61785B46",
  "totalDifficulty": "0x81278",
  "transactions": [],
  "transactionsRoot": "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
  "uncles": []
}`),
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
