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
	ListAbility = "upload/list"
)

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

var List = validator.NewCapability(
	ListAbility,
	schema.DIDString(),
	ListCaveatsReader,
	func(claimed, delegated ucan.Capability[ListCaveats]) failure.Failure {
		if err := ValidateSpaceDID(claimed.With()); err != nil {
			return err
		}
		
		if claimed.With() != delegated.With() {
			return schema.NewSchemaError(fmt.Sprintf(
				"resource '%s' doesn't match delegated '%s'",
				claimed.With(), delegated.With(),
			))
		}
		
		return nil
	},
)