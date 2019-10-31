package util

import "math"

// RelToLevel computes the relational distance given the relatedness score
//
// Examples:
//     relToLevel(0.5)   --> 1
//     relToLevel(0.25)  --> 2
//     relToLevel(0.125) --> 3
//     relToLevel(<=0)   --> 0  // Only "unrelated" case
func RelToLevel(x float64) uint {
	if x <= 0 {
		return 0
	}
	return uint(math.Round(math.Log(1/x) / math.Log(2)))
}
