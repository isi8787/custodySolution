package main

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/btcsuite/btcd/btcec"
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/dustinxie/ecc"

	"github.com/coinbase/kryptology/pkg/tecdsa/dkls"
)

func main() {

	msg := "Hello 123"

	argCount := len(os.Args[1:])

	if argCount > 0 {
		msg = os.Args[1]
	}

	params, _ := dkls.NewParams(btcec.S256(), curves.NewK256Scalar())

	alice := dkls.NewAlice(params)
	bob := dkls.NewBob(params)

	m := []byte(msg)
	digest := sha256.Sum256(m)

	alicePipe, bobPipe := dkls.NewPipeWrappers()
	errors := make(chan error, 2)
	go func() {
		errors <- alice.DKG(alicePipe)
	}()
	go func() {
		errors <- bob.DKG(bobPipe)
	}()
	for i := 0; i < 2; i++ {
		if err := <-errors; err != nil {
			fmt.Printf("Here")
		}
	}

	fmt.Printf("Message to sign: %s\n", msg)
	fmt.Printf("\nAlice Secret Key: %x\n", alice.SkA.Bytes())
	fmt.Printf("Alice Public Key: %x\n", alice.Pk.Bytes())

	fmt.Printf("\nBob Secret Key: %x\n", bob.SkB.Bytes())
	fmt.Printf("Bob Public Key: %x\n", bob.Pk.Bytes())

	errors = make(chan error, 2)
	go func() {
		errors <- alice.Sign(digest[:], alicePipe)
	}()
	go func() {
		errors <- bob.Sign(digest[:], bobPipe)
	}()

	for i := 0; i < 2; i++ {
		if err := <-errors; err != nil {
			fmt.Printf("Signing failed!!!")
		}
	}

	fmt.Printf("Bob Signature: R:%x S:%x\n", bob.Sig.R, bob.Sig.S)

	fmt.Printf("\n=== Now checking with Bob's public key ===\n")
	publicKey := ecdsa.PublicKey{
		Curve: ecc.P256k1(), //secp256k1
		X:     bob.Pk.X,
		Y:     bob.Pk.Y,
	}

	rtn := ecdsa.Verify(&publicKey, digest[:], bob.Sig.R, bob.Sig.S)

	if rtn {
		fmt.Printf("Signature works. Yipee!")
	} else {
		fmt.Printf("Signature does not check out!!!")
	}

}

