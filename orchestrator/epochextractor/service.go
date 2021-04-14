package epochextractor

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/epochextractor/pandora"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/epochextractor/vanguard"
	types "github.com/prysmaticlabs/eth2-types"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"sort"
	"sync"
	"time"
)

type Service struct {
	connectedPandora         bool
	connectedVanguard        bool
	isRunning                bool
	processingLock           sync.RWMutex
	ctx                      context.Context
	cancel                   context.CancelFunc
	headTicker               *time.Ticker
	pandoraHttpEndpoint      string
	vanguardHttpEndpoint     string
	pandoraClient            pandora.PandoraClient
	vanguardClient           vanguard.VanguardClient
	runError                 error
	curEpoch                 types.Epoch
	nextEpoch                types.Epoch
	curSlot                  types.Slot
	lastSlotCurEpoch         types.Slot
	curEpochProposerPubKeys  []string //
	nextEpochProposerPubKeys []string //
	sortedSlots              []types.Slot
	genesisTime              uint64
	secondsPerSlot           uint64
	isEpochProcessed        map[types.Epoch]bool
}

type Config struct {
	pandoraHttpEndpoint  string
	vanguardHttpEndpoint string
}

func NewService(ctx context.Context, pandoraHttpEndpoint string, vanguardHttpEndpoint string, genesisTime uint64) (*Service, error) {

	ctx, cancel := context.WithCancel(ctx)
	_ = cancel // govet fix for lost cancel. Cancel is handled in service.Stop()
	return &Service{
		ctx:                      ctx,
		cancel:                   cancel,
		pandoraHttpEndpoint:      pandoraHttpEndpoint,
		vanguardHttpEndpoint:     vanguardHttpEndpoint,
		headTicker:               time.NewTicker(time.Duration(2) * time.Second),
		curEpochProposerPubKeys:  make([]string, 32),
		nextEpochProposerPubKeys: make([]string, 32),
		secondsPerSlot:           6,
		genesisTime:              genesisTime,
		sortedSlots:              make([]types.Slot, 64),
		isEpochProcessed:        make(map[types.Epoch]bool),
	}, nil
}

// Start a consensus info fetcher service's main event loop.
func (s *Service) Start() {
	// Exit early if eth1 endpoint is not set.
	if s.pandoraHttpEndpoint == "" || s.vanguardHttpEndpoint == "" {
		return
	}
	go func() {
		s.isRunning = true
		s.waitForConnection()
		if s.ctx.Err() != nil {
			log.Info("Context closed, exiting pandora goroutine")
			return
		}
		s.run(s.ctx.Done())
	}()
}

func (s *Service) Stop() error {
	if s.cancel != nil {
		defer s.cancel()
	}
	s.closeClients()
	return nil
}

func (s *Service) Status() error {
	// Service don't start
	if !s.isRunning {
		return nil
	}
	// get error from run function
	if s.runError != nil {
		return s.runError
	}
	return nil
}

// closes down our active eth1 clients.
func (s *Service) closeClients() {
	pandoraClient, ok := s.pandoraClient.(*pandora.RPCClient)
	if ok {
		pandoraClient.Close()
	}

	vanguardClient, ok := s.vanguardClient.(*vanguard.GRPCClient)
	if ok {
		vanguardClient.Close()
	}
}

func (s *Service) waitForConnection() {
	err := s.connectToPandoraChain()
	if err == nil {
		s.connectedPandora = true
	}

	err = s.connectToVanguardChain()
	if err == nil {
		s.connectedVanguard = true
		return
	}

	s.runError = err
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			log.Debugf("Trying to dial pandora and vanguard endpoint: %s and %s", s.pandoraHttpEndpoint, s.vanguardHttpEndpoint)
			errConnect := s.connectToPandoraChain()
			if errConnect != nil {
				log.Debug("Could not connect to pandora endpoint")
				s.runError = errConnect
				continue
			}

			errConnect = s.connectToVanguardChain()
			if errConnect != nil {
				log.Debug("Could not connect to vanguard endpoint")
				s.runError = errConnect
				continue
			}

			s.connectedPandora = true
			s.connectedVanguard = true
			s.runError = nil

			log.WithField("pandoraHttp", s.pandoraHttpEndpoint).WithField(
				"vanguardHttp", s.vanguardHttpEndpoint).Info("Connected to pandora and vanguard chain")
			return
		case <-s.ctx.Done():
			log.Debug("Received cancelled context,closing existing pandora and vanguard client service")
			return
		}
	}
}

