package main

import (
	"fmt"
	"math"
)

type Variable int

func (v Variable) Val(x []float64) float64 { return x[int(v)] }
func (vv Variable) Partial(v Variable) Func {
	if vv != v {
		return Constant(0)
	}
	return Constant(1)
}

type Func interface {
	Val(x []float64) float64
	Partial(v Variable) Func
}

type Constant float64

func (c Constant) Val(x []float64) float64 { return float64(c) }
func (c Constant) Partial(v Variable) Func { return Constant(0) }

type Sum []Func

func (s Sum) Val(x []float64) float64 {
	tot := 0.0
	for _, fn := range s {
		tot += fn.Val(x)
	}
	return tot
}

func (s Sum) Partial(v Variable) Func {
	var deriv Sum
	for _, fn := range s {
		deriv = append(deriv, fn.Partial(v))
	}
	return deriv
}

type Mult []Func

func (m Mult) Val(x []float64) float64 {
	tot := 1.0
	for _, fn := range m {
		tot *= fn.Val(x)
	}
	return tot
}

func (m Mult) Partial(v Variable) Func {
	if len(m) == 0 {
		return Constant(0)
	}
	deriv := Sum{
		Mult{m[0].Partial(v), Mult(m[1:])},
		Mult{m[0], Mult(m[1:]).Partial(v)},
	}
	return deriv
}

type Branch struct {
	Cond   func(x []float64) bool
	IfTrue Func
	Else   Func
}

func (b *Branch) Val(x []float64) float64 {
	if b.Cond(x) {
		return b.IfTrue.Val(x)
	}
	return b.Else.Val(x)
}

func (b *Branch) Partial(v Variable) Func {
	return &Branch{
		Cond:   func(x []float64) bool { return b.Val(x) >= 0 },
		IfTrue: b.Partial(v),
		Else:   Negative(b).Partial(v),
	}
}

type Ln struct {
	Func
}

func (ln Ln) Val(x []float64) float64 { return math.Log(ln.Func.Val(x)) }
func (ln Ln) Partial(v Variable) Func { return Mult{ln.Func.Partial(v), Inverse(ln.Func)} }

type Abs struct {
	Func
}

func (a Abs) Val(x []float64) float64 { return math.Abs(a.Func.Val(x)) }
func (a Abs) Partial(v Variable) Func {
	return &Branch{
		Cond:   func(x []float64) bool { return a.Func.Val(x) >= 0 },
		IfTrue: a.Func.Partial(v),
		Else:   Negative(a.Func).Partial(v),
	}
}

type Pow struct {
	Base     Func
	Exponent Func
}

func (p *Pow) Val(x []float64) float64 {
	return math.Pow(p.Base.Val(x), p.Exponent.Val(x))
}

func (p *Pow) Partial(v Variable) Func {
	return Mult{
		p,
		Sum{
			Mult{p.Exponent.Partial(v), Ln{Abs{p.Base}}},
			Mult{p.Base.Partial(v), Inverse(p.Base), p.Exponent},
		},
	}
}

func Negative(f Func) Func { return Mult{Constant(-1), f} }
func Inverse(f Func) Func  { return &Pow{f, Constant(-1)} }

func main() {
	fmt.Println("what to do?")
}
