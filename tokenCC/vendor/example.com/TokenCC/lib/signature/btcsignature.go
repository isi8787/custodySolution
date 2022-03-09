package signature


import (
	"crypto/ecdsa"

	"encoding/hex"
    "math/big"
	"bytes"
	"github.com/btcsuite/btcutil"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"



)


func SignatureScript(sigbytes []byte,  compressedBytes []byte) ([]byte, error) {
	return txscript.NewScriptBuilder().AddData(sigbytes).AddData(compressedBytes).Script()
}


type BitcoinPublicKey struct {
	X, Y *big.Int
}


const (
	pubkeyUncompressed byte = 0x4 // x coord + y coord
)

func SerializeUncompressed(p *ecdsa.PublicKey) []byte {
	b := make([]byte, 0, 65)
	b = append(b, pubkeyUncompressed)
	b = paddedAppend(32, b, p.X.Bytes())
	return paddedAppend(32, b, p.Y.Bytes())
}

func paddedAppend(size uint, dst, src []byte) []byte {
	for i := 0; i < int(size)-len(src); i++ {
		dst = append(dst, 0)
	}
	return append(dst, src...)
}



func CreateTx(addrPubKey btcutil.Address, destination string, amount int64, txid string , pkScript string) (*wire.MsgTx, []byte, error) {

	// extracting destination address as []byte from function argument (destination string)
	destinationAddr, err := btcutil.DecodeAddress(destination, &chaincfg.TestNet3Params)
	if err != nil {
	   return nil, nil, err
	}
 
	destinationAddrByte, err := txscript.PayToAddrScript(destinationAddr)
	if err != nil {
	   return nil, nil, err
	}
 
 
	// creating a new bitcoin transaction, different sections of the tx, including
	// input list (contain UTXOs) and outputlist (contain destination address and usually our address)
	// in next steps, sections will be field and pass to sign
	redeemTx, err := NewTx()
	if err != nil {
	   return nil, nil, err
	}
 
 
	utxoHash, err := chainhash.NewHashFromStr(txid)
	if err != nil {
	   return nil, nil, err
	}
 
	// the second argument is vout or Tx-index, which is the index
	// of spending UTXO in the transaction that Txid referred to
	// in this case is 0, but can vary different numbers
	outPoint := wire.NewOutPoint(utxoHash, 1)
 
	// making the input, and adding it to transaction
	txIn := wire.NewTxIn(outPoint, nil, nil)
	redeemTx.AddTxIn(txIn)
 
	// adding the destination address and the amount to
	// the transaction as output
	redeemTxOut := wire.NewTxOut(amount, destinationAddrByte)
	redeemTx.AddTxOut(redeemTxOut)

	sourcePKScript, err := hex.DecodeString(pkScript)
	if err != nil {
	   return nil, nil, nil
	}

	hash, err := txscript.CalcSignatureHash(sourcePKScript, txscript.SigHashAll, redeemTx, 0)
	if err != nil {
			return nil, nil, err
	}
	
	return redeemTx, hash , nil
}



func SignTx(redeemTx *wire.MsgTx, signature []byte) (string, error) {

	// since there is only one input, and want to add 
	// signature to it use 0 as index
	redeemTx.TxIn[0].SignatureScript = signature
 
	var signedTx bytes.Buffer
	redeemTx.Serialize(&signedTx)
 
	hexSignedTx := hex.EncodeToString(signedTx.Bytes())
 
	return hexSignedTx, nil
 }



 func NewTx() (*wire.MsgTx, error) {
	return wire.NewMsgTx(wire.TxVersion), nil
 }


