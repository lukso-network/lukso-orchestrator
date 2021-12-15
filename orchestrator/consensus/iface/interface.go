package iface

import (
	"github.com/ethereum/go-ethereum/event"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

type VerifiedSlotInfoFeed interface {
	SubscribeVerifiedSlotInfoEvent(chan<- *types.SlotInfoWithStatus) event.Subscription
}

type ReorgInfoFeed interface {
	SubscribeReorgInfoEvent(chan<- *types.Reorg) event.Subscription
}
