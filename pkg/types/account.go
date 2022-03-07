package types

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Account struct {
	Name    string
	Private string
	Public  string
}

func (a Account) Address() common.Address {
	if a.Public != "" {
		return common.HexToAddress(a.Public)
	}
	pk, err := crypto.HexToECDSA(a.Private)
	if err != nil {
		panic(err)
	}
	return crypto.PubkeyToAddress(pk.PublicKey)
}

func (a Account) PrivateKey() *ecdsa.PrivateKey {
	pk, _ := crypto.HexToECDSA(a.Private)
	return pk
}
