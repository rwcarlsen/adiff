package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
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
		if fn.Val(x) == 0 {
			return 0
		}
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

type Tanh struct {
	Func
}

func (t *Tanh) SetInner(f Func)        { t.Func = f }
func (t Tanh) Val(x []float64) float64 { return math.Tanh(t.Func.Val(x)) }
func (t Tanh) Partial(v Variable) Func {
	return Mult{
		t.Func.Partial(v),
		Sum{
			Constant(1),
			Negative(&Pow{&Tanh{t.Func}, Constant(2)}),
		},
	}
}

type Passthrough struct{ Func }

func (p *Passthrough) SetInner(f Func) { p.Func = f }

type Network struct {
	nextVarIndex int
	Vars         []Variable
	Weights      []Variable
	CostFunc     Func
	State        []float64
	Outputs      []*Neuron
}

func (n *Network) Train(learnRate float64, varData [][]float64) {
	if len(n.State) == 0 {
		// initialize weights and vars input vector and set weights to 1
		n.State = make([]float64, n.NVars())
		for _, w := range n.Weights {
			n.State[int(w)] = 1
		}
	}

	// train network using residual (cost function)
	for _, pos := range varData {
		for i, index := range n.Vars {
			n.State[int(index)] = pos[i]
		}

		currNeurons := n.Outputs
		nextNeurons := []*Neuron{}
		for len(currNeurons) > 0 {
			nextNeurons = nextNeurons[:0]
			for _, neuron := range currNeurons {
				for _, w := range neuron.Weights {
					n.State[int(w)] += -learnRate * n.CostFunc.Partial(w).Val(n.State)
					//fmt.Printf("new weight%v=%v\n", w, n.State[int(w)])
				}
				for _, f := range neuron.Inputs {
					if neur, ok := f.(*Neuron); ok {
						nextNeurons = append(nextNeurons, neur)
					}
				}
			}
			currNeurons = nextNeurons
		}
	}
}

func (n *Network) NVars() int { return n.nextVarIndex }

func (n *Network) addVar() Variable {
	v := Variable(n.nextVarIndex)
	n.Vars = append(n.Vars, v)
	n.nextVarIndex++
	return v
}

func (n *Network) addWeight() Variable {
	v := Variable(n.nextVarIndex)
	n.Weights = append(n.Weights, v)
	n.nextVarIndex++
	return v
}

func (n *Network) NewNeuron() *Neuron {
	return &Neuron{network: n, Activation: &Tanh{}}
}

func (n *Network) NewInput() (*Neuron, Variable) {
	v := n.addVar()
	neuron := n.NewNeuron()
	neuron.Inputs = append(neuron.Inputs, v)
	neuron.Weights = append(neuron.Weights, n.addWeight())
	return neuron, v
}

func (n *Network) NewOutput(a ActivationFunc) *Neuron {
	neuron := &Neuron{network: n, Activation: a}
	n.Outputs = append(n.Outputs, neuron)
	return neuron
}

type ActivationFunc interface {
	Func
	SetInner(f Func)
}

type Neuron struct {
	network    *Network
	Inputs     []Func
	Weights    []Variable
	Activation ActivationFunc
}

func (n *Neuron) PullFrom(neurons ...*Neuron) *Neuron {
	for _, src := range neurons {
		n.Inputs = append(n.Inputs, src)
		n.Weights = append(n.Weights, n.network.addWeight())
	}
	return n
}

func (n *Neuron) Val(x []float64) float64 {
	var fn Sum
	for i := range n.Weights {
		fn = append(fn, Mult{n.Weights[i], n.Inputs[i]})
	}
	n.Activation.SetInner(fn)
	return n.Activation.Val(x)
}

func (n *Neuron) Partial(v Variable) Func {
	var fn Sum
	for i := range n.Weights {
		fn = append(fn, Mult{n.Weights[i], n.Inputs[i]})
	}
	n.Activation.SetInner(fn)
	return n.Activation.Partial(v)
}

func Laplace(f Func, vars ...Variable) Func {
	var sum Sum
	for _, v := range vars {
		sum = append(sum, f.Partial(v).Partial(v))
	}
	return sum
}

var plot = flag.String("plot", "", "'svg' to create svg plot with gnuplot")

func main() {
	flag.Parse()
	//prob1d()
	prob2d()
}

func prob2d() {
	var net Network
	in1, var1 := net.NewInput()
	in2, var2 := net.NewInput()
	out1 := net.NewOutput(&Passthrough{}).PullFrom(in1, in2)

	// a PDE would be defined like follows
	u, x, y := out1, var1, var2
	forcingFunc := Constant(10)
	diffusionCoeff := Constant(2)
	residual := Sum{Mult{Laplace(u, x, y), diffusionCoeff}, Negative(forcingFunc)}
	net.CostFunc = residual

	// we can then use the residual as a cost function to update the weights using a
	// backpropogation algorithm.

	// build training data (input variable combos) and train the network
	trainingPositions := [][]float64{}
	for xv := 0.0; xv < 5; xv += .1 {
		for yv := 0.0; yv < 5; yv += .1 {
			trainingPositions = append(trainingPositions, []float64{xv, yv})
		}
	}

	learnRate := .98
	net.Train(learnRate, trainingPositions)

	// look at the results
	var buf bytes.Buffer
	for xv := 0.0; xv < 5; xv += .1 {
		for yv := 0.0; yv < 5; yv += .1 {
			net.State[int(x)] = xv
			net.State[int(y)] = yv
			fmt.Fprintf(&buf, "%v %v %v\n", xv, yv, u.Val(net.State))
		}
	}

	fmt.Println("Solution (x y u):")
	fmt.Print(buf.String())

	if *plot != "" {
		cmd := exec.Command("gnuplot", "-e", `set terminal svg; set output "`+*plot+`"; plot "-" u 1:2:3 w image`)
		cmd.Stdin = &buf
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func prob1d() {
	var net Network
	in1, var1 := net.NewInput()
	// This is a dummy input and variable to enable the network to output nonzero values when all
	// inputs are zero.
	dummyin, dummyvar := net.NewInput()

	out1 := net.NewOutput(&Passthrough{}).PullFrom(in1, dummyin)

	// a PDE would be defined like follows
	u, x := out1, var1
	// we want to approximate u(x) = 3 so error=(u-3)^2
	residual := &Pow{Sum{u, Constant(-3)}, Constant(2)}
	net.CostFunc = residual

	// build training data (input variable combos) and train the network
	trainingPositions := [][]float64{}
	for xv := 0.0; xv < 5; xv += .1 {
		// dummy input value corresponding to our dummy variable
		dummy := 1.0
		trainingPositions = append(trainingPositions, []float64{xv, dummy})
	}

	learnRate := .98
	net.Train(learnRate, trainingPositions)

	// look at the results
	var buf bytes.Buffer
	for xv := 0.0; xv < 5; xv += .1 {
		net.State[int(x)] = xv
		net.State[int(dummyvar)] = 1.0
		fmt.Fprintf(&buf, "%v\t%v\n", xv, u.Val(net.State))
	}

	fmt.Println("Solution (x u):")
	fmt.Print(buf.String())
}
