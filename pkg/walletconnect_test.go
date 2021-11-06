package cethacea

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_decrypt(t *testing.T) {
	msg := EncryptionEnvelope{
		Data: "f9952f462930e8f93ad0642b0751cd5717cc587bf92de436ec6deafe0cd4ee38adea678855a90ad716e2e613a032ea88cd1fc28f8f59829c69ac670adad3b42a78e8fd56cb66eb4aeb0b4b701a6e3aebc53a58b32872e3a1d9a9678caff161bf39bef3c4bbfde5cd4c2abe1e7d3bb6a4f0e8e68efc981a97713341b1a4819996704307a35812bfa0dc81ce3037b9b76baa520f1e181780e3f90d83c351e5342ee0635d62cb13a18a7938d0e29acfa22f1e05aa17ae6f2cae04693c42c5962d9b05e1e37d76fed4fd993e08d4bc228ee75aa33ebf7bcc354acebc4a576778ff1d9c64b226f2ed9a1e0c23f363484a8a91f34366ca32036723e9f5f5470e4ef9a2849bdaafe7f6b71134d3eb77eecbda34d9818092a7bed766d4bd97f52a761ba55542f9dfef70617adcaed42647addc976a9f8bf8c9e5f93f2c1fc80e317d7bd3bd01293d4a5bd212633c061b699ca132d61b5b60c24be2d4983be27ab21dba6f3c97c835aaea5fbe49e51d01a65b90a71a0cb7f9ff75758720fa492f20aba2b716d0a4f3e3968c03e58b79729bb0b9a4542f69f244936a2f19da8b8b45664831bd18d5ba96116d571e335e20731023c4",
		Hmac: "4e9b41ea0c48e9d169d46fcd569e6752ad347f2ff5a39aba7b1ace15fe75b589",
		Iv:   "c8906359b0bb5ed115355fc966157793",
	}
	key, err := hex.DecodeString("7ba379f77267bd721e970b411fe745af90725b335de49d29c37da7ce0eb63b3b")
	require.NoError(t, err)
	decrypted, err := decryptMessage(msg, key)
	require.NoError(t, err)
	require.Contains(t, decrypted, "jsonrpc")
	fmt.Println(decrypted)
}

func Test_encrytDecript(t *testing.T) {
	msg := "ahoj ahoj poplacsek"
	key, err := hex.DecodeString("7ba379f77267bd721e970b411fe745af90725b335de49d29c37da7ce0eb63b3b")
	require.NoError(t, err)

	env, err := encryptMessage([]byte(msg), key)
	require.NoError(t, err)

	resp, err := decryptMessage(env, key)
	require.NoError(t, err)
	require.Equal(t, msg, resp)
}

func Test_sign(t *testing.T) {
	pk, err := crypto.HexToECDSA("72e06ca1f2a055a4f531d48616a744ca9e0682c32035fadd9f56d814a9704309")
	require.NoError(t, err)
	message, err := hex.DecodeString("416363657373207a6b53796e63206163636f756e742e0a0a4f6e6c79207369676e2074686973206d65737361676520666f722061207472757374656420636c69656e7421")
	require.NoError(t, err)
	signature, err := personalSign(message, pk)
	require.NoError(t, err)
	expected, err := hex.DecodeString("8be981f0d4356c8ad2e32bc9f68384da154cb6f5bdef16306fc8135fe08751604539a8045c838825def79a71c5508650176fd1cf258a003cf74ea5444c9f86ed1c")
	require.NoError(t, err)
	require.Equal(t, expected, signature)
}
