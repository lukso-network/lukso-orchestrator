package utils

import (
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"testing"
)

func TestStack_RemoveByTime(t *testing.T) {
	s := NewStack()
	headerInfos, _ := testutil.GetHeaderInfosAndShardInfos(1, 25)
	for i := 0; i < 25; i++ {
		s.Push(headerInfos[i].Header.Hash().Bytes())
	}

	for i := 0; i < 25; i++ {
		s.Delete(headerInfos[i].Header.Hash().Bytes())
	}
}