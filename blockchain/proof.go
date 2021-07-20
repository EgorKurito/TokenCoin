package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/big"
)

const targetBits = 16

type ProofOfWork struct {
	Target *big.Int
}

// NewProofOfWork builds and returns a ProofOfWork
func NewProofOfWork() *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{target}

	return pow
}

// Run performs a proof-of-work
func (pow *ProofOfWork) Run(prevBlockHash, TXsHash []byte) (int, []byte) {
	var intHash big.Int
	var hash [32]byte

	nonce := 0

	for nonce < math.MaxInt64 {
		hash = pow.initDataHash(prevBlockHash, TXsHash, nonce)

		fmt.Printf("\r%x", hash)
		intHash.SetBytes(hash[:])

		if intHash.Cmp(pow.Target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Println("\n")

	return nonce, hash[:]
}

func (pow *ProofOfWork) initDataHash(prevBlockHash, TXsHash []byte, nonce int) [32]byte {
	data := bytes.Join(
		[][]byte{
			prevBlockHash,
			TXsHash,
			IntToHex(int64(nonce)),
			IntToHex(int64(targetBits)),
		},
		[]byte{},
	)

	return sha256.Sum256(data)
}

// Validate validates block's Pow
func (pow *ProofOfWork) Validate(prevBlockHash, TXsHash []byte, nonce int) bool {
	var intHash big.Int

	hash := pow.initDataHash(prevBlockHash, TXsHash, nonce)

	intHash.SetBytes(hash[:])

	return intHash.Cmp(pow.Target) == -1
}

// IntToHex converts an int64 to a byte array
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}
