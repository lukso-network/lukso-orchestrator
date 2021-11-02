package main

import "log"

type pandoraCanonicalHead struct {
	blockHash   string
	parentHash  string
	blockNumber uint64
}

func main() {
	experiment1 := staticExperiment1()
	statesAfter := processHeadsWithReorg(
		experiment1,
		"0xabc",
		[]uint64{0, 1, 2, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19},
	)
	log.Fatal(statesAfter)
}

func processHeadsWithReorg(
	statesBeforeReorg map[uint64]map[uint64]pandoraCanonicalHead,
	reorgParentHash string,
	reorgRange []uint64,
) (statesAfterReorg map[uint64]map[uint64]pandoraCanonicalHead) {
	statesAfterReorg = map[uint64]map[uint64]pandoraCanonicalHead{}

	for _, reorgNode := range reorgRange {
		for node, stateBeforeReorg := range statesBeforeReorg {
			if reorgNode != node {
				continue
			}

			for slot, head := range stateBeforeReorg {
				if head.blockHash != reorgParentHash {
					continue
				}

				newState := head
				_, exists := statesAfterReorg[node]

				if !exists {
					statesAfterReorg[node] = map[uint64]pandoraCanonicalHead{}
				}

				statesAfterReorg[node][slot] = newState
			}
		}
	}

	return
}

func staticExperiment1() (statesBeforeReorg map[uint64]map[uint64]pandoraCanonicalHead) {
	return map[uint64]map[uint64]pandoraCanonicalHead{
		//999: {
		//	220: {"0xabc", "0xab", 220},
		//	221: {"0xabc1", "0xabc", 221},
		//	222: {"0xabc2", "0xabc1", 222},
		//	223: {"0xabc3", "0xabc", 221},
		//	224: {"0xabc4", "0xabc3", 222},
		//	225: {"0xabc5", "0xabc4", 223},
		//},
		0: {
			220: {"0xabc", "0xab", 220},
			221: {"0xabc1", "0xabc", 221},
			222: {"0xabc2", "0xabc1", 222},
		},
		1: {
			220: {"0xabc", "0xab", 220},
			221: {"0xabc1", "0xabc", 221},
			222: {"0xabc2", "0xabc1", 222},
		},
		2: {
			220: {"0xabc", "0xab", 220},
			223: {"0xabc3", "0xabc", 221},
		},
		3: {
			220: {"0xabc", "0xab", 220},
			223: {"0xabc3", "0xabc", 221},
			224: {"0xabc4", "0xabc3", 222},
		},
		4: {
			220: {"0xabc", "0xab", 220},
			221: {"0xabc1", "0xabc", 221},
			222: {"0xabc2", "0xabc1", 222},
		},
		5: {
			220: {"0xabc", "0xab", 220},
			221: {"0xabc1", "0xabc", 221},
			222: {"0xabc2", "0xabc1", 222},
		},
		6: {
			220: {"0xabc", "0xab", 220},
			221: {"0xabc1", "0xabc", 221},
			222: {"0xabc2", "0xabc1", 222},
		},
		7: {
			220: {"0xabc", "0xab", 220},
			221: {"0xabc1", "0xabc", 221},
			222: {"0xabc2", "0xabc1", 222},
		},
		8: {
			220: {"0xabc", "0xab", 220},
			221: {"0xabc1", "0xabc", 221},
			222: {"0xabc2", "0xabc1", 222},
		},
		9: {
			220: {"0xabc", "0xab", 220},
			221: {"0xabc1", "0xabc", 221},
			222: {"0xabc2", "0xabc1", 222},
		},
		10: {
			220: {"0xabc", "0xab", 220},
			221: {"0xabc1", "0xabc", 221},
			222: {"0xabc2", "0xabc1", 222},
		},
		11: {
			220: {"0xabc", "0xab", 220},
			221: {"0xabc1", "0xabc", 221},
			222: {"0xabc2", "0xabc1", 222},
		},
		12: {
			220: {"0xabc", "0xab", 220},
			221: {"0xabc1", "0xabc", 221},
		},
		13: {
			220: {"0xabc", "0xab", 220},
		},
		14: {
			220: {"0xabc", "0xab", 220},
		},
		15: {
			220: {"0xabc", "0xab", 220},
		},
		16: {
			220: {"0xabc", "0xab", 220},
		},
		17: {
			220: {"0xabc", "0xab", 220},
		},
		18: {
			220: {"0xabc", "0xab", 220},
		},
		19: {
			220: {"0xabc", "0xab", 220},
		},
	}
}
