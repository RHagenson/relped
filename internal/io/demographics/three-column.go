package demographics

import (
	"encoding/csv"
	"io"
	"os"
	"strings"

	"time"

	"github.com/jszwec/csvutil"
	log "github.com/sirupsen/logrus"
)

var _ CsvInput = new(ThreeColumnCsv)

type ThreeColumnCsv struct {
	ages  map[string]Age
	sexes map[string]Sex
}

func NewThreeColumnCsv(f *os.File) *ThreeColumnCsv {
	y := uint(time.Now().Year())
	inCsv := csv.NewReader(f)
	dec, err := csvutil.NewDecoder(inCsv)
	if err != nil {
		log.Fatal(err)
	}

	type entry struct {
		ID        string `csv:"ID"`
		Sex       string `csv:"Sex"`
		BirthYear uint   `csv:"Birth Year"`
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
		ages:  make(map[string]Age),
		sexes: make(map[string]Sex),
	}

	for _, e := range entries {
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
		c.ages[e.ID] = Age(y - uint(e.BirthYear))
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
