package vanguardchain

import (
	"fmt"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain/client"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"github.com/pkg/errors"
	eth2Types "github.com/prysmaticlabs/eth2-types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

var (
	errShardInfoProcess       = errors.New("Failed to process the pending vanguard shardInfo")
	errConsensusInfoNil       = errors.New("Incoming consensus info is nil")
	errInvalidValidatorLength = errors.New("Incoming consensus info's validator list is invalid")
	errConsensusInfoProcess   = errors.New("Could not process minimal consensus info")
)

// subscribeVanNewPendingBlockHash
func (s *Service) subscribeVanNewPendingBlockHash(client client.VanguardClient, fromSlot uint64) error {
	var blockRoot []byte
	stream, err := client.StreamNewPendingBlocks(blockRoot, eth2Types.Slot(fromSlot))
	if nil != err {
		log.WithError(err).Error("Failed to subscribe to stream of new pending blocks")
		return err
	}
	log.WithField("fromSlot", fromSlot).Info("Successfully subscribed to vanguard blocks")
	for {
		select {
		case <-s.ctx.Done():
			log.Debug("closing subscribeVanNewPendingBlockHash")
			return nil

		case <-s.resubscribePendingBlkCh:
			fromSlot = s.orchestratorDB.LatestLatestFinalizedSlot()
			stream, err = client.StreamNewPendingBlocks(blockRoot, eth2Types.Slot(fromSlot))
			if err != nil {
				log.WithError(err).Error("Failed to re-subscribe to new pending blocks stream")
				return err
			}
			log.WithField("slot", fromSlot).Debug("Re-subscribing to new vanguard pending blocks stream")

		default:
			vanBlock, err := stream.Recv()
			if e, ok := status.FromError(err); ok {
				switch e.Code() {
				case codes.Canceled, codes.Internal, codes.Unavailable:
					log.WithError(err).Infof("Trying to restart connection. rpc status: %v", e.Code())
					s.waitForConnection()
					stream, err = client.StreamNewPendingBlocks(blockRoot, eth2Types.Slot(fromSlot))
					if err != nil {
						log.WithError(err).Error("Failed to subscribe to new pending blocks stream")
						return err
					}
				}
			}
			if err := s.OnNewPendingVanguardBlock(s.ctx, vanBlock); err != nil {
				log.WithError(err).Error("Failed to process the pending vanguard shardInfo. Exiting vanguard pending header subscription")
				return errConsensusInfoProcess
			}
		}
	}
	return nil
}

// subscribeNewConsensusInfoGRPC
func (s *Service) subscribeNewConsensusInfoGRPC(client client.VanguardClient, fromEpoch uint64) error {
	stream, err := client.StreamMinimalConsensusInfo(fromEpoch)
	if nil != err {
		log.WithError(err).Error("Failed to subscribe to stream of new consensus info")
		return err
	}
	log.WithField("fromEpoch", fromEpoch).Info("Successfully subscribed to minimal " +
		"consensus info to vanguard client")
	for {
		select {
		case <-s.ctx.Done():
			log.Info("Received cancelled context, closing existing consensus info subscription")
			return nil

		case <-s.resubscribeEpochInfoCh:
			latestFinalizedEpoch := s.orchestratorDB.LatestLatestFinalizedEpoch()
			fromEpoch = latestFinalizedEpoch

			// checking consensus info db
			for i := latestFinalizedEpoch; i >= 0; {
				epochInfo, _ := s.orchestratorDB.ConsensusInfo(s.ctx, i)
				if epochInfo == nil {
					// epoch info is missing. so subscribe from here. maybe db operation was wrong
					latestFinalizedEpoch = i
					log.WithField("epoch", fromEpoch).Debug("Found missing epoch info in db, so subscription should " +
						"be started from this missing epoch")
				}
				if i == 0 {
					break
				}
				i--
			}
			log.WithField("fromEpoch", fromEpoch).Info("Re-subscribing to new vanguard epoch info stream")
			stream, err = client.StreamMinimalConsensusInfo(fromEpoch)
			if nil != err {
				log.WithError(err).Error("Failed to re-subscribe to new vanguard epoch info info stream, Exiting go routine")
				return err
			}

		default:
			vanMinimalConsensusInfo, err := stream.Recv()
			if e, ok := status.FromError(err); ok {
				switch e.Code() {
				case codes.Canceled, codes.Internal, codes.Unavailable:
					log.WithError(err).Infof("Trying to restart connection. rpc status: %v", e.Code())
					s.waitForConnection()
					stream, err = client.StreamMinimalConsensusInfo(fromEpoch)
					if nil != err {
						log.WithError(err).Error("Failed to subscribe to stream of new consensus info, Exiting go routine")
						return err
					}
				}
			}

			if nil == vanMinimalConsensusInfo {
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
			if err := s.OnNewConsensusInfo(s.ctx, consensusInfo); err != nil {
				log.WithError(err).Error("Closing epoch info subscription, Exiting go routine")
				return err
			}
		}
	}

	return nil
}
