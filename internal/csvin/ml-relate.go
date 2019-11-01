package csvin

import (
	"encoding/csv"
	"os"
	"strconv"

	"github.com/rhagenson/relped/internal/util"
	log "github.com/sirupsen/logrus"
)

var _ CsvInput = new(MLRelateCsv)

type MLRelateCsv struct {
	rels     map[string]map[string]float64
	dists    map[string]map[string]uint
	indvs    map[string]struct{}
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
		rels:  make(map[string]map[string]float64, len(records)),
		dists: make(map[string]map[string]uint, len(records)),
		indvs: make(map[string]struct{}, len(records)),
		min:   0,
		max:   1,
	}

	for i := range records {
		i1 := records[i][0]
		i2 := records[i][1]
		rel := 0.0

		// Set relatedness value
		if val, err := strconv.ParseFloat(records[i][9], 64); err == nil {
			if val < 0 { // Negative value just means unrelated
				rel = 0
			} else {
				rel = val
			}
		}
		if _, ok := c.rels[i1]; ok {
			c.rels[i1][i2] = rel
		} else {
			c.rels[i1] = make(map[string]float64, len(records))
			c.rels[i1][i2] = rel
		}
		if rel < c.min {
			c.min = rel
		}
		if c.max < rel {
			c.max = rel
		}
		// Add individuals to set for building non-redundant list of individuals
		c.indvs[i1] = struct{}{}
		c.indvs[i2] = struct{}{}

		// Set relational distances
		if dist, err := util.MLRelateToDist(records[i][2]); err == nil {
			if _, ok := c.dists[i1]; ok {
				c.dists[i1][i2] = dist
			} else {
				c.dists[i1] = make(map[string]uint, len(records))
				c.dists[i1][i2] = dist
			}
		} else {
			log.Errorf("Problem reading ML-Relate R entry: %s", err)
		}
	}

	if normalize {
		for i1, m := range c.rels {
			for i2, rel := range m {
				c.rels[i1][i2] = (rel - c.min) / (c.max - c.min)
			}
		}
	}

	return c
}

func (c *MLRelateCsv) Indvs() map[string]struct{} {
	return c.indvs
}

func (c *MLRelateCsv) Relatedness(i1, i2 string) float64 {
	return c.rels[i1][i2]
}

func (c *MLRelateCsv) RelDistance(i1, i2 string) uint {
	return c.dists[i1][i2]
}
