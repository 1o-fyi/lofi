package cmd

import (
	"math/big"

	"filippo.io/age"
	"git.sr.ht/~lofi/lib"
	"github.com/keep-network/keep-core/pkg/bls"
)

// returns the public key for G2.
func mapToKeyShare(id *age.X25519Identity) *bls.SecretKeyShare {
	sk := lib.Ss(id.String())
	skShare := bls.GetSecretKeyShare([]*big.Int{sk.V}, 1)
	return skShare
}
