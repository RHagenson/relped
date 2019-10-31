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
		rels:  make(map[string]map[string]float64, len(records)),
		dists: make(map[string]map[string]uint, len(records)),
		indvs: make([]string, 0, len(records)),
		min:   0,
		max:   1,
	}

	indvMap := make(map[string]struct{}, len(records))
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

		// Determine minimum and maximum
		if rel < c.min {
			c.min = rel
		}
		if c.max < rel {
			c.max = rel
		}

		// Add individuals to set for building non-redundant list of individuals
		indvMap[i1] = struct{}{}
		indvMap[i2] = struct{}{}
	}
	for indv := range indvMap {
		c.indvs = append(c.indvs, indv)
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

func (c *MLRelateCsv) Indvs() []string {
	return c.indvs
}

func (c *MLRelateCsv) Relatedness(i1, i2 string) float64 {
	return c.rels[i1][i2]
}

func (c *MLRelateCsv) RelDistance(i1, i2 string) uint {
	return c.dists[i1][i2]
}
