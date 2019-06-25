package adiff

import (
	"testing"
)

func BenchmarkNumber(b *testing.B) {
	tmp := NewSimple(50, 0)
	for i := 0; i < b.N; i++ {
		_ = Add(tmp, Sin(tmp, Mul(tmp, Const(3), Pow(tmp, Variable{0, 5}, Const(2)))), Const(7))
	}
}

//func TestNumber(t *testing.T) {
//	NDims = 1
//
//	tmp1 := NewNumber(0)
//	tmp2 := NewNumber(0)
//
//	var tests = []struct {
//		Num       Number
//		Want      float64
//		WantDeriv []float64
//	}{
//		{ // 3*x^2+7 where x=5
//			Num:       Add(tmp1, Mul(tmp1, Const(tmp2, 3), Pow(tmp1, NewVariable(0, 5), Const(tmp2, 2))), Const(tmp2, 7)),
//			Want:      82,
//			WantDeriv: []float64{30},
//		},
//		{ // sin(3*x^2)+7 where x=5
//			Num:       Add(tmp1, Sin(tmp1, Mul(tmp1, Const(tmp2, 3), Pow(tmp1, NewVariable(0, 5), Const(tmp2, 2)))), Const(tmp2, 7)),
//			Want:      math.Sin(3*math.Pow(5, 2)) + 7,
//			WantDeriv: []float64{math.Cos(3*math.Pow(5, 2)) * 6 * 5},
//		},
//		{ // x^2
//			Num:       Pow(tmp1, NewVariable(0, 5), Const(tmp2, 2)),
//			Want:      25,
//			WantDeriv: []float64{10},
//		},
//		{ // 3*x
//			Num:       Mul(tmp1, Const(tmp2, 3), NewVariable(0, 5)),
//			Want:      15,
//			WantDeriv: []float64{3},
//		},
//	}
//
//	const tol = 1e-8
//	for i, test := range tests {
//		t.Logf("case %v:", i+1)
//		if math.Abs(test.Num.Val-test.Want) > tol {
//			t.Errorf("    Value: want %v, got %v", test.Want, test.Num.Val)
//		} else {
//			t.Logf("    Value: PASS")
//		}
//		for j := 0; j < NDims; j++ {
//			if math.Abs(test.Num.Derivs[j]-test.WantDeriv[j]) > tol {
//				t.Errorf("    Deriv[%v]: want %v, got %v", j, test.WantDeriv[j], test.Num.Derivs[j])
//			} else {
//				t.Logf("    Deriv[%v]: PASS", j)
//			}
//		}
//	}
//}
