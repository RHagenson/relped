package unit

type Relatedness float64

func (r Relatedness) Weight() Weight {
	return Weight(1.0 / r)
}
