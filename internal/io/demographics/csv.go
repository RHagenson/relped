package demographics

type Age uint
type Sex uint

const (
	Unknown Sex = iota // Unknown is the default
	Female
	Male
)

type CsvInput interface {
	Age(string) (Age, bool)
	Sex(string) (Sex, bool)
}
