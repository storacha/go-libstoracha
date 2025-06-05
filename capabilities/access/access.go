package access

import (
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/validator"
)

const AccessAbility = "access/*"

// Access capability definition
// This capability can only be delegated (but not invoked) allowing audience to
// derive any `access/` prefixed capability for the space identified by the DID
// in the `with` field.
var Access = validator.NewCapability(
	AccessAbility,
	schema.DIDString(),
	schema.Struct[struct{}](nil, nil, types.Converters...),
	nil,
)
