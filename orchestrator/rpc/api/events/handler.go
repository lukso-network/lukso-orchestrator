package events

import (
	"github.com/ethereum/go-ethereum/event"
	ethLog "github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"sync"
	"time"
)

// Type determines the kind of filter and is used to put the filter in to
// the correct bucket when added.
type Type byte

const (
	// UnknownSubscription indicates an unknown subscription type
	UnknownSubscription Type = iota

	// MinConsensusInfoSubscription
	MinConsensusInfoSubscription

	// LastSubscription keeps track of the last index
	LastIndexSubscription
)

// subscription
type subscription struct {
	id        rpc.ID
	typ       Type
	created   time.Time
	installed chan struct{} // closed when the filter is installed
	err       chan error    // closed when the filter is uninstalled

	epoch         uint64 // last served epoch number
	consensusInfo chan *types.MinimalEpochConsensusInfo
}

// EventSystem creates subscriptions, processes events and broadcasts them to the
// subscription which match the subscription criteria.
type EventSystem struct {
	backend Backend

	// Subscriptions
	consensusInfoSub event.Subscription // Subscription for new epoch validator list

	// Channels
	install         chan *subscription                    // install filter for event notification
	uninstall       chan *subscription                    // remove filter for event notification
	consensusInfoCh chan *types.MinimalEpochConsensusInfo // Channel to receive new new consensus info event
}

// NewEventSystem creates a new manager that listens for event on the given mux,
// parses and filters them. It uses the all map to retrieve filter changes. The
// work loop holds its own index that is used to forward events to filters.
//
// The returned manager has a loop that needs to be stopped with the Stop function
// or by stopping the given mux.
func NewEventSystem(backend Backend) *EventSystem {
	m := &EventSystem{
		backend:         backend,
		install:         make(chan *subscription),
		uninstall:       make(chan *subscription),
		consensusInfoCh: make(chan *types.MinimalEpochConsensusInfo, 1),
	}

	// Subscribe events
	m.consensusInfoSub = m.backend.SubscribeNewEpochEvent(m.consensusInfoCh)

	// Make sure none of the subscriptions are empty
	if m.consensusInfoSub == nil {
		ethLog.Crit("Subscribe for event system failed")
	}

	go m.eventLoop()
	return m
}

// Subscription is created when the client registers itself for a particular event.
type Subscription struct {
	ID        rpc.ID
	f         *subscription
	es        *EventSystem
	unsubOnce sync.Once
}

// Err returns a channel that is closed when unsubscribed.
func (sub *Subscription) Err() <-chan error {
	return sub.f.err
}

// Unsubscribe uninstalls the subscription from the event broadcast loop.
func (sub *Subscription) Unsubscribe() {
	sub.unsubOnce.Do(func() {
	uninstallLoop:
		for {
			// write uninstall request and consume logs/hashes. This prevents
			// the eventLoop broadcast method to deadlock when writing to the
			// filter event channel while the subscription loop is waiting for
			// this method to return (and thus not reading these events).
			select {
			case sub.es.uninstall <- sub.f:
				break uninstallLoop
			case <-sub.f.consensusInfo:
			}
		}

		// wait for filter to be uninstalled in work loop before returning
		// this ensures that the manager won't use the event channel which
		// will probably be closed by the client asap after this method returns.
		<-sub.Err()
	})
}

// subscribe installs the subscription in the event broadcast loop.
func (es *EventSystem) subscribe(sub *subscription) *Subscription {
	es.install <- sub
	<-sub.installed
	return &Subscription{ID: sub.id, f: sub, es: es}
}

// SubscribePendingTxs creates a subscription that writes transaction hashes for
// transactions that enter the transaction pool.
func (es *EventSystem) SubscribeConsensusInfo(consensusInfo chan *types.MinimalEpochConsensusInfo, epoch uint64) *Subscription {
	sub := &subscription{
		id:            rpc.NewID(),
		typ:           MinConsensusInfoSubscription,
		created:       time.Now(),
		epoch:         epoch,
		consensusInfo: consensusInfo,
		installed:     make(chan struct{}),
		err:           make(chan error),
	}
	return es.subscribe(sub)
}

type filterIndex map[Type]map[rpc.ID]*subscription

// handleConsensusInfoEvent
func (es *EventSystem) handleConsensusInfoEvent(filters filterIndex, ev *types.MinimalEpochConsensusInfo) {
	for _, f := range filters[MinConsensusInfoSubscription] {
		//if f.isNew {
		//	es.sendConsensusInfo(f, ev)
		//	f.isNew = false
		//	continue
		//}
		f.consensusInfo <- ev
	}
}

// eventLoop (un)installs filters and processes mux events.
func (es *EventSystem) eventLoop() {
	// Ensure all subscriptions get cleaned up
	defer func() {
		es.consensusInfoSub.Unsubscribe()
	}()

	index := make(filterIndex)
	for i := UnknownSubscription; i < LastIndexSubscription; i++ {
		index[i] = make(map[rpc.ID]*subscription)
	}

	for {
		select {
		case ev := <-es.consensusInfoCh:
			es.handleConsensusInfoEvent(index, ev)
		case f := <-es.install:
			index[f.typ][f.id] = f
			close(f.installed)
		case f := <-es.uninstall:
			delete(index[f.typ], f.id)
			close(f.err)

		// System stopped
		case <-es.consensusInfoSub.Err():
			return
		}
	}
}

// sendConsensusInfo
//func (es *EventSystem) sendConsensusInfo(f *subscription, ev *types.MinimalEpochConsensusInfo) {
//	curEpoch := es.backend.CurrentEpoch()
//	log.WithField("curEpoch", curEpoch).Debug("current epoch in epoch extractor")
//	log.WithField("f.epoch", f.epoch).Debug("subscriber's epoch status")
//
//	// when requested from epoch is greater or equal than current epoch.
//	if f.epoch >= curEpoch {
//		log.WithField("consensusInfo", ev).Debug("Sending consensus info to subscriber")
//		f.consensusInfo <- ev
//		return
//	}
//
//	consensusInfos := es.backend.ConsensusInfoByEpochRange(f.epoch)
//	log.WithField("consensusInfos", consensusInfos).Debug("start sending consensus infos")
//
//	// sending previous epoch consensus infos
//	for _, consensusInfo := range consensusInfos {
//		log.WithField("consensusInfo", consensusInfo).Debug("Sending consensus info to subscriber")
//		f.consensusInfo <- consensusInfo
//	}
//}
