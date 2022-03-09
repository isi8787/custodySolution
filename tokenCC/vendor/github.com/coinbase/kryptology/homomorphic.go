package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/coinbase/kryptology/pkg/verenc/elgamal"
)

func main() {

	val1 := 5
	val2 := 4
	argCount := len(os.Args[1:])

	if argCount > 0 {
		val1, _ = strconv.Atoi(os.Args[1])

	}
	if argCount > 1 {
		val2, _ = strconv.Atoi(os.Args[2])
	}

	curve := curves.ED25519()
	pk, sk, _ := elgamal.NewKeys(curve)

	x1 := curve.Scalar.New(val1)

	x2 := curve.Scalar.New(val2)

	res1, _ := pk.HomomorphicEncrypt(x1)

	res2, _ := pk.HomomorphicEncrypt(x2)

	res1 = res1.Add(res2)


	dec, _ := res1.Decrypt(sk)

	fmt.Printf("Val1: %d\n", val1)
	fmt.Printf("Val2: %d\n", val2)

	fmt.Printf("Result: %x\n", dec.ToAffineCompressed())

	for i := 0; i < 100000; i++ {
		val := curve.Scalar.New(i)
		Val := curve.Point.Generator().Mul(val)

		rtn := Val.Equal(dec)
		if rtn == true {
			fmt.Printf("Result is %d \n", i)
			break
		}

	}

}

