package demographics_test

import (
	"testing"

	"github.com/rhagenson/relped/internal/io/demographics"
)

// TODO: Make into a property test using gopter
func TestCalculateAge(t *testing.T) {
	tt := []struct {
		name  string
		cur   uint
		birth uint
		exp   demographics.Age
	}{
		{
			name:  "Birth in current year is zero",
			cur:   2019,
			birth: 2019,
			exp:   demographics.Age(0),
		},
		{
			name:  "Birth in future year is zero",
			cur:   2019,
			birth: 3000,
			exp:   demographics.Age(0),
		},
		{
			name:  "Birth in past year assumes birthday has passed",
			cur:   2019,
			birth: 2018,
			exp:   demographics.Age(1),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got := demographics.CalculateAge(tc.cur, tc.birth)
			if got != tc.exp {
				t.Errorf("Got %v, Expected %v", got, tc.exp)
			}
		})
	}
}
