package standards

import _ "embed"

//go:embed ERC20.abi
var erc20 []byte

func GetPredefinedContract(name string) ([]byte, bool) {
	switch name {
	case "@ERC-20", "<ERC-20>", "<ERC20>", "erc20":
		return erc20, true
	}
	return nil, false
}
