package demographics

import (
	"encoding/csv"
	"os"
	"strconv"

	"time"

	"github.com/rhagenson/relped/internal/util"
	log "github.com/sirupsen/logrus"
)

var _ CsvInput = new(ThreeColumnCsv)

type ThreeColumnCsv struct {
	ages  map[string]Age
	sexes map[string]Sex
}

func NewThreeColumnCsv(f *os.File) *ThreeColumnCsv {
	c := &ThreeColumnCsv{
		ages:  make(map[string]Age),
		sexes: make(map[string]Sex),
	}
	inCsv := csv.NewReader(f)
	// Simple three column format:
	// ID, Sex, Birth Year
	inCsv.FieldsPerRecord = 3
	records, err := inCsv.ReadAll()
	if err != nil {
		log.Errorf("Problem parsing line: %s\n", err)
	}
	records = util.RmHeader(records)
	y := uint(time.Now().Year())

	for i := range records {
		id := records[i][0]
		sex := records[i][1]

		if birthYear, err := strconv.ParseUint(records[i][2], 10, 64); err != nil {
			c.ages[id] = Age(y - uint(birthYear))
		}
		switch sex {
		case "Female", "F", "FEMALE", "female":
			c.sexes[id] = Female
		case "Male", "M", "MALE", "male":
			c.sexes[id] = Male
		case "Unknown", "U", "UNKNOWN", "unknown":
			c.sexes[id] = Unknown
		default:
			log.Warnf("Did not understand Sex in entry: %v; setting Sex to Unknown\n", records[i])
			c.sexes[id] = Unknown
		}
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
