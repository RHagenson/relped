package util

import (
	"github.com/rhagenson/relped/internal/unit/relational"
)

// MLRelateToDist converts the category used by ML-Relate to
// its relational distance. Errors on unrecognized categories.
func MLRelateToDist(cat string) relational.Degree {
	switch cat {
	case "PO":
		return relational.First
	case "FS":
		return relational.Second
	case "HS":
		return relational.Third
	case "U":
		return relational.Unrelated
	default:
		return relational.Unrelated
	}
}
