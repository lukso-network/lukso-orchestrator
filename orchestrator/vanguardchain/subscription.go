package vanguardchain

import (
	"context"
	"fmt"
	"time"

	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"github.com/pkg/errors"
	eth2Types "github.com/prysmaticlabs/eth2-types"
	ethpb "github.com/prysmaticlabs/prysm/proto/eth/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	errConsensusInfoNil       = errors.New("Incoming consensus info is nil")
	errInvalidValidatorLength = errors.New("Incoming consensus info's validator list is invalid")
	errConsensusInfoProcess   = errors.New("Could not process minimal consensus info")
)

// subscribeVanNewPendingBlockHash
func (s *Service) subscribeVanNewPendingBlockHash(ctx context.Context, fromSlot uint64) error {
	var blockRoot []byte
	stream, err := s.beaconClient.StreamNewPendingBlocks(ctx,
		&ethpb.StreamPendingBlocksRequest{
			BlockRoot: blockRoot,
			FromSlot:  eth2Types.Slot(fromSlot),
		})
	if nil != err {
		log.WithError(err).Error("Failed to subscribe to stream of new pending blocks")
		return err
	}
	log.WithField("fromSlot", fromSlot).Info("Successfully subscribed to vanguard blocks")
	for {
		select {
		case <-ctx.Done():
			log.Info("Received cancelled context, exiting vanguard pending block streaming subscription!")
			return nil

		case <-s.stopPendingBlkSubCh:
			log.Info("Received re-org event, exiting vanguard pending block streaming subscription!")
			return nil

		default:
			vanBlockInfo, err := stream.Recv()
			if err != nil {
				if e, ok := status.FromError(err); ok {
					switch e.Code() {
					case codes.Canceled, codes.Internal, codes.Unavailable:
						log.WithError(err).Infof("Trying to restart connection. rpc status: %v", e.Code())
						s.waitForConnection()
						// Re-try subscription from latest finalized slot
						latestFinalizedSlot := s.db.LatestLatestFinalizedSlot()
						stream, err = s.beaconClient.StreamNewPendingBlocks(ctx,
							&ethpb.StreamPendingBlocksRequest{
								BlockRoot: blockRoot,
								FromSlot:  eth2Types.Slot(latestFinalizedSlot),
							})
						if err != nil {
							log.WithError(err).Error("Failed to subscribe to new pending blocks stream")
							return err
						}
					}
				} else {
					log.WithError(err).Error("Could not receive pending blocks from vanguard node")
					return err
				}
			}

			if err := s.onNewPendingVanguardBlock(ctx, vanBlockInfo); err != nil {
				log.WithError(err).Error("Failed to process the pending vanguard shardInfo. Exiting vanguard pending header subscription")
				return errConsensusInfoProcess
			}
		}
	}
	return nil
}

// subscribeNewConsensusInfoGRPC
func (s *Service) subscribeNewConsensusInfoGRPC(ctx context.Context, fromEpoch uint64) error {
	stream, err := s.beaconClient.StreamMinimalConsensusInfo(ctx, &ethpb.MinimalConsensusInfoRequest{FromEpoch: eth2Types.Epoch(fromEpoch)})
	if nil != err {
		log.WithError(err).Error("Failed to subscribe to stream of new consensus info")
		return err
	}

	log.WithField("fromEpoch", fromEpoch).Info("Successfully subscribed to minimal consensus info to vanguard client")

	for {
		select {
		case <-ctx.Done():
			log.Info("Received cancelled context, closing existing consensus info subscription")
			return nil

		case <-s.stopEpochInfoSubCh:
			log.Info("Received re-org event, exiting vanguard consensus info streaming subscription!")
			return nil

		default:
			vanMinimalConsensusInfo, err := stream.Recv()
			if err != nil {
				if e, ok := status.FromError(err); ok {
					switch e.Code() {
					case codes.Canceled, codes.Internal, codes.Unavailable:
						log.WithError(err).Infof("Trying to restart connection. rpc status: %v", e.Code())
						s.waitForConnection()
						latestFinalizedEpoch := s.db.LatestLatestFinalizedEpoch()
						stream, err = s.beaconClient.StreamMinimalConsensusInfo(ctx, &ethpb.MinimalConsensusInfoRequest{FromEpoch: eth2Types.Epoch(latestFinalizedEpoch)})
						if nil != err {
							log.WithError(err).Error("Failed to subscribe to stream of new consensus info, Exiting go routine")
							return err
						}
					}
				} else {
					log.WithError(err).Error("Could not receive epoch info from vanguard")
					return err
				}
			}

			if vanMinimalConsensusInfo == nil {
				log.Error("Received nil consensus info, Exiting go routine")
				return errConsensusInfoNil
			}

			// Only non empty check for now
			if len(vanMinimalConsensusInfo.ValidatorList) < 1 {
				log.WithField("epochInfo", fmt.Sprintf("%+v", vanMinimalConsensusInfo)).
					Error("Incoming consensus info's validator list is invalid, Exiting go routine")
				return errInvalidValidatorLength
			}

			consensusInfo := &types.MinimalEpochConsensusInfoV2{
				Epoch:            uint64(vanMinimalConsensusInfo.Epoch),
				ValidatorList:    vanMinimalConsensusInfo.ValidatorList,
				EpochStartTime:   vanMinimalConsensusInfo.EpochTimeStart,
				SlotTimeDuration: time.Duration(vanMinimalConsensusInfo.SlotTimeDuration.Seconds),
				FinalizedSlot:    s.db.LatestLatestFinalizedSlot(),
			}

			// if re-org happens then we get this info not nil
			if vanMinimalConsensusInfo.ReorgInfo != nil {
				reorgInfo := &types.Reorg{
					VanParentHash: vanMinimalConsensusInfo.ReorgInfo.VanParentHash,
					PanParentHash: vanMinimalConsensusInfo.ReorgInfo.PanParentHash,
					NewSlot:       uint64(vanMinimalConsensusInfo.ReorgInfo.NewSlot),
				}
				consensusInfo.ReorgInfo = reorgInfo
			}

			log.WithField("epoch", vanMinimalConsensusInfo.Epoch).WithField("epochInfo", fmt.Sprintf("%+v", vanMinimalConsensusInfo)).
				Debug("Received new consensus info")
			if err := s.onNewConsensusInfo(ctx, consensusInfo); err != nil {
				log.WithError(err).Error("Failed to handle consensus info. Closing epoch info subscription, Exiting go routine")
				return err
			}
		}
	}

	return nil
}
