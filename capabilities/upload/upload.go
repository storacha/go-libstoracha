package upload

import (
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"
)

const (
	UploadAbility  = "upload/*"
	AddAbility     = "upload/add"
	GetAbility     = "upload/get"
	RemoveAbility  = "upload/remove"
	ListAbility    = "upload/list"
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

func (go GetOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&go, GetOkType(), types.Converters...)
}

var GetOkReader = schema.Struct[GetOk](GetOkType(), nil, types.Converters...)

type RemoveCaveats struct {
	Root cid.Cid `ipld:"root"`
}

func (rc RemoveCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&rc, RemoveCaveatsType(), types.Converters...)
}

var RemoveCaveatsReader = schema.Struct[RemoveCaveats](RemoveCaveatsType(), nil, types.Converters...)

type RemoveOk struct {
	Root   cid.Cid   `ipld:"root"`
	Shards []cid.Cid `ipld:"shards,omitempty"`
}

func (ro RemoveOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ro, RemoveOkType(), types.Converters...)
}

var RemoveOkReader = schema.Struct[RemoveOk](RemoveOkType(), nil, types.Converters...)

type ListCaveats struct {
	Cursor *string `ipld:"cursor,omitempty"`
	Size   *int    `ipld:"size,omitempty"`
	Pre    *bool   `ipld:"pre,omitempty"`
}

func (lc ListCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&lc, ListCaveatsType(), types.Converters...)
}

var ListCaveatsReader = schema.Struct[ListCaveats](ListCaveatsType(), nil, types.Converters...)

type ListItem struct {
	Root       cid.Cid   `ipld:"root"`
	Shards     []cid.Cid `ipld:"shards,omitempty"`
	InsertedAt string    `ipld:"insertedAt"`
	UpdatedAt  string    `ipld:"updatedAt"`
}

type ListOk struct {
	Cursor  *string    `ipld:"cursor,omitempty"`
	Before  *string    `ipld:"before,omitempty"`
	After   *string    `ipld:"after,omitempty"`
	Size    int        `ipld:"size"`
	Results []ListItem `ipld:"results"`
}

func (lo ListOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&lo, ListOkType(), types.Converters...)
}

var ListOkReader = schema.Struct[ListOk](ListOkType(), nil, types.Converters...)


var Upload = validator.NewCapability(
	UploadAbility,
	schema.DIDString(),
	schema.Struct[struct{}](basicnode.Prototype.Any, nil, types.Converters...),
	func(claimed, delegated ucan.Capability[struct{}]) failure.Failure {
		if claimed.With() != delegated.With() {
			return schema.NewSchemaError(fmt.Sprintf(
				"resource '%s' doesn't match delegated '%s'",
				claimed.With(), delegated.With(),
			))
		}
		return nil
	},
)

var Add = validator.NewCapability(
	AddAbility,
	schema.DIDString(),
	AddCaveatsReader,
	func(claimed, delegated ucan.Capability[AddCaveats]) failure.Failure {
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

var Get = validator.NewCapability(
	GetAbility,
	schema.DIDString(),
	GetCaveatsReader,
	func(claimed, delegated ucan.Capability[GetCaveats]) failure.Failure {
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

var Remove = validator.NewCapability(
	RemoveAbility,
	schema.DIDString(),
	RemoveCaveatsReader,
	func(claimed, delegated ucan.Capability[RemoveCaveats]) failure.Failure {
		if fail := equalWith(claimed.With(), delegated.With()); fail != nil {
			return fail
		}

		if delegated.Can() == UploadAbility {
			return nil
		}

		if fail := equalRoot(claimed.Nb().Root, delegated.Nb().Root); fail != nil {
			return fail
		}

		return nil
	},
)

var List = validator.NewCapability(
	ListAbility,
	schema.DIDString(),
	ListCaveatsReader,
	func(claimed, delegated ucan.Capability[ListCaveats]) failure.Failure {
		if claimed.With() != delegated.With() {
			return schema.NewSchemaError(fmt.Sprintf(
				"resource '%s' doesn't match delegated '%s'",
				claimed.With(), delegated.With(),
			))
		}
		
		return nil
	},
)


func equalWith(claimed, delegated string) failure.Failure {
	if claimed != delegated {
		return schema.NewSchemaError(fmt.Sprintf(
			"resource '%s' doesn't match delegated '%s'",
			claimed, delegated,
		))
	}

	return nil
}

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