// run subscribes to all the services for the ETH1.0 chain.
func (s *Service) run(done <-chan struct{}) {
	s.runError = nil

	curSlot, err := s.vanguardClient.CanonicalHeadSlot()
	if err != nil {
		log.WithError(err).Debug("could not fetch the head slot from pandora chain")
	}

	s.curSlot = curSlot
	s.curEpoch = types.Epoch(curSlot.DivSlot(32))
	s.nextEpoch = s.curEpoch.Add(1)
	s.lastSlotCurEpoch = types.Slot(s.curEpoch.Add(1).Mul(32).Sub(1))

	if err := s.initProposerList(); err != nil {
		log.WithError(err).Error("Error when bootstrapping consensus info")
	}

	for {
		select {
		case <-done:
			s.isRunning = false
			s.runError = nil
			s.connectedVanguard = false
			s.connectedPandora = false
			log.Debug("Context closed, exiting goroutine")
			return
		case <-s.headTicker.C:
			curSlot, err := s.vanguardClient.CanonicalHeadSlot()
			if err != nil {
				log.WithError(err).Debug("could not fetch the head slot from pandora chain")
				continue
			}

			s.curSlot = curSlot
			s.curEpoch = types.Epoch(curSlot.DivSlot(32))
			s.nextEpoch = s.curEpoch.Add(1)
			s.lastSlotCurEpoch = types.Slot(s.curEpoch.Add(1).Mul(32).Sub(1))

			log.WithField("headSlot", curSlot).WithField(
				"lastSlotCurEpoch", s.lastSlotCurEpoch).Debug("Chain head info")

			if s.lastSlotCurEpoch == curSlot && !s.isEpochProcessed[s.curEpoch] {
				// getting proposer list for next epoch(curEpoch + 1)
				assignments, err := s.vanguardClient.NextEpochProposerList()
				if err != nil {
					log.WithError(err).Debug("could not fetch the assignments from vanguard chain")
					continue
				}
				if err := s.processAssignments(assignments); err != nil {
					log.WithError(err).Debug("could not process the assignments")
					continue
				}
				if err := s.sendConsensusInfoToPandora(false); err != nil {
					log.WithError(err).Debug("could not insert the consensus info into pandora chain")
					continue
				}
				s.isEpochProcessed[s.curEpoch] = true
			}
		}
	}
}

func (s *Service) initProposerList() error {
	// getting proposer list for next epoch(curEpoch + 1)
	assignments, err := s.vanguardClient.NextEpochProposerList()
	if err != nil {
		log.WithError(err).Error("got error from NextEpochProposerList api")
		return err
	}

	if err := s.processAssignments(assignments); err != nil {
		return err
	}

	if err := s.sendConsensusInfoToPandora(true); err != nil {
		return err
	}

	return nil
}

func (s *Service) processAssignments(assignments *ethpb.ValidatorAssignments) error {
	slotToPubKey := make(map[types.Slot]string, 64)
	for _, assignment := range assignments.Assignments {
		for _, proposerSlot := range assignment.ProposerSlots {
			slotToPubKey[proposerSlot] = common.Bytes2Hex(assignment.PublicKey)
			//log.WithField("Slot", proposerSlot).WithField("pubKey", common.Bytes2Hex(assignment.PublicKey)[:2]).Info("proposer info")
		}
	}

	//log.WithField("slotToPubKey", slotToPubKey).Info("slotToPubkey")
	if slotToPubKey[types.Slot(0)] == "" && s.curEpoch == 0 {
		slotToPubKey[0] = "0x"
	}

	if len(slotToPubKey) != 64 {
		log.Error("Invalid length! len: ", len(slotToPubKey))
		return errors.New("Invalid length of proposer list")
	}

	s.sortSlots(slotToPubKey)
	//curChainEpoch := types.Epoch(s.sortedSlots[31].DivSlot(32))
	//if curChainEpoch != s.curEpoch {
	//	log.WithField("curChainEpoch", curChainEpoch).WithField(
	//		"finalizedEpoch", s.curEpoch).Error(
	//			"Vanguard chain does not achieve any finalization or checkpoint")
	//	return errors.New("Vanguard chain not advancing")
	//}

	for i, slot := range s.sortedSlots {
		if i < 32 {
			s.curEpochProposerPubKeys[i] = "0x" + slotToPubKey[slot]
		} else {
			s.nextEpochProposerPubKeys[i-32] = "0x" + slotToPubKey[slot]
		}
	}

	s.logProposerSchedule(slotToPubKey)
	return nil
}

