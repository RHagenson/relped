package unit

type Weight float64

func (r Weight) Relatedness() Relatedness {
	return Relatedness(1.0 / r)
}
