package util

import (
	"github.com/rhagenson/relped/internal/unit/relational"
)

// CategoryToDist converts the category used by ML-Relate to
// its relational distance. Errors on unrecognized categories.
func CategoryToDist(cat string) relational.Degree {
	switch cat {
	case "PO":
		return relational.First // PO should have no nodes between them: direct link
	case "FS":
		return relational.Second // FS should have have paths of one node between them: both shared parents
	case "HS":
		return relational.Second // HS should only have one node between them: the shared parent
	case "U":
		return relational.Unrelated
	default:
		return relational.Unrelated
	}
}
