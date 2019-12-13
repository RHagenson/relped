package demographics

type CsvInput interface {
	Age(string) (Age, bool)
	Sex(string) (Sex, bool)
	Indvs() []string
}
