package sign

import (
	"math/big"

	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/validator"
)

const DataSetDeleteAbility = "pdp/sign/dataset/delete"

type DataSetDeleteCaveats struct {
	DataSet *big.Int
}

func (c DataSetDeleteCaveats) ToIPLD() (ipld.Node, error) {
	return ipld.WrapWithRecovery(&c, DataSetDeleteCaveatsType(), types.Converters...)
}

var DataSetDeleteCaveatsReader = schema.Struct[DataSetDeleteCaveats](DataSetDeleteCaveatsType(), nil, types.Converters...)

type DataSetDeleteOk = AuthSignature

var DataSetDeleteOkReader = AuthSignatureReader

var DataSetDelete = validator.NewCapability(
	DataSetDeleteAbility,
	schema.DIDString(),
	DataSetDeleteCaveatsReader,
	validator.DefaultDerives,
)
