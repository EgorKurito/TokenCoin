package main

import (
	"egorkurito/myBlockChain/blockchain"
	"fmt"
)

func main() {
	chain := blockchain.InitBlockChain()

	chain.AddBlock("first block after genesis")
	chain.AddBlock("second block after genesis")
	chain.AddBlock("third block after genesis")

	for _, block := range chain.Blocks {
		fmt.Printf("Previous hash: %x\n", block.PrevHash)
		fmt.Printf("data: %s\n", block.Data)
		fmt.Printf("hash: %x\n\n", block.Hash)
	}
}