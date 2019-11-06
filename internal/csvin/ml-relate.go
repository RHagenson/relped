package csvin

import (
	"encoding/csv"
	"os"
	"strconv"

	"github.com/rhagenson/relped/internal/unit"
	"github.com/rhagenson/relped/internal/unit/relational"
	"github.com/rhagenson/relped/internal/util"
	log "github.com/sirupsen/logrus"
)

var _ CsvInput = new(MLRelateCsv)

type MLRelateCsv struct {
	rels     map[string]map[string]unit.Relatedness
	dists    map[string]map[string]relational.Degree
	indvs    []string
	min, max float64
}

func NewMLRelateCsv(f *os.File, normalize bool) *MLRelateCsv {
	inCsv := csv.NewReader(f)
	// Columns:
	// Ind1, Ind2, R, LnL.R., U, HS, FS, PO, Relationships, Relatedness
	inCsv.FieldsPerRecord = 10
	records, err := inCsv.ReadAll()
	if err != nil {
		log.Errorf("Problem parsing line: %s\n", err)
	}
	records = util.RmHeader(records)

	c := &MLRelateCsv{
		rels:  make(map[string]map[string]unit.Relatedness, len(records)),
		dists: make(map[string]map[string]relational.Degree, len(records)),
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
		if val, err := strconv.ParseFloat(records[i][9], 64); err == nil {
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

		// Set relational distances
		if _, ok := c.dists[from]; ok {
			c.dists[from][to] = util.MLRelateToDist(records[i][2])
		} else {
			c.dists[from] = make(map[string]relational.Degree, len(records))
			c.dists[from][to] = util.MLRelateToDist(records[i][2])
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

func (c *MLRelateCsv) Indvs() []string {
	return c.indvs
}

func (c *MLRelateCsv) Relatedness(from, to string) unit.Relatedness {
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

func (c *MLRelateCsv) RelDistance(from, to string) relational.Degree {
	if innerRels, ok := c.dists[from]; ok {
		if val, ok := innerRels[to]; ok {
			return val
		}
	}
	if innerRels, ok := c.dists[to]; ok {
		if val, ok := innerRels[from]; ok {
			return val
		}
	}
	return relational.Unrelated
}
