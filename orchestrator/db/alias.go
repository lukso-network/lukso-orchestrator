package db

import "github.com/lukso-network/lukso-orchestrator/orchestrator/db/iface"

type ROnlyConsensusInfoDB = iface.ReadOnlyConsensusInfoDatabase

type ConsensusInfoAccessDB = iface.ConsensusInfoAccessDatabase

type VanguardHeaderHashDB = iface.VanHeaderAccessDatabase

type ROnlyVanHeaderHashDB = iface.VanHeaderAccessDatabase

type PandoraHeaderHashDB = iface.PanHeaderAccessDatabase

type ROnlyPanHeaderHashDB = iface.ReadOnlyPanHeaderAccessDatabase

type ROnlyRealmDB = iface.RealmReadOnlyAccessDatabase

type RealmDB = iface.RealmAccessDatabase

type ROnlyVerifiedSlotInfoDB = iface.ReadOnlyVerifiedSlotInfoDatabase

type ROnlyInvalidSlotInfoDB = iface.ReadOnlyInvalidSlotInfoDatabase

type VerifiedSlotInfoDB = iface.VerifiedSlotDatabase

type InvalidSlotInfoDB = iface.InvalidSlotDatabase

type Database = iface.Database
