package demographics

import (
	"os"
	"strings"

	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/gocarina/gocsv"
	log "github.com/sirupsen/logrus"
)

var _ CsvInput = new(ThreeColumnCsv)

type ThreeColumnCsv struct {
	ages  map[string]Age
	sexes map[string]Sex
}

func NewThreeColumnCsv(f *os.File) *ThreeColumnCsv {
	y := uint(time.Now().Year())
	type entry struct {
		ID        string `csv:"ID"`
		Sex       string `csv:"Sex"`
		BirthYear uint   `csv:"BirthYear"`
	}
	entries := make([]*entry, 0, 100)

	gocsv.FailIfUnmatchedStructTags = true
	if err := gocsv.UnmarshalFile(f, &entries); err != nil {
		log.Fatalf("Misread in CSV: %s, rename column to match names used here\n", err)
	}

	c := &ThreeColumnCsv{
		ages:  make(map[string]Age),
		sexes: make(map[string]Sex),
	}

	ids := mapset.NewSet()
	for _, e := range entries {
		if ids.Contains(e.ID) {
			log.Warnf("Demographics for ID %q duplicated, using: %+v\n", e.ID, e)
		}
		switch {
		case strings.ToUpper(e.Sex) == "F", strings.ToUpper(e.Sex) == "FEMALE":
			c.sexes[e.ID] = Female
		case strings.ToUpper(e.Sex) == "M", strings.ToUpper(e.Sex) == "MALE":
			c.sexes[e.ID] = Male
		case strings.ToUpper(e.Sex) == "U", strings.ToUpper(e.Sex) == "UNKNOWN":
			c.sexes[e.ID] = Unknown
		default:
			log.Warnf("Did not understand Sex in entry: %v; setting Sex to Unknown\n", e)
			c.sexes[e.ID] = Unknown
		}
		c.ages[e.ID] = Age(y - e.BirthYear)
		ids.Add(e.ID)
	}

	return c
}

func (c *ThreeColumnCsv) Age(id string) (Age, bool) {
	age, ok := c.ages[id]
	return age, ok
}

func (c *ThreeColumnCsv) Sex(id string) (Sex, bool) {
	sex, ok := c.sexes[id]
	return sex, ok
}
