package main

import (
	"fmt"
	"math/big"
	"os"

	"github.com/coinbase/kryptology/pkg/core"
)

type (
	PublicKey struct {
		N  *big.Int
		N2 *big.Int
	}

	SecretKey struct {
		PublicKey
		Lambda  *big.Int
		Totient *big.Int
		U       *big.Int
	}
	Ciphertext *big.Int
)

func add(pk PublicKey, a Ciphertext, b Ciphertext) Ciphertext {

	c := new(big.Int).Mul(a, b)
	return (new(big.Int).Mod(c, pk.N2))

}
func encrypt(pk PublicKey, msg *big.Int) *big.Int {

	r, _ := core.Rand(pk.N)

	a := new(big.Int).Add(pk.N, big.NewInt(1))
	a.Exp(a, msg, pk.N2)
	b := new(big.Int).Exp(r, pk.N, pk.N2) // b = r^N (mod N^2)

	// ciphertext = a*b = (N+1)^m * r^N  (mod N^2)
	c := Mul(a, b, pk.N2)

	return c
}
func L(n, x *big.Int) *big.Int {
	// (x - 1) / n
	b := new(big.Int).Sub(x, big.NewInt(1))
	return b.Div(b, n)
}

func decrypt(sk SecretKey, c *big.Int) *big.Int {

	// a ≡ c^{lambda(N)}		mod N^2
	a := new(big.Int).Exp(c, sk.Lambda, sk.N2)

	// l = L(ɑ, N)
	ell := L(sk.N, a)

	// m ≡ lu = L(ɑ)*u = L(c^{λ(N)})*u	mod N
	m := Mul(ell, sk.U, sk.N)

	return m
}

func Mul(x, y, m *big.Int) *big.Int {

	z := new(big.Int).Mul(x, y)

	return z.Mod(z, m)
}
func Inv(x, m *big.Int) *big.Int {

	z := new(big.Int).ModInverse(x, m)
	return (z)
}
func mul(a, c, m *big.Int) *big.Int {

	z := new(big.Int).Exp(c, a, m)
	return (z)
}
func setupKeys(sec *SecretKey, pub *PublicKey) {
	pub.N, _ = new(big.Int).SetString("22203902867524505059996239340306362808852805402888214954381553003002718752808306965243974655390219346481612755387890570991182385566928749760445875916800573782909761881261515602762049819293013811136510263722491329215251675663091154175860620927146517652389408089110716148633480085801107700968384078929774277970426932561081560231010426294975678729992804063220974701278229766883426991469078323539488917623430196595127834729964807458110080684240115196595760172158113810254192728271785178985307185853395355962836026777351498860874006114137632167254987479651229489157192247478252351962954320801263428208801271515398015887801", 10)
	pub.N2 = new(big.Int).Mul(pub.N, pub.N)

	sec.N = pub.N
	sec.N2 = pub.N2
	sec.Lambda, _ = new(big.Int).SetString("11101951433762252529998119670153181404426402701444107477190776501501359376404153482621987327695109673240806377693945285495591192783464374880222937958400286891454880940630757801381024909646506905568255131861245664607625837831545577087930310463573258826194704044555358074316740042900553850484192039464887138985064438949068643503538028104882092520753872226183177268421975002892541203962036580811698912148170624870917841537346985483337432253649017635885024033744586145456966646545432316660152135614842196027111652507352356170725302821683928442675667004336523805058723372589095589316741830468728743532156406225121890756346", 10)
	sec.Totient, _ = new(big.Int).SetString("22203902867524505059996239340306362808852805402888214954381553003002718752808306965243974655390219346481612755387890570991182385566928749760445875916800573782909761881261515602762049819293013811136510263722491329215251675663091154175860620927146517652389408089110716148633480085801107700968384078929774277970128877898137287007076056209764185041507744452366354536843950005785082407924073161623397824296341249741835683074693970966674864507298035271770048067489172290913933293090864633320304271229684392054223305014704712341450605643367856885351334008673047610117446745178191178633483660937457487064312812450243781512692", 10)
	sec.U, _ = new(big.Int).SetString("11108720355657041776647490476262423100273444107890028034525926371450220865341816894255404548947647952186614924131107824022774134150177636593890919577479093776330856863330832621779073688637362125712752218152358601679371477743308049274665882845748928958872854339383876978533793119837835146167726752137945969833510158025058440018805812646143848801020667307117190186323497007267362203459968413761903710129427082211518450427328378521338965975284389455704146581178957343955351641425309976491396722227064973929033480821132777267246266490713852090985090755453080857556125803448778856380844874040092451545474091429770838805", 10)

}
func main() {

	var sec SecretKey
	var pub PublicKey

	//	pub, sec, _ := paillier.NewKeys()
	setupKeys(&sec, &pub)

	v1 := "223"
	v2 := "224"
	argCount := len(os.Args[1:])

	if argCount > 0 {
		v1 = os.Args[1]
	}
	if argCount > 1 {
		v2 = os.Args[2]
	}

	val1, _ := new(big.Int).SetString(v1, 10)
	val2, _ := new(big.Int).SetString(v2, 10)

	cipher1 := encrypt(pub, val1)

	cipher2 := encrypt(pub, val2)

	// Adding
	cipher3 := add(pub, cipher1, cipher2)

	decrypted1 := decrypt(sec, cipher3)

	// Subtraction
	inv_cipher2 := Inv(cipher2, pub.N2)

	cipher4 := add(pub, cipher1, inv_cipher2)

	decrypted2 := decrypt(sec, cipher4)

	// Multiply
	cipher5 := mul(val1, cipher2, pub.N2)

	decrypted3 := decrypt(sec, cipher5)

	fmt.Printf("\na=%s, b=%s\n", val1, val2)

	fmt.Printf("\nEnc(a)=%v, Enc(b)=%v\n", cipher1, cipher2)

	fmt.Printf("\n%s + %s = %s\n", val1, val2, decrypted1)
	fmt.Printf("\n%s * %s = %s\n", val1, val2, decrypted3)

	if decrypted2.Cmp(new(big.Int).Div(pub.N, big.NewInt(2))) > 0 {
		// m = m - n
		fmt.Printf("%s - %s = %s\n", val1, val2, new(big.Int).Sub(decrypted2, pub.N))
	} else {
		fmt.Printf("%s - %s = %s\n", val1, val2, decrypted2)
	}

}
