package base58

import (
	"egorkurito/TokenCoin/util"
	"github.com/mr-tron/base58"
)

// Encode base58 encode will give a wallet address that is more readable.
// ToDo: Make more safety
func Encode(input []byte) []byte {
	return []byte(base58.Encode(input))
}

func Decode(input []byte) []byte {
	decode, err := base58.Decode(string(input[:]))
	if err != nil {
		util.LogErrHandle(err)
	}

	return decode
}
