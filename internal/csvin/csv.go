package csvin

import "github.com/rhagenson/relped/internal/unit"

type CsvInput interface {
	Indvs() []string
	Relatedness(i1, i2 string) unit.Relatedness
	RelDistance(i1, i2 string) unit.RelationalDistance
}
