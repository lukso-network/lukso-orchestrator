package db

import "github.com/lukso-network/lukso-orchestrator/orchestrator/db/iface"

type ROnlyConsensusInfoDB = iface.ReadOnlyConsensusInfoDatabase

type ConsensusInfoAccessDB = iface.ConsensusInfoAccessDatabase

type ROnlyVerifiedSlotInfoDB = iface.ReadOnlyVerifiedSlotInfoDatabase

type ROnlyInvalidSlotInfoDB = iface.ReadOnlyInvalidSlotInfoDatabase

type VerifiedSlotInfoDB = iface.VerifiedSlotDatabase

type InvalidSlotInfoDB = iface.InvalidSlotDatabase

type Database = iface.Database
