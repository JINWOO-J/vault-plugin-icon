package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/k0kubun/pp/v3"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSerialize(t *testing.T) {
	b, _ := getBackend(t)
	req := logical.TestRequest(t, logical.UpdateOperation, "accounts")
	data := map[string]interface{}{
		"privateKey": "0xec85999367d32fbbe02dd600a2a44550b95274cc67d14375a9f0bce233f13ad2",
	}
	req.Data = data
	storage := req.Storage
	res, _ := b.HandleRequest(context.Background(), req)

	spew.Dump(res)
	fmt.Printf("res.Data : %+v", res)
	address1 := res.Data["address"].(string)
	assert.Equal(t, "hxbe1833529dae2328156cc834223cdc462e4d129d", address1)

	req = logical.TestRequest(t, logical.CreateOperation, "accounts/"+address1+"/sign")
	req.Storage = storage
	data = map[string]interface{}{
		//"serialize": "icx_sendTransaction.id.1234.jsonrpc.2\\.0.method.icx_sendTransaction.params.{from.hx1bb2825a74ebe30239e669330694b10ded650bbd.nid.0x53.nonce.0x64.stepLimit.0xf4240.timestamp.0x18281f8fe61.to.hxa067296997056e507ac2296573472f3c750d8b62.value.0x16345785d8a0000.version.0x3}",
		"id":      2848,
		"jsonrpc": "2.0",
		"method":  "icx_sendTransaction",
		"params": map[string]interface{}{
			"from":      "hx5a05b58a25a1e5ea0f1d5715e1f655dffc1fb30a",
			"to":        "hx32b5704b766c535c34291c0d10ddd5bbd7b6b9fb",
			"stepLimit": "0x4a817c800",
			"value":     "0x38d7ea4c68000",
			"nid":       "0x53",
			"nonce":     "0xa",
			"version":   "0x3",
			"timestamp": "0x5e5d940e41678",
		},
	}
	req.Data = data
	resp, err := b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	signedTx := resp.Data["signature"].(string)

	fmt.Printf(">>> signedTx: %s", signedTx)
	FPrintln(resp)

	signatureBytes, err := DecodeStringToBytes(signedTx)
	FPrintln("signatureBytes: %v", signatureBytes)

	bs, _ := json.Marshal(resp.Data)

	FPrintln(string(bs))
	FPrintln(ToJsonString(resp.Data))
}

func TestSerializeParams(t *testing.T) {

	//assert := assert.New(t)

	b, _ := getBackend(t)
	req := logical.TestRequest(t, logical.UpdateOperation, "accounts")
	data := map[string]interface{}{
		"privateKey": "0xec85999367d32fbbe02dd600a2a44550b95274cc67d14375a9f0bce233f13ad2",
	}
	req.Data = data
	storage := req.Storage
	res, _ := b.HandleRequest(context.Background(), req)
	spew.Dump(res)
	fmt.Printf("res.Data : %+v", res)
	address1 := res.Data["address"].(string)
	assert.Equal(t, "hxbe1833529dae2328156cc834223cdc462e4d129d", address1)

	req = logical.TestRequest(t, logical.CreateOperation, "accounts/"+address1+"/param_sign")
	req.Storage = storage
	data = map[string]interface{}{
		//"serialize": "icx_sendTransaction.id.1234.jsonrpc.2\\.0.method.icx_sendTransaction.params.{from.hx1bb2825a74ebe30239e669330694b10ded650bbd.nid.0x53.nonce.0x64.stepLimit.0xf4240.timestamp.0x18281f8fe61.to.hxa067296997056e507ac2296573472f3c750d8b62.value.0x16345785d8a0000.version.0x3}",
		"from":      "hx5a05b58a25a1e5ea0f1d5715e1f655dffc1fb30a",
		"to":        "hx32b5704b766c535c34291c0d10ddd5bbd7b6b9fb",
		"stepLimit": "0x4a817c800",
		"value":     "0x38d7ea4c68000",
		"nid":       "0x53",
		"nonce":     "0xa",
		"version":   "0x3",
		"timestamp": "0x5e5d940e41678",
	}
	req.Data = data
	resp, err := b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	signedTx := resp.Data["signature"].(string)

	fmt.Print(">>> signedTx ------------ \n\n")
	pp.Print(resp)
	fmt.Print(">>\n ------------ \n\n")

	signatureBytes, err := DecodeStringToBytes(signedTx)
	pp.Printf("signatureBytes: %v", signatureBytes)

	bs, err := json.Marshal(resp.Data)

	pp.Print(string(bs))
	pp.Print(ToJsonString(resp.Data))
	//var tx types.Transaction
	//err = tx.DecodeRLP(rlp.NewStream(bytes.NewReader(signatureBytes), 0))
	//if err != nil {
	//	t.Fatalf("err: %v", err)
	//}
	//v, _, _ := tx.RawSignatureValues()
	//assert.Equal(true, contains([]*big.Int{big.NewInt(27), big.NewInt(28)}, v))
	//
	//sender, _ := types.Sender(types.HomesteadSigner{}, &tx)
	//assert.Equal(address1, strings.ToLower(sender.Hex()))
}
