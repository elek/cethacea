package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/elek/cethacea/pkg/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"strings"
)

const (
	DefaultAccountFile = ".accounts.yaml"
)

type AccountRepo struct {
	Accounts []types.Account
	Selected string
}

func NewAccountRepo(selected string) (*AccountRepo, error) {
	var accounts []types.Account
	err := LoadYamlConfig(DefaultAccountFile, "accounts", &accounts)
	if err != nil {
		return nil, err
	}

	if len(accounts) == 0 && (selected == "" || selected == "<generated>") {
		key := make([]byte, 32)
		_, _ = rand.Read(key)

		accounts = append(accounts, types.Account{
			Name:    "<generated>",
			Private: hex.EncodeToString(key),
		})
		selected = "<generated>"
	}

	if selected != "" {
		if _, err := os.Stat(selected); err == nil {
			//it's a file
			content, err := ioutil.ReadFile(selected)
			if err != nil {
				return nil, err
			}
			keyString := strings.TrimSpace(string(content))
			_, err = crypto.HexToECDSA(keyString)
			if err != nil {
				return nil, err
			}

			accounts = append(accounts, types.Account{
				Name:    selected,
				Private: keyString,
			})
			selected = selected
		}

		_, err := hex.DecodeString(selected)
		if err == nil {
			accounts = append(accounts, types.Account{
				Name:    "<pk>",
				Private: "0x" + selected,
			})
			selected = "<pk>"
		}

		if strings.HasPrefix(selected, "0x") {
			_, err := hex.DecodeString(selected[2:])
			if err == nil {
				accounts = append(accounts, types.Account{
					Name:    "<pk>",
					Private: "0x" + selected,
				})
				selected = "<pk>"
			}
		}
	}

	if len(accounts) == 1 && selected == "" {
		selected = accounts[0].Name
	}
	return &AccountRepo{
		Accounts: accounts,
		Selected: selected,
	}, nil
}

func (r *AccountRepo) GetCurrentAccount() (types.Account, error) {
	for _, c := range r.Accounts {
		if r.Selected == c.Name {
			return c, nil
		}
	}
	return types.Account{}, errors.New(fmt.Sprintf("Selected account %s is not found", r.Selected))
}

func (r *AccountRepo) ListAccounts() ([]types.Account, error) {
	return r.Accounts, nil
}

func (r *AccountRepo) GetNextName() string {
	for k := 1; k < 100; k++ {
		name := fmt.Sprintf("key%d", k)
		exists, err := r.AccountExists(name)
		if !exists && err == nil {
			return name
		}
	}
	return ""
}

func (r *AccountRepo) AccountExists(name string) (bool, error) {
	for _, acc := range r.Accounts {
		if acc.Name == name {
			return true, nil
		}
	}
	return false, nil
}

func (r *AccountRepo) AddAccount(a types.Account) error {
	var accounts []types.Account
	err := LoadYamlConfig(DefaultAccountFile, "accounts", &accounts)
	if err != nil {
		return err
	}
	accounts = append(accounts, a)
	return SaveYamlConfig(DefaultAccountFile, &accounts)
}

func (r *AccountRepo) GetAccount(name string) (types.Account, error) {
	for _, a := range r.Accounts {
		if a.Name == name {
			return a, nil
		}
	}
	return types.Account{}, errors.New("Default account is couldn't be identified")
}
