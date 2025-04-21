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
	AddAbility = "upload/add"
)

type AddCaveats struct {
	Root   cid.Cid   `ipld:"root"`
	Shards []cid.Cid `ipld:"shards,omitempty"`
}

func (ac AddCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ac, AddCaveatsType(), types.Converters...)
}

var AddCaveatsReader = schema.Struct[AddCaveats](AddCaveatsType(), nil, types.Converters...)

type AddOk struct {
	Root   cid.Cid   `ipld:"root"`
	Shards []cid.Cid `ipld:"shards,omitempty"`
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
		if err := ValidateSpaceDID(claimed.With()); err != nil {
			return err
		}

		if fail := equalWith(claimed.With(), delegated.With()); fail != nil {
			return fail
		}

		if delegated.Can() == UploadAbility {
			return nil
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

func equalRoot(claimed, delegated cid.Cid) failure.Failure {
	if !claimed.Equals(delegated) {
		return schema.NewSchemaError(fmt.Sprintf(
			"root '%s' doesn't match delegated '%s'",
			claimed, delegated,
		))
	}

	return nil
}

func equalShards(claimed, delegated []cid.Cid) failure.Failure {

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
