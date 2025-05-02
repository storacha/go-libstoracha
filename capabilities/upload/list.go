package upload

import (
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"
)

const ListAbility = "upload/list"

type ListCaveats struct {
	Cursor *string
	Size   *uint64
	Pre    *bool
}

func (lc ListCaveats) ToIPLD() (ipld.Node, error) {
	return ipld.WrapWithRecovery(&lc, ListCaveatsType(), types.Converters...)
}

var ListCaveatsReader = schema.Struct[ListCaveats](ListCaveatsType(), nil, types.Converters...)

type ListItem struct {
	Root   ipld.Link
	Shards []ipld.Link
}

type ListOk struct {
	Cursor  *string
	Before  *string
	After   *string
	Size    uint64
	Results []ListItem
}

func (lo ListOk) ToIPLD() (ipld.Node, error) {
	return ipld.WrapWithRecovery(&lo, ListOkType(), types.Converters...)
}

var ListOkReader = schema.Struct[ListOk](ListOkType(), nil, types.Converters...)

var List = validator.NewCapability(
	ListAbility,
	schema.DIDString(),
	ListCaveatsReader,
	func(claimed, delegated ucan.Capability[ListCaveats]) failure.Failure {
		if err := validateSpaceDID(claimed.With()); err != nil {
			return err
		}

		if fail := validator.DefaultDerives(claimed, delegated); fail != nil {
			return fail
		}

		return nil
	},
)
