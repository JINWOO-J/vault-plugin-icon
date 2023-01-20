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
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/helper/logging"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/stretchr/testify/assert"

	"github.com/k0kubun/pp/v3"
)

func getBackend(t *testing.T) (logical.Backend, logical.Storage) {
	config := &logical.BackendConfig{
		Logger:      logging.NewVaultLogger(log.Trace),
		System:      &logical.StaticSystemView{},
		StorageView: &logical.InmemStorage{},
		BackendUUID: "test",
	}

	b, err := Factory(context.Background(), config)
	if err != nil {
		t.Fatalf("unable to create backend: %v", err)
	}
	// Wait for the upgrade to finish
	time.Sleep(time.Second)
	return b, config.StorageView
}

type StorageMock struct {
	switches []int
}

func (s StorageMock) List(c context.Context, path string) ([]string, error) {
	if s.switches[0] == 1 {
		return []string{"key1", "key2"}, nil
	} else {
		return nil, errors.New("StorageMock for List")
	}
}
func (s StorageMock) Get(c context.Context, path string) (*logical.StorageEntry, error) {
	if s.switches[1] == 2 {
		var entry logical.StorageEntry
		return &entry, nil
	} else if s.switches[1] == 1 {
		return nil, nil
	} else {
		return nil, errors.New("StorageMock for Get")
	}
}
func (s StorageMock) Put(c context.Context, se *logical.StorageEntry) error {
	return errors.New("StorageMock for Put")
}
func (s StorageMock) Delete(c context.Context, path string) error {
	return errors.New("StorageMock for Delete")
}

func newStorageMock() StorageMock {
	var sm StorageMock
	sm.switches = []int{0, 0, 0, 0}
	return sm
}

func TestAccounts(t *testing.T) {
	assert := assert.New(t)
	b, _ := getBackend(t)
	// create key1
	req := logical.TestRequest(t, logical.UpdateOperation, "accounts")
	address1AliasName := RandomString(10)
	req.Data = map[string]interface{}{
		"name": address1AliasName,
	}
	storage := req.Storage
	//spew.Dump("=== storage", storage)
	res, err := b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	pp.Print("res.Data : ", res.Data)
	address1 := res.Data["address"].(string)
	aliasName1 := res.Data["alias_name"].(string)

	assert.Equal(len(address1), 42)
	assert.Equal(aliasName1, address1AliasName)

	fmt.Printf("Generate Account Address = %s \n", address1)

	address1Expected := &logical.Response{
		Data: map[string]interface{}{
			"address":    address1,
			"alias_name": address1AliasName,
		},
	}
	req = logical.TestRequest(t, logical.ReadOperation, "accounts/"+address1)
	req.Storage = storage
	address1Resp, err := b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !reflect.DeepEqual(address1Resp, address1Expected) {
		t.Fatalf("bad response.\n\nexpected: %#v\n\nGot: %#v", address1Expected, address1Resp)
	}

	req = logical.TestRequest(t, logical.CreateOperation, "accounts/"+address1+"/sign")
	req.Storage = storage
	data := map[string]interface{}{
		//"serialize": "icx_sendTransaction.id.1234.jsonrpc.2\\.0.method.icx_sendTransaction.params.{from.hx1bb2825a74ebe30239e669330694b10ded650bbd.nid.0x53.nonce.0x64.stepLimit.0xf4240.timestamp.0x18281f8fe61.to.hxa067296997056e507ac2296573472f3c750d8b62.value.0x16345785d8a0000.version.0x3}",
		"id":      2848,
		"jsonrpc": "2.0",
		"method":  "icx_sendTransaction",
		"params": map[string]interface{}{
			"from":      address1,
			"to":        "hxc1d72af5b89ea6594a7e17ca7a804d52d2474462",
			"stepLimit": "0x4a817c800",
			"value":     "0x2386f26fc10000",
		},
	}
	req.Data = data
	resp, err := b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	signedTx := resp.Data["signature"].(string)
	pp.Print(">>> [signedTx]", signedTx)
	//signatureBytes, err := hexutil.Decode(signedTx)
	//var signatureBytes Signature
	//signatureBytes, err = base64.StdEncoding.DecodeString(signedTx)
	signatureBytes := toSignatureBS(signedTx)
	//signatureBytes.RecoverPublicKey()
	if err != nil {
		t.Fatal("error:", err)
	}
	pp.Print("signatureBytes", signatureBytes)
	// delete key by name
	pp.Print("Delete key by address: ", address1, "\n\n")
	req = logical.TestRequest(t, logical.DeleteOperation, "accounts/"+address1)
	req.Storage = storage
	if _, err := b.HandleRequest(context.Background(), req); err != nil {
		t.Fatalf("delete key by name err: %v", err)
	}

	expected := &logical.Response{
		Data: map[string]interface{}{},
	}

	pp.Print("\n\n Delete : ", expected)

	// check the deleted address
	req = logical.TestRequest(t, logical.ReadOperation, "accounts/"+address1)
	req.Storage = storage
	address1Resp, err = b.HandleRequest(context.Background(), req)
	//assert.Equal(err.Error(), "[read] Account does not exist")
	assert.Equal(err.Error(), fmt.Sprintf("[READ][FAIL] Account does not exist - %s", address1))

	req = logical.TestRequest(t, logical.ListOperation, "accounts")
	req.Storage = storage
	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if !reflect.DeepEqual(resp, expected) {
		t.Fatalf("bad response.\n\nexpected: %#v\n\nGot: %#v", expected, resp)
	}
}

