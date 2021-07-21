// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by chain BSD-style
// license that can be found in the LICENSE file.

package ed25519

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"encoding/hex"
	"github.com/aead/ecdh"
	"testing"

	"github.com/vitelabs/go-vite/crypto/ed25519/internal/edwards25519"
)

type zeroReader struct{}

func (zeroReader) Read(buf []byte) (int, error) {
	for i := range buf {
		buf[i] = 0
	}
	return len(buf), nil
}

func UnmarshalMarshal(pub []byte, t *testing.T) {

	println("Generate Pub key : ", hex.EncodeToString(pub))
	var A edwards25519.ExtendedGroupElement
	var pubBytes [32]byte
	copy(pubBytes[:], pub)
	if !A.FromBytes(&pubBytes) {
		t.Fatalf("ExtendedGroupElement.FromBytes failed")
	}

	var pub2 [32]byte
	A.ToBytes(&pub2)

	if pubBytes != pub2 {
		t.Errorf("FromBytes(%v)->ToBytes does not round-trip, got %x\n", pubBytes, pub2)
	}
}

func TestIsValidPrivateKey(t *testing.T) {
	if IsValidPrivateKey([]byte("123")) {
		t.Fatal()
	}
	if IsValidPrivateKey([]byte("1234567812345678123456781234567812345678123456781234567812345678")) {
		t.Fatal()
	}
	_, pri, _ := GenerateKey(rand.Reader)
	if !IsValidPrivateKey(pri) {
		t.Fatal()
	}
}

func TestUnmarshalMarshal(t *testing.T) {
	pub, _, _ := GenerateKey(rand.Reader)
	UnmarshalMarshal(pub, t)
}

func TestUnmarshalMarshalDeterministic(t *testing.T) {
	{
		var zero [32]byte
		pub, _, _ := GenerateKeyFromD(zero)
		UnmarshalMarshal(pub, t)
	}

	{
		var D1 [32]byte
		for i, _ := range D1 {
			D1[i] = byte(i)
		}
		pub, _, _ := GenerateKeyFromD(D1)
		UnmarshalMarshal(pub, t)
	}
}

func TestSignVerify(t *testing.T) {

	for i := 0; i < 10; i++ {
		public, private, _ := GenerateKey(nil)
		println()
		println("priv: ", hex.EncodeToString(private))
		println("pub: ", hex.EncodeToString(public))
		message := "1234567890TEST"
		messageByte := []byte(message)
		println("message: ", hex.EncodeToString(messageByte))
		sig := Sign(private, messageByte)
		println("signdata: ", hex.EncodeToString(sig))
	}

	//message := "12345678901234567890"
	//pub := "5AD4455C87AF117B3A56AC816AE9BF9C92E566803C177FB206669F0F53609471"
	//signdata := "3761078C2BDBF90807A22E0309A6F0D5AD6765455466B840662A32F6AC044B98BC46726C4905449537DB5AA88CCA0F8B93F8C0249B3A826E0CE6A6F7D2019504"
	//M, _ := hex.DecodeString(message)
	//P, _ := hex.DecodeString(pub)
	//S, _ := hex.DecodeString(signdata)
	//if !Verify(P, M, S) {
	//	t.Fatal("not pass")
	//}

}

func TestSignVerifyRandom(t *testing.T) {
	for i := 0; i < 10000; i++ {
		public, private, _ := GenerateKey(nil)

		message := []byte("test message")
		sig := Sign(private, message)
		if !Verify(public, message, sig) {
			t.Errorf("valid signature rejected")
		}

		wrongMessage := []byte("wrong message")
		if Verify(public, wrongMessage, sig) {
			t.Errorf("signature of different message accepted")
		}
	}
}

func TestCryptoSigner(t *testing.T) {
	var zero zeroReader
	public, private, _ := GenerateKey(zero)

	signer := crypto.Signer(private)

	publicInterface := signer.Public()
	public2, ok := publicInterface.(PublicKey)
	if !ok {
		t.Fatalf("expected PublicKey from Public() but got %T", publicInterface)
	}

	if !bytes.Equal(public, public2) {
		t.Errorf("public keys do not match: original:%x vs Public():%x", public, public2)
	}

	message := []byte("message")
	var noHash crypto.Hash
	signature, err := signer.Sign(zero, message, noHash)
	if err != nil {
		t.Fatalf("error from Sign(): %s", err)
	}

	if !Verify(public, message, signature) {
		t.Errorf("Verify failed on signature from Sign()")
	}
}

