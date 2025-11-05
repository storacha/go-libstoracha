package sign

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/schema"
)

type Metadata struct {
	Keys   []string
	Values map[string]string
}

type AuthSignature struct {
	Signature  []byte
	V          uint8
	R          common.Hash
	S          common.Hash
	SignedData []byte
	Signer     common.Address
}

func (as AuthSignature) ToIPLD() (ipld.Node, error) {
	return ipld.WrapWithRecovery(&as, AuthSignatureType(), Converters...)
}

var AuthSignatureReader = schema.Struct[AuthSignature](AuthSignatureType(), nil, Converters...)