func TestCreateAccountsByPrivateKey(t *testing.T) {
	assert := assert.New(t)
	b, _ := getBackend(t)
	req := logical.TestRequest(t, logical.UpdateOperation, "accounts")
	data := map[string]interface{}{
		"privateKey": "d25d2854b9d7d73f3119017e0159d5ee9500be8da5b93e2ddd812e20c2094c32",
	}
	req.Data = data
	res, err := b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("TestCreateAccountsByPrivateKey  err: %v", err)
	}
	address3 := res.Data["address"].(string)
	assert.Equal("hx10b1c9aee600db8eea807c77b5040ab294cdcf99", address3)
}

func TestCreateAccountsByPrivateKeyWith0x(t *testing.T) {
	assert := assert.New(t)
	b, _ := getBackend(t)
	req := logical.TestRequest(t, logical.UpdateOperation, "accounts")
	data := map[string]interface{}{
		"privateKey": "0xd25d2854b9d7d73f3119017e0159d5ee9500be8da5b93e2ddd812e20c2094c32",
	}
	req.Data = data
	res, err := b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("TestCreateAccountsByPrivateKeyWith0x  err: %v", err)
	}
	address3 := res.Data["address"].(string)
	assert.Equal("hx10b1c9aee600db8eea807c77b5040ab294cdcf99", address3)
}

func TestCreateAccountsOK(t *testing.T) {
	assert := assert.New(t)

	b, _ := getBackend(t)
	req := logical.TestRequest(t, logical.UpdateOperation, "accounts")
	data := map[string]interface{}{
		// use N for the secp256k1 curve to trigger an error
		"privateKey": "0xec85999367d32fbbe02dd600a2a44550b95274cc67d14375a9f0bce233f13ad2",
	}
	req.Data = data
	res, _ := b.HandleRequest(context.Background(), req)
	address4 := res.Data["address"].(string)
	assert.Equal("hxbe1833529dae2328156cc834223cdc462e4d129d", address4)
}


func createAccountFunc(t *testing.T, b logical.Backend, storage logical.Storage, name string) (logical.Storage, error)  {

	//b, _ := getBackend(t)

	accountReq := logical.TestRequest(t, logical.UpdateOperation, "accounts")
	data := map[string]interface{}{
		"name": name,
	}
	accountReq.Data = data
	accountReq.Storage = storage
	//storage := accountReq.Storage
	_, _ = b.HandleRequest(context.Background(), accountReq)
	return storage, nil
}

