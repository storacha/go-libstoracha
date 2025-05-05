package upload

import (
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"

	"github.com/storacha/go-libstoracha/capabilities/types"
)

const AddAbility = "upload/add"

type AddCaveats struct {
	Root   ipld.Link
	Shards []ipld.Link
}

func (ac AddCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ac, AddCaveatsType(), types.Converters...)
}

var AddCaveatsReader = schema.Struct[AddCaveats](AddCaveatsType(), nil, types.Converters...)

type AddOk struct {
	Root   ipld.Link
	Shards []ipld.Link
}

func (ao AddOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ao, AddOkType(), types.Converters...)
}

var AddOkReader = schema.Struct[AddOk](AddOkType(), nil, types.Converters...)

var Add = validator.NewCapability(
	AddAbility,
	schema.DIDString(),
	AddCaveatsReader,
	func(claimed, delegated ucan.Capability[AddCaveats]) failure.Failure {
		if err := validateSpaceDID(claimed.With()); err != nil {
			return err
		}

		if fail := validator.DefaultDerives(claimed, delegated); fail != nil {
			return fail
		}

		if fail := equalRoot(claimed.Nb().Root, delegated.Nb().Root); fail != nil {
			return fail
		}

		if len(delegated.Nb().Shards) > 0 {
			if fail := equalShards(claimed.Nb().Shards, delegated.Nb().Shards); fail != nil {
				return fail
			}
		}
		return nil
	},
)
