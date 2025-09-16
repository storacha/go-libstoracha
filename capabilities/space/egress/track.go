package egress

import (
	"net/url"

	"github.com/ipld/go-ipld-prime/datamodel"
	captypes "github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/receipt"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"
)

const TrackAbility = "space/egress/track"

type TrackCaveats struct {
	Receipts ucan.Link
	Endpoint *url.URL
}

func (tc TrackCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&tc, TrackCaveatsType(), captypes.Converters...)
}

var TrackCaveatsReader = schema.Struct[TrackCaveats](TrackCaveatsType(), nil, captypes.Converters...)

type TrackOk struct {
}

func (to TrackOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&to, TrackOkType(), captypes.Converters...)
}

var TrackOkReader = schema.Struct[TrackOk](TrackOkType(), nil, captypes.Converters...)

type TrackError struct {
	Name    string
	Message string
}

func NewTrackError(msg string) TrackError {
	return TrackError{
		Name:    "EgressTrackError",
		Message: msg,
	}
}

func (te TrackError) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&te, TrackErrorType(), captypes.Converters...)
}

var TrackErrorReader = schema.Struct[TrackError](TrackErrorType(), nil, captypes.Converters...)

type TrackReceipt receipt.Receipt[TrackOk, TrackError]

type TrackReceiptReader receipt.ReceiptReader[TrackOk, TrackError]

func NewTrackReceiptReader() (TrackReceiptReader, error) {
	return receipt.NewReceiptReader[TrackOk, TrackError](egressSchema)
}

// EgressTrack capability definition
// This capability allows a storage node to request recording egress for content it has served.
var Track = validator.NewCapability(
	TrackAbility,
	schema.DIDString(),
	TrackCaveatsReader,
	validator.DefaultDerives,
)