func TestListAccountsOK_1(t *testing.T) {
	assert := assert.New(t)
	b, _ := getBackend(t)
	req := logical.TestRequest(t, logical.ListOperation, "accounts")
	sm := req.Storage
	req.Storage = sm
	maxWalletCount := 5
	data := map[string]interface{}{
		"detail": false,
	}
	req.Data = data

	for i := 0; i < maxWalletCount; i++ {
		_, err := createAccountFunc(t, b, sm, fmt.Sprintf("test_wallet_%d" , i))
		if err != nil {
			return
		}
	}
	res, err := b.HandleRequest(context.Background(), req)
	if err != nil{
		pp.Print(err)
	}
	pp.Print(res.Data["keys"])
	assert.Equal(len(res.Data["keys"].([]string)), maxWalletCount )
}

func TestListAccountsDetailOK_1(t *testing.T) {
	assert := assert.New(t)
	b, _ := getBackend(t)
	req := logical.TestRequest(t, logical.ListOperation, "accounts")
	sm := req.Storage
	req.Storage = sm
	maxWalletCount := 5
	data := map[string]interface{}{
		"detail": true,
	}
	req.Data = data

	for i := 0; i < maxWalletCount; i++ {
		_, err := createAccountFunc(t, b, sm, fmt.Sprintf("test_wallet_%d" , i))
		if err != nil {
			return
		}
	}
	res, err := b.HandleRequest(context.Background(), req)
	if err != nil{
		pp.Print(err)
	}
	pp.Print(res.Data["keys"])
	assert.Equal( len(res.Data["keys"].([]map[string]interface{})), maxWalletCount )
}

func TestListAccountsFailure1(t *testing.T) {
	assert := assert.New(t)

	b, _ := getBackend(t)
	req := logical.TestRequest(t, logical.ListOperation, "accounts")
	sm := newStorageMock()
	req.Storage = sm
	_, err := b.HandleRequest(context.Background(), req)

	assert.Equal("StorageMock for List", err.Error())
}

func TestCreateAccountsFailure1(t *testing.T) {
	assert := assert.New(t)

	b, _ := getBackend(t)
	req := logical.TestRequest(t, logical.UpdateOperation, "accounts")
	sm := newStorageMock()
	req.Storage = sm
	_, err := b.HandleRequest(context.Background(), req)

	assert.Equal("StorageMock for Put", err.Error())
}

func TestCreateAccountsFailure2(t *testing.T) {
	assert := assert.New(t)

	b, _ := getBackend(t)
	req := logical.TestRequest(t, logical.UpdateOperation, "accounts")
	data := map[string]interface{}{
		"privateKey": "abc",
	}
	req.Data = data
	sm := newStorageMock()
	req.Storage = sm
	_, err := b.HandleRequest(context.Background(), req)
	assert.Equal("privateKey must be a 32-byte hexidecimal string - input: 3-bytes", err.Error())
}

func TestCreateAccountsFailure3(t *testing.T) {
	assert := assert.New(t)

	b, _ := getBackend(t)
	req := logical.TestRequest(t, logical.UpdateOperation, "accounts")
	data := map[string]interface{}{
		// use N for the secp256k1 curve to trigger an error
		"privateKey": "fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141",
		//"privateKey": "0xec85999367d32fbbe02dd600a2a44550b95274cc67d14375a9f0bce233f13ad2",
	}
	req.Data = data
	sm := newStorageMock()
	req.Storage = sm
	_, err := b.HandleRequest(context.Background(), req)
	assert.Equal("Error reconstructing private key from input hex", err.Error())
}

func TestCreateAccountsFailureInvalidLength(t *testing.T) {
	assert := assert.New(t)

	b, _ := getBackend(t)
	req := logical.TestRequest(t, logical.UpdateOperation, "accounts")
	data := map[string]interface{}{
		// use N for the secp256k1 curve to trigger an error
		"privateKey": "0xec85999367d32fbbe02dd600a2a44550b95274cc67d14375a9f0bce233f13ad2sdsdsdsdsds",
	}
	req.Data = data
	sm := newStorageMock()
	req.Storage = sm
	_, err := b.HandleRequest(context.Background(), req)
	assert.Equal("privateKey must be a 32-byte hexidecimal string - input: 77-bytes", err.Error())
}

