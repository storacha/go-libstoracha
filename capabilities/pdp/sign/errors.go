package sign

import (
	"fmt"

	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
)

type SignError struct {
	ErrorName string
	Message   string
}

func (se SignError) Name() string {
	return se.ErrorName
}

func (se SignError) Error() string {
	return se.Message
}

func (ge SignError) ToIPLD() (ipld.Node, error) {
	return ipld.WrapWithRecovery(&ge, SignErrorType(), types.Converters...)
}

var SignErrorReader = schema.Struct[SignError](SignErrorType(), nil, Converters...)

const InvalidResourceErrorName = "InvalidResource"

func NewInvalidResourceError(expected ucan.Resource, actual ucan.Resource) SignError {
	return SignError{
		ErrorName: InvalidResourceErrorName,
		Message:   fmt.Sprintf("invalid resource, expected: %s got: %s", expected, actual),
	}
}
