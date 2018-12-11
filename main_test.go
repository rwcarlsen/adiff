package main

import (
	"fmt"
	"math"
	"testing"
)

type Problem struct {
	Eqn             Func
	Nvars           int
	WantFunc        func(x []float64) float64
	WantDerivFunc   func(v Variable, x []float64) float64
	CheckDerivs     [][]Variable // each entry is a list of independent vars to take partial deriv for
	CheckDerivsWant []func(x []float64) float64
	Xmin, Xmax      float64
	Tol             float64
}

var x Variable = 0
var y Variable = 1

var problems []*Problem = []*Problem{
	&Problem{
		Nvars:       1,
		Eqn:         x,
		WantFunc:    func(x []float64) float64 { return x[0] },
		CheckDerivs: [][]Variable{{x}},
		CheckDerivsWant: []func(x []float64) float64{
			func(x []float64) float64 { return 1 },
		},
		Xmin: 0, Xmax: 1,
		Tol: 1e-10,
	},
	&Problem{
		Nvars:       1,
		Eqn:         &Pow{x, Constant(2)},
		WantFunc:    func(x []float64) float64 { return x[0] * x[0] },
		CheckDerivs: [][]Variable{{x}},
		CheckDerivsWant: []func(x []float64) float64{
			func(x []float64) float64 { return 2 * x[0] },
		},
		Xmin: 0, Xmax: 1,
		Tol: 1e-10,
	},
	&Problem{
		Nvars:       2,
		Eqn:         &Pow{x, Constant(2)},
		WantFunc:    func(x []float64) float64 { return x[0] * x[0] },
		CheckDerivs: [][]Variable{{x}, {y}},
		CheckDerivsWant: []func(x []float64) float64{
			func(x []float64) float64 { return 2 * x[0] },
			func(x []float64) float64 { return 0 },
		},
		Xmin: 0, Xmax: 1,
		Tol: 1e-10,
	},
	&Problem{
		Nvars:       2,
		Eqn:         Mult{x, y},
		WantFunc:    func(x []float64) float64 { return x[0] * x[1] },
		CheckDerivs: [][]Variable{{x}, {y}},
		CheckDerivsWant: []func(x []float64) float64{
			func(x []float64) float64 { return x[1] },
			func(x []float64) float64 { return x[0] },
		},
		Xmin: 0, Xmax: 1,
		Tol: 1e-10,
	},
	&Problem{
		Nvars:       2,
		Eqn:         Mult{&Pow{x, Constant(2)}, y},
		WantFunc:    func(x []float64) float64 { return x[0] * x[0] * x[1] },
		CheckDerivs: [][]Variable{{x}, {y}},
		CheckDerivsWant: []func(x []float64) float64{
			func(x []float64) float64 { return 2 * x[0] * x[1] },
			func(x []float64) float64 { return x[0] * x[0] },
		},
		Xmin: 0, Xmax: 1,
		Tol: 1e-10,
	},
	&Problem{
		Nvars:       2,
		Eqn:         Sum{Mult{&Pow{x, Constant(2)}, y}, &Pow{y, Constant(2)}, Constant(7)},
		WantFunc:    func(x []float64) float64 { return x[0]*x[0]*x[1] + x[1]*x[1] + 7 },
		CheckDerivs: [][]Variable{{x}, {y}, {x, x}, {x, y}, {y, x}},
		CheckDerivsWant: []func(x []float64) float64{
			func(x []float64) float64 { return 2 * x[0] * x[1] },
			func(x []float64) float64 { return x[0]*x[0] + 2*x[1] },
			func(x []float64) float64 { return 2 * x[1] },
			func(x []float64) float64 { return 2 * x[0] },
			func(x []float64) float64 { return 2 * x[0] },
		},
		Xmin: 0, Xmax: 1,
		Tol: 1e-10,
	},
}

func TestProblems(t *testing.T) {
	for i, prob := range problems {
		t.Run(fmt.Sprintf("Problem %v", i+1), testProb(prob))
	}
}

func testProb(p *Problem) func(t *testing.T) {
	return func(t *testing.T) {
		ndivs := 10
		divs := make([]int, p.Nvars)
		for i := range divs {
			divs[i] = ndivs
		}

		perms := Permute(0, divs...)

		x := make([]float64, p.Nvars)
		for _, perm := range perms {
			for i := range perm {
				x[i] = float64(perm[i])/float64(ndivs)*(p.Xmax-p.Xmin) + p.Xmin
			}

			got := p.Eqn.Val(x)
			if math.Abs(got-p.WantFunc(x)) > p.Tol {
				t.Errorf("FAIL f%v: want %v, got %v", x, p.WantFunc(x), got)
			} else {
				t.Logf("     f%v = %v", x, got)
			}

			for i, deriv := range p.CheckDerivs {
				fn := p.Eqn
				dname := ""
				for _, jvar := range deriv {
					dname += fmt.Sprintf("dv%v", jvar)
					fn = fn.Partial(jvar)
				}

				want := p.CheckDerivsWant[i](x)
				got := fn.Val(x)
				if math.Abs(got-want) > p.Tol {
					t.Errorf("FAIL     df/%v: want %v, got %v", dname, want, got)
				} else {
					t.Logf("         df/%v = %v", dname, got)
				}
			}
		}
	}
}

func Permute(maxsum int, dimensions ...int) [][]int {
	return permute(maxsum, dimensions, make([]int, 0, len(dimensions)))
}

func sum(vals ...int) int {
	tot := 0
	for _, val := range vals {
		tot += val
	}
	return tot
}

func permute(maxsum int, dimensions []int, prefix []int) [][]int {
	set := make([][]int, 0)

	if maxsum > 0 {
		if tot := sum(prefix...); tot == maxsum {
			return [][]int{append(append([]int{}, prefix...), make([]int, len(dimensions))...)}
		} else if tot > maxsum {
			return set
		}
	}

	if len(dimensions) == 1 {
		for i := 0; i < dimensions[0]; i++ {
			val := append(append([]int{}, prefix...), i)
			if maxsum == 0 || sum(val...) <= maxsum {
				set = append(set, val)
			}
		}
		return set
	}
	max := dimensions[0]
	for i := 0; i < max; i++ {
		newprefix := append(prefix, i)
		moresets := permute(maxsum, dimensions[1:], newprefix)
		set = append(set, moresets...)
	}
	return set
}
