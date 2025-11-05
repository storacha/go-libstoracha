package sign

import (
	"math/big"

	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/schema"
)

type Metadata struct {
	Keys   []string
	Values map[string]string
}

type AuthSignature struct {
	Signature  []byte
	V          *big.Int
	R          []byte
	S          []byte
	SignedData []byte
	Signer     []byte
}

func (as AuthSignature) ToIPLD() (ipld.Node, error) {
	return ipld.WrapWithRecovery(&as, AuthSignatureType(), types.Converters...)
}

var AuthSignatureReader = schema.Struct[AuthSignature](AuthSignatureType(), nil, types.Converters...)
