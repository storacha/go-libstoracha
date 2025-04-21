package upload

import (
	"fmt"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"
)

const GetAbility = "upload/get"

type GetCaveats struct {
	Root datamodel.Link `ipld:"root"`
}

func (gc GetCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&gc, GetCaveatsType(), types.Converters...)
}

var GetCaveatsReader = schema.Struct[GetCaveats](GetCaveatsType(), nil, types.Converters...)

type GetOk struct {
	Root   datamodel.Link   `ipld:"root"`
	Shards []datamodel.Link `ipld:"shards"`
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
		if err := validateSpaceDID(claimed.With()); err != nil {
			return err
		}

		if fail := equalWith(claimed.With(), delegated.With()); fail != nil {
			return fail
		}

		if delegated.Can() == UploadAbility {
			return nil
		}

		if delegated.Can() == UploadAbility {
			return nil
		}

		if claimed.Nb().Root.String() != delegated.Nb().Root.String() {
			return schema.NewSchemaError(fmt.Sprintf(
				"root '%s' doesn't match delegated '%s'",
				claimed.Nb().Root, delegated.Nb().Root,
			))
		}
		return nil
	},
)
