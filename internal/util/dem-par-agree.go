package util

import (
	"fmt"
	"strings"

	"github.com/rhagenson/relped/internal/io/demographics"
	"github.com/rhagenson/relped/internal/io/parentage"
)

func DemsAndParsAgree(dems demographics.CsvInput, pars parentage.CsvInput) string {
	if dems == nil || pars == nil {
		return ""
	}
	var s strings.Builder
	for _, indv := range pars.Indvs() {
		if sire, ok := pars.Sire(indv); ok {
			if sex, ok := dems.Sex(sire); ok {
				if sex != demographics.Male {
					s.WriteString(fmt.Sprintf("Sire %s for ID %s should be male, but is stated as %s in demographics\n", sire, indv, sex))
				}
			}
		}
		if dam, ok := pars.Dam(indv); ok {
			if sex, ok := dems.Sex(dam); ok {
				if sex != demographics.Female {
					s.WriteString(fmt.Sprintf("Dam %s for ID %s should be female, but is stated as %s in demographics\n", dam, indv, sex))
				}
			}
		}
	}
	return s.String()
}
