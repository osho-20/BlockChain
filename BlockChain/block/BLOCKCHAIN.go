package BLOCKCHAIN

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	neigh "neighbor.go"
	"utils.go"
)

const (
	MINNING_DIFFICULTY = 3
	MINING_SENDER      = "1MjYwTgjcNrwHPBs6bZyUn2rWG2di8smgZ"
	MINING_REWARD      = 1.0
	MINING_TIME_SEC    = 10

	BLOCKCHAIN_PORT_RANGE_START = 5000
	BLOCKCHAIN_PORT_RANGE_END = 5003
	NEIGHBOR_IP_RANGE_START = 0 
	NEIGHBOR_IP_RANGE_END = 1
	BLOCKCHAIN_NEIGHBOR_SYNCTIME = 20
)

// Defining Block.
type Block struct {
	timeStamp     int64
	nonce         int
	previousBlock [32]byte
	transactions  []*Transaction
}

// Creating New Block
func NewBlock(nonce int, previousBlock [32]byte, transact []*Transaction) *Block {
	block := new(Block)
	block.timeStamp = time.Now().UnixNano()
	block.previousBlock = previousBlock
	block.nonce = nonce
	block.transactions = transact
	return block
}

func (block *Block)PreviousHash() [32]byte{
	return block.previousBlock
}

func (block *Block)Nonce() int{
	return block.nonce
}

func (block *Block) Transactions() []*Transaction{
	return block.transactions
}

// Printing Block Values
func (block *Block) Print() {
	fmt.Printf("Time Stamp:      %d\n", block.timeStamp)
	fmt.Printf("Previous Hash:   %x\n", block.previousBlock)
	fmt.Printf("Nonce:           %d\n", block.nonce)
	for _, transact := range block.transactions {
		transact.Print()
	}
}

// Defining BlockChain
type BlockChain struct {
	transaction       []*Transaction
	chain             []*Block
	blockChainAddress string
	port              uint16
	mux               sync.Mutex
	neighbors []string
	muxNeighbors sync.Mutex
}

// Getting last Block
func (blockchain *BlockChain) LastBlock() *Block {
	return blockchain.chain[len(blockchain.chain)-1]
}

// Converting it to json
func (block *Block) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		TimeStamp     int64          `json:"timeStamp"`
		Nonce         int            `json:"nonce"`
		PreviousBlock string       `json:"previousBlock"`
		Transaction   []*Transaction `json:"transactions"`
	}{
		TimeStamp:     block.timeStamp,
		Nonce:         block.nonce,
		PreviousBlock: fmt.Sprintf("%x",block.previousBlock),
		Transaction:   block.transactions,
	})
}

// Creating Hash for previous blocks
func (block *Block) Hash() [32]byte {
	hash, _ := block.MarshalJSON()
	return sha256.Sum256([]byte(hash))
}

// Adding New Block to Chain
func NewBlockChain(blockChainAddress string,port uint16) *BlockChain {
	blockchain := new(BlockChain)
	block := &Block{}
	blockchain.blockChainAddress = blockChainAddress
	blockchain.CreateBlock(0, block.Hash())
	blockchain.port=port
	return blockchain
}

func (blockchain *BlockChain) Run(){
	blockchain.startSyncNeighbors()
	blockchain.ResolveConflicts()
}

func (blockchain *BlockChain) SetNeighbor(){
	blockchain.neighbors = neigh.FindNeighbors(
		neigh.GetHost(),blockchain.port,NEIGHBOR_IP_RANGE_START , NEIGHBOR_IP_RANGE_END,BLOCKCHAIN_PORT_RANGE_START,BLOCKCHAIN_PORT_RANGE_END,
	)
	log.Printf("%v\n",blockchain.neighbors)
}

func (blockchain *BlockChain) SyncNeighbors() {
	blockchain.muxNeighbors.Lock()
	defer blockchain.muxNeighbors.Unlock()
	blockchain.SetNeighbor()
}

func (blockchain* BlockChain) startSyncNeighbors(){
	blockchain.SyncNeighbors()
	_ = time.AfterFunc(time.Second* BLOCKCHAIN_NEIGHBOR_SYNCTIME,blockchain.startSyncNeighbors)
}

func (blockchain *BlockChain) TransactionPool() []*Transaction{
	return blockchain.transaction
}

func (blockchain* BlockChain) UnMarshalJSON(data []byte) error{
	v:=&struct{
		Block []*Block `json:"chain"`
	}{
		Block: blockchain.chain,
	}
	if err:=json.Unmarshal(data,&v); err!=nil{
		return err;
	}
	return nil
}

