package base58

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncode(t *testing.T) {
	nameString := "testbase58"
	base58String := "7YHPeiWKwxqNaT"

	encodeResult := string(Encode([]byte(nameString)))

	assert.Equal(t, base58String, encodeResult)
}

func TestDecode(t *testing.T) {
	nameString := "testbase58"
	base58String := "7YHPeiWKwxqNaT"

	decodeResult := string(Decode([]byte(base58String)))

	assert.Equal(t, nameString, decodeResult)
}
