package util_test

import (
	"testing"

	"github.com/rhagenson/relped/internal/unit"
	"github.com/rhagenson/relped/internal/util"
)

// TODO: Make into a property test using gopter
func TestNormalizeRelatedness(t *testing.T) {
	tt := []struct {
		name string
		rels map[string]map[string]unit.Relatedness
		exp  map[string]map[string]unit.Relatedness
	}{
		{
			name: "Value over 1 becomes 1",
			rels: map[string]map[string]unit.Relatedness{
				"I1": map[string]unit.Relatedness{
					"I2": unit.Relatedness(100),
				},
			},
			exp: map[string]map[string]unit.Relatedness{
				"I1": map[string]unit.Relatedness{
					"I2": unit.Relatedness(1),
				},
			},
		},
		{
			name: "Value under 0 becomes 0",
			rels: map[string]map[string]unit.Relatedness{
				"I1": map[string]unit.Relatedness{
					"I2": unit.Relatedness(-100),
				},
			},
			exp: map[string]map[string]unit.Relatedness{
				"I1": map[string]unit.Relatedness{
					"I2": unit.Relatedness(0),
				},
			},
		},
		{
			name: "Values are forced into range [0,1]",
			rels: map[string]map[string]unit.Relatedness{
				"I1": map[string]unit.Relatedness{
					"I2": unit.Relatedness(100),
				},
				"I2": map[string]unit.Relatedness{
					"I3": unit.Relatedness(-100),
				},
			},
			exp: map[string]map[string]unit.Relatedness{
				"I1": map[string]unit.Relatedness{
					"I2": unit.Relatedness(1),
				},
				"I2": map[string]unit.Relatedness{
					"I3": unit.Relatedness(0),
				},
			},
		},
		{
			name: "Values in range do not change",
			rels: map[string]map[string]unit.Relatedness{
				"I1": map[string]unit.Relatedness{
					"I2": unit.Relatedness(0.25),
				},
				"I2": map[string]unit.Relatedness{
					"I3": unit.Relatedness(0.75),
				},
			},
			exp: map[string]map[string]unit.Relatedness{
				"I1": map[string]unit.Relatedness{
					"I2": unit.Relatedness(0.25),
				},
				"I2": map[string]unit.Relatedness{
					"I3": unit.Relatedness(0.75),
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got := util.NormalizeRelatedness(tc.rels)
			for from, m := range got {
				for to := range m {
					if got[from][to] != tc.exp[from][to] {
						t.Errorf("Got %v, Expected %v", got, tc.exp)
					}
				}
			}
		})
	}
}
