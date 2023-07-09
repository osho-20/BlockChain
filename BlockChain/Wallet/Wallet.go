package Wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
	"utils.go"
)

type wallet struct {
	privateKey    *ecdsa.PrivateKey
	publicKey     *ecdsa.PublicKey
	walletAddress string
}

func NewWallet() *wallet {
	// 1. Create ECDSA private key (32 bytes) public key (64 bytes)
	walt := new(wallet)
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	walt.privateKey = privateKey
	walt.publicKey = &walt.privateKey.PublicKey
	// 2. Perform SHA-256 hsah9ing on the public key (32 bytes)
	hash2 := sha256.New()
	hash2.Write(walt.publicKey.X.Bytes())
	hash2.Write(walt.publicKey.Y.Bytes())
	digest2 := hash2.Sum(nil)
	// 3. Perform RIPEMD-160 hashing on the result of SHA-256 (20 bytes)
	hash3 := ripemd160.New()
	hash3.Write(digest2)
	digest3 := hash3.Sum(nil)
	// 4. Add version Byte in front of RIPEMD-160 ash (0x00 for main network)
	version := make([]byte, 21)
	version[0] = 0x00
	copy(version[1:], digest3[:])
	// 5. Perform SHA-256 hash on extended RIPEMD-160 Result
	hash5 := sha256.New()
	hash5.Write(version)
	digest5 := hash5.Sum(nil)
	// 6. Perform SHA-256 hash on the result of the previous hash SHA-256 hash.
	hash6 := sha256.New()
	hash6.Write(digest5)
	digest6 := hash6.Sum(nil)
	// 7. Take first 4 bytes of the second SHA-256 hash for checksum.
	checksum1 := digest6[:4]
	// 8. Add the 4 checksum bytes from 7 at the end of extended RIPEMD-160 hash from 4 (25 bytes)
	bytes25 := make([]byte, 25)
	copy(bytes25[:21], version[:])
	copy(bytes25[21:], checksum1[:])
	// 9. Convert the result from a bytes sting into base58
	address := base58.Encode(bytes25)
	walt.walletAddress = address
	return walt
}

func (walt *wallet) PrivateKey() *ecdsa.PrivateKey {
	return walt.privateKey
}

func (walt *wallet) PrivateKeyStr() string {
	return fmt.Sprintf("%x", walt.privateKey.D.Bytes())
}

func (walt *wallet) PublicKey() *ecdsa.PublicKey {
	return walt.publicKey
}

func (walt *wallet) PublicKeyStr() string {
	return fmt.Sprintf("%064x%064x", walt.publicKey.X.Bytes(), walt.publicKey.Y.Bytes())
}

func (walt *wallet) Address() string {
	return walt.walletAddress
}

type Transaction struct {
	senderPrivateKey *ecdsa.PrivateKey
	senderPublicKey  *ecdsa.PublicKey
	senderAddress    string
	receiverAddress  string
	value            float32
} 

func NewTransaction(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey, sender string, receiver string, value float32) *Transaction {
	return &Transaction{privateKey, publicKey, sender, receiver, value}
}

func (transact *Transaction) GenerateSignature() *utils.Signature {
	m, _ := json.Marshal(transact)
	hash := sha256.Sum256([]byte(m))
	r, s, _ := ecdsa.Sign(rand.Reader, transact.senderPrivateKey, hash[:])
	return &utils.Signature{r, s}
}

func (transact *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Sender   string  `json:"senderAddress"`
		Receiver string  `json:"receiverAddress"`
		Value    float32 `json:"value"`
	}{
		Sender:   transact.senderAddress,
		Receiver: transact.receiverAddress,
		Value:    transact.value,
	})
}

func (wallet* wallet) MarshalJSON()([]byte,error){
	return json.Marshal(struct{
		PrivateKey string `json:"private_key"`
		PublicKey string `json:"public_key"`
		WalletAddress string `json:"wallet_address"`
	}{
		PrivateKey: wallet.PrivateKeyStr(),
		PublicKey: wallet.PublicKeyStr(),
		WalletAddress: wallet.walletAddress,
	})
}

type TransactionReq struct{
	SenderPrivateKey *string `json:"sender_private_key"`
	SenderBlockChainAddress *string `json:"sender_address"`
	ReceiverBlockChainAddress *string `json:"receiver_address"`
	SenderPublicKey *string `json:"sender_public_key"`
	Value *string `json:"value"`
}

func (transact * TransactionReq) Validate() bool{
	if (transact.SenderBlockChainAddress ==nil || transact.ReceiverBlockChainAddress==nil || transact.SenderPrivateKey==nil || transact.SenderPublicKey == nil || transact.Value ==nil){
		return false
	}
	return true
}