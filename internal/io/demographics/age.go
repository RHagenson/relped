package demographics

type Age uint

func CalculateAge(cur, birth uint) Age {
	return Age(cur - birth)
}
