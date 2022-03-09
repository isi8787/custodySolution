package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"fmt"
	"math/big"

	"github.com/dustinxie/ecc"

	"os"
	"strconv"

	"github.com/btcsuite/btcd/btcec"
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/coinbase/kryptology/pkg/paillier"
	"github.com/coinbase/kryptology/pkg/tecdsa/gg20/dealer"
	"github.com/coinbase/kryptology/pkg/tecdsa/gg20/participant"

)

var (
	testPrimes = []*big.Int{
		B10("186141419611617071752010179586510154515933389116254425631491755419216243670159714804545944298892950871169229878325987039840135057969555324774918895952900547869933648175107076399993833724447909579697857041081987997463765989497319509683575289675966710007879762972723174353568113668226442698275449371212397561567"),
		B10("94210786053667323206442523040419729883258172350738703980637961803118626748668924192069593010365236618255120977661397310932923345291377692570649198560048403943687994859423283474169530971418656709749020402756179383990602363122039939937953514870699284906666247063852187255623958659551404494107714695311474384687"),
		B10("62028909880050184794454820320289487394141550306616974968340908736543032782344593292214952852576535830823991093496498970213686040280098908204236051130358424961175634703281821899530101130244725435470475135483879784963475148975313832483400747421265545413510460046067002322131902159892876739088034507063542087523"),
		B10("321804071508183671133831207712462079740282619152225438240259877528712344129467977098976100894625335474509551113902455258582802291330071887726188174124352664849954838358973904505681968878957681630941310372231688127901147200937955329324769631743029415035218057960201863908173045670622969475867077447909836936523"),
		B10("52495647838749571441531580865340679598533348873590977282663145916368795913408897399822291638579504238082829052094508345857857144973446573810004060341650816108578548997792700057865473467391946766537119012441105169305106247003867011741811274367120479722991749924616247396514197345075177297436299446651331187067"),
		B10("118753381771703394804894143450628876988609300829627946826004421079000316402854210786451078221445575185505001470635997217855372731401976507648597119694813440063429052266569380936671291883364036649087788968029662592370202444662489071262833666489940296758935970249316300642591963940296755031586580445184253416139"),
	}
	dealerParams = &dealer.ProofParams{
		N:  B10("135817986946410153263607521492868157288929876347703239389804036854326452848342067707805833332721355089496671444901101084429868705550525577068432132709786157994652561102559125256427177197007418406633665154772412807319781659630513167839812152507439439445572264448924538846645935065905728327076331348468251587961"),
		H1: B10("130372793360787914947629694846841279927281520987029701609177523587189885120190605946568222485341643012763305061268138793179515860485547361500345083617939280336315872961605437911597699438598556875524679018909165548046362772751058504008161659270331468227764192850055032058007664070200355866555886402826731196521"),
		H2: B10("44244046835929503435200723089247234648450309906417041731862368762294548874401406999952605461193318451278897748111402857920811242015075045913904246368542432908791195758912278843108225743582704689703680577207804641185952235173475863508072754204128218500376538767731592009803034641269409627751217232043111126391"),
	}
	k256Verifier = func(pubKey *curves.EcPoint, hash []byte, sig *curves.EcdsaSignature) bool {
		btcPk := &btcec.PublicKey{
			Curve: btcec.S256(),
			X:     pubKey.X,
			Y:     pubKey.Y,
		}
		btcSig := btcec.Signature{
			R: sig.R,
			S: sig.S,
		}
		return btcSig.Verify(hash, btcPk)
	}
)

func getParams(msg *string, t, n *uint32) {
	argCount := len(os.Args[1:])

	if argCount > 0 {
		*msg = os.Args[1]

	}
	if argCount > 1 {
		val, _ := strconv.Atoi(os.Args[2])
		*t = uint32(val)
	}
	if argCount > 2 {
		val, _ := strconv.Atoi(os.Args[3])
		*n = uint32(val)
	}

}
func genPrimesArray(count int) []struct{ p, q *big.Int } {
	primesArray := make([]struct{ p, q *big.Int }, 0, count)
	for len(primesArray) < count {
		for i := 0; i < len(testPrimes) && len(primesArray) < count; i++ {
			for j := 0; j < len(testPrimes) && len(primesArray) < count; j++ {
				if i == j {
					continue
				}
				keyPrime := struct {
					p, q *big.Int
				}{
					testPrimes[i], testPrimes[j],
				}
				primesArray = append(primesArray, keyPrime)
			}
		}
	}
	return primesArray
}

func B10(s string) *big.Int {
	x, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic("Couldn't derive big.Int from string")
	}
	return x
}

type signingSetup struct {
	curve        elliptic.Curve
	pk           *curves.EcPoint
	sharesMap    map[uint32]*dealer.Share
	pubSharesMap map[uint32]*dealer.PublicShare
	pubkeys      map[uint32]*paillier.PublicKey
	privkeys     map[uint32]*paillier.SecretKey
	proofParams  *dealer.TrustedDealerKeyGenType
}



