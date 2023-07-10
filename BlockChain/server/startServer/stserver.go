package main

import (
	"flag"
	"fmt"
	"os"

	server "server.go"
)

func main(){
	port := flag.Uint("port",5000,"TCP Port Number for BlockChain Server")
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
	app:=server.BCServer(uint16 (*port))
	app.Run()
}