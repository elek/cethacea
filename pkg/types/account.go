package types

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Account struct {
	Name    string
	Private string
}

func (a Account) Address() common.Address {
	pk, _ := crypto.HexToECDSA(a.Private)
	return crypto.PubkeyToAddress(pk.PublicKey)
}

func (a Account) PrivateKey() *ecdsa.PrivateKey {
	pk, _ := crypto.HexToECDSA(a.Private)
	return pk
}
