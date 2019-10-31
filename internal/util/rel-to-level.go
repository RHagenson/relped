package util

import "math"

// RelToLevel computes the relational distance given the relatedness score
//
// Examples:
//     relToLevel(0.5)   --> (1, true)
//     relToLevel(0.25)  --> (2, true)
//     relToLevel(0.125) --> (3, true)
//     relToLevel(<=0)   --> (0, false) // Only "unrelated" case
func RelToLevel(x float64) (uint, bool) {
	if x <= 0 {
		return 0, false
	}
	return uint(math.Round(math.Log(1/x) / math.Log(2))), true
}
