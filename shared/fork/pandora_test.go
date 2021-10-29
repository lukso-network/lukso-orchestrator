package fork

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGuardAllUnsupportedPandoraForks(t *testing.T) {
	t.Run("should return no error when not blacklisted", func(t *testing.T) {
		headerHash := common.HexToHash("0xfccd8d44c6554f390556e7a6d48670fa13147dade6824993725bdb27868f7e04")
		assert.Nil(t, GuardAllUnsupportedPandoraForks(headerHash, 52))
	})

	t.Run("should return error when blacklisted", func(t *testing.T) {
		assert.Greater(t, len(unsupportedForkL15PandoraProd), 1)

		for slot, hash := range unsupportedForkL15PandoraProd {
			assert.Error(t, GuardAllUnsupportedPandoraForks(hash, slot))
		}
	})
}
