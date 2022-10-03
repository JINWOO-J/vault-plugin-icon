package backend

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/haltingstate/secp256k1-go"
	secp256k1_go "github.com/haltingstate/secp256k1-go/secp256k1-go2"
	"strings"
)

const (
	// InvalidAddress intends to prevent empty address_to
	InvalidAddress string = "InvalidAddress"
)

const (
	// PublicKeyLenCompressed is the byte length of a compressed public key
	PublicKeyLenCompressed = 33
	// PublicKeyLenUncompressed is the byte length of an uncompressed public key
	PublicKeyLenUncompressed      = 65
	publicKeyCompressed      byte = 0x2 // y_bit + x coord
	publicKeyUncompressed    byte = 0x4 // x coord + y coord
	AddressIDBytes                = 20
	AddressBytes                  = AddressIDBytes + 1
	// PrivateKeyLen is the byte length of a private key
	PrivateKeyLen = 32
)

const (
	Version2 = 2
	Version3 = 3
)

// Account is an ICON account
type Account struct {
	Address    string `json:"address"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	AliasName  string `json:"alias_name"`
}

type PrivateKey struct {
	bytes []byte // 32-byte
}

type PublicKey struct {
	bytes []byte // 33-byte compressed format to use halting state library efficiently
}

var (
	transactionSaltBytes = []byte("icx_sendTransaction.")
	transactionFields    = map[int]struct {
		inclusion map[string]bool
		exclusion map[string]bool
	}{
		Version2: {
			exclusion: map[string]bool{
				"method":    true,
				"signature": true,
				"tx_hash":   true,
			},
		},
		Version3: {
			exclusion: map[string]bool{
				"signature": true,
				"txHash":    true,
			},
		},
	}
)

// Bytes returns bytes form of private key.
func (key *PrivateKey) Bytes() []byte {
	kb := make([]byte, PrivateKeyLen)
	copy(kb, key.bytes)
	return kb
}

func (key *PrivateKey) String() string {
	return "0x" + hex.EncodeToString(key.bytes)
}

func (key *PublicKey) String() string {
	return "0x" + hex.EncodeToString(key.bytes)
}

func (key *PublicKey) SerializeUncompressed() []byte {
	return secp256k1.UncompressPubkey(key.bytes)
}

func (key *PrivateKey) PublicKey() *PublicKey {
	pkBytes := secp256k1.PubkeyFromSeckey(key.bytes)
	pk, err := ParsePublicKey(pkBytes)
	if err != nil {
		panic(err)
	}
	return pk
}

func (key *PublicKey) Address() string {
	pubKey, _ := ParsePublicKey(key.bytes)
	uncompressed := pubKey.SerializeUncompressed()
	digest := SHA3Sum256(uncompressed[1:])
	address := "hx" + hex.EncodeToString(digest[len(digest)-AddressIDBytes:])
	//fmt.Printf(" pubKey=%s \n address=%s \n", pubKey, address)
	return address
}

func GenerateKey() (*PrivateKey, *PublicKey) {
	pub, priv := secp256k1.GenerateKeyPair()
	privKey := &PrivateKey{priv}
	pubKey, _ := ParsePublicKey(pub)
	//address := PublicKeyExtractAddress(pub)
	return privKey, pubKey
}

func PublicKeyExtractAddress(pub []byte) string {
	pubKey, _ := ParsePublicKey(pub)
	uncompressed := pubKey.SerializeUncompressed()
	digest := SHA3Sum256(uncompressed[1:])
	address := "hx" + hex.EncodeToString(digest[len(digest)-AddressIDBytes:])
	//fmt.Printf(" pubKey=%s \n address=%s \n", pubKey, address)
	return address
}

func uncompToCompPublicKey(uncomp []byte) (comp []byte) {
	comp = make([]byte, PublicKeyLenCompressed)
	// skip to check the validity of uncompressed key
	format := publicKeyCompressed
	if uncomp[64]&0x1 == 0x1 {
		format |= 0x1
	}
	comp[0] = format
	copy(comp[1:], uncomp[1:33])
	return
}

func ParsePrivateKey(b []byte) (*PrivateKey, error) {

	if len(b) != PrivateKeyLen {
		return nil, errors.New("InvalidKeyLength")
	}
	if secp256k1_go.SeckeyIsValid(b) != 1 {
		return nil, errors.New("InvalidSeckey")
	}

	b2 := make([]byte, len(b))

	copy(b2, b)
	return &PrivateKey{b2}, nil
}

// ParsePublicKey parse private key and return private key object.
func ParsePublicKey(pubKey []byte) (*PublicKey, error) {
	switch len(pubKey) {
	case 0:
		return nil, errors.New("public key bytes are empty")
	case PublicKeyLenCompressed:
		return &PublicKey{pubKey}, nil
	case PublicKeyLenUncompressed:
		return &PublicKey{uncompToCompPublicKey(pubKey)}, nil
	default:
		return nil, nil
		// 		, errors.New("wrong format")
	}
}

func ParsePrivateKeyFromString(privateKeyStr string) (*PrivateKey, error) {
	privateKeyStr = strings.TrimLeft(privateKeyStr, "0x")
	privateKey, decodeErr := hex.DecodeString(privateKeyStr)
	if decodeErr != nil {
		return nil, fmt.Errorf("[ERROR] ParsePrivateKeyFromString decode_error %s", decodeErr)
	}
	privateKeyBytes, err := ParsePrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] ParsePrivateKeyFromString %s", err)
	}
	return privateKeyBytes, nil
}

func SignFromPrivateKey(privateKeyStr string, requestSign []byte) (string, error) {
	privateKey, err := ParsePrivateKeyFromString(privateKeyStr)
	if err != nil {
		return "", fmt.Errorf("[ERROR] ParsePrivateKeyFromString %s", err)
	}
	requestSignBytes := SHA3Sum256(requestSign)
	signedTx, err := NewSignature(requestSignBytes, privateKey)
	if err != nil {
		return "", fmt.Errorf("[ERROR] NewSignature error -  %s", err)
	}
	verifySign := signedTx.Verify(requestSignBytes, privateKey.PublicKey())
	if verifySign == false {
		return "", fmt.Errorf("[ERROR] Failed to Verify the Transaction")
	}
	b64Signature, _ := signedTx.EncodeBase64()
	return b64Signature, err
}
