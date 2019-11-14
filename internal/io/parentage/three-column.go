package parentage

import (
	"encoding/csv"
	"io"
	"os"

	mapset "github.com/deckarep/golang-set"
	"github.com/jszwec/csvutil"
	log "github.com/sirupsen/logrus"
)

var _ CsvInput = new(ThreeColumnCsv)

type ThreeColumnCsv struct {
	sires map[string]string
	dams  map[string]string
	indvs []string
}

func NewThreeColumnCsv(f *os.File) *ThreeColumnCsv {
	inCsv := csv.NewReader(f)
	dec, err := csvutil.NewDecoder(inCsv)
	if err != nil {
		log.Fatal(err)
	}

	type entry struct {
		ID   string `csv:"ID"`
		Sire string `csv:"Sire"`
		Dam  string `csv:"Dam"`
	}

	entries := make([]entry, 0, 100)
	for {
		var e entry

		if err := dec.Decode(&e); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		entries = append(entries, e)
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