func TestMalleability(t *testing.T) {
	// https://tools.ietf.org/html/rfc8032#section-5.1.7 adds an additional test
	// that s be in [0, order). This prevents someone from adding chain multiple of
	// order to s and obtaining chain second valid signature for the same message.
	msg := []byte{0x54, 0x65, 0x73, 0x74}
	sig := []byte{
		0x7c, 0x38, 0xe0, 0x26, 0xf2, 0x9e, 0x14, 0xaa, 0xbd, 0x05, 0x9a,
		0x0f, 0x2d, 0xb8, 0xb0, 0xcd, 0x78, 0x30, 0x40, 0x60, 0x9a, 0x8b,
		0xe6, 0x84, 0xdb, 0x12, 0xf8, 0x2a, 0x27, 0x77, 0x4a, 0xb0, 0x67,
		0x65, 0x4b, 0xce, 0x38, 0x32, 0xc2, 0xd7, 0x6f, 0x8f, 0x6f, 0x5d,
		0xaf, 0xc0, 0x8d, 0x93, 0x39, 0xd4, 0xee, 0xf6, 0x76, 0x57, 0x33,
		0x36, 0xa5, 0xc5, 0x1e, 0xb6, 0xf9, 0x46, 0xb3, 0x1d,
	}
	publicKey := []byte{
		0x7d, 0x4d, 0x0e, 0x7f, 0x61, 0x53, 0xa6, 0x9b, 0x62, 0x42, 0xb5,
		0x22, 0xab, 0xbe, 0xe6, 0x85, 0xfd, 0xa4, 0x42, 0x0f, 0x88, 0x34,
		0xb1, 0x08, 0xc3, 0xbd, 0xae, 0x36, 0x9e, 0xf5, 0x49, 0xfa,
	}

	if Verify(publicKey, msg, sig) {
		t.Fatal("non-canonical signature accepted")
	}
}

func BenchmarkKeyGeneration(b *testing.B) {
	var zero zeroReader
	for i := 0; i < b.N; i++ {
		if _, _, err := GenerateKey(zero); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSigning(b *testing.B) {
	var zero zeroReader
	_, priv, err := GenerateKey(zero)
	if err != nil {
		b.Fatal(err)
	}
	message := []byte("Hello, world!")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sign(priv, message)
	}
}

func BenchmarkVerification(b *testing.B) {
	var zero zeroReader
	pub, priv, err := GenerateKey(zero)
	if err != nil {
		b.Fatal(err)
	}
	message := []byte("Hello, world!")
	signature := Sign(priv, message)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Verify(pub, message, signature)
	}
}

func TestPrivateKey_Clear(t *testing.T) {
	priv := make([]byte, PrivateKeySize)
	for i := range priv {
		priv[i] = 1
	}
	var P PrivateKey
	P = priv
	P.Clear()
	if IsValidPrivateKey(P) {
		t.Fatal()
	}
	for i := range priv {
		if priv[i] == 1 {
			t.Fatal()
		}
	}

}

func TestX25519Exchange(t *testing.T) {
	pub1, priv1, _ := GenerateKey(nil)
	xpub1 := pub1.ToX25519Pk()
	println(hex.EncodeToString(xpub1))
	xpriv1 := priv1.ToX25519Sk()
	println(hex.EncodeToString(xpriv1))

	pub2, priv2, _ := GenerateKey(nil)
	xpub2 := pub2.ToX25519Pk()
	xpriv2 := priv2.ToX25519Sk()

	println(hex.EncodeToString(ecdh.X25519().ComputeSecret(&xpriv1, &xpub2)))

	println(hex.EncodeToString(ecdh.X25519().ComputeSecret(&xpriv2, &xpub1)))
}
