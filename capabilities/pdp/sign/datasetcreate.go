package sign

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/validator"
)

const DataSetCreateAbility = "pdp/sign/dataset/create"

type DataSetCreateCaveats struct {
	DataSet  *big.Int
	Payee    common.Address
	Metadata Metadata
}

func (c DataSetCreateCaveats) ToIPLD() (ipld.Node, error) {
	return ipld.WrapWithRecovery(&c, DataSetCreateCaveatsType(), Converters...)
}

var DataSetCreateCaveatsReader = schema.Struct[DataSetCreateCaveats](DataSetCreateCaveatsType(), nil, Converters...)

type DataSetCreateOk = AuthSignature

var DataSetCreateOkReader = AuthSignatureReader

var DataSetCreate = validator.NewCapability(
	DataSetCreateAbility,
	schema.DIDString(),
	DataSetCreateCaveatsReader,
	validator.DefaultDerives,
)
