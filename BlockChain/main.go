package main

import (
	"flag"
	"fmt"
	"os"

	"wallet_server.go"
)

// bc "BLOCKCHAIN.go"
// "Wallet.go"

func main() {
	// walletM := Wallet.NewWallet()
	// walletA := Wallet.NewWallet()
	// walletB := Wallet.NewWallet()
	// transact := Wallet.NewTransaction(walletA.PrivateKey(), walletA.PublicKey(), walletA.Address(), walletB.Address(), 10.0)
	// blockchain := bc.NewBlockChain(walletM.Address())
	// Added := blockchain.AddTransaction(walletA.Address(), walletB.Address(), 10, walletA.PublicKey(), transact.GenerateSignature())
	// fmt.Println("Added : ", Added)
	// blockchain.Mining()
	// blockchain.PrintBlockChain()
	// fmt.Printf("A %.1f\n", blockchain.TotalAmount(walletA.Address()))
	// fmt.Printf("B %.1f\n", blockchain.TotalAmount(walletB.Address()))
	// fmt.Printf("M %.1f\n", blockchain.TotalAmount(walletM.Address()))

	// port := flag.Uint("port",8080,"TCP Port Number for Wallet Server")
	// gateway := flag.String("gateway","http://127.0.0.1:5000","BlockChain Gateway")
	
	// flag.Parse()
	// app:=wallet_server.NewWalletServer(uint16 (*port),*gateway)
	// fmt.Println(app)
	// app.Run()
	port := flag.Uint("port", 8080, "TCP Port Number for Wallet Server")
	gateway := flag.String("gateway", "http://127.0.0.1:5000", "BlockChain Gateway")
	bind := flag.String("bind", "", "Bind address for the server")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "Error: Unexpected positional arguments\n")
		flag.Usage()
		os.Exit(1)
	}

	if *bind != "" {
		// Handle the bind address logic here
		fmt.Printf("Using bind address: %s\n", *bind)
	}

	app := wallet_server.NewWalletServer(uint16(*port), *gateway)
	fmt.Println(app)
	app.Run()
}