func (block *Block) UnMarshalJSON(data []byte) error{
	var preblock string
	v:=&struct{
		TimeStamp *int64 `json:"timestamp"`
		Nonce *int `json:"nonce"`
		PreviousBlock *string `json:"previousblock"`
		Transactions []*Transaction `json:"transactions"`
	}{
		TimeStamp: &block.timeStamp,
		Nonce: &block.nonce,
		PreviousBlock: &preblock,
		Transactions: block.transactions,
	}
	if err:=json.Unmarshal(data,&v); err!=nil{
		return err;
	}
	pb,_:= hex.DecodeString((*v.PreviousBlock))
	copy(block.previousBlock[:],pb[:32])
	return nil
}

// 
func (blockchain* BlockChain) MarshalJSON() ([]byte,error){
	return json.Marshal(struct{
		Blocks []*Block `json:"chain"`
	}{
		Blocks: blockchain.chain,
	})

}

// Creating New Block for Chain
func (blockchain *BlockChain) CreateBlock(nonce int, previousBlock [32]byte) *Block {
	block := NewBlock(nonce, previousBlock, blockchain.transaction)
	blockchain.chain = append(blockchain.chain, block)

	// Once the block is created the transaction history is saved in that block and blockchain transaction history is cleared.
	blockchain.transaction = []*Transaction{}
	return block
}

// Printing Values of BlockChain
func (blockchain *BlockChain) PrintBlockChain() {
	for i, block := range blockchain.chain {
		fmt.Printf("%s Block no. %d  %s\n", strings.Repeat("=", 25), i+1, strings.Repeat("=", 25))
		block.Print()
		fmt.Println()
	}
}

// Transaction Type
type Transaction struct {
	senderAddress   string
	receiverAddress string
	value           float32
}

// Creating New Transaction
func NewTransaction(sender string, receiver string, value float32) *Transaction {
	transact := new(Transaction)
	transact.senderAddress = sender
	transact.receiverAddress = receiver
	transact.value = value
	return transact
}

// Adding Transaction
func (blockchain *BlockChain) CreateTransaction(sender string, receiver string, value float32, senderPublicKey *ecdsa.PublicKey, sign *utils.Signature) bool {
	isTransacted:=blockchain.AddTransaction(sender,receiver,value,senderPublicKey,sign)
	return isTransacted
}


func (blockchain *BlockChain) AddTransaction(sender string, receiver string, value float32, senderPublicKey *ecdsa.PublicKey, sign *utils.Signature) bool {
	transact := NewTransaction(sender, receiver, value)
	if sender == MINING_SENDER {
		blockchain.transaction = append(blockchain.transaction, transact)
		return true
	}
	if blockchain.VerifyTransaction(senderPublicKey, sign, transact) {
		// if(blockchain.TotalAmount(sender) < value){
		// 	log.Println("Error : Not Enough balance in a wallet")
		// 	return false;
		// }
		blockchain.transaction = append(blockchain.transaction, transact)
		return true
	} 
	log.Println("Error: Verify Transaction")
	return false
}

// Verify Transaction Signature
func (blockChain *BlockChain) VerifyTransaction(senderPublicKey *ecdsa.PublicKey, sign *utils.Signature, transact *Transaction) bool {
	m, _ := json.Marshal(transact)
	hash := sha256.Sum256([]byte(m))
	return ecdsa.Verify(senderPublicKey, hash[:], sign.R, sign.S)
}

// Printing Transaction
func (transact *Transaction) Print() {
	fmt.Printf("%s\n", strings.Repeat("-", 60))
	fmt.Printf("Senders Address: %s\n", transact.senderAddress)
	fmt.Printf("Receiver Address: %s\n", transact.receiverAddress)
	fmt.Printf("Value: %.10f\n", transact.value)
}

func (transact *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		SenderAddress   string  `json:"senderAddress"`
		ReceiverAddress string  `json:"receiverAddress"`
		Value           float32 `json:"value"`
	}{
		SenderAddress:   transact.senderAddress,
		ReceiverAddress: transact.receiverAddress,
		Value:           transact.value,
	})
}
func (transact *Transaction) UnMarshalJSON(data []byte) error {
	v:= &struct {
		SenderAddress   string  `json:"senderAddress"`
		ReceiverAddress string  `json:"receiverAddress"`
		Value           float32 `json:"value"`
	}{
		SenderAddress:   transact.senderAddress,
		ReceiverAddress: transact.receiverAddress,
		Value:           transact.value,
	}
	if err:=json.Unmarshal(data, &v);err!=nil{
		return err;
	}
	return nil
}

// Making a copy of block chain transaction history
func (blockchain *BlockChain) CopyTransactionList() []*Transaction {
	transactions := make([]*Transaction, 0)
	for _, transact := range blockchain.transaction {
		transactions = append(transactions, NewTransaction(transact.senderAddress, transact.receiverAddress, transact.value))
	}
	return transactions
}

