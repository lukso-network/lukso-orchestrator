package vanguardchain

import (
	"context"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

// OnNewConsensusInfo
func (s *Service) OnNewConsensusInfo(ctx context.Context, consensusInfo *types.MinimalEpochConsensusInfo) {
	nsent := s.consensusInfoFeed.Send(consensusInfo)
	log.WithField("nsent", nsent).Trace("Send consensus info to subscribers")

	if err := s.db.SaveConsensusInfo(ctx, consensusInfo); err != nil {
		log.WithError(err).Warn("failed to save consensus info into db!")
		return
	}
}

// OnConsensusSubError
func (s *Service) OnConsensusSubError(err error) {
	s.conInfoSubErrCh <- err
}
