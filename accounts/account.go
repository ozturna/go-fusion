package accounts

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/go-fusion/accounts/keystore"
	"github.com/go-fusion/protocol/types"
)

// Account ss
type Account interface {
	Address() types.Address
	String() string
	Lock() error
	Unlock(passphrase string, timeout time.Duration) error
	Sign(hash types.Hash) ([]byte, error)
}

type account struct {
	address  types.Address
	key      *keystore.Key
	filename string
	ks       keystore.KeyStore
	mu       sync.RWMutex
}

func (m *account) Address() types.Address {
	return m.address
}

func (m *account) String() string {
	return fmt.Sprintf("%s At %s", m.address, m.filename)
}

func (m *account) Lock() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.key != nil && m.key.PrivateKey != nil {
		clearKey(m.key.PrivateKey)
	}
	m.key = nil
	return nil
}

func (m *account) Unlock(passphrase string, timeout time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key, err := m.ks.GetKey(m.filename, m.address, passphrase)
	if err != nil {
		return err
	}
	m.key = key
	if timeout > 0 {
		go func() {
			t := time.NewTimer(timeout)
			defer t.Stop()
			select {
			case <-t.C:
				m.Lock()
			}
		}()
	}
	return nil
}

func (m *account) Sign(hash types.Hash) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.key == nil {
		return nil, errors.New("Account Locked,Plaese Unlock First")
	}
	prv := m.key.PrivateKey
	return btcec.SignCompact(btcec.S256(), (*btcec.PrivateKey)(prv), hash[:], false)
}

func clearBytes(bytes []byte) {
	count, err := rand.Read(bytes)
	if count != len(bytes) || err != nil {
		for i := range bytes {
			bytes[i] = 0
		}
	}
}

func clearKey(k *ecdsa.PrivateKey) {
	z := k.D.Bits()
	for i := range z {
		z[i] = big.Word(rand.Int())
	}
}

func newAccount(dir string, ks keystore.KeyStore, auth string) (Account, error) {
	address, key, filename, err := ks.NewKey(dir, auth)
	if err != nil {
		return nil, err
	}
	var ret = &account{
		address:  address,
		key:      key,
		filename: filename,
		ks:       ks,
	}
	ret.Lock()
	return ret, nil
}