func sign2p(msg string, setup *signingSetup) {
	// Hash of message for signature
	m := []byte(msg)
	msgHash := sha256.Sum256(m)

	// Create signers
	p1 := participant.Participant{*setup.sharesMap[1], setup.privkeys[1]}
	s1, _ := p1.PrepareToSign(
		setup.pk,
		k256Verifier,
		setup.curve,
		setup.proofParams,
		setup.pubSharesMap,
		setup.pubkeys)

	p2 := participant.Participant{*setup.sharesMap[2], setup.privkeys[2]}
	s2, _ := p2.PrepareToSign(
		setup.pk,
		k256Verifier,
		setup.curve,
		setup.proofParams,
		setup.pubSharesMap,
		setup.pubkeys)

	//
	// Sign Round 1
	//
	r1_s1_bcast, r1_s1_p2p, _ := s1.SignRound1()

	r1_s2_bcast, r1_s2_p2p, _ := s2.SignRound1()

	//
	// Sign Round 2
	//
	r2_s1_p2p, _ := s1.SignRound2(
		map[uint32]*participant.Round1Bcast{2: r1_s2_bcast},
		map[uint32]*participant.Round1P2PSend{2: r1_s2_p2p[1]},
	)

	r2_s2_p2p, _ := s2.SignRound2(
		map[uint32]*participant.Round1Bcast{1: r1_s1_bcast},
		map[uint32]*participant.Round1P2PSend{1: r1_s1_p2p[2]},
	)

	//
	// Sign Round 3
	//
	r3_s1_bcast, _ := s1.SignRound3(
		map[uint32]*participant.P2PSend{2: r2_s2_p2p[1]},
	)

	r3_s2_bcast, _ := s2.SignRound3(
		map[uint32]*participant.P2PSend{1: r2_s1_p2p[2]},
	)

	//
	// Sign Round 4
	//
	r4_s1_bcast, _ := s1.SignRound4(
		map[uint32]*participant.Round3Bcast{2: r3_s2_bcast})

	r4_s2_bcast, _ := s2.SignRound4(
		map[uint32]*participant.Round3Bcast{1: r3_s1_bcast})

	//
	// Sign Round 5
	//
	r5_s1_bcast, r5_s1_p2p, _ := s1.SignRound5(
		map[uint32]*participant.Round4Bcast{2: r4_s2_bcast},
	)

	r5_s2_bcast, r5_s2_p2p, _ := s2.SignRound5(
		map[uint32]*participant.Round4Bcast{1: r4_s1_bcast},
	)

	//
	// Sign Round 6
	//
	r6_s1_bcast, _ := s1.SignRound6Full(msgHash[:],
		map[uint32]*participant.Round5Bcast{2: r5_s2_bcast},
		map[uint32]*participant.Round5P2PSend{2: r5_s2_p2p[1]},
	)

	r6_s2_bcast, _ := s2.SignRound6Full(msgHash[:],
		map[uint32]*participant.Round5Bcast{1: r5_s1_bcast},
		map[uint32]*participant.Round5P2PSend{1: r5_s1_p2p[2]},
	)

	//
	// Compute signature
	//

	s1_sig, _ := s1.SignOutput(
		map[uint32]*participant.Round6FullBcast{2: r6_s2_bcast},
	)

	s2_sig, _ := s2.SignOutput(
		map[uint32]*participant.Round6FullBcast{1: r6_s1_bcast},
	)

	publicKey := ecdsa.PublicKey{
		Curve: ecc.P256k1(), //secp256k1
		X:     setup.pk.X,
		Y:     setup.pk.Y,
	}

	fmt.Printf("\nNode 1 signature: (%d %d)\n", s1_sig.R, s1_sig.S)

	fmt.Printf("Node 2 signature: (%d %d)\n", s2_sig.R, s2_sig.S)

	rtn := ecdsa.Verify(&publicKey, msgHash[:], s1_sig.R, s1_sig.S)
	fmt.Printf("\nSignature Verified: %v", rtn)

}

func main() {

	tshare := uint32(2)
	nshare := uint32(2)
	msg := "Hello"

	getParams(&msg, &tshare, &nshare)

	k256 := btcec.S256()

	ikm, _ := dealer.NewSecret(k256)

	pk, sharesMap, _ := dealer.NewDealerShares(k256, tshare, nshare, ikm)

	fmt.Printf("Message: %s\n", msg)
	fmt.Printf("Sharing scheme: Any %d from %d\n", tshare, nshare)
	fmt.Printf("Random secret: (%x)\n\n", ikm)
	fmt.Printf("Public key: (%s %s)\n\n", pk.X, pk.Y)

	for i := range sharesMap {
		fmt.Printf("Share: %x\n", sharesMap[i].Bytes())
	}

	pubSharesMap, _ := dealer.PreparePublicShares(sharesMap)

	keyPrimesArray := genPrimesArray(2)
	paillier1, _ := paillier.NewSecretKey(keyPrimesArray[0].p, keyPrimesArray[0].q)

	paillier2, _ := paillier.NewSecretKey(keyPrimesArray[1].p, keyPrimesArray[1].q)

	dealerSetup := &signingSetup{
		curve:        k256,
		pk:           pk,
		sharesMap:    sharesMap,
		pubSharesMap: pubSharesMap,
		pubkeys: map[uint32]*paillier.PublicKey{
			1: &paillier1.PublicKey,
			2: &paillier2.PublicKey,
		},
		privkeys: map[uint32]*paillier.SecretKey{
			1: paillier1,
			2: paillier2,
		},
		proofParams: &dealer.TrustedDealerKeyGenType{
			ProofParams: dealerParams,
		},
	}

	sign2p(msg, dealerSetup)

}


