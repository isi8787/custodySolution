package main

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"

	crypto "github.com/coinbase/kryptology/pkg/core"
)

func main() {

	p, _ := rand.Prime(rand.Reader, 128)
	q, _ := rand.Prime(rand.Reader, 128)
	N := new(big.Int).Mul(p, q)
	PHI := new(big.Int).Mul(new(big.Int).Sub(p, big.NewInt(1)), new(big.Int).Sub(q, big.NewInt(1)))

	PsfProofLength := 5

	// 1. M = N^{-1} mod \phi(N)
	M, _ := crypto.Inv(N, PHI)

	// 2. Generate challenges
	x := make([]*big.Int, PsfProofLength)
	for i := 0; i < PsfProofLength; i++ {
		x[i], _ = rand.Int(rand.Reader, new(big.Int).SetUint64(uint64(math.Pow(2, 64))))
	}

	proof := make([]*big.Int, PsfProofLength)
	// 3. Create proofs y_i = x_i^M mod N
	for i, xj := range x {

		yi, _ := crypto.Exp(xj, M, N)
		proof[i] = yi
	}

	fmt.Printf("p=%s\n", p)
	fmt.Printf("q=%s\n", q)
	fmt.Printf("N=%s\n", N)
	fmt.Printf("PHI=%s\n", PHI)
	fmt.Printf("\nChallenges:\t%v\n", x)
	fmt.Printf("Proof:\t%v\n", proof)

	proven := true
	for j, xj := range x {
		// 4. Proof: yj^N == x mod N return false

		lhs, _ := crypto.Exp(proof[j], N, N)

		if lhs.Cmp(xj) != 0 {
			fmt.Printf("Failed at %d\n", j)
			proven = false
		}
	}
	if proven == true {
		fmt.Printf("\nZero knowledge proof of safe Paillier\n")
	} else {
		fmt.Printf("\nNo Zero knowledge proof of safe Paillier\n")
	}

	fmt.Printf("\n\n== Now trying with an incorrect value of N=p^2 ==\n")

	N = new(big.Int).Mul(p, p)
	PHI = new(big.Int).Mul(new(big.Int).Sub(p, big.NewInt(1)), new(big.Int).Sub(p, big.NewInt(1)))

	proof = make([]*big.Int, PsfProofLength)
	// 3. Create proofs y_i = x_i^M mod N
	for i, xj := range x {

		yi, _ := crypto.Exp(xj, M, N)
		proof[i] = yi
	}

	fmt.Printf("p=%s\n", p)
	fmt.Printf("N=%s\n", N)
	fmt.Printf("PHI=%s\n", PHI)
	fmt.Printf("\nChallenges:\t%v\n", x)
	fmt.Printf("Proof:\t%v\n", proof)

	proven = true
	for j, xj := range x {
		// 4. Proof: yj^N == x mod N return false

		lhs, _ := crypto.Exp(proof[j], N, N)

		if lhs.Cmp(xj) != 0 {
			fmt.Printf("[Failed at %d]", j)
			proven = false
		}
	}
	if proven == true {
		fmt.Printf("\nZero knowledge proof of safe Paillier\n")
	} else {
		fmt.Printf("\nNo Zero knowledge proof of safe Paillier\n")
	}

}


