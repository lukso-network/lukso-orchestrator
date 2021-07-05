package consensus

import (
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	log "github.com/sirupsen/logrus"
	"reflect"
)

func CompareShardingInfo(ob1, ob2 *types.HeaderHash) bool {
	if ob1 == nil && ob2 == nil {
		// in existing code this will happen. as some part may have no sharding info for testing.
		return true
	}
	if !reflect.DeepEqual(ob1.ParentHash, ob2.ParentHash) {
		log.WithField("object1 parentHash", ob1.ParentHash).
			WithField("object2 parentHash", ob2.ParentHash).
			Error("parent hash mismatched")
		return false
	}
	if !reflect.DeepEqual(ob1.ReceiptHash, ob2.ReceiptHash) {
		log.WithField("object1 ReceiptHash", ob1.ReceiptHash).
			WithField("object2 ReceiptHash", ob2.ReceiptHash).
			Error("ReceiptHash mismatched")
		return false
	}
	//if !reflect.DeepEqual(ob1.HeaderHash, ob2.HeaderHash) {
	//	log.WithField("object1 HeaderHash", ob1.HeaderHash).
	//		WithField("object2 HeaderHash", ob2.HeaderHash).
	//		Error("HeaderHash mismatched")
	//	return false
	//}
	if !reflect.DeepEqual(ob1.Signature, ob2.Signature) {
		log.WithField("object1 Signature", ob1.Signature).
			WithField("object2 Signature", ob2.Signature).
			Error("Signature mismatched")
		return false
	}
	if !reflect.DeepEqual(ob1.TxHash, ob2.TxHash) {
		log.WithField("object1 TxHash", ob1.TxHash).
			WithField("object2 TxHash", ob2.TxHash).
			Error("TxHash mismatched")
		return false
	}
	if !reflect.DeepEqual(ob1.StateRoot, ob2.StateRoot) {
		log.WithField("object1 StateRoot", ob1.StateRoot).
			WithField("object2 StateRoot", ob2.StateRoot).
			Error("StateRoot mismatched")
		return false
	}
	if ob1.BlockNumber != ob2.BlockNumber {
		log.WithField("object1 blockNumber", ob1.BlockNumber).
			WithField("object2 blockNumber", ob2.BlockNumber).
			Error("blockNumber mismatched")
		return false
	}
	return true
}
