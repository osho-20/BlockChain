package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"math/big"
)

type Signature struct {
	R *big.Int
	S *big.Int
}

func SignatureFromString(str string) *Signature{
	r,s:=String2BigInt(str)
	return &Signature{&r,&s}
}

func (sign *Signature) Sign() string {
	return fmt.Sprintf("%064x%064x", sign.R, sign.S)
}

func String2BigInt(s string)(big.Int,big.Int){
	bx,_:=hex.DecodeString(s[:64])
	by,_:=hex.DecodeString(s[64:])
	var bix big.Int
	var biy big.Int
	_ = bix.SetBytes(bx)
	_ = biy.SetBytes(by)
	return bix,biy
}

func PublicKeyfromString(s string)*ecdsa.PublicKey{
	x,y:= String2BigInt(s)
	return &ecdsa.PublicKey{elliptic.P256(),&x,&y}
}

func PrivateKeyfromString(s string , publicKey *ecdsa.PublicKey)*ecdsa.PrivateKey{
	b,_:= hex.DecodeString(s[:])
	var big big.Int
	_ = big.SetBytes(b)
	return &ecdsa.PrivateKey{*publicKey,&big}
}