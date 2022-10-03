package backend

import (
	"github.com/k0kubun/pp/v3"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestValidHexCheck(t *testing.T) {
	assert := assert.New(t)
	hexList := map[string]interface{}{
		"23dff":   true,
		"zzzz":    false,
		"kkkkkkk": false,
		"fffff":   true,
		"0xfffff": true,
	}
	for hexValue, expected := range hexList {
		res := IsValidHexString(hexValue)
		pp.Printf("res = %v / val = %v expected = %v \n", res, hexValue, expected)
		assert.Equal(res, expected)
	}
}

func TestValidNumber(t *testing.T) {
	assert := assert.New(t)
	res := ValidNumber("23232")
	assert.Equal(res, big.NewInt(23232))
}

func TestIsInt(t *testing.T) {
	assert := assert.New(t)
	res := IsInt("23232")
	assert.Equal(res, true)

	res = IsInt("sdsdsd")
	assert.Equal(res, false)
}

func TestToJsonString(t *testing.T) {
	data := map[string]interface{}{
		"id":      2848,
		"jsonrpc": "2.0",
		"method":  "icx_sendTransaction",
	}
	res := ToJsonString(data)
	expected := "{\"id\":2848,\"jsonrpc\":\"2.0\",\"method\":\"icx_sendTransaction\"}"
	assert.Equal(t, res, expected)
}

func TestPrintType(t *testing.T) {
	res := PrintType(2132323)
	assert.Equal(t, res, "int")

	res = PrintType(1.222)
	assert.Equal(t, res, "float64")
}

func TestIsValidIconAddress(t *testing.T) {

	hexList := map[string]interface{}{
		"hx643971796de7a7a74c631bb4f0794995670f3c34": true,
		"43971796de7a7a74c631bb4f0794995670f3c34":    false,
	}
	for hexValue, expected := range hexList {
		res := IsValidIconAddress(hexValue)
		pp.Printf("res = %v / val = %v expected = %v \n", res, hexValue, expected)
		assert.Equal(t, res, expected)
	}

}
