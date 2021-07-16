package iface

import (
	"github.com/ethereum/go-ethereum/event"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

type PandoraHeaderFeed interface {
	SubscribeHeaderInfoEvent(chan<- *types.PandoraHeaderInfo) event.Subscription
}
