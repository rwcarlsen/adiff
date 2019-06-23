package adiff

import (
	"math"
)

type Number struct {
	Val    float64
	Derivs []float64
}

var NDims int

func NewNumber(val float64) Number { return Number{Val: val, Derivs: make([]float64, NDims)} }
func NewVariable(index int, val float64) Number {
	n := Number{Val: val, Derivs: make([]float64, NDims)}
	n.Derivs[index] = 1
	return n
}

func Const(dst Number, val float64) Number {
	dst.Val = val
	for i := 0; i < NDims; i++ {
		dst.Derivs[i] = 0
	}
	return dst
}

func Log(dst, a Number) Number {
	result := NewNumber(math.Log(a.Val))
	for i := 0; i < NDims; i++ {
		result.Derivs[i] = a.Derivs[i] / a.Val
	}
	return result
}

func Add(dst Number, nums ...Number) Number {
	if len(nums) == 0 {
		return NewNumber(0)
	}

	result := NewNumber(0)
	for _, n := range nums {
		result.Val += n.Val
		for i := 0; i < NDims; i++ {
			result.Derivs[i] += n.Derivs[i]
		}
	}
	return result
}

func Mul(dst Number, nums ...Number) Number {
	result := NewNumber(nums[0].Val)
	for i := 0; i < NDims; i++ {
		result.Derivs[i] = nums[0].Derivs[i]
	}

	for _, n := range nums[1:] {
		for i := 0; i < NDims; i++ {
			result.Derivs[i] = result.Derivs[i]*n.Val + result.Val*n.Derivs[i]
		}
		result.Val *= n.Val
	}
	return result
}

func Abs(dst, n Number) Number {
	if n.Val < 0 {
		dst.Val = -n.Val
		for i := 0; i < NDims; i++ {
			dst.Derivs[i] = -n.Derivs[i]
		}
	} else {
		dst.Val = n.Val
		for i := 0; i < NDims; i++ {
			dst.Derivs[i] = n.Derivs[i]
		}
	}
	return dst
}

func Sin(dst, a Number) Number {
	for i := 0; i < NDims; i++ {
		dst.Derivs[i] = math.Cos(a.Val) * a.Derivs[i]
	}
	dst.Val = math.Sin(a.Val)
	return dst
}

func Cos(dst, a Number) Number {
	for i := 0; i < NDims; i++ {
		dst.Derivs[i] = -math.Sin(a.Val) * a.Derivs[i]
	}
	dst.Val = math.Cos(a.Val)
	return dst
}

func Pow(dst, a, b Number) Number {
	result := math.Pow(a.Val, b.Val)
	for i := 0; i < NDims; i++ {
		dst.Derivs[i] = result * (b.Derivs[i]*math.Log(math.Abs(a.Val)) + a.Derivs[i]*b.Val/a.Val)
	}
	dst.Val = result
	return dst
}
