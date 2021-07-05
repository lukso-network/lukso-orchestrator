package consensus

import (
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

func CompareShardingInfo(ob1, ob2 *types.HeaderHash) bool {
	if ob1 == nil && ob2 == nil {
		// in existing code this will happen. as some part may have no sharding info for testing.
		return true
	}
	// TODO: IT WILL OPEN AFTER RESOLVING HASHING PROBLEM IN VANGUARD
	//if !reflect.DeepEqual(ob1.PandoraShardHash, ob2.PandoraShardHash) {
	//	log.WithField("object1 hash", ob1.PandoraShardHash).
	//		WithField("object2 hash", ob2.PandoraShardHash).
	//		Error("hash mismatched")
	//	return false
	//}
	//if !reflect.DeepEqual(ob1.Signature, ob2.Signature) {
	//	log.WithField("object1 Signature", ob1.Signature).
	//		WithField("object2 Signature", ob2.Signature).
	//		Error("Signature mismatched")
	//	return false
	//}
	return true
}
