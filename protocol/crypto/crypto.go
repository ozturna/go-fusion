package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"fmt"
	"math/big"

	"github.com/btcsuite/btcd/btcec"
	"github.com/go-fusion/protocol/types"
	"golang.org/x/crypto/blake2b"
)

// Hash256 Common Hash
func Hash256(data ...[]byte) types.Hash {
	d, e := blake2b.New256(nil)
	if e != nil {
		return types.Hash{}
	}
	for _, b := range data {
		d.Write(b)
	}
	return types.BytesToHash(d.Sum(nil))
}

// ToECDSA creates a private key with the given D value.
func ToECDSA(d []byte) (*ecdsa.PrivateKey, error) {
	return toECDSA(d, true)
}

// ToECDSAUnsafe blindly converts a binary blob to a private key. It should almost
// never be used unless you are sure the input is valid and want to avoid hitting
// errors due to bad origin encoding (0 prefixes cut off).
func ToECDSAUnsafe(d []byte) *ecdsa.PrivateKey {
	priv, _ := toECDSA(d, false)
	return priv
}

// toECDSA creates a private key with the given D value. The strict parameter
// controls whether the key's length should be enforced at the curve size or
// it can also accept legacy encodings (0 prefixes).
func toECDSA(d []byte, strict bool) (*ecdsa.PrivateKey, error) {
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = btcec.S256()
	if strict && 8*len(d) != priv.Params().BitSize {
		return nil, fmt.Errorf("invalid length, need %d bits", priv.Params().BitSize)
	}
	priv.D = new(big.Int).SetBytes(d)

	// The priv.D must < N
	if priv.D.Cmp(btcec.S256().N) >= 0 {
		return nil, fmt.Errorf("invalid private key, >=N")
	}
	// The priv.D must not be zero or negative.
	if priv.D.Sign() <= 0 {
		return nil, fmt.Errorf("invalid private key, zero or negative")
	}

	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(d)
	if priv.PublicKey.X == nil {
		return nil, errors.New("invalid private key")
	}
	return priv, nil
}

// FromECDSAPub  public key to bytes
func FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(btcec.S256(), pub.X, pub.Y)
}

// PubkeyToAddress ss
func PubkeyToAddress(p *ecdsa.PublicKey) types.Address {
	pubBytes := FromECDSAPub(p)
	var data = Hash256(pubBytes[1:])
	return types.BytesToAddress(data[12:])
}

// SigToPub ss
func SigToPub(hash types.Hash, sig []byte) (*ecdsa.PublicKey, error) {
	pub, _, err := btcec.RecoverCompact(btcec.S256(), sig, hash[:])
	return (*ecdsa.PublicKey)(pub), err
}

// Sender ss
func Sender(hash types.Hash, sig []byte) (types.Address, error) {
	pub, err := SigToPub(hash, sig)
	if err != nil {
		return types.Address{}, err
	}
	return PubkeyToAddress(pub), nil
}
