package csvin

import (
	"encoding/csv"
	"os"
	"strconv"

	"github.com/rhagenson/relped/internal/util"
	log "github.com/sirupsen/logrus"
)

var _ CsvInput = new(ThreeColumnCsv)

type ThreeColumnCsv struct {
	rels     map[string]map[string]float64
	indvs    []string
	min, max float64
}

func NewThreeColumnCsv(f *os.File, normalize bool) *ThreeColumnCsv {
	inCsv := csv.NewReader(f)
	inCsv.FieldsPerRecord = 3 // Simple three column format: Indv1, Indv2, Relatedness
	records, err := inCsv.ReadAll()
	if err != nil {
		log.Errorf("Problem parsing line: %s\n", err)
	}
	records = util.RmHeader(records)

	c := &ThreeColumnCsv{
		rels:  make(map[string]map[string]float64, len(records)),
		indvs: make([]string, 0, len(records)),
		min:   0,
		max:   1,
	}

	indvMap := make(map[string]struct{}, len(records))

	for i := range records {
		i1 := records[i][0]
		i2 := records[i][1]
		rel := 0.0
		if val, err := strconv.ParseFloat(records[i][2], 64); err == nil {
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

func (c *ThreeColumnCsv) Indvs() []string {
	return c.indvs
}

func (c *ThreeColumnCsv) Relatedness(i1, i2 string) float64 {
	return c.rels[i1][i2]
}
func (c *ThreeColumnCsv) RelDistance(i1, i2 string) uint {
	return util.RelToLevel(c.Relatedness(i1, i2))
}
