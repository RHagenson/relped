package relatedness

import (
	mapset "github.com/deckarep/golang-set"
	"github.com/rhagenson/relped/internal/unit"
	"github.com/rhagenson/relped/internal/unit/relational"
)

type CsvInput interface {
	Indvs() mapset.Set
	Relatedness(i1, i2 string) unit.Relatedness
	RelDistance(i1, i2 string) relational.Degree
}
