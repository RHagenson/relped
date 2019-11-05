package unit

type RelationalDistance uint
type GraphDistance uint

func (r RelationalDistance) GraphDistance() GraphDistance {
	if r == 0 {
		return 0
	}
	return GraphDistance(r - 1)
}

func (r GraphDistance) RelationalDistance() RelationalDistance {
	return RelationalDistance(r + 1)
}
