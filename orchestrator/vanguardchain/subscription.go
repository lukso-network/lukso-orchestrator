package vanguardchain

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
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
func (s *Service) subscribeVanNewPendingBlockHash(
	client client.VanguardClient,
) error {

	latestVerifiedSlot := s.orchestratorDB.LatestSavedVerifiedSlot()
	latestVerifiedSlotInfo, err := s.orchestratorDB.VerifiedSlotInfo(latestVerifiedSlot)
	var blockRoot []byte
	if err != nil {
		log.WithField("latestVerifiedSlot", latestVerifiedSlot).
			WithError(err).
			Warn("Failed to retrieve latest verified slot info for pending block subscription")
		return err
	}

	if latestVerifiedSlotInfo != nil {
		blockRoot = latestVerifiedSlotInfo.VanguardBlockHash.Bytes()
	}

	if latestVerifiedSlot == 0 {
		latestVerifiedSlot = latestVerifiedSlot + 1
	}

	// subscribe from a safe location.
	if latestVerifiedSlot > 32 {
		latestVerifiedSlot -= 32
	}

	stream, err := client.StreamNewPendingBlocks(blockRoot, eth2Types.Slot(latestVerifiedSlot))
	if nil != err {
		log.WithError(err).Error("Failed to subscribe to stream of new pending blocks")
		return err
	}

	log.WithField("fromSlot", latestVerifiedSlot).
		WithField("blockRoot", hexutil.Encode(blockRoot)).
		Info("Successfully subscribed to vanguard blocks")

	go func() {
		for {
			select {
			case <-s.ctx.Done():
				log.Debug("closing subscribeVanNewPendingBlockHash")
				return
			default:
				vanBlock, err := stream.Recv()

				if e, ok := status.FromError(err); ok {
					switch e.Code() {
					case codes.Canceled, codes.Internal, codes.Unavailable:
						log.WithError(err).Infof("Trying to restart connection. rpc status: %v", e.Code())
						s.conInfoSubErrCh <- err
						return
					}
				}

				if err := s.OnNewPendingVanguardBlock(s.ctx, vanBlock); err != nil {
					log.WithError(err).Error("Failed to process the pending vanguard shardInfo")
					s.conInfoSubErrCh <- errShardInfoProcess
					return
				}
			}
		}
	}()
	return nil
}

// subscribeNewConsensusInfoGRPC
func (s *Service) subscribeNewConsensusInfoGRPC(client client.VanguardClient) error {
	fromEpoch := s.orchestratorDB.LatestSavedEpoch()
	log.WithField("fromEpoch", fromEpoch).Debug("initial from value subscribeNewConsensusInfoGRPC")
	for i := s.orchestratorDB.LatestSavedEpoch(); i >= 0; {
		epochInfo, _ := s.orchestratorDB.ConsensusInfo(s.ctx, i)
		if epochInfo == nil {
			// epoch info is missing. so subscribe from here. maybe db operation was wrong
			fromEpoch = i
			log.WithField("fromEpoch", fromEpoch).Debug("setting from Epoch inside subscribeNewConsensusInfoGRPC")
		}
		if i == 0 {
			break
		}
		i--
	}
	log.WithField("fromEpoch", fromEpoch).Debug("requesting from value subscribeNewConsensusInfoGRPC")
	stream, err := client.StreamMinimalConsensusInfo(fromEpoch)
	if nil != err {
		log.WithError(err).Error("Failed to subscribe to stream of new consensus info")
		return err
	}

	log.WithField("fromEpoch", fromEpoch).
		Info("Successfully subscribed to minimal consensus info to vanguard client")

	go func() {
		for {
			select {
			case <-s.ctx.Done():
				log.Info("Received cancelled context, closing existing consensus info subscription")
				return

			default:
				vanMinimalConsensusInfo, err := stream.Recv()
				if e, ok := status.FromError(err); ok {
					switch e.Code() {
					case codes.Canceled, codes.Internal, codes.Unavailable:
						log.WithError(err).Infof("Trying to restart connection. rpc status: %v", e.Code())
						s.conInfoSubErrCh <- err
						return
					}
				}

				if nil == vanMinimalConsensusInfo {
					log.Error("Received nil consensus info")
					s.conInfoSubErrCh <- errConsensusInfoNil
					return
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
						NewSlot: uint64(vanMinimalConsensusInfo.ReorgInfo.NewSlot),
					}
					consensusInfo.ReorgInfo = reorgInfo
				}
				// Only non empty check for now
				if len(consensusInfo.ValidatorList) < 1 {
					log.WithField("consensusInfo", consensusInfo).WithField("err", err).Error(
						"empty validator list")
					s.conInfoSubErrCh <- errInvalidValidatorLength
					return
				}

				log.WithField("epoch", vanMinimalConsensusInfo.Epoch).
					WithField("epochInfo", fmt.Sprintf("%+v", vanMinimalConsensusInfo)).
					Debug("Received new consensus info for next epoch")
				if err := s.OnNewConsensusInfo(s.ctx, consensusInfo); err != nil {
					s.conInfoSubErrCh <- errConsensusInfoProcess
					return
				}
			}
		}
	}()

	return nil
}
