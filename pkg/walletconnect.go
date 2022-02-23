package cethacea

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/elek/cethacea/pkg/chain"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gorilla/websocket"
	"github.com/manifoldco/promptui"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"math/big"
	"net/url"
	"strconv"
	"strings"
)

func init() {

	walletConnectCmd := cobra.Command{
		Use:     "wallet-connect",
		Aliases: []string{"wc"},
		Short:   "WalletConnect related commands",
	}
	{
		cmd := cobra.Command{
			Use:   "connect <url>",
			Args:  cobra.ExactArgs(1),
			Short: "Create new wallet connect connection",
			RunE: func(cmd *cobra.Command, args []string) error {
				ceth, err := NewCethContext(&Settings)
				if err != nil {
					return err
				}
				return walletConnect(ceth, args[0])
			},
		}
		walletConnectCmd.AddCommand(&cmd)
	}
	RootCmd.AddCommand(&walletConnectCmd)
}

type EncryptionEnvelope struct {
	Data string `json:"data"`
	Iv   string `json:"iv"`
	Hmac string `json:"hmac"`
}
type Message struct {
	Topic   string `json:"topic"`
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

type JsonRpcRequest struct {
	Id      int64       `json:"id"`
	JsonRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type JsonRpcResponse struct {
	Id      int64       `json:"id"`
	JsonRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
}

func walletConnect(ceth *Ceth, wcUrl string) error {
	account, err := ceth.AccountRepo.GetCurrentAccount()
	if err != nil {
		return err
	}

	client, err := ceth.GetChainClient()
	if err != nil {
		return err
	}

	parsed, err := url.Parse(wcUrl)
	if err != nil {
		return errors.Wrap(err, "Wrong url")
	}
	topic := strings.Split(parsed.Opaque, "@")[0]
	bridge := parsed.Query().Get("bridge")
	key, err := hex.DecodeString(parsed.Query().Get("key"))
	if err != nil {
		return errors.Wrap(err, "Key is invalid")
	}
	c, _, err := websocket.DefaultDialer.Dial(strings.ReplaceAll(bridge, "http", "ws"), nil)
	if err != nil {
		return errors.Wrap(err, "Couldn't connect to "+bridge)
	}
	defer c.Close()

	uid, err := uuid.NewV4()
	if err != nil {
		return err
	}

	done := make(chan struct{})
	peerId := ""

	handleMessage := func(message []byte) (interface{}, error) {
		k := Message{}
		err := json.Unmarshal(message, &k)
		if err != nil {
			return nil, err
		}

		msg := EncryptionEnvelope{}
		err = json.Unmarshal([]byte(k.Payload), &msg)
		if err != nil {
			return nil, err
		}

		decrypted, err := decryptMessage(msg, key)
		if err != nil {
			return nil, err
		}

		jsonRPC := JsonRpcRequest{}
		fmt.Println(decrypted)
		err = json.Unmarshal([]byte(decrypted), &jsonRPC)
		if err != nil {
			return nil, errors.Wrap(err, "invalid json")
		}
		switch jsonRPC.Method {
		case "wc_sessionRequest":
			firstParams := jsonRPC.Params.([]interface{})[0]
			params := firstParams.(map[string]interface{})
			peerId = params["peerId"].(string)
			resp := JsonRpcResponse{
				Id:      jsonRPC.Id,
				JsonRPC: "2.0",
				Result: map[string]interface{}{
					"peerId":   uid.String(),
					"approved": true,
					"chainId":  params["chainId"],
					"accounts": []string{account.Address().String()},
				},
			}
			if err = confirm(); err != nil {
				return nil, err
			}
			return resp, nil
		case "eth_sendTransaction":
			firstParams := jsonRPC.Params.([]interface{})[0]
			params := firstParams.(map[string]interface{})
			to := common.HexToAddress(params["to"].(string))
			var opts []interface{}

			fmt.Println("Transaction to " + to.String())
			data := params["data"].(string)
			if data != "" {
				fmt.Println("Data: " + data)
				opts = append(opts, chain.WithData{Data: common.FromHex(data)})
			}

			value := ""
			if params["value"] != nil {
				value = params["value"].(string)
			}
			if value != "" {
				value = strings.TrimPrefix(value, "0x")
				parsedValue, ok := new(big.Int).SetString(value, 16)
				if !ok {
					return nil, errors.New("Couldn't parse value: " + value)
				}
				fmt.Println("Value: " + value)
				opts = append(opts, chain.WithValue{Value: parsedValue})
			}

			if err = confirm(); err != nil {
				return nil, err
			}
			hash, err := client.SendTransaction(context.Background(), account, &to, opts...)
			if err != nil {
				return nil, err
			}
			resp := JsonRpcResponse{
				Id:      jsonRPC.Id,
				JsonRPC: "2.0",
				Result:  "0x" + hash.Hex(),
			}
			return resp, nil
		case "personal_sign":
			params := jsonRPC.Params.([]interface{})
			hexMessage := params[0].(string)

			hexMessage = strings.TrimPrefix(hexMessage[2:], "0x")
			message, err := hex.DecodeString(hexMessage)
			if err != nil {
				return nil, err
			}
			fmt.Println("Message to sign: " + string(message))
			if err = confirm(); err != nil {
				return nil, err
			}
			signature, err := personalSign(message, account.PrivateKey())
			if err != nil {
				return nil, err
			}
			resp := JsonRpcResponse{
				Id:      jsonRPC.Id,
				JsonRPC: "2.0",
				Result:  "0x" + hex.EncodeToString(signature),
			}
			return resp, nil
		default:
			fmt.Println("No implementation for " + jsonRPC.Method)
		}
		return nil, nil
	}

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				fmt.Println("Error on receiving message: " + err.Error())
				break
			}
			resp, err := handleMessage(message)
			if err != nil {
				fmt.Println("Error on message processing: " + err.Error())
			}
			if resp != nil {
				out, _ := json.Marshal(resp)
				fmt.Println(string(out))

				msg, err := encryptMessage(out, key)
				if err != nil {
					fmt.Println("Error on encrypting message: " + err.Error())
					continue
				}

				payload, err := json.Marshal(msg)
				if err != nil {
					fmt.Println("Error on marshal response: " + err.Error())
					continue
				}

				err = c.WriteJSON(Message{
					Topic:   peerId,
					Type:    "pub",
					Payload: string(payload),
				})
				if err != nil {
					fmt.Println("Error on sending back response: " + err.Error())
				}
			}
		}
	}()

	err = subscribe(c, topic)
	if err != nil {
		return err
	}

	err = subscribe(c, uid.String())
	if err != nil {
		return err
	}

	<-done
	return nil
}

