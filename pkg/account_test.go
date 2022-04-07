package cethacea

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
	"unsafe"
)

func TestGethKeyStore(t *testing.T) {
	store := keystore.NewKeyStore("/tmp/keystore", keystore.StandardScryptN, keystore.StandardScryptP)
	address := common.HexToAddress("0x158d2c25ba6107b622f288663f50f53601ab6710")
	var account accounts.Account

	for _, w := range store.Wallets() {
		s, err := w.Status()
		require.NoError(t, err)
		fmt.Println(s)
		for _, a := range w.Accounts() {
			if a.Address == address {
				account = a
			}
		}
	}
	err := store.Unlock(account, "Welcome1")
	require.NoError(t, err)

	rs := reflect.ValueOf(store).Elem()
	rf := rs.FieldByName("unlocked")
	// rf can't be read or set.
	rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
	key := rf.MapIndex(reflect.ValueOf(address))
	pk := key.Elem().FieldByName("PrivateKey").Interface()
	fmt.Println(hex.EncodeToString(crypto.FromECDSA(pk.(*ecdsa.PrivateKey))))
}