func TestExportAccount(t *testing.T) {
	b, _ := getBackend(t)
	req := logical.TestRequest(t, logical.UpdateOperation, "accounts")
	aliasName := "ExportAccount"
	data := map[string]interface{}{
		"name": aliasName,
	}
	req.Data = data
	storage := req.Storage
	res, _ := b.HandleRequest(context.Background(), req)
	address := res.Data["address"].(string)
	req = logical.TestRequest(t, logical.ReadOperation, "export/accounts/"+address)
	req.Storage = storage
	resp, err := b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	assert.Equal(t, resp.Data["address"], address)
	assert.Equal(t, resp.Data["alias_name"], aliasName)
}

func TestSignTransaction(t *testing.T) {
	//assert := assert.New(t)
	b, _ := getBackend(t)
	req := logical.TestRequest(t, logical.UpdateOperation, "accounts")
	data := map[string]interface{}{
		"name": "SignAccount",
	}
	req.Data = data
	storage := req.Storage
	res, _ := b.HandleRequest(context.Background(), req)
	address := res.Data["address"].(string)
	pp.Print(address)

	req = logical.TestRequest(t, logical.CreateOperation, "accounts/"+address+"/sign")
	req.Storage = storage

	txData := map[string]interface{}{
		//"serialize": "icx_sendTransaction.id.1234.jsonrpc.2\\.0.method.icx_sendTransaction.params.{from.hx1bb2825a74ebe30239e669330694b10ded650bbd.nid.0x53.nonce.0x64.stepLimit.0xf4240.timestamp.0x18281f8fe61.to.hxa067296997056e507ac2296573472f3c750d8b62.value.0x16345785d8a0000.version.0x3}",
		"id":      2848,
		"jsonrpc": "2.0",
		"method":  "icx_sendTransaction",
		"params": map[string]interface{}{
			"from":      address,
			"to":        "hxc1d72af5b89ea6594a7e17ca7a804d52d2474462",
			"stepLimit": "0x4a817c800",
			"value":     "0x2386f26fc10000",
		},
	}

	req.Data = txData
	resp, err := b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	pp.Print(resp)
}

func TestSignParamTransaction(t *testing.T) {
	//assert := assert.New(t)
	b, _ := getBackend(t)
	req := logical.TestRequest(t, logical.UpdateOperation, "accounts")
	data := map[string]interface{}{
		//"privateKey": "82445cedfb35eb7390f35d71fe4b589bd102e70796dccebb5a2584cddca15c17",
		"privateKey": "01399e416f938def824ed0785ec90133374ad0d7eb97bf82fa46d5163e951501",
	}
	req.Data = data
	storage := req.Storage
	res, _ := b.HandleRequest(context.Background(), req)
	address := res.Data["address"].(string)
	pp.Print(address)

	req = logical.TestRequest(t, logical.CreateOperation, "accounts/"+address+"/param_sign")
	req.Storage = storage

	txData := map[string]interface{}{
		//"serialize": "icx_sendTransaction.id.1234.jsonrpc.2\\.0.method.icx_sendTransaction.params.{from.hx1bb2825a74ebe30239e669330694b10ded650bbd.nid.0x53.nonce.0x64.stepLimit.0xf4240.timestamp.0x18281f8fe61.to.hxa067296997056e507ac2296573472f3c750d8b62.value.0x16345785d8a0000.version.0x3}",
		"id":      2848,
		"jsonrpc": "2.0",
		"method":  "icx_sendTransaction",
		"params": map[string]interface{}{
			"from":      "sdsdsds",
			"to":        "hxc1d72af5b89ea6594a7e17ca7a804d52d2474462",
			"stepLimit": "0x4a817c800",
			"value":     "0x2386f26fc10000",
			"nid":       "0x53",
			"nonce":     "0x1d",
			"version":   "0x3",
			"timestamp": TimeStampNow(),
		},
	}
	req.Data = txData["params"].(map[string]interface{})
	resp, err := b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	pp.Print(resp)
	sig_len := len(resp.Data["signature"].(string))
	assert.NotEqual(t, sig_len, 0)
}

