package backend

import (
	"github.com/k0kubun/pp/v3"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSignFromPrivateKey(t *testing.T) {
	privateKey := "82445cedfb35eb7390f35d71fe4b589bd102e70796dccebb5a2584cddca15c17"
	res, err := SignFromPrivateKey(privateKey, []byte("ttt"))
	pp.Print(res)
	pp.Print(err)
	assert.Equal(t, err, nil)
}