func personalSign(message []byte, key *ecdsa.PrivateKey) ([]byte, error) {
	prefix := []byte("\x19Ethereum Signed Message:\n" + strconv.Itoa(len(message)))
	var fullMessage = append(prefix, message...)
	hash := crypto.Keccak256Hash(fullMessage).Bytes()
	signature, err := crypto.Sign(hash, key)
	if err != nil {
		return nil, err
	}
	signature[len(signature)-1] = signature[len(signature)-1] + 27
	return signature, nil
}

func subscribe(c *websocket.Conn, topic string) error {
	out, err := json.Marshal(Message{
		Topic:   topic,
		Type:    "sub",
		Payload: "",
	})
	if err != nil {
		return err
	}

	err = c.WriteMessage(websocket.TextMessage, out)
	if err != nil {
		return err
	}
	return nil
}

func encryptMessage(message []byte, key []byte) (EncryptionEnvelope, error) {
	data, iv, err := encrypt(message, key)
	if err != nil {
		return EncryptionEnvelope{}, err
	}
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	mac.Write(iv)

	return EncryptionEnvelope{
		Data: hex.EncodeToString(data),
		Iv:   hex.EncodeToString(iv),
		Hmac: hex.EncodeToString(mac.Sum(nil)),
	}, nil
}
func decryptMessage(message EncryptionEnvelope, key []byte) (string, error) {
	iv, err := hex.DecodeString(message.Iv)
	if err != nil {
		return "", err
	}

	msg, err := hex.DecodeString(message.Data)
	if err != nil {
		return "", err
	}

	checksum, err := hex.DecodeString(message.Hmac)
	if err != nil {
		return "", err
	}

	mac := hmac.New(sha256.New, key)
	mac.Write(msg)
	mac.Write(iv)
	if !hmac.Equal(mac.Sum(nil), checksum) {
		return "", errors.New("Hmac mismatch")
	}

	decrypted, err := decrypt(msg, key, iv)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

func decrypt(msg []byte, key []byte, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	var decrypted = make([]byte, len(msg))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(decrypted, []byte(msg))
	padSize := decrypted[len(decrypted)-1]
	woPadding := decrypted[:len(decrypted)-int(padSize)]
	return woPadding, nil
}

func encrypt(msg []byte, key []byte) ([]byte, []byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	blockSize := 16
	padLen := blockSize - len(msg)%blockSize
	padding := bytes.Repeat([]byte{byte(padLen)}, padLen)
	msg = append(msg, padding...)

	iv := make([]byte, 16)
	_, err = rand.Read(iv)
	if err != nil {
		return nil, nil, err
	}
	var encrypted = make([]byte, len(msg))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encrypted, msg)
	return encrypted, iv, nil
}

func confirm() error {
	prompt := promptui.Prompt{
		Label:     "Confirm",
		IsConfirm: true,
	}

	_, err := prompt.Run()
	return err
}
