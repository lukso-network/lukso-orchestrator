package iface

import (
	"github.com/ethereum/go-ethereum/event"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

type PandoraService interface {
	SubscribeHeaderInfoEvent(chan<- *types.PandoraHeaderInfo) event.Subscription
	StopPandoraSubscription()
	ResumePandoraSubscription() error
}
