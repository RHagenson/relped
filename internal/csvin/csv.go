package csvin

type CsvInput interface {
	Indvs() map[string]struct{}
	Relatedness(i1, i2 string) float64
	RelDistance(i1, i2 string) uint
}
