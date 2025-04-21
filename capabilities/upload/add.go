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

const AddAbility = "upload/add"

type AddCaveats struct {
	Root   datamodel.Link   `ipld:"root"`
	Shards []datamodel.Link `ipld:"shards"`
}

func (ac AddCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ac, AddCaveatsType(), types.Converters...)
}

var AddCaveatsReader = schema.Struct[AddCaveats](AddCaveatsType(), nil, types.Converters...)

type AddOk struct {
	Root   datamodel.Link   `ipld:"root"`
	Shards []datamodel.Link `ipld:"shards"`
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

		if fail := equalWith(claimed.With(), delegated.With()); fail != nil {
			return fail
		}

		if delegated.Can() == UploadAbility {
			// Even with parent capability, check that claimed shards are valid
			if len(claimed.Nb().Shards) > 0 && len(delegated.Nb().Shards) == 0 {
				return schema.NewSchemaError("cannot claim shards when delegated capability has none")
			}
			return nil
		}

		if fail := equalRoot(claimed.Nb().Root, delegated.Nb().Root); fail != nil {
			return fail
		}

		if len(delegated.Nb().Shards) > 0 {
			if fail := equalShards(claimed.Nb().Shards, delegated.Nb().Shards); fail != nil {
				return fail
			}
		} else if len(claimed.Nb().Shards) > 0 {
			return schema.NewSchemaError("claimed capability includes shards not present in delegation")
		}

		return nil
	},
)

func equalRoot(claimed, delegated datamodel.Link) failure.Failure {
	if claimed.String() != delegated.String() {
		return schema.NewSchemaError(fmt.Sprintf(
			"root '%s' doesn't match delegated '%s'",
			claimed, delegated,
		))
	}

	return nil
}

func equalShards(claimed, delegated []datamodel.Link) failure.Failure {
	delegatedMap := make(map[string]bool)
	for _, shard := range delegated {
		delegatedMap[shard.String()] = true
	}

	for _, shard := range claimed {
		if !delegatedMap[shard.String()] {
			return schema.NewSchemaError(fmt.Sprintf(
				"shard '%s' not found in delegated shards",
				shard,
			))
		}
	}

	return nil
}
