package upload

import (
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"
)

const (
	GetAbility = "upload/get"
)

type GetCaveats struct {
	Root *cid.Cid `ipld:"root,omitempty"`
}

func (gc GetCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&gc, GetCaveatsType(), types.Converters...)
}

var GetCaveatsReader = schema.Struct[GetCaveats](GetCaveatsType(), nil, types.Converters...)

type GetOk struct {
	Root       cid.Cid   `ipld:"root"`
	Shards     []cid.Cid `ipld:"shards,omitempty"`
	InsertedAt string    `ipld:"insertedAt"`
	UpdatedAt  string    `ipld:"updatedAt"`
}

func (ok GetOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ok, GetOkType(), types.Converters...)
}

var GetOkReader = schema.Struct[GetOk](GetOkType(), nil, types.Converters...)

var Get = validator.NewCapability(
	GetAbility,
	schema.DIDString(),
	GetCaveatsReader,
	func(claimed, delegated ucan.Capability[GetCaveats]) failure.Failure {
		if err := ValidateSpaceDID(claimed.With()); err != nil {
			return err
		}
		
		if fail := equalWith(claimed.With(), delegated.With()); fail != nil {
			return fail
		}

		if delegated.Can() == UploadAbility {
			return nil
		}

		if delegated.Nb().Root != nil {
			if claimed.Nb().Root == nil {
				return schema.NewSchemaError("root must be specified for invocation")
			}

			if !claimed.Nb().Root.Equals(*delegated.Nb().Root) {
				return schema.NewSchemaError(fmt.Sprintf(
					"root '%s' doesn't match delegated '%s'",
					claimed.Nb().Root, delegated.Nb().Root,
				))
			}
		}

		return nil
	},
)