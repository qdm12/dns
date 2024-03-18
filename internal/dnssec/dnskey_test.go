package dnssec

import "testing"

var testGlobalMap = map[uint8]uint8{ //nolint:gochecknoglobals
	1: 1,
	2: 2,
	3: 3,
	4: 4,
	5: 5,
	6: 6,
	7: 7,
	8: 8,
}

func testSwitchStatement(key uint8) uint8 {
	switch key {
	case 1:
		return 1
	case 2:
		return 2
	case 3:
		return 3
	case 4:
		return 4
	case 5:
		return 5
	case 6:
		return 6
	case 7:
		return 7
	case 8:
		return 8
	default:
		panic("invalid key")
	}
}

// This benchmark aims to check if, for algoIDToPreference, it is
// better to:
// 1. have a global map variable
// 2. have a function with a switch statement
// The second point at equal performance is better due to its
// immutability nature, unlike 1.
func Benchmark_globalMap_switch(b *testing.B) {
	b.Run("global_map", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = testGlobalMap[1]
		}
	})

	b.Run("switch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = testSwitchStatement(1)
		}
	})
}
