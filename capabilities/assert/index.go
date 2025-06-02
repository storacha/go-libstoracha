package assert

import (
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/validator"
)

// IndexAbility claims that a content graph can be found in blob(s) that are
// identified and indexed in the given index CID.
const IndexAbility = "assert/index"

type IndexCaveats struct {
	Content ipld.Link
	Index   ipld.Link
}

func (ic IndexCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ic, IndexCaveatsType(), types.Converters...)
}

var IndexCaveatsReader = schema.Struct[IndexCaveats](IndexCaveatsType(), nil, types.Converters...)

var Index = validator.NewCapability(IndexAbility, schema.DIDString(), IndexCaveatsReader, nil)
