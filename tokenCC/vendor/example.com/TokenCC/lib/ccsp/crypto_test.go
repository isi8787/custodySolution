package ccsp

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"testing"
)

func TestVerify(t *testing.T) {
	pubKeyXString := "tozGuxFTzVdl7/vk3kIz3zYbIYSk15e/h0/yIAERH2g="

	//+fjLsUDHCzzs+maIG09Q4W0aT9cgMnGDffp6kuiWzrc=\n" - K1

	pubKeyYString := "XOYx0DbRZAv2eDIhpg0uo4WsM+e73wCqm++Y5xA8xwM="

	//"4gbiHTUONGB64AkROH2uh5RnoEspdGZJUl//esBRj94=" - k1

	data := "MTIzNDU2Nzg="

	signature := "MEQCIHv0K4EjtUM8HVJu+LsyWBpTjShkdZIaBWqQIxMtobyDAiA1cFss/+N/j6WF2Tjl1WiMunW7JPtheLLDPl5mwRf1DA=="

	//"MEUCICzYu8w5n921mbtf5XAiGmGeT3nBDUgUhUiCSCMieSdcAiEAivlSX7RuKC2FEfSJOESzgAT7mcpw9A0Cx5GTNoxso2s=" - k1
	//"MEUCICzYu8w5n921mbtf5XAiGmGeT3nBDUgUhUiCSCMieSdcAiEAivlSX7RuKC2FEfSJOESzgAT7
	//mcpw9A0Cx5GTNoxso2s="
	dataByte, error := base64.StdEncoding.DecodeString(data)
	if error != nil {
		t.Errorf("Data Decode failed with error [%s]", error)
	}

	hash := sha256.Sum256(dataByte)
	signatureByte, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		t.Errorf("signature Decode failed with error [%s]", err)
	}
	pubKeyXBytes, err1 := base64.StdEncoding.DecodeString(pubKeyXString)
	if err1 != nil {
		t.Errorf("pubKeyXBytes Decode failed with error [%s]", err1)
	}
	pubKeyYBytes, err2 := base64.StdEncoding.DecodeString(pubKeyYString)
	if err2 != nil {
		t.Errorf("pubKeyYBytes Decode failed with error [%s]", err2)
	}
	isSignatureValid, _ := VerifySignature("SECP", "P256r1", signatureByte, hash[:], pubKeyXBytes, pubKeyYBytes)

	if !isSignatureValid {
		t.Error("KCSSSignatureValidation failed")
	}

}

func TestEnum(t *testing.T) {

	if "P256r1" == P256r1.String() {
		fmt.Println("AAA")
	}
	if "SECP" == SECP.String() {
		fmt.Println("BBB")
	}

	err := validateSupportedCurve("SECP", "P256r1")
	if err != nil {
		t.Errorf("Enum test failed with error [%s]", err)
	}

}
