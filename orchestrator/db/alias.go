package db

import "github.com/lukso-network/lukso-orchestrator/orchestrator/db/iface"

type ROnlyConsensusInfoDB = iface.ReadOnlyConsensusInfoDatabase

type ConsensusInfoAccessDB = iface.ConsensusInfoAccessDatabase

type ROnlyInvalidSlotInfoDB = iface.ReadOnlyInvalidSlotInfoDatabase

type ROnlyVerifiedShardInfoDB = iface.ReadOnlyVerifiedShardInfoDatabase

type VerifiedShardInfoDB = iface.VerifiedShardInfoDatabase

type InvalidSlotInfoDB = iface.InvalidSlotDatabase

type Database = iface.Database