func (s *Service) sendConsensusInfoToPandora(isBootstrapping bool) error {
	var status bool
	var err error
	if isBootstrapping {
		curEpochStart := s.genesisTime
		if s.curEpoch > 0 {
			curEpochStart = curEpochStart + uint64(s.curEpoch.Mul(s.secondsPerSlot*32))
		}
		log.WithField("epoch", s.curEpoch).WithField(
			"curEpochProposerPubKeys", s.curEpochProposerPubKeys).Debug(
			"Before sending consensus info to pandora in bootstrapping")
		status, err = s.pandoraClient.InsertConsensusInfo(s.ctx, s.curEpoch, s.curEpochProposerPubKeys, curEpochStart)
	} else {
		nextEpochStart := s.genesisTime + uint64(s.nextEpoch.Mul(s.secondsPerSlot*32))
		log.WithField("epoch", s.nextEpoch).WithField(
			"nextEpochProposerPubKeys", s.nextEpochProposerPubKeys).Debug(
			"Before sending consensus info to pandora")
		status, err = s.pandoraClient.InsertConsensusInfo(s.ctx, s.nextEpoch, s.nextEpochProposerPubKeys, nextEpochStart)
	}

	if err != nil {
		return err
	}

	if !status {
		log.Warn("Failed to insert consensus info into pandora chain")
	}

	log.Debug("Successfully inserted consensus info into pandora chain!")
	return nil
}

func (s *Service) logProposerSchedule(slotToPubKey map[types.Slot]string) {
	log.WithField("curEpoch", s.curEpoch).WithField("nextEpoch", s.nextEpoch).Info("Showing epoch info...")
	// To perform the opertion you want
	for _, slot := range s.sortedSlots {
		pubKey := slotToPubKey[slot]
		if len(slotToPubKey[slot]) > 12 {
			pubKey = slotToPubKey[slot][:12]
		}

		log.WithField("slot", slot).WithField(
			"proposerPubKey", "0x"+pubKey).Info(" Proposer schedule")
	}
}

func (s *Service) sortSlots(slotToPubKey map[types.Slot]string) {
	slots := make([]int, 64)
	i := 0
	log.WithField("len(slotToPubKey)", len(slotToPubKey)).Debug("slotToPubKey length")
	for k, _ := range slotToPubKey {
		slots[i] = int(uint64(k))
		i++
	}

	sort.Ints(slots)
	for idx, slot := range slots {
		s.sortedSlots[idx] = types.Slot(slot)
	}
}

func (s *Service) connectToPandoraChain() error {
	pandoraClient, err := pandora.Dial(s.ctx, s.pandoraHttpEndpoint)
	if err != nil {
		return err
	}
	s.pandoraClient = pandoraClient
	// Make a simple call to ensure we are actually connected to a working node.
	chainID, err := s.pandoraClient.ChainID(s.ctx)
	log.WithField("pandoraChainID", chainID).WithField("error", err).Debug("successfully retrieve chain id of pandora client")
	if err != nil {
		s.pandoraClient.Close()
		return err
	}
	return nil
}

func (s *Service) connectToVanguardChain() error {
	vanguardClient, err := vanguard.Dial(s.ctx, s.vanguardHttpEndpoint, 1*time.Second, 5, 4194304)
	if err != nil {
		return err
	}
	s.vanguardClient = vanguardClient
	// Make a simple call to ensure we are actually connected to a working node.
	_, err = s.vanguardClient.CanonicalHeadSlot()
	if err != nil {
		s.vanguardClient.Close()
		return err
	}
	return nil
}
