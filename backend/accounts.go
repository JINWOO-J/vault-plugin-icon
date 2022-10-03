// Copyright © 2022 Jinwoo
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package backend

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/k0kubun/pp/v3"
	"regexp"
)

func paths(b *backend) []*framework.Path {
	return []*framework.Path{
		pathCreateAndList(b),
		pathReadAndDelete(b),
		pathSign(b),
		pathSignAuth(b),
		pathParamSign(b),
		pathExport(b),
	}
}

func (b *backend) listAccounts(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	vals, err := req.Storage.List(ctx, "accounts/")
	if err != nil {
		b.Logger().Error("Failed to retrieve the list of accounts", "error", err)
		return nil, err
	}
	b.Logger().Info("Retrieve the list of accounts", "error", err)
	return logical.ListResponse(vals), nil
}

func (b *backend) createAccount(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	var nameInput = ""
	var keyInput = ""
	var privateKey *PrivateKey
	var publicKey *PublicKey
	var err error

	nameInput = data.Get("name").(string)
	keyInput = data.Get("privateKey").(string)

	if keyInput != "" {
		re := regexp.MustCompile("[0-9a-fA-F]{64}$")
		key := re.FindString(keyInput)

		if key == "" {
			b.Logger().Error("Input private key did not parse successfully", "privateKey", keyInput)
			return nil, fmt.Errorf("privateKey must be a 32-byte hexidecimal string - input: %d-bytes", len(keyInput))
		}
		privkey, _ := hex.DecodeString(key)
		privateKey, err = ParsePrivateKey(privkey)

		if err != nil {
			b.Logger().Error("Error reconstructing private key from input hex", "error", err)
			return nil, fmt.Errorf("Error reconstructing private key from input hex")
		}
		publicKey = privateKey.PublicKey()
		b.Logger().Info("Load private key", "address", publicKey.Address(), "publicKey", publicKey.String())

	} else {
		privateKey, publicKey = GenerateKey()
		b.Logger().Info("Generate new private key", "address", publicKey.Address(), "publicKey", publicKey.String())
	}

	account, err := b.retrieveAccount(ctx, req, publicKey.Address())
	if account != nil && account.Address == publicKey.Address() && account.AliasName == nameInput {
		b.Logger().Info("Already key", "name", nameInput, "address", publicKey.Address(), "publicKey", publicKey.String())
	} else {
		accountPath := fmt.Sprintf("accounts/%s", publicKey.Address())
		accountJSON := &Account{
			Address:    publicKey.Address(),
			PrivateKey: privateKey.String(),
			PublicKey:  publicKey.String(),
			AliasName:  nameInput,
		}
		entry, _ := logical.StorageEntryJSON(accountPath, accountJSON)
		err = req.Storage.Put(ctx, entry)
		if err != nil {
			b.Logger().Error("[ERROR] Failed to save the new account to storage", "error", err)
			return nil, err
		} else {
			b.Logger().Info("[OK] Save the new account", "Address", publicKey.Address())
		}
	}
	return &logical.Response{
		Data: map[string]interface{}{
			"address":    publicKey.Address(),
			"alias_name": nameInput,
		},
	}, nil
}

func (b *backend) readAccount(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	address := data.Get("name").(string)
	b.Logger().Info("[read] Retrieving account for address", "address", address)
	account, err := b.retrieveAccount(ctx, req, address)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, fmt.Errorf("[read] Account does not exist")
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"address":    account.Address,
			"alias_name": account.AliasName,
		},
	}, nil
}

func (b *backend) exportAccount(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	address := data.Get("name").(string)
	b.Logger().Info("[EXPORT][TRY] Retrieving account for address", "address", address)
	account, err := b.retrieveAccount(ctx, req, address)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, fmt.Errorf("[EXPORT][FAIL] Account does not exist - %s", address)
	}

	b.Logger().Info("[EXPORT][OK] ", "address", account.Address, ", alias_name", account.AliasName)

	return &logical.Response{
		Data: map[string]interface{}{
			"address":    account.Address,
			"alias_name": account.AliasName,
			//"privateKey": account.PrivateKey,
		},
	}, nil
}

func (b *backend) deleteAccount(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	address := data.Get("name").(string)
	account, err := b.retrieveAccount(ctx, req, address)
	if err != nil {
		b.Logger().Error("[delete] Failed to retrieve the account by address", "address", address, "error", err)
		return nil, err
	}
	if account == nil {
		return nil, nil
	}
	if err := req.Storage.Delete(ctx, fmt.Sprintf("accounts/%s", account.Address)); err != nil {
		b.Logger().Error("[delete] Failed to delete the account from storage", "address", address, "error", err)
		return nil, err
	}
	return nil, nil
}

