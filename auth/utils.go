package auth

import (
	"github.com/jaimi-io/hypersdk/crypto"

	"github.com/jaimi-io/hypersdk/chain"
)

func GetUser(auth chain.Auth) crypto.PublicKey {
	switch a := auth.(type){
	case *EIP712:
		return a.From
	default:
		return crypto.EmptyPublicKey
	}
}