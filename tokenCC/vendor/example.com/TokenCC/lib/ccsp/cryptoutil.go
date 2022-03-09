// SPDX-License-Identifier: Apache-2.0

/*
  Cryptography Utility functions for cryptographic operations


*/

package ccsp

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/asn1"
	"errors"
	"fmt"
	"math/big"
)

type curveName int
type curveType int

//Only SECP256r1 is supported in Golang

const (
	SECP curveName = iota
)

const (
	P256r1 curveType = iota
)

type ECDSASignature struct {
	R, S *big.Int
}

func (cName curveName) String() string {

	suppportedCurves := [...]string{"SECP"}

	/* To be used when more curve types are supported
		if s < lowerValue || s > UpperValue {
	        return "Unknown"
		}
	*/

	return suppportedCurves[cName]
}

func (cType curveType) String() string {

	supportedCurveTypes := [...]string{"P256r1"}

	return supportedCurveTypes[cType]
}

//validateSupportedCurve checkes whethee the curve is supported or not
func validateSupportedCurve(curveName, curveType string) error {
	if curveName != SECP.String() {
		return fmt.Errorf("Curve [%s] not supported", curveName)
	}
	if curveType != P256r1.String() {
		return fmt.Errorf("Curve Type [%s] of Curve[%s] is not supported ", curveType, curveName)
	}
	return nil
}

//VerifyPublikKeyParam verifies public Key for the suoported curve
//returns ECDSSA public key
func VerifyPublikKeyParam(curveName, curveType string, x, y []byte) (*ecdsa.PublicKey, bool, error) {

	curveValidationError := validateSupportedCurve(curveName, curveType)
	if curveValidationError != nil {
		return nil, false, curveValidationError
	}
	xBigInt := new(big.Int).SetBytes(x)
	yBigInt := new(big.Int).SetBytes(y)

	pubKey := new(ecdsa.PublicKey)
	pubKey.X = xBigInt
	pubKey.Y = yBigInt

	pubKey.Curve = elliptic.P256()

	isOnCurve := elliptic.P256().IsOnCurve(xBigInt, yBigInt)
	return pubKey, isOnCurve, nil
}

//VerifySignature verifies ECDSA Signture
func VerifySignature(curveName, curveType string, signature, messageDigest, x, y []byte) (bool, error) {

	pubKey, isOnCurve, error := VerifyPublikKeyParam(curveName, curveType, x, y)
	if error != nil {
		return false, error
	}
	if !isOnCurve {
		fmt.Println("IsOnCurve not valid")
		return false, errors.New("Point is not on curve")
	}

	sig := &ECDSASignature{}
	_, err := asn1.Unmarshal(signature, sig)
	if err != nil {
		return false, err
	}
	isValid := ecdsa.Verify(pubKey, messageDigest, sig.R, sig.S)

	return isValid, nil

}
