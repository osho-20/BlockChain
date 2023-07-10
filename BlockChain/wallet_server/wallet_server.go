package wallet_server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"path"
	"strconv"

	block "BLOCKCHAIN.go"
	wallet "Wallet.go"
	"utils.go"
)

const tempDir="wallet_server/templates/"

type WalletServer struct{
	port uint16
	gateway string
}

func NewWalletServer(port uint16, gateway string) *WalletServer{
	return &WalletServer{port,gateway}
}

func (server *WalletServer) Port() uint16{
	return server.port
}

func (server *WalletServer) GateWay() string{
	return server.gateway
}

func (server* WalletServer) Index(w http.ResponseWriter,req *http.Request){
	switch req.Method{
		case http.MethodGet:
			t,_:=template.ParseFiles(path.Join(tempDir,"index.html"))
			// fmt.Println(t)
			t.Execute(w,"")
		default:
			log.Printf("Error: Invalid.")
	}
}

func (server *WalletServer) Wallet(w http.ResponseWriter,req *http.Request){
	switch req.Method{
	case http.MethodPost:
		w.Header().Add("Content-Type","application/json")
		myWallet:= wallet.NewWallet()
		m,_:=myWallet.MarshalJSON()
		io.WriteString(w,string(m[:]))
	default:
		w.WriteHeader(http.StatusBadRequest)
		log.Println("Error: Invalid HTTP Method")
	}
}

func (server* WalletServer) Transaction(resp http.ResponseWriter,req* http.Request){
	switch req.Method{
	case http.MethodPost:
		decoder:= json.NewDecoder(req.Body)
		var transact wallet.TransactionReq
		err:=decoder.Decode(&transact)
		if (err!=nil){
			fmt.Printf("Error: %v",err)
			io.WriteString(resp,"Fail.")
			return
		}
		if(!transact.Validate()){
			log.Println("Error: Missing Data.")
			io.WriteString(resp,"Fail.")
			return;
		}
		publicKey:=utils.PublicKeyfromString(*transact.SenderPublicKey)
		privateKey:=utils.PrivateKeyfromString(*transact.SenderPrivateKey,publicKey)
		value,err:=strconv.ParseFloat(*transact.Value,32)
		if(err!=nil){
			log.Println("ERROR: parse error")
			io.WriteString(resp,"Fail.")
			return
		}
		value32:=float32(value)
		resp.Header().Add("ContentType","application/json")
		transaction:=wallet.NewTransaction(privateKey,publicKey,*transact.SenderBlockChainAddress,*transact.ReceiverBlockChainAddress , value32)
		sign:=transaction.GenerateSignature()
		signStr:= sign.Sign()
		bt := &block.TransactionReq{
			SenderBlockChainAddress:    transact.SenderBlockChainAddress,
			ReceiverBlockChainAddress:  transact.ReceiverBlockChainAddress,
			SenderPublicKey:            transact.SenderPublicKey,
			Value:                      &value32,
   			Signature:                  &signStr,
		}
		m,_:=json.Marshal(bt) 
		buf:=bytes.NewBuffer(m)
		fmt.Println(buf);
		res,err1:= http.Post(server.GateWay()+"/transactions","application/json",buf)
		if (err1!=nil){
			fmt.Printf("Error: %v",err)
			io.WriteString(resp,"Fail.")
			return
		}

		fmt.Println(res)
		fmt.Println("resp: ",res.StatusCode)
		if(res.StatusCode==int(201)){
			fmt.Println("success.")
			io.WriteString(resp,"Success")
			return;
		}
		fmt.Println("Fail.")
		io.WriteString(resp,"Fail.")
		return
		// fmt.Println(privateKey)
		// fmt.Println(publicKey)
		// fmt.Println(*transact.SenderBlockChainAddress)
		// fmt.Printf("%0.1f\n",value32)
	default:
		resp.WriteHeader(http.StatusBadRequest)
		log.Println("Error: Invalid http request.")
	}
}

func (server* WalletServer) WalletAmount(resp http.ResponseWriter,req* http.Request){
	switch req.Method{
	case http.MethodGet:
		bca:=req.URL.Query().Get("blockchain_address")
		endpoint:= fmt.Sprintf("%s/amount",server.GateWay())
		client:=&http.Client{}
		bcsReq,_:=http.NewRequest("GET",endpoint,nil)
		query:=bcsReq.URL.Query()
		query.Add("blockchain_address",bca)
		bcsReq.URL.RawQuery=query.Encode()

		bcsResp ,err:=client.Do(bcsReq)
		if(err!=nil){
			log.Printf("Error: %v\n",err)
			io.WriteString(resp,"Fail.")
			return
		}
		fmt.Println("Success.",bcsReq.URL.RawQuery)
		// resp.Header().Add("ContentType","application/json")
		if bcsResp.StatusCode == 200{
			decoder:=json.NewDecoder(bcsResp.Body)
			var bar block.AmountResponse
			err:=decoder.Decode(&bar)
			if(err!=nil){
				log.Printf("Error: %v\n",err)
				io.WriteString(resp,"Fail.")
				return
			}
			m,_:=json.Marshal(struct{
				Message string `json:"message"`
				Amount float32 `json:"amount"`
			}{
				Message: "Success",
				Amount: bar.Amount,
			})
			io.WriteString(resp,string(m[:]))
			return;
		}else {
			io.WriteString(resp,"Fail.")
		}

	default:
		resp.WriteHeader(http.StatusBadRequest)
		log.Println("Error: 1 Invalid http request.")
	}
}

func (server *WalletServer) Run(){
	fmt.Println(1,server)
	http.HandleFunc("/",server.Index)
	http.HandleFunc("/wallet",server.Wallet)
	http.HandleFunc("/wallet/amount",server.WalletAmount)
	http.HandleFunc("/transaction",server.Transaction)
	// log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(server.Port())),nil))
	// log.Fatal(http.ListenAndServe("https://blockchain-2vc1.onrender.com/"+strconv.Itoa(int(server.Port())),nil))
	log.Fatal(http.ListenAndServe(":80", nil))
}