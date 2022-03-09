package main

import (
	"fmt"
	"math/big"
	"os"
	"strconv"

	bls12381 "github.com/coinbase/kryptology/pkg/core/curves/native/bls12-381"
)

func Abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
func show_values(a int64, b int64, x int64) {
	fmt.Printf("a=%d\n", a)
	fmt.Printf("b=%d\n", b)
	fmt.Printf("Solution: x=%d\n\n", x)
	if b < 0 && a < 0 {
		fmt.Printf("\nEqn: x^2 - %d x - %d\n", Abs(a), Abs(b))
	} else if b < 0 {
		fmt.Printf("\nEqn: x^2 + %d x + %d\n", a, Abs(b))
	} else if a < 0 {
		fmt.Printf("\nEqn: x^2 + %d x + %d\n", Abs(a), b)
	} else {
		fmt.Printf("\nEqn: x^2 + %d x + %d\n", a, b)
	}
}
func main() {

	xval := int64(7)
	aval := int64(-1)
	bval := int64(-42)

	argCount := len(os.Args[1:])

	if argCount > 0 {
		xval, _ = strconv.ParseInt(os.Args[1], 10, 64)
	}
	if argCount > 1 {
		aval, _ = strconv.ParseInt(os.Args[2], 10, 64)
	}
	if argCount > 2 {
		bval, _ = strconv.ParseInt(os.Args[3], 10, 64)
	}

	x := big.NewInt(Abs(xval))
	a := big.NewInt(Abs(aval))
	b := big.NewInt(Abs(bval))

	show_values(aval, bval, xval)

	bls := bls12381.NewEngine()
	g1, g2 := bls.G1, bls.G2

	G1, G2 := g1.One(), g2.One()

	xG1 := g1.New()
	xG2 := g2.New()

	Ga := g2.New()
	Gb := g2.New()

	g1.MulScalar(xG1, G1, x)
	g2.MulScalar(xG2, G2, x)

	if xval < 0 {
		g1.Neg(xG1, xG1)
		g2.Neg(xG2, xG2)
	}

	g2.MulScalar(Ga, G2, a)
	g2.MulScalar(Gb, G2, b)

	if aval < 0 {
		g2.Neg(Ga, Ga)
	}
	if bval < 0 {
		g2.Neg(Gb, Gb)
	}

	bls.AddPair(xG1, xG2)
	bls.AddPair(xG1, Ga)
	bls.AddPair(G1, Gb)

	rtn, _ := bls.Check()
	if rtn {
		fmt.Printf("\nYou have proven it!")
	} else {
		fmt.Printf("\nYou have NOT proven it!")
	}

}