func (b *backend) retrieveAccount(ctx context.Context, req *logical.Request, address string) (*Account, error) {
	var path string
	matched, err := regexp.MatchString("^(hx)?[0-9a-fA-F]{40}$", address)
	if !matched || err != nil {
		b.Logger().Error("Failed to retrieve the account, malformatted account address", "address", address, "error", err)
		return nil, fmt.Errorf("Failed to retrieve the account, malformatted account address")
	} else {
		if address[:2] != "hx" {
			address = "hx" + address
		}
		path = fmt.Sprintf("accounts/%s", address)
		entry, err := req.Storage.Get(ctx, path)
		if err != nil {
			b.Logger().Error("Failed to retrieve the account by address", "path", path, "error", err)
			return nil, err
		}
		if entry == nil {
			// could not find the corresponding key for the address
			return nil, nil
		}
		var account Account
		_ = entry.DecodeJSON(&account)
		return &account, nil
	}
}

func (b *backend) signTx(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {

	b.Logger().Info("Start signTx")
	var txHash []byte
	serializeText := data.Get("serialize").(string)
	params := data.Get("params").(map[string]interface{})
	params["from"] = data.Get("name")
	data.Raw["params"] = params
	from := data.Get("name").(string)
	delete(data.Raw, "name")
	pp.Print(data.Raw)
	dataInput := data.Get("data").(string)
	if dataInput == "" {
		dataInput = data.Get("input").(string)
	}
	if len(dataInput) > 2 && dataInput[0:2] != "0x" {
		dataInput = "0x" + dataInput
	}
	if serializeText != "" {
		pp.Print("[INPUT] serializeText")
		pp.Print(serializeText)
		txHash = SHA3Sum256([]byte(serializeText))
	} else {
		pp.Print("[INPUT] JSON")

		fields, _ := transactionFields[3]
		res, _ := SerializeMap(params, fields.inclusion, fields.exclusion)
		//res = append([]byte("icx_sendTransaction."), res...)
		res = append(transactionSaltBytes, res...)
		//txHash = SHA3Sum256([]byte(BytesToString(res)))
		txHash = SHA3Sum256(res)
		pp.Print(BytesToString(res))
		serializeText = BytesToString(res)
		b.Logger().Info("Serialized text", "serialize", BytesToString(res))
	}

	account, err := b.retrieveAccount(ctx, req, from)
	fmt.Print("\nGet account \n")
	pp.Print(account)

	if err != nil {
		b.Logger().Error("Failed to retrieve the signing account", "address", from, "error", err)
		return nil, fmt.Errorf("Error retrieving signing account %s", from)
	}
	if account == nil {
		return nil, fmt.Errorf("Signing account %s does not exist", from)
	}

	amount := ValidNumber(data.Get("value").(string))
	if amount == nil {
		b.Logger().Error("Invalid amount for the 'value' field", "value", data.Get("value").(string))
		return nil, fmt.Errorf("Invalid amount for the 'value' field")
	}

	if err != nil {
		b.Logger().Error("===>> Error reconstructing private key from retrieved hex", "error", err)
		return nil, fmt.Errorf("Error reconstructing private key from retrieved hex")
	}

	//var nonce uint64
	//nonce = nonceIn.Uint64()
	//txHash = SHA3Sum256([]byte("icx_sendTransaction.fee.0x2386f26fc10000.from.hx57b8365292c115d3b72d948272cc4d788fa91f64.timestamp.1538976759263551.to.hx57b8365292c115d3b72d948272cc4d788fa91f64.value.0xde0b6b3a7640000"))
	privateKey, _ := ParsePrivateKeyFromString(account.PrivateKey)
	signedTx, err := NewSignature(txHash, privateKey)
	pp.Printf("\n\n account.PrivateKey: %v \n", account.PrivateKey)
	pp.Printf("\n\n signedTx: %v \n", signedTx.String())

	//signature, _ := signedTx.SerializeRSV()
	b64_sig, _ := signedTx.EncodeBase64()
	//b64_sig, _ := signedTx.MarshalJSON()

	b.Logger().Info("Account Address", "address", account.Address)
	b.Logger().Info("Signed Transaction", "signedTx", signedTx.String())
	b.Logger().Info("Signed Transaction based encoded 64", "signedTx_b64", b64_sig)

	//publicKey := privateKey.PublicKey()
	VerifySign := signedTx.Verify(txHash, privateKey.PublicKey())

	b.Logger().Info("Verify Transaction", "VerifySign", VerifySign)

	if err != nil {
		b.Logger().Error("Failed to sign the transaction object", "error", err)
		return nil, err
	}
	params = data.Get("params").(map[string]interface{})
	params["signature"] = b64_sig
	data.Raw["params"] = params
	b.Logger().Info("Payload", "payload", pp.Sprintf(ToJsonString(data.Raw)))
	return &logical.Response{
		Data: map[string]interface{}{
			"transaction_hash": "0x" + hex.EncodeToString(txHash),
			"signature":        b64_sig,
			"serializeText":    serializeText,
		},
	}, nil
}

func (b *backend) signAuth(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {

	b.Logger().Info(">>>> Start signAuth")
	walletAddress := data.Get("walletAddress").(string)
	time := data.Get("time")
	if IsValidIconAddress(walletAddress) == false {
		return nil, fmt.Errorf("invalid 'walletAddress' value=%s, len=%d", walletAddress, len(walletAddress))
	}
	if time == 0 {
		return nil, fmt.Errorf("required invalid 'time' value : %d", time)
	}
	b.Logger().Info("Params", "walletAddress", walletAddress, "time", time)

	account, err := b.retrieveAccount(ctx, req, walletAddress)
	if err != nil {
		b.Logger().Error("Failed to retrieve the signing account", "walletAddress", walletAddress, "error", err)
		return nil, fmt.Errorf("error retrieving signing account %s", walletAddress)
	}
	if account == nil {
		return nil, fmt.Errorf("signing account %s does not exist", walletAddress)
	}

	requestSignText := fmt.Sprintf("%s%d", walletAddress, time)
	signature, err := SignFromPrivateKey(account.PrivateKey, []byte(requestSignText))

	if err != nil {
		return nil, fmt.Errorf("signing error, address=%s, err=%v", walletAddress, err)
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"walletAddress": walletAddress,
			"time":          time,
			"signature":     signature,
		},
	}, nil

}

func (b *backend) signTransaction(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {

	b.Logger().Info(">>>> Start signTransaction")
	var txHash []byte
	var serializeByte []byte

	serializeText := data.Get("serialize").(string)
	from := data.Get("from").(string)

	if IsValidIconAddress(from) == false {
		return nil, fmt.Errorf("Invalid 'address' value=%s, len=%d", from, len(from))
	}
	b.Logger().Info("data.Raw", fmt.Sprintf("%v", data.Raw))

	if serializeText != "" {
		serializeByte = []byte(serializeText)
	} else {
		fields, _ := transactionFields[3]
		res, err := SerializeMap(data.Raw, fields.inclusion, fields.exclusion)
		if err != nil {
			b.Logger().Error("Serialize Error", "err", err)
			return nil, fmt.Errorf("serialize error")
		}
		serializeByte = append(transactionSaltBytes, res...)
	}

	account, err := b.retrieveAccount(ctx, req, from)
	if err != nil {
		b.Logger().Error("Failed to retrieve the signing account", "address", from, "error", err)
		return nil, fmt.Errorf("Error retrieving signing account %s", from)
	}
	if account == nil {
		return nil, fmt.Errorf("Signing account %s does not exist", from)
	}

	timestamp := data.Get("timestamp")
	if timestamp == "" {
		b.Logger().Error("Invalid timestamp field", "timestamp", data.Get("timestamp").(string))
		return nil, fmt.Errorf("Invalid timestamp")
	}

	amount := ValidNumber(data.Get("value").(string))
	if amount == nil {
		b.Logger().Error("Invalid amount for the 'value' field", "value", data.Get("value").(string))
		return nil, fmt.Errorf("Invalid amount for the 'value' field")
	}

	toAddr := data.Get("to").(string)
	if toAddr == "" {
		b.Logger().Error("Invalid 'to' value, required 'to' address")
		return nil, fmt.Errorf("Invalid 'to' value, required 'to' address")
	}

	if len(toAddr) != 42 {
		b.Logger().Error("Invalid to address, length not 42", len(toAddr))
		return nil, fmt.Errorf("Invalid 'to' value 42 != %d", len(toAddr))
	}

	stepLimit := data.Get("stepLimit").(string)
	if stepLimit == "" {
		b.Logger().Error("Invalid stepLimit")
		return nil, fmt.Errorf("Invalid stepLimit")
	}

	if err != nil {
		b.Logger().Error("===>> Error reconstructing private key from retrieved hex", "error", err)
		return nil, fmt.Errorf("Error reconstructing private key from retrieved hex")
	}

	b64Signature, err := SignFromPrivateKey(account.PrivateKey, serializeByte)

	if err != nil {
		return nil, fmt.Errorf("signing error, address=%s, err=%v", account.PrivateKey, err)
	}

	b.Logger().Info("Account Address", "address", account.Address)
	b.Logger().Info("Signed Transaction based encoded 64", "signedTx_b64", b64Signature)

	if err != nil {
		b.Logger().Error("Failed to sign the transaction object", "error", err)
		return nil, err
	}

	data.Raw["signature"] = b64Signature

	b.Logger().Info("Payload", "payload", ToJsonString(data.Raw))
	return &logical.Response{
		Data: map[string]interface{}{
			"txHash":        "0x" + hex.EncodeToString(txHash),
			"signature":     b64Signature,
			"serialize":     BytesToString(serializeByte),
			"account":       account.Address,
			"signed_params": data.Raw,
		},
	}, nil
}
