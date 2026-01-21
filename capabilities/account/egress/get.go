package egress

import (
	"fmt"
	"slices"
	"time"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/receipt"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/did"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"

	"github.com/storacha/go-libstoracha/capabilities/types"
)

const GetAbility = "account/egress/get"

// Period is a time range to filter egress results (optional)
// From is inclusive and To is exclusive.
// Currently only the date portion of `time.Time` is used because the resolution of the data is always daily.
type Period struct {
	From time.Time
	To   time.Time
}

// GetCaveats allows filtering the egress results.
// Both caveats are optional.
// An empty `Spaces` will return egress stats for all spaces owned by the account.
// A nil `Period` will return egress stats from the first day of the last complete month to today by default.
type GetCaveats struct {
	Spaces []did.DID
	Period *Period
}

func (gc GetCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&gc, GetCaveatsType(), types.Converters...)
}

var GetCaveatsReader = schema.Struct[GetCaveats](GetCaveatsType(), nil, types.Converters...)

// DailyStats contains the egress stats for a single day, in number of bytes.
// Only the date part of `time.Time` is used.
type DailyStats struct {
	Date   time.Time
	Egress uint64
}

// SpaceEgress contains the egress stats for a single space.
// Total is the total egress in bytes for the given period.
// DailyStats contains the stats for each day in the period. Sorted by date ascending.
type SpaceEgress struct {
	Total      uint64
	DailyStats []DailyStats
}

type SpacesModel struct {
	Keys   []did.DID
	Values map[did.DID]SpaceEgress
}

// GetOk contains the egress stats for the given period.
// Total is the total egress in bytes for the requested period.
// Spaces offers a detailed daily breakdown of the egress for each space in the requested period.
type GetOk struct {
	Total  uint64
	Spaces SpacesModel
}

func (gok GetOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&gok, GetOkType(), types.Converters...)
}

type GetError struct {
	ErrorName string
	Message   string
}

const AccountNotFoundErrorName = "AccountNotFound"

func NewAccountNotFoundError(msg string) GetError {
	return GetError{
		ErrorName: AccountNotFoundErrorName,
		Message:   msg,
	}
}

const SpaceUnauthorizedErrorName = "SpaceUnauthorized"

func NewSpaceUnauthorizedError(msg string) GetError {
	return GetError{
		ErrorName: SpaceUnauthorizedErrorName,
		Message:   msg,
	}
}

const PeriodNotAcceptableErrorName = "PeriodNotAcceptable"

func NewPeriodNotAcceptableError(msg string) GetError {
	return GetError{
		ErrorName: PeriodNotAcceptableErrorName,
		Message:   msg,
	}
}

func (ge GetError) Name() string {
	return ge.ErrorName
}

func (ge GetError) Error() string {
	return ge.Message
}

func (ge GetError) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ge, GetErrorType(), types.Converters...)
}

var GetErrorReader = schema.Mapped(
	schema.Struct[GetError](GetErrorType(), nil, types.Converters...),
	func(ge GetError) (GetError, failure.Failure) {
		if ge.Name() != AccountNotFoundErrorName {
			return GetError{}, failure.FromError(fmt.Errorf("incorrect name: %s, expected: %s", ge.Name(), AccountNotFoundErrorName))
		}
		return ge, nil
	},
)

type GetReceipt receipt.Receipt[GetOk, GetError]
type GetReceiptReader receipt.ReceiptReader[GetOk, GetError]

func NewGetReceiptReader() (GetReceiptReader, error) {
	return receipt.NewReceiptReaderFromTypes[GetOk, GetError](GetOkType(), GetErrorType(), types.Converters...)
}

var GetOkReader = schema.Struct[GetOk](GetOkType(), nil, types.Converters...)

var Get = validator.NewCapability(
	GetAbility,
	schema.DIDString(),
	GetCaveatsReader,
	getDerives,
)

func getDerives(claimed, delegated ucan.Capability[GetCaveats]) failure.Failure {
	if claimed.With() != delegated.With() {
		return failure.FromError(fmt.Errorf("can not derive %s with %s from %s", claimed.Can(), claimed.With(), delegated.With()))
	}

	if delegated.Nb().Spaces != nil {
		if claimed.Nb().Spaces == nil {
			return failure.FromError(fmt.Errorf("constraint violation: violates imposed spaces constraint %v because it asks for all spaces", delegated.Nb().Spaces))
		}

		for _, s := range claimed.Nb().Spaces {
			if !slices.Contains(delegated.Nb().Spaces, s) {
				return failure.FromError(fmt.Errorf("constraint violation: violates imposed spaces constraint %v because it asks for space %s", delegated.Nb().Spaces, s))
			}
		}
	}

	if delegated.Nb().Period != nil {
		if claimed.Nb().Period == nil {
			return failure.FromError(fmt.Errorf("constraint violation: violates imposed period constraint %v because it doesn't have a period constraint", delegated.Nb().Period))
		}
		if claimed.Nb().Period.From.Before(delegated.Nb().Period.From) {
			return failure.FromError(fmt.Errorf("constraint violation: violates imposed period constraint because it requests dates before %s", delegated.Nb().Period.From))
		}
		if claimed.Nb().Period.To.After(delegated.Nb().Period.To) {
			return failure.FromError(fmt.Errorf("constraint violation: violates imposed period constraint because it requests dates after %s", delegated.Nb().Period.To))
		}
	}

	return nil
}
