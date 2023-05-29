package auth

import (
	"github.com/jaimi-io/hypersdk/crypto"

	"github.com/jaimi-io/hypersdk/chain"
)

func GetUser(auth chain.Auth) crypto.PublicKey {
	switch auth.(type){
	case *EIP712:
		return crypto.EmptyPublicKey
	default:
		return crypto.EmptyPublicKey
	}
}