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

type Period struct {
	From time.Time
	To   time.Time
}

type GetCaveats struct {
	Spaces []did.DID
	Period *Period
}

func (gc GetCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&gc, GetCaveatsType(), types.Converters...)
}

var GetCaveatsReader = schema.Struct[GetCaveats](GetCaveatsType(), nil, types.Converters...)

type DailyStats struct {
	Date   time.Time
	Egress uint64
}

type SpaceEgress struct {
	Total      uint64
	DailyStats []DailyStats
}

type SpacesModel struct {
	Keys   []did.DID
	Values map[did.DID]SpaceEgress
}

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

const AccountNotFoundErrorName = "AccountNotFoundError"

func NewAccountNotFoundError(msg string) GetError {
	return GetError{
		ErrorName: AccountNotFoundErrorName,
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
		return failure.FromError(fmt.Errorf("Can not derive %s with %s from %s", claimed.Can(), claimed.With(), delegated.With()))
	}

	if delegated.Nb().Spaces != nil {
		if claimed.Nb().Spaces == nil {
			return failure.FromError(fmt.Errorf("Constraint violation: violates imposed spaces constraint %v because it asks for all spaces", delegated.Nb().Spaces))
		}

		for _, s := range claimed.Nb().Spaces {
			if !slices.Contains(delegated.Nb().Spaces, s) {
				return failure.FromError(fmt.Errorf("Constraint violation: violates imposed spaces constraint %v because it asks for space %s", delegated.Nb().Spaces, s))
			}
		}
	}

	if delegated.Nb().Period != nil {
		if claimed.Nb().Period == nil {
			return failure.FromError(fmt.Errorf("Constraint violation: violates imposed period constraint %v because it doesn't have a period constraint", delegated.Nb().Period))
		}
		if claimed.Nb().Period.From.Before(delegated.Nb().Period.From) {
			return failure.FromError(fmt.Errorf("Constraint violation: violates imposed period constraint because it requests dates before %s", delegated.Nb().Period.From))
		}
		if claimed.Nb().Period.To.After(delegated.Nb().Period.To) {
			return failure.FromError(fmt.Errorf("Constraint violation: violates imposed period constraint because it requests dates after %s", delegated.Nb().Period.To))
		}
	}

	return nil
}
