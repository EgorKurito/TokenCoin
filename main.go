package main

import (
	"egorkurito/myBlockChain/blockchain"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
)

type CommandLine struct {
	blockchain *blockchain.BlockChain
}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage: ")
	fmt.Println(" add -block <BLOCK_DATA> - add a block to the chain")
	fmt.Println(" print - prints the blocks in the chain")
}

func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) addBlock(data string) {
	cli.blockchain.AddBlock(data)
	fmt.Println("Added Block!")
}

func (cli *CommandLine) printChain() {
	iterator := cli.blockchain.Iterator()

	for {
		block := iterator.Next()
		fmt.Printf("Previous hash: %x\n", block.PrevHash)
		fmt.Printf("data: %s\n", block.Data)
		fmt.Printf("hash: %x\n\n", block.Hash)

		pow := blockchain.NewProofOfWork(block)
		fmt.Printf("Pow: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *CommandLine) run() {
	cli.validateArgs()

	addBlockCmd := flag.NewFlagSet("add", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
	addBlockData := addBlockCmd.String("block", "", "Block data")

	switch os.Args[1] {
	case "add":
		if err := addBlockCmd.Parse(os.Args[2:]); err != nil {
			blockchain.LogErrHandle(err)
		}
	case "print":
		if err := printChainCmd.Parse(os.Args[2:]); err != nil {
			blockchain.LogErrHandle(err)
		}
	default:
		cli.printUsage()
		runtime.Goexit()
	}

	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			runtime.Goexit()
		}
		cli.addBlock(*addBlockData)
	}
	if printChainCmd.Parsed() {
		cli.printChain()
	}
}

func main() {
	defer os.Exit(0)

	chain := blockchain.InitBlockChain()
	defer chain.Database.Close()

	cli := CommandLine{chain}

	cli.run()
}
