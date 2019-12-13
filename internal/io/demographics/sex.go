package demographics

type Sex uint

const (
	Unknown Sex = iota // Unknown is the default
	Female
	Male
)

func (s Sex) String() string {
	switch s {
	case Female:
		return "Female"
	case Male:
		return "Male"
	case Unknown:
		return "Unknown"
	default:
		return "N/A"
	}
}
