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

func (v Variable) Simplify() Func { return v }
func (v Variable) String() string { return fmt.Sprintf("v%v", int(v)) }

type Func interface {
	Val(x []float64) float64
	Partial(v Variable) Func
	Simplify() Func
	String() string
}

type Constant float64

func (c Constant) Val(x []float64) float64 { return float64(c) }
func (c Constant) Partial(v Variable) Func { return Constant(0) }
func (c Constant) String() string          { return fmt.Sprintf("%v", float64(c)) }
func (c Constant) Simplify() Func          { return c }

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

func (s Sum) Simplify() Func {
	nonzero := Sum{}
	// skip all zero terms
	for _, f := range s {
		simple := f.Simplify()
		if c, ok := simple.(Constant); ok && float64(c) == 0 {
			continue
		}
		nonzero = append(nonzero, simple)
	}

	simpler := Sum{}
	constTot := 0.0
	// merge all constant terms into a single term
	for _, f := range nonzero {
		simple := f.Simplify()
		if c, ok := simple.(Constant); ok {
			constTot += float64(c)
			continue
		}
		simpler = append(simpler, simple)
	}
	if constTot != 0 {
		simpler = append(simpler, Constant(constTot))
	}

	if len(simpler) == 0 {
		return Constant(0)
	} else if len(simpler) == 1 {
		return simpler[0]
	}
	return simpler
}

func (s Sum) String() string {
	simple := s.Simplify()

	if sum, ok := simple.(Sum); ok {
		if len(sum) == 1 {
			return sum[0].String()
		}

		str := "(" + sum[0].String()
		for _, f := range sum[1:] {
			str += " + " + f.String()
		}
		return str + ")"
	}
	return simple.String()
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

func (m Mult) Simplify() Func {
	simplified := Mult{}
	for _, f := range m {
		simple := f.Simplify()
		if c, ok := simple.(Constant); ok && float64(c) == 0 {
			// if any term is zero, the entire mult expression is zero
			return Constant(0)
		} else if c, ok := simple.(Constant); ok && float64(c) == 1 {
			// if any term is "1", skip it - it doesn't affect the mult expression's value
			continue
		} else if inner, ok := simple.(Mult); ok {
			// if any term is itself a mult expression, just merge it with this one
			for _, in := range inner {
				simplified = append(simplified, in.Simplify())
			}
		} else {
			// otherwise, just add/keep the term
			simplified = append(simplified, simple)
		}
	}

	dups := map[Variable]*Pow{}
	simpler := Mult{}
	constTot := 1.0
	// if any term is a Pow expression with a variable base, track it so we can merge e.g.
	// v^f*v^g style expressions to be v^(f+g). This applies also to straight variable terms
	// which are equivalent to Pow{v, 1}.
	for _, f := range simplified {
		p, pok := f.(*Pow)
		sv, svok := f.(Variable)
		if pok {
			if v, ok := p.Base.(Variable); ok {
				if _, ok := dups[v]; ok {
					dups[v] = &Pow{dups[v].Base, Sum{dups[v].Exponent, p.Exponent}.Simplify()}
				} else {
					dups[v] = p
				}
				continue
			}
		} else if svok {
			if _, ok := dups[sv]; ok {
				dups[sv] = &Pow{dups[sv].Base, Sum{dups[sv].Exponent, Constant(1)}.Simplify()}
			} else {
				dups[sv] = &Pow{sv, Constant(1)}
			}
			continue
		}
		// also merge all constant's together into a single constant term.
		if c, ok := f.(Constant); ok {
			constTot *= float64(c)
			continue
		}
		simpler = append(simpler, f)
	}

	// add back the merged constant and pow expressions
	simpler = append(simpler, Constant(constTot))
	for _, f := range dups {
		simpler = append(simpler, f.Simplify())
	}

	if len(simpler) == 0 {
		return Constant(1)
	} else if len(simpler) == 1 {
		return simpler[0]
	}
	return simpler
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

func (m Mult) String() string {
	simple := m.Simplify()

	if mult, ok := simple.(Mult); ok {
		if len(mult) == 1 {
			return mult[0].String()
		}

		str := "(" + mult[0].String()
		for _, f := range mult[1:] {
			str += " * " + f.String()
		}
		return str + ")"
	}
	return simple.String()
}

type Ln struct {
	Func
}

func (ln Ln) Val(x []float64) float64 { return math.Log(ln.Func.Val(x)) }
func (ln Ln) Partial(v Variable) Func { return Mult{ln.Func.Partial(v), Inverse(ln.Func)} }
func (ln Ln) String() string          { return fmt.Sprintf("ln(%v)", ln.Func) }
func (ln Ln) Simplify() Func          { return Ln{ln.Func.Simplify()} }

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
			Mult{p.Exponent.Partial(v), Ln{Abs(p.Base)}},
			Mult{p.Base.Partial(v), Inverse(p.Base), p.Exponent},
		},
	}
}

