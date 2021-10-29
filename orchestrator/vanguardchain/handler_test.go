package vanguardchain

import (
	"context"
	"encoding/json"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/lukso-network/lukso-orchestrator/shared/fork"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	types2 "github.com/lukso-network/lukso-orchestrator/shared/types"
	types "github.com/prysmaticlabs/eth2-types"
	eth "github.com/prysmaticlabs/prysm/proto/eth/v1alpha1"
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
	require.NoError(t, vanSvc.OnNewPendingVanguardBlock(ctx, beaconBlock))
	time.Sleep(100 * time.Millisecond)
	assert.LogsContain(t, hook, "New vanguard shard info has arrived")

	require.NoError(t, vanSvc.OnNewPendingVanguardBlock(ctx, beaconBlock))

	time.Sleep(100 * time.Millisecond)
	assert.LogsContain(t, hook, "New vanguard shard info has arrived")

	var (
		supportedSlot uint64
	)

	for currentSlot := range fork.SupportedForkL15PandoraProd {
		supportedSlot = currentSlot

		break
	}

	supportedHeader := &eth1Types.Header{}
	jsonBytes := fork.SupportedL15HeadersJson[supportedSlot]
	require.NoError(t, json.Unmarshal(jsonBytes, supportedHeader))
	assert.NotNil(t, supportedHeader.Number)

	beaconBlock.Slot = types.Slot(supportedSlot)
	beaconBlock.Body.PandoraShard = []*eth.PandoraShard{
		{

			ParentHash:  supportedHeader.ParentHash.Bytes(),
			TxHash:      supportedHeader.TxHash.Bytes(),
			StateRoot:   make([]byte, 32),
			BlockNumber: supportedHeader.Number.Uint64(),
			ReceiptHash: make([]byte, 32),
			Signature:   make([]byte, 96),
			Hash:        supportedHeader.Hash().Bytes(),
			SealHash:    make([]byte, 32),
		},
	}

	require.NoError(t, vanSvc.OnNewPendingVanguardBlock(ctx, beaconBlock))
	time.Sleep(100 * time.Millisecond)
	assert.LogsContain(t, hook, "no consensus info for supported fork")

	require.NoError(t, vanSvc.orchestratorDB.SaveConsensusInfo(context.Background(), &types2.MinimalEpochConsensusInfo{
		Epoch:            uint64(beaconBlock.Slot.Div(32)),
		ValidatorList:    nil,
		EpochStartTime:   0,
		SlotTimeDuration: 0,
	}))

	require.NoError(t, vanSvc.OnNewPendingVanguardBlock(ctx, beaconBlock))
	time.Sleep(100 * time.Millisecond)
	assert.LogsContain(t, hook, "forced trigger of reorg")
}
