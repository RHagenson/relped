package util

import "fmt"

// MLRelateToDist converts the category used by ML-Relate to
// its relational distance. Errors on unrecognized categories.
func MLRelateToDist(cat string) (uint, error) {
	switch cat {
	case "PO":
		return 1, nil
	case "FS":
		return 2, nil
	case "HS":
		return 3, nil
	case "U":
		return 0, nil
	default:
		return 0, fmt.Errorf("entry %q not understood", cat)
	}
}
