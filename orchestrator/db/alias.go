package db

import "github.com/lukso-network/lukso-orchestrator/orchestrator/db/iface"

type ROnlyConsensusInfoDB = iface.ReadOnlyConsensusInfoDatabase

type ConsensusInfoAccessDB = iface.ConsensusInfoAccessDatabase

type VanguardHeaderHashDB = iface.VanHeaderAccessDatabase

type ROnlyVanHeaderHashDB = iface.VanHeaderAccessDatabase

type PandoraHeaderHashDB = iface.PanHeaderAccessDatabase

type ROnlyPanHeaderHashDB = iface.ReadOnlyPanHeaderAccessDatabase

type Database = iface.Database
