package adiff

import (
	"math"
)

type Number interface {
	Deriv(i int) float64
	Value() float64
}

func IsConst(ndims int, n Number) bool {
	for i := 0; i < ndims; i++ {
		if n.Deriv(i) != 0 {
			return false
		}
	}
	return true
}

type Simple struct {
	Val    float64
	Derivs []float64
}

func NewSimple(ndims int, val float64) Simple {
	return Simple{Val: val, Derivs: make([]float64, ndims)}
}

func (s Simple) Value() float64      { return s.Val }
func (s Simple) Deriv(i int) float64 { return s.Derivs[i] }

type Const float64

func (c Const) Value() float64      { return float64(c) }
func (c Const) Deriv(i int) float64 { return 0 }

type Variable struct {
	Index int
	Val   float64
}

func (v Variable) Value() float64 { return float64(v.Val) }
func (v Variable) Deriv(i int) float64 {
	if i == v.Index {
		return 1
	}
	return 0
}

func Log(dst Simple, a Number) Number {
	for i := 0; i < len(dst.Derivs); i++ {
		dst.Derivs[i] = a.Deriv(i) / a.Value()
	}
	dst.Val = math.Log(a.Value())
	return dst
}

func Add(dst Simple, a, b Number) Number {
	for i := 0; i < len(dst.Derivs); i++ {
		dst.Derivs[i] = a.Deriv(i) + b.Deriv(i)
	}
	dst.Val = a.Value() + b.Value()
	return dst
}

func Mul(dst Simple, a, b Number) Number {
	for i := 0; i < len(dst.Derivs); i++ {
		dst.Derivs[i] = a.Deriv(i)*b.Value() + a.Value()*b.Deriv(i)
	}
	dst.Val = a.Value() * b.Value()
	return dst
}

func Abs(dst Simple, n Number) Number {
	if n.Value() < 0 {
		dst.Val = -n.Value()
		for i := 0; i < len(dst.Derivs); i++ {
			dst.Derivs[i] = -n.Deriv(i)
		}
	} else {
		dst.Val = n.Value()
		for i := 0; i < len(dst.Derivs); i++ {
			dst.Derivs[i] = n.Deriv(i)
		}
	}
	return dst
}

func Sin(dst Simple, a Number) Number {
	for i := 0; i < len(dst.Derivs); i++ {
		dst.Derivs[i] = math.Cos(a.Value()) * a.Deriv(i)
	}
	dst.Val = math.Sin(a.Value())
	return dst
}

func Cos(dst Simple, a Number) Number {
	for i := 0; i < len(dst.Derivs); i++ {
		dst.Derivs[i] = -math.Sin(a.Value()) * a.Deriv(i)
	}
	dst.Val = math.Cos(a.Value())
	return dst
}

func Pow(dst Simple, a, b Number) Number {
	result := math.Pow(a.Value(), b.Value())
	for i := 0; i < len(dst.Derivs); i++ {
		dst.Derivs[i] = result * (b.Deriv(i)*math.Log(math.Abs(a.Value())) + a.Deriv(i)*b.Value()/a.Value())
	}
	dst.Val = result
	return dst
}
