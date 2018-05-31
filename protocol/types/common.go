package types

import (
	"encoding/hex"

	"golang.org/x/crypto/blake2b"
)

const (
	// HashBytesNumber ss
	HashBytesNumber = 32
	// AddressBytesNumber ss
	AddressBytesNumber = 20
)

//Hash ss
type Hash [HashBytesNumber]byte

// AssetID ss
type AssetID Hash

//Address ss
type Address [AddressBytesNumber]byte

// BytesToHash ss
func BytesToHash(b []byte) Hash {
	var h Hash
	h.SetBytes(b)
	return h
}

// Hex ss
func (h Hash) Hex() string { return hex.EncodeToString(h[:]) }

// String implements the stringer interface and is used also by the logger when
// doing full logging into a file.
func (h Hash) String() string {
	return "0x" + h.Hex()
}

// SetBytes Sets the hash to the value of b. If b is larger than len(h), 'b' will be cropped (from the left).
func (h *Hash) SetBytes(b []byte) {
	if len(b) > len(h) {
		b = b[len(b)-HashBytesNumber:]
	}
	copy(h[HashBytesNumber-len(b):], b)
}

// BytesToAddress ss
func BytesToAddress(b []byte) Address {
	var a Address
	a.SetBytes(b)
	return a
}

// SetBytes sets the address to the value of b. If b is larger than len(a) it will panic
func (a *Address) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-AddressBytesNumber:]
	}
	copy(a[AddressBytesNumber-len(b):], b)
}

// Hex returns an EIP55-compliant hex string representation of the address.
func (a Address) Hex() string {
	unchecksummed := hex.EncodeToString(a[:])
	hash := blake2b.Sum256([]byte(unchecksummed))

	result := []byte(unchecksummed)
	for i := 0; i < len(result); i++ {
		hashByte := hash[i/2]
		if i%2 == 0 {
			hashByte = hashByte >> 4
		} else {
			hashByte &= 0xf
		}
		if result[i] > '9' && hashByte > 7 {
			result[i] -= 32
		}
	}
	return string(result)
}

// String implements the stringer interface and is used also by the logger.
func (a Address) String() string {
	return "0x" + a.Hex()
}
