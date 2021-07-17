package vanguardchain

import (
	"context"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	types "github.com/prysmaticlabs/eth2-types"
	eth "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"testing"
	"time"
)

func TestService_OnNewPendingVanguardBlock(t *testing.T) {
	hook := logTest.NewGlobal()
	ctx := context.Background()
	vanSvc, _ := SetupVanguardSvc(ctx, t, GRPCFunc)
	slot := uint64(5)

	beaconBlock := &eth.BeaconBlock{
		Slot:       types.Slot(slot),
		ParentRoot: make([]byte, 32),
		StateRoot:  make([]byte, 32),
		Body: &eth.BeaconBlockBody{
			RandaoReveal: make([]byte, 96),
			Eth1Data: &eth.Eth1Data{
				DepositRoot: make([]byte, 32),
				BlockHash:   make([]byte, 32),
			},
			Graffiti:          make([]byte, 32),
			Attestations:      []*eth.Attestation{},
			AttesterSlashings: []*eth.AttesterSlashing{},
			Deposits:          []*eth.Deposit{},
			ProposerSlashings: []*eth.ProposerSlashing{},
			VoluntaryExits:    []*eth.SignedVoluntaryExit{},
			PandoraShard: []*eth.PandoraShard{{
				ParentHash:  make([]byte, 32),
				TxHash:      make([]byte, 32),
				StateRoot:   make([]byte, 32),
				BlockNumber: slot,
				ReceiptHash: make([]byte, 32),
				Signature:   make([]byte, 96),
				Hash:        make([]byte, 32),
				SealHash:    make([]byte, 32),
			}},
		},
	}
	vanSvc.OnNewPendingVanguardBlock(ctx, beaconBlock)
	time.Sleep(100 * time.Millisecond)
	assert.LogsContain(t, hook, "Sharding info pushed to consensus service")

	//	 Should return error when possible reorg will happen
	require.NoError(t, vanSvc.orchestratorDB.SaveLatestVerifiedRealmSlot(6))
	vanSvc.OnNewPendingVanguardBlock(ctx, beaconBlock)

	time.Sleep(100 * time.Millisecond)
	assert.LogsContain(t, hook, "Sharding info pushed to consensus service")
}
