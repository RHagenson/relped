package relational

type Degree uint

const (
	Unrelated Degree = iota
	First
	Second
	Third
	Fourth
	Fifth
	Sixth
	Seventh
	Eighth
	Ninth // Maximum estimatable relational distance accroding to [@doi:10.1016/j.ajhg.2016.05.020]
)
