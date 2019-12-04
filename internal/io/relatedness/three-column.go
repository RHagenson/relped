package relatedness

import (
	"os"
	"strconv"

	mapset "github.com/deckarep/golang-set"
	"github.com/gocarina/gocsv"
	"github.com/rhagenson/relped/internal/unit"
	"github.com/rhagenson/relped/internal/unit/relational"
	"github.com/rhagenson/relped/internal/util"
	log "github.com/sirupsen/logrus"
)

var _ CsvInput = new(ThreeColumnCsv)

type ThreeColumnCsv struct {
	rels     map[string]map[string]unit.Relatedness
	dists    map[string]map[string]relational.Degree
	indvs    mapset.Set
	min, max float64
}

func NewThreeColumnCsv(f *os.File, normalize bool) *ThreeColumnCsv {
	type entry struct {
		ID1 string `csv:"ID1"`
		ID2 string `csv:"ID2"`
		Rel string `csv:"Rel"`
	}
	entries := make([]*entry, 0, 100)

	gocsv.FailIfUnmatchedStructTags = true
	if err := gocsv.UnmarshalFile(f, &entries); err != nil {
		log.Fatalf("Misread in CSV: %s, rename column to match names used here\n", err)
	}

	c := &ThreeColumnCsv{
		rels:  make(map[string]map[string]unit.Relatedness, len(entries)),
		dists: make(map[string]map[string]relational.Degree, len(entries)),
		indvs: mapset.NewSet(),
		min:   0,
		max:   1,
	}

	pairs := make(map[string][]string, len(entries))
	for _, e := range entries {
		from := e.ID1
		to := e.ID2
		rel := e.Rel

		if vs, ok := pairs[from]; ok {
			for _, v := range vs {
				if v == to {
					log.Warnf("Relatedness pair ID %q and ID %q duplicated, using: %+v\n", from, to, e)
				}
			}
		}

		if _, ok := c.rels[from]; !ok {
			c.rels[from] = make(map[string]unit.Relatedness, len(entries))
		}
		if _, ok := c.dists[from]; !ok {
			c.dists[from] = make(map[string]relational.Degree, len(entries))
		}

		// Set relatedness and distance values
		if val, err := strconv.ParseFloat(rel, 64); err == nil {
			c.dists[from][to] = util.RelToLevel(val)
			if 0 < val {
				c.addRelatedness(from, to, val)
			} else { // Negative value just means unrelated
				c.addRelatedness(from, to, 0.0)
			}
		} else {
			c.dists[from][to] = util.CategoryToDist(rel)
			switch rel {
			case "PO":
				c.addRelatedness(from, to, 0.5)
			case "FS":
				c.addRelatedness(from, to, 0.25)
			case "HS":
				c.addRelatedness(from, to, 0.125)
			case "U":
				c.addRelatedness(from, to, 0.0)
			default:
				c.addRelatedness(from, to, 0.0)
			}
		}

		c.indvs.Add(from)
		c.indvs.Add(to)
		if _, ok := pairs[from]; ok {
			pairs[from] = append(pairs[from], to)
		} else {
			pairs[from] = make([]string, 0, len(entries))
			pairs[from] = append(pairs[from], to)
		}
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

func (c *ThreeColumnCsv) addRelatedness(from, to string, rel float64) {
	c.rels[from][to] = unit.Relatedness(rel)
	if rel < c.min {
		c.min = rel
	}
	if c.max < rel {
		c.max = rel
	}
}

func (c *ThreeColumnCsv) Indvs() mapset.Set {
	return c.indvs.Clone()
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
