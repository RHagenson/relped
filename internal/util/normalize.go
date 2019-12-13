package util

import (
	"github.com/rhagenson/relped/internal/unit"
)

// NormalizeRelatedness normalizes the Relatedness values to be [0,1]-bounded
// if all values are already between [0,1] NormalizeRelatedness does nothing
func NormalizeRelatedness(rels map[string]map[string]unit.Relatedness) map[string]map[string]unit.Relatedness {
	var min, max = 0.0, 1.0
	var relVal float64

	for _, m := range rels {
		for _, rel := range m {
			relVal = float64(rel)
			if relVal < min {
				min = relVal
			} else if max < relVal {
				max = relVal
			}
		}
	}
	if min == 0.0 && max == 1.0 {
		return rels
	}

	cp := make(map[string]map[string]unit.Relatedness, len(rels))
	for from, m := range rels {
		for to, rel := range m {
			if _, ok := cp[from]; !ok {
				cp[from] = make(map[string]unit.Relatedness, len(m))
			}
			cp[from][to] = unit.Relatedness((float64(rel) - min) / (max - min))
		}
	}
	return cp
}
