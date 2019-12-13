package demographics

type Age uint

func CalculateAge(cur, birth uint) Age {
	if cur > birth {
		return Age(cur - birth)
	}
	return Age(0)
}
