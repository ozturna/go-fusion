package accounts

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/go-fusion/accounts/keystore"
	"github.com/go-fusion/protocol/types"
)

// Manager ss
type Manager interface {
	Accounts() []Account
	NewAccount(passphrase string) (Account, error)
}

type manager struct {
	dir      string
	accounts map[types.Address]Account
	ks       keystore.KeyStore
}

// NewManager ss
func NewManager(dir string) Manager {
	var m = &manager{dir: dir}
	m.ks = keystore.NewPassphraseKeyStore(keystore.StandardScryptN, keystore.StandardScryptP)
	m.accounts = make(map[types.Address]Account)
	m.scanAccounts()
	return m
}

func (m *manager) scanAccounts() {
	var (
		buf = new(bufio.Reader)
		key struct {
			Address string
		}
	)
	readAccount := func(path string) Account {
		fd, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer fd.Close()
		buf.Reset(fd)

		key.Address = ""
		err = json.NewDecoder(buf).Decode(&key)

		if err != nil {
			return nil
		}
		addr, err := hex.DecodeString(key.Address)
		if err != nil || len(addr) != types.AddressBytesNumber {
			return nil
		}
		return &account{
			address:  types.BytesToAddress(addr),
			filename: path,
			ks:       m.ks,
		}
	}

	filepath.Walk(m.dir, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		account := readAccount(path)
		if account != nil {
			m.accounts[account.Address()] = account
		}
		return nil
	})
}

func (m *manager) Accounts() []Account {
	ret := make([]Account, len(m.accounts))
	var index = 0
	for _, v := range m.accounts {
		ret[index] = v
		index++
	}
	return ret
}

func (m *manager) NewAccount(passphrase string) (Account, error) {
	account, err := newAccount(m.dir, m.ks, passphrase)
	if err != nil {
		return nil, err
	}
	m.accounts[account.Address()] = account
	return account, nil
}
