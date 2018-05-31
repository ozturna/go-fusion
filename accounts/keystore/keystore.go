package keystore

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/go-fusion/protocol/crypto"
	"github.com/go-fusion/protocol/types"
)

// Key ss
type Key struct {
	Address types.Address

	PrivateKey *ecdsa.PrivateKey
}

// KeyStore interface
type KeyStore interface {
	// Loads and decrypts the key from disk.
	GetKey(filename string, addr types.Address, auth string) (*Key, error)

	// Writes and encrypts the key.
	StoreKey(filename string, k *Key, auth string) error

	// Create a key and writes encrypts the key to file.
	NewKey(dir string, auth string) (types.Address, *Key, string, error)

	EncryptKey(k *Key, auth string) ([]byte, error)

	DecryptKey(data []byte, auth string) (*Key, error)
}

type keyStore struct {
	impl KeyStore
}

func (m *keyStore) GetKey(filename string, addr types.Address, auth string) (*Key, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var key *Key

	if m.impl != nil {
		key, err = m.impl.DecryptKey(data, auth)
	} else {
		key, err = m.DecryptKey(data, auth)
	}

	if err != nil {
		return nil, err
	}

	if key.Address != addr {
		return nil, fmt.Errorf("key content mismatch: have address %x, want %x", key.Address, addr)
	}
	return key, nil
}

func (m *keyStore) StoreKey(filename string, k *Key, auth string) error {
	var (
		data []byte
		err  error
	)
	if m.impl != nil {
		data, err = m.impl.EncryptKey(k, auth)
	} else {
		data, err = m.EncryptKey(k, auth)
	}

	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0600)
}

func (m *keyStore) NewKey(dir string, auth string) (types.Address, *Key, string, error) {
	var key, err = newKey(rand.Reader)
	if err != nil {
		return types.Address{}, nil, "", err
	}
	var address = crypto.PubkeyToAddress(&key.PrivateKey.PublicKey)
	var filename = filepath.Join(dir, keyFileName(address))

	err = m.StoreKey(filename, key, auth)

	if err != nil {
		return types.Address{}, nil, "", err
	}

	return address, key, filename, nil
}

// plain
func (m *keyStore) EncryptKey(k *Key, auth string) ([]byte, error) {
	return json.Marshal(k)
}

// plain
func (m *keyStore) DecryptKey(data []byte, auth string) (*Key, error) {
	key := new(Key)
	if err := json.Unmarshal(data, key); err != nil {
		return nil, err
	}
	return key, nil
}

func newKey(rand io.Reader) (*Key, error) {
	privateKeyECDSA, err := ecdsa.GenerateKey(btcec.S256(), rand)
	if err != nil {
		return nil, err
	}
	return newKeyFromECDSA(privateKeyECDSA), nil
}

func newKeyFromECDSA(privateKeyECDSA *ecdsa.PrivateKey) *Key {
	key := &Key{
		Address:    crypto.PubkeyToAddress(&privateKeyECDSA.PublicKey),
		PrivateKey: privateKeyECDSA,
	}
	return key
}

func keyFileName(keyAddr types.Address) string {
	ts := time.Now().UTC()
	return fmt.Sprintf("UTC--%s--%s.json", toISO8601(ts), hex.EncodeToString(keyAddr[:]))
}

func toISO8601(t time.Time) string {
	var tz string
	name, offset := t.Zone()
	if name == "UTC" {
		tz = "Z"
	} else {
		tz = fmt.Sprintf("%03d00", offset/3600)
	}
	return fmt.Sprintf("%04d-%02d-%02dT%02d-%02d-%02d.%09d%s", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), tz)
}
