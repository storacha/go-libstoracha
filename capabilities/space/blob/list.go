package blob

import (
	"fmt"
	"time"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"
)

const ListAbility = "space/blob/list"

// ListCaveats represents the caveats required to perform a space/blob/list invocation.
type ListCaveats struct {
	// Cursor is a pointer that can be moved back and forth on the list.
	// It can be used to paginate a list for instance.
	Cursor *string
	// Size is the maximum number of items per page.
	Size *uint64
}

func (lc ListCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&lc, ListCaveatsType(), types.Converters...)
}

var ListCaveatsReader = schema.Struct[ListCaveats](ListCaveatsType(), nil, types.Converters...)

// ListOk represents the successful response for a space/blob/list invocation.
type ListOk struct {
	Cursor  *string
	Before  *string
	After   *string
	Size    uint64
	Results []ListBlobItem
}

type ListBlobItem struct {
	Blob       Blob
	InsertedAt time.Time
}

func (lo ListOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&lo, ListOkType(), types.Converters...)
}

var ListOkReader = schema.Struct[ListOk](ListOkType(), nil, types.Converters...)

// List is a capability that allows the agent to list stored Blobs in a space identified by did:key in the `with` field.
var List = validator.NewCapability(
	ListAbility,
	schema.DIDString(),
	ListCaveatsReader,
	func(claimed, delegated ucan.Capability[ListCaveats]) failure.Failure {
		// Check if the space matches
		if claimed.With() != delegated.With() {
			return schema.NewSchemaError(fmt.Sprintf(
				"Resource '%s' doesn't match delegated '%s'",
				claimed.With(), delegated.With(),
			))
		}

		return nil
	},
)
