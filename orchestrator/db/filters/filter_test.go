package filters

import (
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"testing"

	types "github.com/prysmaticlabs/eth2-types"
)

func TestQueryFilter_ChainsCorrectly(t *testing.T) {
	f := NewFilter().
		SetStartEpoch(2).
		SetEndEpoch(4)

	filterSet := f.Filters()
	assert.Equal(t, 3, len(filterSet), "Unexpected number of filters")
	for k, v := range filterSet {
		switch k {
		case StartEpoch:
			t.Log(v.(types.Epoch))
		case EndEpoch:
			t.Log(v.(types.Epoch))
		default:
			t.Log("Unknown filter type")
		}
	}
}
