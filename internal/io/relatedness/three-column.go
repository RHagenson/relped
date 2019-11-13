package relatedness

import (
	"encoding/csv"
	"os"
	"strconv"

	"github.com/rhagenson/relped/internal/unit"
	"github.com/rhagenson/relped/internal/unit/relational"
	"github.com/rhagenson/relped/internal/util"
	log "github.com/sirupsen/logrus"
)

var _ CsvInput = new(ThreeColumnCsv)

type ThreeColumnCsv struct {
	rels     map[string]map[string]unit.Relatedness
	indvs    []string
	min, max float64
}

func NewThreeColumnCsv(f *os.File, normalize bool) *ThreeColumnCsv {
	inCsv := csv.NewReader(f)
	// Simple three column format:
	// Indv1, Indv2, Relatedness
	inCsv.FieldsPerRecord = 3
	records, err := inCsv.ReadAll()
	if err != nil {
		log.Errorf("Problem parsing line: %s\n", err)
	}
	records = util.RmHeader(records)

	c := &ThreeColumnCsv{
		rels:  make(map[string]map[string]unit.Relatedness, len(records)),
		indvs: make([]string, 0, len(records)),
		min:   0,
		max:   1,
	}

	indvMap := make(map[string]struct{}, len(records))

	for i := range records {
		from := records[i][0]
		to := records[i][1]
		rel := 0.0

		// Set relatedness value
		if val, err := strconv.ParseFloat(records[i][2], 64); err == nil {
			if val < 0 { // Negative value just means unrelated
				rel = 0
			} else {
				rel = val
			}
		}
		if _, ok := c.rels[from]; ok {
			c.rels[from][to] = unit.Relatedness(rel)
		} else {
			c.rels[from] = make(map[string]unit.Relatedness, len(records))
			c.rels[from][to] = unit.Relatedness(rel)
		}

		// Determine max and min
		if rel < c.min {
			c.min = rel
		}
		if c.max < rel {
			c.max = rel
		}

		// Add individuals to set for building non-redundant list of individuals
		indvMap[from] = struct{}{}
		indvMap[to] = struct{}{}
	}

	for indv := range indvMap {
		c.indvs = append(c.indvs, indv)
	}

	if normalize {
		for from, m := range c.rels {
			for to, rel := range m {
				c.rels[from][to] = unit.Relatedness((float64(rel) - c.min) / (c.max - c.min))
			}
		}
	}

	return c
}

func (c *ThreeColumnCsv) Indvs() []string {
	return c.indvs
}

func (c *ThreeColumnCsv) Relatedness(from, to string) unit.Relatedness {
	if innerRels, ok := c.rels[from]; ok {
		if val, ok := innerRels[to]; ok {
			return unit.Relatedness(val)
		}
	}
	if innerRels, ok := c.rels[to]; ok {
		if val, ok := innerRels[from]; ok {
			return unit.Relatedness(val)
		}
	}
	return unit.Relatedness(0)
}

func (c *ThreeColumnCsv) RelDistance(from, to string) relational.Degree {
	return util.RelToLevel(float64(c.Relatedness(from, to)))
}
