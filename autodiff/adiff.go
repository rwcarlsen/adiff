package autodiff

import "math"

type Number interface {
	Value() float64
	Deriv(i int)
}

type GeneralNumber struct {
	Val    float64
	Derivs []float64
}

var NDims int

func NewNumber() GeneralNumber         { return GeneralNumber{Derivs: make([]float64, NDims)} }
func (g GeneralNumber) Value() float64 { return g.Val }
func (g GeneralNumber) Deriv(i int) float64 {
	if i < len(g.Derivs) {
		return g.Derivs[i]
	}
	return 0
}

type Variable struct {
	Val   float64
	Index int
}

func NewVariable(index int, val float64) Variable { return Variable{Val: val, Index: index} }
func (v Variable) Value() float64                 { return v.Val }
func (v Variable) Deriv(i int) float64 {
	if i == v.Index {
		return 1
	}
	return 0
}

type Constant struct {
	Val float64
}

func NewConstant(val float64) Constant { return Constant{Val: val} }
func (c Constant) Value() float64      { return c.Val }
func (c Constant) Deriv(i int) float64 { return 0 }

func Ln(a Number) Number {
	result := NewNumber()
	result.Val = math.Log(a.Value())
	for i := 0; i < NDims; i++ {
		results.Deriv(i) = a.Deriv(i) / a.Value()
	}
	return result
}

func Add(nums ...Number) Number {
	if len(nums) == 0 {
		return GeneralNumber{}
	}

	result := NewNumberLike(nums[0])
	for _, n := range nums {
		result.Val += n.Val
		for i := 0; i < NDims; i++ {
			result.Deriv(i) += n.Deriv(i)
		}
	}
	return result
}

func Mult(nums ...Number) Number {
	if len(nums) == 0 {
		return Number{}
	}

	result := NewNumberLike(nums[0])
	for _, n := range nums {
		result.Val *= n.Val
		for i := 0; i < NDims; i++ {
			term := n.Deriv(i)
			for j := 0; j < NDims; j++ {
				if i != j {
					term *= n.Deriv(j)
				}
			}
			result.Deriv(i) += term
		}
	}
	return result
}

func Pow(a, b Number) Number {
	result := NewNumberLike(a)
	result.Val = math.Pow(a, b)
	for i := 0; i < NDims; i++ {
		results.Deriv(i) = result.Val * (b.Deriv(i)*math.Log(math.Abs(a.Value())) + a.Deriv(i)*b.Value()/a.Value())
	}
	return result
}
