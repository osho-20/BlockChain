package main

import (
	"flag"

	server "server.go"
)

func main(){
	port := flag.Uint("port",5000,"TCP Port Number for BlockChain Server")
	flag.Parse()
	app:=server.BCServer(uint16 (*port))
	app.Run()
}