func (p *Pow) String() string { return fmt.Sprintf("(%v^%v)", p.Base, p.Exponent) }
func (p *Pow) Simplify() Func {
	base := p.Base.Simplify()
	exp := p.Exponent.Simplify()
	if expc, ok := exp.(Constant); ok && float64(expc) == 0 {
		return Constant(1)
	} else if expc, ok := exp.(Constant); ok && float64(expc) == 1 {
		return base
	}
	return &Pow{base, exp}
}

type Branch func(x []float64) Func

func (b Branch) Val(x []float64) float64 { return b(x).Val(x) }

func (b Branch) Partial(v Variable) Func {
	return Branch(func(x []float64) Func { return b(x).Partial(v) })
}

func (b Branch) Simplify() Func { return b }
func (b Branch) String() string { return "Branch(???)" }

func Negative(f Func) Func { return Mult{Constant(-1), f} }
func Inverse(f Func) Func  { return &Pow{f, Constant(-1)} }
func Abs(f Func) Func {
	return Branch(func(x []float64) Func {
		if f.Val(x) >= 0 {
			return f
		}
		return Negative(f)
	})
}

type Tanh struct {
	Func
}

func (t *Tanh) Simplify() Func          { return &Tanh{t.Func.Simplify()} }
func (t *Tanh) String() string          { return fmt.Sprintf("tanh(%v)", t.Func) }
func (t *Tanh) SetInner(f Func)         { t.Func = f }
func (t *Tanh) Val(x []float64) float64 { return math.Tanh(t.Func.Val(x)) }
func (t *Tanh) Partial(v Variable) Func {
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
func (p *Passthrough) Simplify() Func  { return &Passthrough{p.Func.Simplify()} }

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
	derivs := map[Variable]Func{}

	// train network using residual (cost function) evaluated at each training data point.
	for _, pos := range varData {
		for i, index := range n.Vars {
			n.State[int(index)] = pos[i]
		}

		fmt.Printf("weights: %.3f", n.State[int(n.Weights[0])])
		for _, w := range n.Weights[1:] {
			fmt.Printf(", %.3f", n.State[int(w)])
		}
		fmt.Println()

		// calculate a delta weight for each weight in the network
		dweight := make([]float64, len(n.Weights))
		for i, w := range n.Weights {
			if _, ok := derivs[w]; !ok {
				derivs[w] = n.CostFunc.Partial(w).Simplify()
			}
			partialcost := derivs[w]
			dweight[i] = -learnRate * partialcost.Val(n.State)
		}
		// update all weights together
		for i, w := range n.Weights {
			n.State[int(w)] += dweight[i]
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

func (n *Network) NewOutputFunc(a ActivationFunc) *Neuron {
	neuron := &Neuron{network: n, Activation: a}
	n.Outputs = append(n.Outputs, neuron)
	return neuron
}

func (n *Network) NewOutput() *Neuron {
	neuron := &Neuron{network: n, Activation: &Passthrough{}}
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

func (n *Neuron) getFunc() Func {
	var fn Sum
	for i := range n.Weights {
		fn = append(fn, Mult{n.Weights[i], n.Inputs[i]})
	}
	n.Activation.SetInner(fn)
	return n.Activation
}

func (n *Neuron) Val(x []float64) float64 {
	var fn Sum
	for i := range n.Weights {
		fn = append(fn, Mult{n.Weights[i], n.Inputs[i]})
	}
	n.Activation.SetInner(fn)
	return n.getFunc().Val(x)
}

func (n *Neuron) Partial(v Variable) Func {
	var fn Sum
	for i := range n.Weights {
		fn = append(fn, Mult{n.Weights[i], n.Inputs[i]})
	}
	n.Activation.SetInner(fn)
	return n.getFunc().Partial(v)
}

func (n *Neuron) String() string { return n.getFunc().String() }
func (n *Neuron) Simplify() Func { return n.getFunc().Simplify() }

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
	prob1dDiscont()
	//prob1d()
	//prob2d()
}

func prob2d() {
	var net Network
	in1, var1 := net.NewInput()
	in2, var2 := net.NewInput()
	out1 := net.NewOutput().PullFrom(in1, in2)

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

	fmt.Println("Approximation Eqn: ", out1)
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

	out1 := net.NewOutput().PullFrom(in1, dummyin)

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

	fmt.Println("Approximation Eqn: ", out1)
	fmt.Println("Solution (x u):")
	fmt.Print(buf.String())
}

func prob1dDiscont() {
	var net Network
	in1, var1 := net.NewInput()
	// This is a dummy input and variable to enable the network to output nonzero values when all
	// inputs are zero.
	dummyin, dummyvar := net.NewInput()

	// hidden layer
	//n1 := net.NewNeuron().PullFrom(in1, dummyin)
	//n2 := net.NewNeuron().PullFrom(in1, dummyin)
	//n3 := net.NewNeuron().PullFrom(in1, dummyin)

	//out1 := net.NewOutput().PullFrom(n1, n2, n3)
	out1 := net.NewOutput().PullFrom(in1, dummyin)
	fmt.Println("networkFunc: ", out1)

	// convenient vars/names for building our PDE and BCs
	u, x := out1, var1

	// define boundary conditions
	penalty := Constant(1.0)
	bcs := Branch(func(xv []float64) Func {
		if xv[int(x)] == 0 {
			return Sum{Constant(1), Negative(u)}
		} else if xv[int(x)] == 1 {
			return Sum{Constant(7), Negative(u)}
		}
		return Constant(0)
	})

	k := Constant(1)
	heatSource := Constant(0)
	// define our PDE: -k*laplace(u) = S
	residual := Sum{Mult{k, Laplace(u, x)}, heatSource}

	net.CostFunc = Sum{&Pow{residual, Constant(2)}, &Pow{Mult{penalty, bcs}, Constant(2)}}.Simplify()
	fmt.Println("costfunc: ", net.CostFunc)

	// build training data (input variable combos) and train the network
	trainingPositions := [][]float64{}
	for xv := 0.01; xv < 1; xv += .01 {
		dummy := 1.0 // dummy input value corresponding to our dummy variable
		trainingPositions = append(trainingPositions, []float64{xv, dummy})
	}
	// manually add boundary positions
	trainingPositions = append(trainingPositions, []float64{0, 1})
	trainingPositions = append(trainingPositions, []float64{1, 1})

	learnRate := .9
	net.Train(learnRate, trainingPositions)

	// look at the results
	var buf bytes.Buffer
	for xv := 0.0; xv <= 1.1; xv += .1 {
		net.State[int(x)] = xv
		net.State[int(dummyvar)] = 1.0
		fmt.Fprintf(&buf, "%v\t%v\n", xv, u.Val(net.State))
	}

	fmt.Println("Approximation Eqn: ", out1)
	fmt.Println("Solution (x u):")
	fmt.Print(buf.String())
}
