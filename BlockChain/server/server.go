package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	block "BLOCKCHAIN.go"
	wallet "Wallet.go"
	utils "utils.go"
)

var cache map[string]*block.BlockChain = make(map[string] *block.BlockChain)

type BlockChainServer struct{
	port uint16

}
func BCServer(port uint16) *BlockChainServer{
	return &BlockChainServer{port}
}

func(server* BlockChainServer) GetBlockChain() *block.BlockChain{
	bc,ok:= cache["blockchain"]
	if(!ok){
		miner:=wallet.NewWallet()
		bc = block.NewBlockChain(miner.Address(),server.Port())
		cache["blockchain"]=bc
		// log.Printf("private key %s",miner.PrivateKeyStr())
		// log.Printf("public key %s",miner.PublicKeyStr())
		// log.Printf("Addres %s",miner.Address())

	}
	return bc
}


func (server *BlockChainServer) Port() uint16{
	return server.port
}

func (server *BlockChainServer) BChain(w http.ResponseWriter, req *http.Request){
	switch req.Method{
	case http.MethodGet:
		w.Header().Add("Content-Type","application/json")
		bc:=server.GetBlockChain()
		m,_:=bc.MarshalJSON()
		io.WriteString(w,string(m[:]))
	default:
	}
}

func (server *BlockChainServer) Transaction(rep http.ResponseWriter,req*http.Request){
	switch req.Method{
	case http.MethodGet:
		rep.Header().Add("Content-Type","appliaction/json")
		bc:=server.GetBlockChain()
		transact:=bc.TransactionPool()
		m,_:=json.Marshal(struct{
			Transactions []*block.Transaction `json:"transactions"`
			Length int `json:"length"`
		}{
			Transactions: transact,
			Length: len(transact),
		})
		io.WriteString(rep,string(m[:]))
	case http.MethodPost:
		decode:= json.NewDecoder(req.Body)
		var transact block.TransactionReq
		err:=decode.Decode(&transact)
		if(err!=nil){
			log.Printf("Error: %v\n",err)
			io.WriteString(rep,"Fail.")
			return;
		}
		if(!transact.Valid()){
			log.Println("Error: missing field(s)")
			io.WriteString(rep,"Fail.")
			return
		}
		publickey:=utils.PublicKeyfromString(*transact.SenderPublicKey)
		sign:=utils.SignatureFromString(*transact.Signature)
		bc:=server.GetBlockChain()
		isCreated:=bc.CreateTransaction(*transact.SenderBlockChainAddress,*transact.ReceiverBlockChainAddress,*transact.Value,publickey,sign)
		rep.Header().Add("Content-Type","application/json")
		if !isCreated{
			fmt.Println("Not Created.")
			rep.WriteHeader(http.StatusBadRequest)
			io.WriteString(rep,"Success")

		}else{
			fmt.Println("Created")
			rep.WriteHeader(http.StatusCreated)
			io.WriteString(rep,"Fail.")
		}
	default:
		log.Println("Error: Invalid HTTP Method")
		rep.WriteHeader(http.StatusBadRequest)
	}
}

func (server * BlockChainServer) Mine(w http.ResponseWriter,req *http.Request){
	switch req.Method{
	case http.MethodGet:
		bc:=server.GetBlockChain()
		isMined:=bc.Mining()
		var m string
		if !isMined{
			w.WriteHeader(http.StatusBadRequest)
			m = "Fail."
		}else{
			m = "Success"
		}
		w.Header().Add("ContentType","application/json")
		io.WriteString(w,m)
	default:
		log.Println("Error: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (server * BlockChainServer) StartMine(w http.ResponseWriter,req *http.Request){
	switch req.Method{
	case http.MethodGet:
		bc:=server.GetBlockChain()
		bc.StartMine()
		var m string = "Success"
		w.Header().Add("ContentType","application/json")
		io.WriteString(w,m)
	default:
		log.Println("Error: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (server * BlockChainServer) Amount(w http.ResponseWriter,req *http.Request){
	switch req.Method{
	case http.MethodGet:
		blockchainAddress:=req.URL.Query().Get("blockchain_address")
		amount:=server.GetBlockChain().TotalAmount(blockchainAddress)
		ar := &block.AmountResponse{amount}
		m,_:=ar.MarshalJSON()
		w.Header().Add("ContentType","application/json")
		io.WriteString(w,string(m[:]))
	default:
		log.Println("Error: 2 Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (server *BlockChainServer) Consensus (resp http.ResponseWriter,req *http.Request){
	switch req.Method{
	case http.MethodPut:
		bc:=server.GetBlockChain()
		replaced:=bc.ResolveConflicts()
		resp.Header().Add("ContentType","application/json")
		if replaced {
			io.WriteString(resp,"Success.")
		}else{
			io.WriteString(resp,"Fail.")
		}
	default:
		log.Println("Error: 2 Invalid HTTP Method")
		resp.WriteHeader(http.StatusBadRequest)
	}

}


func (server *BlockChainServer) Run(){
	server.GetBlockChain().Run()
	http.HandleFunc("/server",server.BChain)
	http.HandleFunc("/transactions",server.Transaction)
	http.HandleFunc("/mine",server.Mine)
	http.HandleFunc("/",server.StartMine)
	http.HandleFunc("/amount",server.Amount)
	http.HandleFunc("/consesus",server.Consensus)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(server.Port())),nil))
}