func TestSignParamTransactionFailure1(t *testing.T) {
	b, _ := getBackend(t)
	req := logical.TestRequest(t, logical.UpdateOperation, "accounts")
	data := map[string]interface{}{
		"privateKey": "82445cedfb35eb7390f35d71fe4b589bd102e70796dccebb5a2584cddca15c17",
	}
	req.Data = data
	storage := req.Storage
	res, _ := b.HandleRequest(context.Background(), req)
	address := res.Data["address"].(string)
	pp.Print(address)

	req = logical.TestRequest(t, logical.CreateOperation, "accounts/"+address+"/param_sign")
	req.Storage = storage

	txData := map[string]interface{}{
		//"serialize": "icx_sendTransaction.id.1234.jsonrpc.2\\.0.method.icx_sendTransaction.params.{from.hx1bb2825a74ebe30239e669330694b10ded650bbd.nid.0x53.nonce.0x64.stepLimit.0xf4240.timestamp.0x18281f8fe61.to.hxa067296997056e507ac2296573472f3c750d8b62.value.0x16345785d8a0000.version.0x3}",
		"id":      2848,
		"jsonrpc": "2.0",
		"method":  "icx_sendTransaction",
		"params": map[string]interface{}{
			"from": "sdsdsds",
			//"to":        "hxc1d72af5b89ea6594a7e17ca7a804d52d2474462",
			"stepLimit": "0x4a817c800",
			"value":     "0x2386f26fc10000",
			"nid":       "0x53",
			"nonce":     "0x1d",
			"version":   "0x3",
			"timestamp": TimeStampNow(),
		},
	}
	req.Data = txData["params"].(map[string]interface{})
	_, err := b.HandleRequest(context.Background(), req)
	assert.Equal(t, err.Error(), "Invalid 'to' value, required 'to' address")
}

func TestSignParamTransactionFailure2(t *testing.T) {
	b, _ := getBackend(t)
	req := logical.TestRequest(t, logical.UpdateOperation, "accounts")
	data := map[string]interface{}{
		"privateKey": "82445cedfb35eb7390f35d71fe4b589bd102e70796dccebb5a2584cddca15c17",
	}
	req.Data = data
	storage := req.Storage
	res, _ := b.HandleRequest(context.Background(), req)
	address := res.Data["address"].(string)
	pp.Print(address)

	req = logical.TestRequest(t, logical.CreateOperation, "accounts/"+address+"/param_sign")
	req.Storage = storage

	txData := map[string]interface{}{
		//"serialize": "icx_sendTransaction.id.1234.jsonrpc.2\\.0.method.icx_sendTransaction.params.{from.hx1bb2825a74ebe30239e669330694b10ded650bbd.nid.0x53.nonce.0x64.stepLimit.0xf4240.timestamp.0x18281f8fe61.to.hxa067296997056e507ac2296573472f3c750d8b62.value.0x16345785d8a0000.version.0x3}",
		"id":      2848,
		"jsonrpc": "2.0",
		"method":  "icx_sendTransaction",
		"params": map[string]interface{}{
			"from": "sdsdsds",
			"to":   "hxc1d72af5b89ea6594a7e17ca7a804d52d2474462",
			//"stepLimit": "0x4a817c800",
			"value":     "0x2386f26fc10000",
			"nid":       "0x53",
			"nonce":     "0x1d",
			"version":   "0x3",
			"timestamp": TimeStampNow(),
		},
	}
	req.Data = txData["params"].(map[string]interface{})
	_, err := b.HandleRequest(context.Background(), req)
	assert.Equal(t, err.Error(), "Invalid stepLimit")
}

func TestSignParamTransactionFailure200(t *testing.T) {
	b, _ := getBackend(t)
	address := "INVALID_ADDRESS"
	req := logical.TestRequest(t, logical.CreateOperation, "accounts/"+address+"/param_sign")
	txData := map[string]interface{}{
		"id":      2848,
		"jsonrpc": "2.0",
		"method":  "icx_sendTransaction",
		"params": map[string]interface{}{
			"to":        "hxc1d72af5b89ea6594a7e17ca7a804d52d2474462",
			"stepLimit": "0x4a817c800",
			"value":     "0x2386f26fc10000",
			"nid":       "0x53",
			"nonce":     "0x1d",
			"version":   "0x3",
			"timestamp": TimeStampNow(),
		},
	}
	req.Data = txData["params"].(map[string]interface{})
	_, err := b.HandleRequest(context.Background(), req)
	assert.Equal(t, err.Error(), fmt.Sprintf("Invalid 'address' value=%s, len=%d", address, len(address)))
}
