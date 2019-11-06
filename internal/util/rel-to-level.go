package util

import (
	"math"

	"github.com/rhagenson/relped/internal/unit/relational"
)

// RelToLevel computes the relational distance given the relatedness score
//
// Examples:
//     relToLevel(0.5)   --> First
//     relToLevel(0.25)  --> Second
//     relToLevel(0.125) --> Third
//	   ...
//     relToLevel(<=0)   --> Unrelated
func RelToLevel(x float64) relational.Degree {
	if x <= 0 {
		return 0
	}
	switch uint(math.Round(math.Log(1/x) / math.Log(2))) {
	case 1:
		return relational.First
	case 2:
		return relational.Second
	case 3:
		return relational.Third
	case 4:
		return relational.Fourth
	case 5:
		return relational.Fifth
	case 6:
		return relational.Sixth
	case 7:
		return relational.Seventh
	case 8:
		return relational.Eighth
	case 9:
		return relational.Ninth
	default:
		return relational.Unrelated
	}
}
