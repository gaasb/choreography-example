package domain

type Money int

func NewPrice(value int) Money {
	if value < 0 {
		return Money(0)
	}
	return Money(value)
}

func (m *Money) CalculateSum(money Money) {
	*m += money
}
func (m *Money) Multiply(n int) Money {
	return *m * Money(n)
}
func (m Money) ToInt() int {
	return int(m)
}
