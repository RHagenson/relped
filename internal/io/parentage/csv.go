package parentage

type CsvInput interface {
	Indvs() []string
	Sire(string) (string, bool)
	Dam(string) (string, bool)
}
