package parentage

import (
	"os"

	mapset "github.com/deckarep/golang-set"
	"github.com/gocarina/gocsv"
	log "github.com/sirupsen/logrus"
)

var _ CsvInput = new(ThreeColumnCsv)

type ThreeColumnCsv struct {
	sires map[string]string
	dams  map[string]string
	indvs []string
}

func NewThreeColumnCsv(f *os.File) *ThreeColumnCsv {
	type entry struct {
		ID   string `csv:"ID"`
		Sire string `csv:"Sire"`
		Dam  string `csv:"Dam"`
	}

	entries := make([]entry, 0, 100)

	gocsv.FailIfUnmatchedStructTags = true
	if err := gocsv.UnmarshalFile(f, &entries); err != nil {
		log.Fatalf("Misread in CSV: %s, rename column to match names used here\n", err)
	}

	c := &ThreeColumnCsv{
		sires: make(map[string]string),
		dams:  make(map[string]string),
	}

	indvSet := mapset.NewSet()

	for i, e := range entries {
		if e.ID != "" || e.Sire != "" || e.Dam != "" {
			switch e.Sire {
			case "0", "?":
				// Do nothing
			default:
				indvSet.Add(e.ID)
				c.sires[e.ID] = e.Sire
			}
			switch e.Dam {
			case "0", "?":
				// Do nothing
			default:
				indvSet.Add(e.ID)
				c.dams[e.ID] = e.Dam
			}
		} else {
			log.Warnf("Problem reading entry #%d: ID: %s, Sire: %s, Dam: %s\n", i+1, e.ID, e.Sire, e.Dam)
		}
	}

	for _, indv := range indvSet.ToSlice() {
		c.indvs = append(c.indvs, indv.(string))
	}

	return c
}

func (c *ThreeColumnCsv) Sire(id string) (string, bool) {
	sire, ok := c.sires[id]
	return sire, ok
}
func (c *ThreeColumnCsv) Dam(id string) (string, bool) {
	dam, ok := c.dams[id]
	return dam, ok
}

func (c *ThreeColumnCsv) Indvs() []string {
	return c.indvs
}