// Check for mining
func (blockchain *BlockChain) Proof(nonce int, previousBlock [32]byte, transactions []*Transaction, difficulty int) bool {
	zeros := strings.Repeat("0", difficulty)
	guessBlock := Block{0, nonce, previousBlock, transactions}
	guessHash := fmt.Sprintf("%x", guessBlock.Hash())
	return guessHash[0:difficulty] == zeros
}

// Mining block chain and giving rewards.
func (blockchain *BlockChain) Mining() bool {
	blockchain.mux.Lock()
	defer blockchain.mux.Unlock()
	if(len(blockchain.TransactionPool())==0){
		fmt.Println("false")
		return false
	}
	fmt.Println("running")
	blockchain.AddTransaction(MINING_SENDER, blockchain.blockChainAddress, MINING_REWARD, nil, nil)
	fmt.Println(blockchain.blockChainAddress)
	nonce := blockchain.MineTransaction()
	previousblock := blockchain.LastBlock().Hash()
	blockchain.CreateBlock(nonce, previousblock)
	log.Println("action = mining, status = success")

	for _,n:= range blockchain.neighbors{
		endpoint:=fmt.Sprintf("http://%s/consensus",n)
		client:=&http.Client{}
		req,_:=http.NewRequest("PUT",endpoint,nil)
		resp,_:=client.Do(req)
		log.Printf("%v\n",resp)
	}
	return true
}

func (blockchain *BlockChain) StartMine(){
	blockchain.Mining()
	_ = time.AfterFunc(time.Second*MINING_TIME_SEC,blockchain.StartMine)
}
// Mining trasaction and making it valid.
func (blockchain *BlockChain) MineTransaction() int {
	transaction := blockchain.transaction
	previous := blockchain.LastBlock().Hash()
	nonce := 0
	for !blockchain.Proof(nonce, previous, transaction, MINNING_DIFFICULTY) {
		nonce += 1
	}
	return nonce
}

// Calculating Total amount of crypto
func (blockchain *BlockChain) TotalAmount(blockChainAddress string) float32 {
	var totalAmount float32 = 0.0
	for _, block := range blockchain.chain {
		for _, transact := range block.transactions {
			if blockChainAddress == transact.receiverAddress {
				totalAmount += transact.value
			}
			if blockChainAddress == transact.senderAddress {
				totalAmount -= transact.value
			}
		}
	}
	return totalAmount
}

type TransactionReq struct{
	SenderBlockChainAddress *string `json:"sender_bc_address"`
	ReceiverBlockChainAddress *string `json:"receiver_bc_address"`
	SenderPublicKey* string `json:"sender_public_key"`
	Value *float32 `json:"value"`
	Signature *string `json:"signature"`
}

func (tr *TransactionReq) Valid() bool{
	if(tr.Signature ==  nil || tr.ReceiverBlockChainAddress==nil || tr.SenderPublicKey==nil || tr.SenderBlockChainAddress==nil || tr.Value==nil){
		return false
	}
	return true
}

type AmountResponse struct{
	Amount float32 `json:"amount"`
}

func (amount *AmountResponse) MarshalJSON() ([]byte,error){
	return json.Marshal(struct{
		Amount float32 `json:"amount"`
	}{
		Amount: amount.Amount,
	})
}

func (blockchain *BlockChain) ValidChain(chain []*Block) bool{
	preblock:=chain[0]
	ci:=1
	for ci<len(chain){
		b:=chain[ci]
		if(b.previousBlock!=preblock.Hash()){
			return false
		}
		if !blockchain.Proof( b.Nonce(), b.PreviousHash(), b.Transactions(), MINNING_DIFFICULTY) {
			return false
		}
		preblock=b
		ci+=1
	}
	return true
}

func (blockchain * BlockChain) Chain() []*Block{
	return blockchain.chain
}

func (blockchain *BlockChain) ResolveConflicts() bool{
	var longestchain []*Block = nil
	maxl:=len(blockchain.chain)
	for _,n:=range blockchain.neighbors {
		endpoint :=fmt.Sprintf("http://%s/chain",n)
		resp,_:= http.Get(endpoint)
		if resp.StatusCode==200 {
			var bcResp BlockChain
			decoder:=json.NewDecoder(resp.Body)
			_ = decoder.Decode(&bcResp)
			chain:=bcResp.Chain()
			if(len(chain)>maxl && blockchain.ValidChain(chain)){
				maxl=len(chain)
				longestchain=chain
			}
		}
	}
	fmt.Println(maxl)
	if longestchain!=nil{
		blockchain.chain=longestchain
		log.Println("Resolved Conflicts Replaced")
		return true
	}
	log.Println("Resolved Conflicts Retained")
	return false
}