package keystore

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/go-fusion/common/math"
	"github.com/go-fusion/version"

	"github.com/go-fusion/protocol/crypto"
	"github.com/go-fusion/protocol/crypto/randentropy"
	"github.com/go-fusion/protocol/types"
	"golang.org/x/crypto/scrypt"
)

const (

	// StandardScryptN is the N parameter of Scrypt encryption algorithm, using 256MB
	// memory and taking approximately 1s CPU time on a modern processor.
	StandardScryptN = 1 << 18

	// StandardScryptP is the P parameter of Scrypt encryption algorithm, using 256MB
	// memory and taking approximately 1s CPU time on a modern processor.
	StandardScryptP = 1

	// LightScryptN is the N parameter of Scrypt encryption algorithm, using 4MB
	// memory and taking approximately 100ms CPU time on a modern processor.
	LightScryptN = 1 << 12

	// LightScryptP is the P parameter of Scrypt encryption algorithm, using 4MB
	// memory and taking approximately 100ms CPU time on a modern processor.
	LightScryptP = 6

	scryptR     = 8
	scryptDKLen = 32
)

type passphraseKeyStore struct {
	keyStore
	scryptN int
	scryptP int
}

type encryptedKeyJSON struct {
	Address    string
	N          int
	R          int
	P          int
	DKlen      int
	Salt       string
	IV         string
	Mac        string
	CipherText string
	Version    uint64
}

func (m *passphraseKeyStore) EncryptKey(k *Key, auth string) ([]byte, error) {
	authArray := []byte(auth)
	salt := randentropy.GetEntropyCSPRNG(32)
	derivedKey, err := scrypt.Key(authArray, salt, m.scryptN, scryptR, m.scryptP, scryptDKLen)
	if err != nil {
		return nil, err
	}
	encryptKey := derivedKey[:16]
	keyBytes := math.PaddedBigBytes(k.PrivateKey.D, 32)
	iv := randentropy.GetEntropyCSPRNG(aes.BlockSize) // 16

	aesBlock, err := aes.NewCipher(encryptKey)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(aesBlock, iv)
	cipherText := make([]byte, len(keyBytes))
	stream.XORKeyStream(cipherText, keyBytes)
	mac := crypto.Hash256(derivedKey[16:32], cipherText)

	if err != nil {
		return nil, err
	}

	return json.Marshal(&encryptedKeyJSON{
		Address:    hex.EncodeToString(k.Address[:]),
		N:          m.scryptN,
		R:          scryptR,
		P:          m.scryptP,
		DKlen:      scryptDKLen,
		Salt:       hex.EncodeToString(salt),
		IV:         hex.EncodeToString(iv),
		Mac:        hex.EncodeToString(mac[:]),
		CipherText: hex.EncodeToString(cipherText),
		Version:    version.KeyStoreVersion.Value,
	})
}

func (m *passphraseKeyStore) DecryptKey(data []byte, auth string) (*Key, error) {
	key := new(Key)

	var encryptedKeyJSON = &encryptedKeyJSON{}

	if err := json.Unmarshal(data, &encryptedKeyJSON); err != nil {
		return nil, err
	}

	if err := version.FromUint64(encryptedKeyJSON.Version).Compatible(version.KeyStoreVersion); err != nil {
		return nil, err
	}

	authArray := []byte(auth)
	salt, err := hex.DecodeString(encryptedKeyJSON.Salt)
	if err != nil {
		return nil, err
	}

	cipherText, err := hex.DecodeString(encryptedKeyJSON.CipherText)
	if err != nil {
		return nil, err
	}

	macBytes, err := hex.DecodeString(encryptedKeyJSON.Mac)
	if err != nil {
		return nil, err
	}
	var mac = types.BytesToHash(macBytes)

	iv, err := hex.DecodeString(encryptedKeyJSON.IV)
	if err != nil {
		return nil, err
	}

	addressBytes, err := hex.DecodeString(encryptedKeyJSON.Address)
	if err != nil {
		return nil, err
	}
	var address = types.BytesToAddress(addressBytes)

	derivedKey, err := scrypt.Key(authArray, salt, encryptedKeyJSON.N, encryptedKeyJSON.R, encryptedKeyJSON.P, encryptedKeyJSON.DKlen)

	if crypto.Hash256(derivedKey[16:32], cipherText) != mac {
		return nil, errors.New("Mac mismactch")
	}

	aesBlock, err := aes.NewCipher(derivedKey[:16])
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(aesBlock, iv)
	plainText := make([]byte, len(cipherText))
	stream.XORKeyStream(plainText, cipherText)

	key.Address = address
	key.PrivateKey = crypto.ToECDSAUnsafe(plainText)
	return key, nil
}

// NewPassphraseKeyStore ss
func NewPassphraseKeyStore(n, p int) KeyStore {

	return &keyStore{impl: &passphraseKeyStore{
		scryptN: n,
		scryptP: p,
	}}
}
